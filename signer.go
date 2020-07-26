package main

import (
	"sort"
	"strconv"
	"sync"
)
type StrArr []string

func (a StrArr) Len() int {
	return len(a)
}
func (a StrArr)Less(i, j int) bool {
	return a[i] < a[j]
}
func (a StrArr)Swap(i, j int) {
	a[i],a[j] = a[j], a[i]
}

// сюда писать код
func ExecutePipeline(jobs ...job) {
	floatChannels := len(jobs) - 1
	
	channels := make([]chan interface{}, 0, floatChannels)
	firstIn := make(chan interface{}, 1)
	lastOut := make(chan interface{}, 1)
	wg := &sync.WaitGroup{}
	for i := 0; i < len(jobs); i++ {
		if (i < len(jobs) - 1) {
			channels = append(channels, make(chan interface{}))
		}
		if i == 0 {
			wg.Add(1)
			go func(num int) {
				defer wg.Done()
				jobs[num](firstIn, channels[num])
				close(channels[num])
			}(i)
		} else if (i < len(jobs) - 1) {
			wg.Add(1)
			go func(num int) {
				defer wg.Done()
				jobs[num](channels[num - 1], channels[num])
				close(channels[num])
			}(i)
		} else {
			wg.Add(1)
			go func (num int) {
				defer wg.Done()
				jobs[num](channels[num - 1], lastOut)
			}(i)
		}
	}
	close(firstIn)
	wg.Wait()
	close(lastOut)
}

func SingleHash(in, out chan interface{}) {
	mutex := &sync.Mutex{}
	waitGroup := &sync.WaitGroup{}
	for data := range in {
		dataStr := strconv.Itoa(data.(int))
		waitGroup.Add(1)
		go func(mu *sync.Mutex, wg *sync.WaitGroup, dataStr string) {
			defer wg.Done()
			results := make([]string, 2, 2)
			innerwg := &sync.WaitGroup{}
			innerwg.Add(1)
			go func(iwg *sync.WaitGroup, dataStr string) {
				defer iwg.Done()
				mu.Lock()
				md5Data := DataSignerMd5(dataStr)
				mu.Unlock()
				results[1] = DataSignerCrc32(md5Data)
			}(innerwg, dataStr)
			results[0] = DataSignerCrc32(dataStr)
			innerwg.Wait()
			out <- results[0] + "~" + results[1]
		}(mutex, waitGroup, dataStr)
	}
	waitGroup.Wait()
}

func MultiHash(in, out chan interface{}) {
	waitgroup := &sync.WaitGroup{}
	for data := range in {
		waitgroup.Add(1)
		go func(wg *sync.WaitGroup, dataStr string) {
			defer wg.Done()
			results := make([]string, 6, 6)
			innerwg := &sync.WaitGroup{}
			for i := 0; i < 6; i++ {
				innerwg.Add(1)
				go func(iwg *sync.WaitGroup, num int) {
					execResult := DataSignerCrc32(strconv.Itoa(num) + dataStr)
					results[num] = execResult
					iwg.Done()
				}(innerwg, i)
			}
			innerwg.Wait()
			var result string
			for _, entry := range results {
				result += entry
			}
			out <- result
		}(waitgroup, data.(string))
	}
	waitgroup.Wait()
}

func CombineResults(in, out chan interface{}) {
	buffer := make(StrArr, 0, 1)
	for data := range in {
		buffer = append(buffer, data.(string))
	}
	sort.Sort(buffer)
	result := buffer[0]
	for i := 1; i < len(buffer); i++ {
		result += "_" + buffer[i]
	}
	out <- result
}