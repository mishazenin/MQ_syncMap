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

//timeout, err := time.ParseDuration(r.FormValue("timeout"))
//	if err!=nil{
//
//
//	}
//только там но нужно будет задавачать количесво секунд
//var (
//	_      context.Context
//	cancel context.CancelFunc
//)
//timeout, err := time.ParseDuration(r.FormValue("timeout"))
//fmt.Fprintf(w, "Been waiting for %v seconds", timeout)
//if err == nil {
//	time.Sleep(2 * time.Second)
//	_, cancel = context.WithTimeout(context.Background(), timeout)
//} else {
//	_, cancel = context.WithCancel(context.Background())
//}
//defer cancel()

///////////////////////////////////////////////////////////////////////////////
////timeout = time.ParseDuration("10h")
//if err == nil {
//	ctx, Cancel = context.WithTimeout(context.Background(), timeout)
//	fmt.Fprintf(w, "%q\n", timeout)
//} else {
//	_, Cancel = context.WithCancel(context.Background())
//}
//fmt.Fprintf(w, "%v",ctx)
//defer Cancel()

//w.WriteHeader(http.StatusBadRequest)
//encoder := json.NewEncoder(w)
//encoder.Encode(w)

//
//func AddHealthCheck() (string, error) {
//
//	//convert go struct to json
//	payload := "bob"
//	jsonPayload, err := json.Marshal(payload)
//	panicError(err)
//
//	// Create client & set timeout
//	client := &http.Client{}
//	client.Timeout = time.Second * 3
//
//	// Create request
//	req, err := http.NewRequest("POST", "http://localhost:8080", bytes.NewBuffer(jsonPayload))
//	panicError(err)
//	req.Header.Set("Content-Type", "application/json")
//
//	// Fetch Request
//	resp, err := client.Do(req)
//	panicError(err)
//	defer resp.Body.Close()
//
//	// Read Response Body
//	respBody, err := ioutil.ReadAll(resp.Body)
//	panicError(err)
//
//	fmt.Println("response Status : ", resp.Status)
//	fmt.Println("response Headers : ", resp.Header)
//	fmt.Println("response Body : ", string(respBody))
//
//	return string(respBody), nil
//}

//var step = 0
//
//for {
//	time.Sleep(time.Microsecond * 100)
//
//	step++
//
//	err := requestWithClose()
//	if err != nil {
//		fmt.Printf("[%d] requestWithClose failed: %s", step, err)
//		continue
//	}
//
//	fmt.Printf("[%d] ok\n", step)
//}

//func timur(str1 string, str2 string) []string {
//
//	finalArr := []string{}
//	start := 0
//	for i := 0; i <= len(str1)-len(str2); {
//
//		j := i + len(str2)
//		if str1[i:j] == str2 {
//			outStr := str1[start:i]
//			finalArr = append(finalArr, outStr)
//			start = j
//			i = start
//		} else {
//			i++
//		}
//
//	}
//	if start < len(str1) {
//		finalArr = append(finalArr, str1[start:])
//	}
//	return finalArr
//}

//
//func GetFibonacciArr(n int) int {
//
//	t1 := 0
//	t2 := 1
//	nextTerm := 0
//
//	arr := make([]int, 0)
//	for i := 1; i <= n+1; i++ {
//		if i == 1 {
//			continue
//		}
//		if i == 2 {
//			arr = append(arr, t2)
//			continue
//		}
//		nextTerm = t1 + t2
//		t1 = t2
//		t2 = nextTerm
//		arr = append(arr, nextTerm)
//	}
//	//fibArr := arr[(s - 1):]
//
//	lastNumber := arr[len(arr)-1]
//	fmt.Println(lastNumber)
//
//	ln := lastNumber % 10
//	fmt.Println(ln)
//	return ln
//
//}
//
//func leftDigits(number, n int) int {
//	digits := make([]byte, 20)
//	i := -1
//	for number != 0 {
//		i++
//		digits[i] = byte(number % 10)
//		number /= 10
//	}
//	r := 0
//	for ; n != 0; n-- {
//		r = r * 10
//		r += int(digits[i])
//		i--
//	}
//	return r
//}

//a := make([]int, 5)
//for i := 0; i < 5; i++ {
//a[i] = i + 1
//}
//
//b := make([]*int, 5)
//for i, el := range a {
//b[i] = &el   //&a[i]
//}
//for _, el := range b {
//fmt.Printf("%d\n", *el)
//}

//const timeDuration = 5 * time.Second
//ctx,cancel:=context.WithTimeout(context.Background(),timeDuration)
//defer cancel()
//
//select {
//case <- time.After(1*time.Millisecond):
//fmt.Println("overspept")
//case<-ctx.Done():
//fmt.Println("timeput exeeced")
//}

//func test2(){
//	t0:=time.Now()
//	numbers:=make([]int,10000000)
//	for i:=0; i<cap(numbers);i++{
//		numbers[i]=i
//	}
//	t1:=time.Now()
//	fmt.Println(t1.Sub(t0))
//
//}

//func test1 (){
//	t0 := time.Now()
//	numbers := []int{}
//	for i := 0; i < 10000000; i++ {
//		numbers = append(numbers, i)
//	}
//	t1 := time.Now()
//	fmt.Println(t1.Sub(t0))
//}

//func BubbleSort()
//{
//array := []int{1, 45, 23, 47, 234, 547, 234, 2, 4, 7, 3, 1, 0}
//
//for i := 0; i < len(array)-1; i++ {
//for j := 0; j < len(array)-i-1; j++ {
//if array[j] > array[j+1] {
//array[j], array[j+1] = array[j+1], array[j]
//}
//}
//}
//
//for i := range array {
//fmt.Println(array[i])
//}
//}
