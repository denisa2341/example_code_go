package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func main() {
	return
}

func ExecutePipeline(task ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})
	for _, f := range task {
		wg.Add(1)
		out := make(chan interface{})
		go workerExecute(f, in, out, wg)
		in = out
	}
	wg.Wait()
}
func workerExecute(oneTask job, in, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(out)
	oneTask(in, out)
}
func SingleHash(in, out chan interface{}) {
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for i := range in {
		wg.Add(1)
		str := strconv.Itoa(i.(int))
		go func(str string, out chan interface{}, wg *sync.WaitGroup, mu *sync.Mutex) {
			defer wg.Done()
			wg1 := &sync.WaitGroup{}
			fmt.Println(str + " SingleHash data " + str)
			mu.Lock()
			wordMd5 := DataSignerMd5(str)
			mu.Unlock()
			fmt.Println(str + " SingleHash md5(data) %s\n" + wordMd5)
			var wordCrc32 string
			wg1.Add(1)
			go func(s *string, data string, wg1 *sync.WaitGroup) {
				defer wg1.Done()
				*s = DataSignerCrc32(data)
			}(&wordCrc32, str, wg1)
			wordMd5crc32 := DataSignerCrc32(wordMd5)
			wg1.Wait()
			fmt.Println(str + " SingleHash crc32(md5(data)) " + wordMd5crc32)
			fmt.Println(str + " SingleHash crc32(data) " + wordCrc32)
			fmt.Println(str + " SingleHash result " + wordCrc32 + "~" + wordMd5crc32)
			out <- wordCrc32 + "~" + wordMd5crc32
		}(str, out, wg, mu)
	}
	wg.Wait()
}
func MultiHash(in, out chan interface{}) {
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for i := range in {
		wg.Add(1)
		str := i.(string)
		go funcMulti(str, out, wg, mu)
	}
	wg.Wait()
}
func CombineResults(in, out chan interface{}) {
	var allResult []string
	for i := range in {
		str := i.(string)
		allResult = append(allResult, str)
	}
	sort.Strings(allResult)
	answ := strings.Join(allResult, "_")
	out <- answ
}
func funcMulti(str string, out chan interface{}, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()
	wg1 := &sync.WaitGroup{}
	strArr := make([]string, 6, 6)
	wg1.Add(6)
	par32 := func(s *string, data string, wg1 *sync.WaitGroup, mu *sync.Mutex) {
		defer wg1.Done()
		s1 := DataSignerCrc32(data)
		mu.Lock()
		*s = s1
		mu.Unlock()
	}
	for ind := 0; ind < 6; ind++ {
		wordMd5 := strconv.Itoa(ind) + str
		go par32(&(strArr[ind]), wordMd5, wg1, mu)
	}
	wg1.Wait()
	for ind := 0; ind < 6; ind++ {
		fmt.Println(str + " MultiHash: crc32(th+step1)) " + strconv.Itoa(ind) + " " + strArr[ind])
	}
	allResult := strings.Join(strArr, "")
	out <- allResult
}
