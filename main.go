package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

//KeyCounter is used as a key in Data map
type Cache struct {
	lock       sync.RWMutex
	KeyCounter int
	Data       map[int]string
}

//Get from cache
func (c *Cache) Get(key int) (string, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	d, ok := c.Data[key]
	return d, ok
}

//Set into cache
func (c *Cache) Set(d *string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.Data[c.KeyCounter] = *d
	c.KeyCounter++
}

// go run main.go <port> port
func main() {

	flag.Int("port", 8000, "set port:")
	flag.Parse()
	port := flag.Arg(0)

	newCache := Cache{KeyCounter: 0, Data: map[int]string{}}
	http.HandleFunc("/color", newCache.handler)
	http.HandleFunc("/name", newCache.handler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%v", port), nil))
}

func (c *Cache) handler(w http.ResponseWriter, r *http.Request) {
	// Check method, whether it's PUT or Get
	switch r.Method {
	case http.MethodPut:
		if r.FormValue("v") == "" {
			w.WriteHeader(http.StatusBadRequest)
			encoder := json.NewEncoder(w)
			err := encoder.Encode(w)
			if err != nil {
				fmt.Println(err)
			}
			return
		}

		req := fmt.Sprintf("curl -X%v http://%v%v\n", r.Method, r.Host, r.URL)
		c.Set(&req)
		_, err := fmt.Fprintf(w, "Value = %q\n,%v", r.FormValue("v"), c.KeyCounter)
		if err != nil {
			fmt.Println(err)
		}
		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		err = encoder.Encode(w)
		if err != nil {
			fmt.Println(err)
		}

	case http.MethodGet:
		for i := 0; i <= c.KeyCounter-1; i++ {
			res, ok := c.Get(i)

			//timeout, _ := strconv.Atoi(r.FormValue("timeout"))
			//could not set timeout into duration, it can't be used as a integer example (6 * time.Millisecond)
			//Solution:
			//In that case you should set key with time value with "timeout" key,
			// example: {"timeout":4s} where "s" stands for seconds
			timeout, _ := time.ParseDuration(r.FormValue("timeout"))

			if !ok {
				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				select {
				case <-time.After(5 * time.Second):
					fmt.Println("Overslept")
					err := response(res, w)
					if err != nil {
						fmt.Println(err)
						w.WriteHeader(http.StatusNotFound)
						encoder := json.NewEncoder(w)
						err := encoder.Encode(w)
						if err != nil {
							fmt.Println(err)
						}
					}

				case <-ctx.Done():
					err := response(res, w)
					fmt.Println("Done")
					if err != nil {
						fmt.Println(err)
					}
				}

				cancel()

			} else {
				fmt.Println("with key")
				time.Sleep(timeout)
				err := response(res, w)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	default:

		w.WriteHeader(http.StatusNotFound)
		encoder := json.NewEncoder(w)
		err := encoder.Encode(w)
		if err != nil {
			fmt.Println(err)
		}
		_, err = fmt.Fprintf(w, "Method %q is not supported", r.Method)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func response(res string, w http.ResponseWriter) error {
	respStr := strings.Replace(res, " -XPUT ", " ", 1)
	if respStr == "" {
		newErrs := errors.New("There is no value")
		return newErrs
	} else {
		_, err := fmt.Fprintf(w, "%q\n", respStr)
		if err != nil {
			return nil
		}
	}

	return nil
}
