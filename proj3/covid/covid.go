package main

import (
	"fmt"
	"os"
	"proj3/balancer"
	"proj3/utils"
	"strconv"
	"sync"
	"time"
)

const usage = "Usage: covid threads zipcode month year\n" +
	"    threads = the number of threads (i.e., goroutines to spawn)\n" +
	"    zipcode = a possible Chicago zipcode\n" +
	"    month = the month to display for that zipcode\n" +
	"    year = the year to display for that zipcode \n"

func Sequential(zipcode, month, year int) {
	records := make(map[string]utils.ZipcodeInfo)
	var totalCases, totalTests, totalDeaths int
	for idx := 1; idx <= utils.NumOfFiles; idx++ {
		utils.ReadData(fmt.Sprintf("data/covid_%v.csv", idx), records, zipcode, month, year)
	}

	for _, value := range records {
		totalCases += value.Cases
		totalTests += value.Tests
		totalDeaths += value.Deaths
	}
	fmt.Printf("%v,%v,%v\n", totalCases, totalTests, totalDeaths)
}

func main() {

	// Check to make sure we received the correct number of arguments
	if len(os.Args) < 5 {
		fmt.Println(usage)
		return
	}

	var threads, zipcode, month, year int
	var err error

	// Retrieve the command-line arguemnts and perform conversion if needed
	if threads, err = strconv.Atoi(os.Args[1]); err != nil {
		fmt.Println(usage)
		return
	}
	if zipcode, err = strconv.Atoi(os.Args[2]); err != nil {
		fmt.Println(usage)
		return
	}
	if month, err = strconv.Atoi(os.Args[3]); err != nil {
		fmt.Println(usage)
		return
	}
	if year, err = strconv.Atoi(os.Args[4]); err != nil {
		fmt.Println(usage)
		return
	}

	start := time.Now()

	if threads <= 1 {
		Sequential(zipcode, month, year)
	} else {

		/** Work Balancing **/
		// initialize the waitgroup and shared goroutine contexts
		var wg sync.WaitGroup
		var rwm sync.RWMutex
		context := utils.SharedContext{
			WgContext:    &wg,
			RWmutex:      &rwm,
			Records:      make(map[string]bool),
			Zipcode:      zipcode,
			Month:        month,
			Year:         year,
			TotalCases:   0,
			TotalTests:   0,
			TotalDeaths:  0,
			FilesCounter: int32(utils.NumOfFiles),
		}

		// Step 1: create a slice of channels
		queues := make([]chan interface{}, threads)

		// Step 2: create a SharingWorker[Threads]
		workers := make([]*balancer.SharingWorker, threads)

		// Step 3: fill up the queues with Runnable tasks
		var start, end int
		var curBT balancer.BalanceTask

		for i := 0; i < threads; i++ {

			workers[i] = balancer.NewSharingWorker(i, &context, &queues, utils.THRESHOLD)

			if i != threads-1 {
				end = end + utils.NumOfFiles/threads
			} else {
				end = utils.NumOfFiles
			}

			queues[i] = make(chan interface{}, end-start)

			for iter := start; iter < end && iter < utils.NumOfFiles; iter++ {
				file := fmt.Sprintf("data/covid_%v.csv", iter+1)
				curBT = balancer.BalanceTask{Filepath: file}
				queues[i] <- curBT
			}

			start = end
		}

		// Step 4: call Run() on each worker
		for i := 0; i < threads; i += 1 {
			context.WgContext.Add(1)
			go workers[i].Run()
		}

		// make the main goroutine wait until others have completed
		wg.Wait()

		// Step 5: print out total counts
		fmt.Printf("%v,%v,%v\n", context.TotalCases, context.TotalTests, context.TotalDeaths)

	}

	fmt.Printf("time: %.5f\n", time.Since(start).Seconds())
}
