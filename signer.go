package main

import (
	"sort"
	"strconv"
	"sync"
)

// сюда писать код
func ExecutePipeline(jobs ...job) {
	floatChannels := len(jobs) - 1

	pipelineItem := func(wg *sync.WaitGroup, job job, in, out chan interface{}) {
		defer wg.Done()
		job(in, out)
		if out != nil {
			close(out)
		}
	}
	channels := make([]chan interface{}, 0, floatChannels)
	wg := &sync.WaitGroup{}
	for i := 0; i < len(jobs); i++ {
		if i < floatChannels {
			channels = append(channels, make(chan interface{}))
		}
		wg.Add(1)
		if i == 0 {
			go pipelineItem(wg, jobs[i], nil, channels[i])
		} else if i < floatChannels {
			go pipelineItem(wg, jobs[i], channels[i-1], channels[i])
		} else {
			go pipelineItem(wg, jobs[i], channels[i-1], nil)
		}
	}
	wg.Wait()
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
	buffer := make(sort.StringSlice, 0, 1)
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
