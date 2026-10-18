package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {

	job := make(chan string, 1024)

	wg := sync.WaitGroup{}
	wg.Add(3)
	go worker(job, &wg)
	go worker(job, &wg)
	go worker(job, &wg)

	for i:=97;i<103;i++{
		job<-"pdf "+string(i)
	}

	close(job)

	wg.Wait()

}

func worker(jobs <-chan string, wg *sync.WaitGroup) {

	defer wg.Done()

	for job := range jobs {
		processPdf(job)
	}

}

func processPdf(file string){
	fmt.Println("Processing pdf",file)
	time.Sleep(3*time.Second)
	fmt.Println("Processing finished")
}
