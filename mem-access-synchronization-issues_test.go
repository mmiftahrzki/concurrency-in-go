package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type value struct {
	mu    sync.Mutex
	value int
}

func greedyWorker(wg *sync.WaitGroup, shared_lock *sync.Mutex, runtime time.Duration) {
	defer wg.Done()

	var count int
	for begin := time.Now(); time.Since(begin) <= runtime; {
		shared_lock.Lock()
		time.Sleep(1 * time.Nanosecond)
		shared_lock.Unlock()

		count++
	}

	fmt.Printf("Greedy worker was able to execute %v work loops\n", count)
}

func politeWorker(wg *sync.WaitGroup, shared_lock *sync.Mutex, runtime time.Duration) {
	defer wg.Done()

	var count int
	for begin := time.Now(); time.Since(begin) <= runtime; {
		shared_lock.Lock()
		time.Sleep(1 * time.Nanosecond)
		shared_lock.Unlock()

		shared_lock.Lock()
		time.Sleep(1 * time.Nanosecond)
		shared_lock.Unlock()

		shared_lock.Lock()
		time.Sleep(1 * time.Nanosecond)
		shared_lock.Unlock()

		count++
	}

	fmt.Printf("Polite worker was able to execute %v work loops\n", count)
}

func printSum(wg *sync.WaitGroup, v1, v2 *value) {
	defer wg.Done()
	v1.mu.Lock()
	defer v1.mu.Unlock()

	time.Sleep(2 * time.Second)
	v2.mu.Lock()
	defer v2.mu.Unlock()

	fmt.Printf("sum = %v\n", v1.value+v2.value)
}

func TestStarvation(t *testing.T) {
	var wg sync.WaitGroup
	var shared_lock sync.Mutex
	var runtime = 1 * time.Second

	wg.Add(2)
	go greedyWorker(&wg, &shared_lock, runtime)
	go politeWorker(&wg, &shared_lock, runtime)
	wg.Wait()
}

func TestDeadlock(t *testing.T) {
	var wg sync.WaitGroup
	var a, b value

	wg.Add(2)
	go printSum(&wg, &a, &b)
	go printSum(&wg, &b, &a)
	wg.Wait()
}
