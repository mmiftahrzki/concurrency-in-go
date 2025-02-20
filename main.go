package main

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"sync"
	"text/tabwriter"
	"time"
)

func main() {
	// deadlock()
	// goroutines()
	// mutexRWMutexComparison()
	// cond()
	// condBroadcast()
	// once()
	// once2()
	// onceDeadlock()
	// bufferedChannel()
	// ownershipChannel()
	// selectStatement()
	// selectStatementMultiCase()
	// selectStatementMultiCaseWithTimeout()
	// selectStatementMultiCaseWithDefault()
	// confinementChannelOwnership()
	// confinementBufferAsChannel()
	// goroutinesCancellationWithChannel()
	// goroutineCancellationWriteBlockingIssue()
	// goroutineCancellationWriteBlockingSolution()
	// theOrChannel()
	// errorHandlingIssue()
	// errorHandlingSolution()
	// errorHandlingBreakingWhenTooManyErrorOccurs()
	// pipelineWithFunction()
	// pipelineWithChannel()
	// pipelineGenerator()
	// fanOutFanInIssue()
	// fanOutFanInSolution()
	// theTeeChannel()
	// theBridgeChannel()
}

func deadlock() {
	type value struct {
		mu    sync.Mutex
		value int
	}

	var wg sync.WaitGroup
	printSum := func(v1, v2 *value) {
		defer wg.Done()
		v1.mu.Lock()
		defer v1.mu.Unlock()

		time.Sleep(2 * time.Second)
		v2.mu.Lock()
		defer v2.mu.Unlock()

		fmt.Printf("sum = %v\n", v1.value+v2.value)
	}

	var a, b value
	wg.Add(2)
	go printSum(&a, &b)
	go printSum(&b, &a)
	wg.Wait()
}

func goroutines() {
	var c <-chan interface{}
	var wg sync.WaitGroup

	mem_consumed := func() uint64 {
		runtime.GC()
		var s runtime.MemStats
		runtime.ReadMemStats(&s)

		return s.Sys
	}

	noop := func() {
		wg.Done()
		<-c
	}

	const num_of_gorountines = 1e4
	wg.Add(num_of_gorountines)
	before := mem_consumed()
	for i := num_of_gorountines; i > 0; i-- {
		go noop()
	}
	wg.Wait()
	after := mem_consumed()
	fmt.Printf("%.3fkb", float64(after-before)/num_of_gorountines/1000)
}

func mutexRWMutexComparison() {
	producer := func(wg *sync.WaitGroup, l sync.Locker) {
		defer wg.Done()
		for i := 5; i > 0; i-- {
			l.Lock()
			l.Unlock()
			time.Sleep(1)
		}
	}

	observer := func(wg *sync.WaitGroup, l sync.Locker) {
		defer wg.Done()
		l.Lock()
		defer l.Unlock()
	}

	test := func(count int, mutex, rw_mutex sync.Locker) time.Duration {
		var wg sync.WaitGroup
		wg.Add(count + 1)
		begin_test_time := time.Now()

		go producer(&wg, mutex)
		for i := count; i > 0; i-- {
			go observer(&wg, rw_mutex)
		}

		wg.Wait()

		return time.Since(begin_test_time)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	defer tw.Flush()

	var m sync.RWMutex
	fmt.Fprintf(tw, "Readers\tRWMutex\tMutex\n")
	for i := 0; i < 20; i++ {
		count := int(math.Pow(2, float64(i)))
		fmt.Fprintf(tw, "%d\t%v\t%v\n", count, test(count, &m, m.RLocker()), test(count, &m, &m))
	}
}

func cond() {
	c := sync.NewCond(&sync.Mutex{})
	queue := make([]int, 0, 10)

	removeFromQueue := func(delay time.Duration, value int) {
		time.Sleep(delay)
		c.L.Lock()
		queue = queue[1:]
		fmt.Printf("Removed %d from queue...\n", value)
		c.L.Unlock()
		c.Signal()
	}

	for i := 0; i < 10; i++ {
		value := i + 1
		c.L.Lock()
		for len(queue) == 2 {
			c.Wait()
		}

		fmt.Printf("Adding %d to queue...\n", value)
		queue = append(queue, value)
		go removeFromQueue(1*time.Second, value)
		c.L.Unlock()
	}
}

func condBroadcast() {
	type Button struct {
		Clicked *sync.Cond
	}

	button := Button{Clicked: sync.NewCond(&sync.Mutex{})}

	subscribe := func(c *sync.Cond, fn func()) {
		var gorouting_running sync.WaitGroup
		gorouting_running.Add(1)
		go func() {
			gorouting_running.Done()
			c.L.Lock()
			defer c.L.Unlock()
			c.Wait()
			fn()
		}()

		gorouting_running.Wait()
	}

	var clickRegistered sync.WaitGroup
	clickRegistered.Add(3)

	subscribe(button.Clicked, func() {
		defer clickRegistered.Done()
		fmt.Println("Maximizing window.")
	})

	subscribe(button.Clicked, func() {
		defer clickRegistered.Done()
		fmt.Println("Displaying annoying dialog box!")
	})

	subscribe(button.Clicked, func() {
		defer clickRegistered.Done()
		fmt.Println("Mouse clicked.")
	})

	button.Clicked.Broadcast()
	clickRegistered.Wait()
}

func once() {
	var count int
	var increment func()
	var once sync.Once
	var increments sync.WaitGroup

	increment = func() {
		count++
	}

	increments.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer increments.Done()
			once.Do(increment)
		}()
	}
	increments.Wait()
	fmt.Printf("Count is %d\n", count)
}

func once2() {
	var count int
	var increment func()
	var decrement func()
	var once sync.Once

	increment = func() {
		count++
	}
	decrement = func() {
		count--
	}

	once.Do(increment)
	once.Do(decrement)

	fmt.Printf("Count: %d\n", count)
}

func onceDeadlock() {
	var once_a, once_b sync.Once
	var initA func()
	var initB func()

	initA = func() {
		once_b.Do(initB)
	}
	initB = func() {
		once_a.Do(initA)
	}

	once_a.Do(initA)
}

func bufferedChannel() {
	var std_out_buffer bytes.Buffer
	var int_stream chan int

	defer std_out_buffer.WriteTo(os.Stdout)
	int_stream = make(chan int, 4)

	go func() {
		defer close(int_stream)
		defer fmt.Fprintln(&std_out_buffer, "Producer done.")

		for i := 0; i < 5; i++ {
			fmt.Fprintf(&std_out_buffer, "Sending: %d\n", i)
			int_stream <- i
		}
	}()

	for integer := range int_stream {
		fmt.Fprintf(&std_out_buffer, "Received %v.\n", integer)
	}
}

func ownershipChannel() {
	var wg sync.WaitGroup
	var newDataStream func() <-chan int

	newDataStream = func() <-chan int {
		var int_stream chan int

		int_stream = make(chan int, 5)
		go func() {
			defer close(int_stream)

			for i := 0; i < 5; i++ {
				int_stream <- i
			}
		}()

		return int_stream
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for data := range newDataStream() {
			fmt.Printf("Received: %d\n", data)
		}
	}()

	wg.Wait()
	fmt.Println("Done receiving!")
}

func selectStatement() {
	start := time.Now()
	c := make(chan interface{})

	go func() {
		time.Sleep(5 * time.Second)
		close(c)
	}()

	fmt.Println("Blocking on read...")
	select {
	case <-c:
		fmt.Printf("Unblokcing %v later\n", time.Since(start))
	}

}

func selectStatementMultiCase() {
	c1 := make(chan int)
	c2 := make(chan int)

	close(c1)
	close(c2)

	var c1Count, c2Count int
	for i := 1000; i > 0; i-- {
		select {
		case <-c1:
			c1Count++
		case <-c2:
			c2Count++
		}
	}

	fmt.Printf("c1Count: %d\nc2Count: %d\n", c1Count, c2Count)
}

func selectStatementMultiCaseWithTimeout() {
	var c1 <-chan int

	select {
	case <-c1:
	case <-time.After(1 * time.Second):
		fmt.Println("Timed out.")
	}
}

func selectStatementMultiCaseWithDefault() {
	done := make(chan interface{})

	go func() {
		time.Sleep(5 * time.Second)
		close(done)
	}()

	work_counter := 0

loop:
	for {
		select {
		case <-done:
			break loop
		default:
		}

		work_counter++
		time.Sleep(1 * time.Second)
	}

	fmt.Printf("Achieved %v cycles of work before signalled to stop.", work_counter)
}

func confinementChannelOwnership() {
	channelOwner := func() <-chan int {
		integer_stream := make(chan int)

		go func() {
			defer close(integer_stream)

			for i := 0; i < 5; i++ {
				integer_stream <- i
			}
		}()

		return integer_stream
	}

	channelConsumer := func(integer_stream <-chan int) {
		for integer := range integer_stream {
			fmt.Printf("Received: %d\n", integer)
		}

		fmt.Println("Done receiving!")
	}

	results := channelOwner()
	channelConsumer(results)
}

func confinementBufferAsChannel() {
	printData := func(wg *sync.WaitGroup, data []byte) {
		defer wg.Done()

		var buffer bytes.Buffer
		for _, b := range data {
			fmt.Fprintf(&buffer, "%c", b)
		}

		fmt.Println(buffer.String())
	}

	var wg sync.WaitGroup
	wg.Add(2)
	data := []byte("golang")

	go printData(&wg, data[:3])
	go printData(&wg, data[3:])

	wg.Wait()
}

func goroutinesCancellationWithChannel() {
	var doWork func(done <-chan any, strings <-chan string) <-chan any
	var done chan any
	var terminated <-chan any

	doWork = func(done <-chan any, strings <-chan string) <-chan any {
		var terminated chan any
		terminated = make(chan any)

		go func() {
			defer fmt.Println("doWork exited...")
			defer close(terminated)

			for {
				select {
				case s := <-strings:
					fmt.Println(s)
				case <-done:
					return
				}
			}
		}()

		return terminated
	}

	done = make(chan any)
	terminated = doWork(done, nil)

	go func() {
		// time.Sleep(1 * time.Second)
		fmt.Println("Cancelling doWork goroutine...")

		close(done)
	}()

	<-terminated
	fmt.Println("Done.")
}

func goroutineCancellationWriteBlockingIssue() {
	var newStream func() <-chan int
	var rand_stream <-chan int

	newStream = func() <-chan int {
		var stream chan int
		stream = make(chan int)

		go func() {
			defer fmt.Println("newStream closure exited.")
			defer close(stream)

			for {
				stream <- rand.Int()
			}
		}()

		return stream
	}

	rand_stream = newStream()
	fmt.Println("3 random ints: ")

	for i := 1; i <= 3; i++ {
		fmt.Printf("%d: %d\n", i, <-rand_stream)
	}
}

func goroutineCancellationWriteBlockingSolution() {
	var newIntStream func(done <-chan any) <-chan int
	var rand_int_stream <-chan int

	newIntStream = func(done <-chan any) <-chan int {
		var stream chan int
		stream = make(chan int)

		go func() {
			defer fmt.Println("Integer stream closure exited.")
			defer close(stream)

			for {
				select {
				case stream <- rand.Int():
				case <-done:
					return
				}
			}
		}()

		return stream
	}

	done := make(chan any)
	rand_int_stream = newIntStream(done)
	fmt.Println("3 random ints: ")

	for i := 1; i <= 3; i++ {
		fmt.Printf("%d: %d\n", i, <-rand_int_stream)
	}
	close(done)
	time.Sleep(1 * time.Second)
}

func theOrChannel() {
	var or func(channels ...<-chan any) <-chan any

	or = func(channels ...<-chan any) <-chan any {
		switch len(channels) {
		case 0:
			return nil
		case 1:
			return channels[0]
		}

		orDone := make(chan any)
		go func() {
			defer close(orDone)

			switch len(channels) {
			case 2:
				select {
				case <-channels[0]:
				case <-channels[1]:
				}
			default:
				select {
				case <-channels[0]:
				case <-channels[1]:
				case <-channels[2]:
				case <-or(append(channels[3:], orDone)...):
				}
			}
		}()

		return orDone
	}

	sig := func(after time.Duration) <-chan any {
		stream := make(chan any)

		go func() {
			defer close(stream)
			time.Sleep(after)
		}()

		return stream
	}

	start := time.Now()
	<-or(
		sig(1*time.Second),
		sig(3*time.Second),
		sig(5*time.Second),
	)

	fmt.Printf("done after %v", time.Since(start))
}

func errorHandlingIssue() {
	checkStatus := func(urls ...string) <-chan *http.Response {
		responses := make(chan *http.Response)

		go func() {
			defer close(responses)

			for _, url := range urls {
				resp, err := http.Get(url)
				if err != nil {
					fmt.Println(err)

					continue
				}

				responses <- resp
			}
		}()

		return responses
	}

	urls := []string{"https://google.com", "https://badhost", "https://google.com"}
	for response := range checkStatus(urls...) {
		fmt.Printf("Response: %v\n", response.Status)
	}
}

func errorHandlingSolution() {
	type Result struct {
		Error    error
		Response *http.Response
	}

	checkStatus := func(urls ...string) <-chan Result {
		results := make(chan Result)

		go func() {
			defer close(results)

			for _, url := range urls {
				resp, err := http.Get(url)
				results <- Result{Error: err, Response: resp}
			}
		}()

		return results
	}

	urls := []string{"https://google.com", "https://badhost", "https://x.com", "https://anotherbadhost"}
	for result := range checkStatus(urls...) {
		if result.Error != nil {
			fmt.Printf("error: %v\n", result.Error)

			continue
		}

		fmt.Printf("Response: %v\n", result.Response.Status)
	}

}

func errorHandlingBreakingWhenTooManyErrorOccurs() {
	type Result struct {
		Error    error
		Response *http.Response
	}

	checkStatus := func(done <-chan any, urls ...string) <-chan Result {
		results := make(chan Result)

		go func() {
			defer close(results)

			for _, url := range urls {
				var result Result
				resp, err := http.Get(url)
				result = Result{Error: err, Response: resp}

				select {
				case <-done:
					return
				case results <- result:
				}
			}
		}()

		return results
	}

	done := make(chan any)
	defer close(done)

	err_count := 0
	urls := []string{
		"https://google.com",
		"https://youtube.com",
		"https://1111ft4h.xyz",
		"https://badhost",
		"https://anotherbadhost",
		"https://x.com",
	}
	for result := range checkStatus(done, urls...) {
		if result.Error != nil {
			fmt.Printf("error: %v\n", result.Error)
			err_count++
			if err_count >= 2 {
				fmt.Println("Too many errors, breaking!")

				break
			}

			continue
		}

		fmt.Printf("Response: %v\n", result.Response.Status)
	}
}

func pipelineWithFunction() {
	multiply := func(values []int, multiplier int) []int {
		multiplied_values := make([]int, len(values))

		for i, v := range values {
			multiplied_values[i] = v * multiplier
		}

		return multiplied_values
	}

	add := func(values []int, additive int) []int {
		added_values := make([]int, len(values))

		for i, v := range values {
			added_values[i] = v + additive
		}

		return added_values
	}

	ints := []int{1, 2, 3, 4}
	for _, v := range multiply(add(multiply(ints, 2), 1), 2) {
		fmt.Println(v)
	}
}

func pipelineWithChannel() {
	generator := func(done <-chan bool, ints ...int) <-chan int {
		int_stream := make(chan int)

		go func() {
			defer close(int_stream)

			for _, v := range ints {
				select {
				case <-done:
					return
				case int_stream <- v:
				}
			}
		}()

		return int_stream
	}

	multiply := func(done <-chan bool, int_stream <-chan int, multiplier int) <-chan int {
		multiplied_stream := make(chan int)

		go func() {
			defer close(multiplied_stream)

			for v := range int_stream {
				multiplied_v := v * multiplier

				select {
				case <-done:
					return
				case multiplied_stream <- multiplied_v:
				}
			}
		}()

		return multiplied_stream
	}

	add := func(done <-chan bool, int_stream <-chan int, additive int) <-chan int {
		added_stream := make(chan int)

		go func() {
			defer close(added_stream)

			for v := range int_stream {
				added_v := v + additive

				select {
				case <-done:
					return
				case added_stream <- added_v:
				}
			}
		}()

		return added_stream
	}

	done := make(chan bool)
	defer close(done)

	int_stream := generator(done, 1, 2, 3, 4)
	pipeline := multiply(done, add(done, multiply(done, int_stream, 2), 1), 2)
	for v := range pipeline {
		fmt.Println(v)
	}
}

func pipelineGenerator() {
	repeat := func(
		done <-chan any,
		values ...any,
	) <-chan any {
		value_stream := make(chan any)

		go func() {
			defer close(value_stream)

			for {
				for _, v := range values {
					select {
					case <-done:
						return
					case value_stream <- v:
					}
				}
			}
		}()

		return value_stream
	}

	repeatFn := func(done <-chan any, fn func() any) <-chan any {
		value_stream := make(chan any)

		go func() {
			defer close(value_stream)

			for {
				select {
				case <-done:
					return
				case value_stream <- fn():
				}
			}
		}()

		return value_stream
	}

	take := func(done <-chan any, value_stream <-chan any, num int) <-chan any {
		take_stream := make(chan any)

		go func() {
			defer close(take_stream)

			for i := 0; i < num; i++ {
				select {
				case <-done:
					return
				case take_stream <- <-value_stream:
				}
			}
		}()

		return take_stream
	}

	random_integer := func() any { return rand.Int() }

	done := make(chan any)
	defer close(done)

	for num := range take(done, repeat(done, 1), 10) {
		fmt.Printf("%v ", num)
	}

	for num := range take(done, repeatFn(done, random_integer), 10) {
		fmt.Println(num)
	}
}

func fanOutFanInIssue() {
	rand := func() any {
		return rand.Intn(50000000)
	}

	repeatFn := func(done <-chan any, fn func() any) <-chan any {
		value_stream := make(chan any)

		go func() {
			defer close(value_stream)

			for {
				select {
				case <-done:
					return
				case value_stream <- fn():
				}
			}
		}()

		return value_stream
	}

	toInt := func(done <-chan any, value_stream <-chan any) <-chan int {
		int_stream := make(chan int)

		go func() {
			defer close(int_stream)

			for v := range value_stream {
				select {
				case <-done:
					return
				case int_stream <- v.(int):
				}
			}
		}()

		return int_stream
	}

	isPrime := func(n int) bool {
		if n < 2 {
			return false
		}

		// Periksa setiap pembagi dari 2 hingga n-1
		for i := 2; i < n; i++ {
			if n%i == 0 {
				return false
			}
		}

		return true
	}

	primeFinder := func(done <-chan any, int_stream <-chan int) <-chan int {
		prime_stream := make(chan int)

		go func() {
			defer close(prime_stream)

			for v := range int_stream {
				if !isPrime(v) {
					continue
				}

				select {
				case <-done:
					return
				case prime_stream <- v:
				}
			}
		}()

		return prime_stream
	}

	take := func(done <-chan any, value_stream <-chan int, num int) <-chan any {
		take_stream := make(chan any)

		go func() {
			defer close(take_stream)

			for i := 0; i < num; i++ {
				select {
				case <-done:
					return
				case take_stream <- <-value_stream:
				}
			}
		}()

		return take_stream
	}

	done := make(chan any)
	defer close(done)

	start := time.Now()

	rand_int_stream := toInt(done, repeatFn(done, rand))
	fmt.Println("Primes:")
	i := 0
	for prime := range take(done, primeFinder(done, rand_int_stream), 10) {
		i++
		fmt.Printf("\t%d: %d\n", i, prime)
	}

	fmt.Printf("Search took: %v\n", time.Since(start))
}

func fanOutFanInSolution() {
	rand := func() any {
		return rand.Intn(50000000)
	}

	repeatFn := func(done <-chan any, fn func() any) <-chan any {
		value_stream := make(chan any)

		go func() {
			defer close(value_stream)

			for {
				select {
				case <-done:
					return
				case value_stream <- fn():
				}
			}
		}()

		return value_stream
	}

	toInt := func(done <-chan any, value_stream <-chan any) <-chan int {
		int_stream := make(chan int)

		go func() {
			defer close(int_stream)

			for v := range value_stream {
				select {
				case <-done:
					return
				case int_stream <- v.(int):
				}
			}
		}()

		return int_stream
	}

	isPrime := func(n int) bool {
		if n < 2 {
			return false
		}

		// Periksa setiap pembagi dari 2 hingga n-1
		for i := 2; i < n; i++ {
			if n%i == 0 {
				return false
			}
		}

		return true
	}

	primeFinder := func(done <-chan any, int_stream <-chan int) <-chan any {
		prime_stream := make(chan any)

		go func() {
			defer close(prime_stream)

			for v := range int_stream {
				if !isPrime(v) {
					continue
				}

				select {
				case <-done:
					return
				case prime_stream <- v:
				}
			}
		}()

		return prime_stream
	}

	take := func(done <-chan any, value_stream <-chan int, num int) <-chan any {
		take_stream := make(chan any)

		go func() {
			defer close(take_stream)

			for i := 0; i < num; i++ {
				select {
				case <-done:
					return
				case take_stream <- <-value_stream:
				}
			}
		}()

		return take_stream
	}

	fanIn := func(done <-chan any, channels ...<-chan any) <-chan int {
		var wg sync.WaitGroup

		multiplexed_stream := make(chan int)

		multiplex := func(c <-chan any) {
			defer wg.Done()

			for i := range c {
				select {
				case <-done:
					return
				case multiplexed_stream <- i.(int):
				}
			}
		}

		wg.Add(len(channels))
		for _, c := range channels {
			go multiplex(c)
		}

		go func() {
			wg.Wait()
			close(multiplexed_stream)
		}()

		return multiplexed_stream
	}

	done := make(chan any)
	defer close(done)

	start := time.Now()

	rand_int_stream := toInt(done, repeatFn(done, rand))
	num_finders := runtime.NumCPU()
	fmt.Printf("Spinning up %d prime finders.\n", num_finders)
	finders := make([]<-chan any, num_finders)
	fmt.Println("Primes:")
	for i := 0; i < num_finders; i++ {
		finders[i] = primeFinder(done, rand_int_stream)
	}

	i := 0
	for prime := range take(done, fanIn(done, finders...), 10) {
		i++
		fmt.Printf("\t%d: %d\n", i, prime)
	}

	fmt.Printf("Search took: %v", time.Since(start))
}

/*
The orDone channel pattern is used in the same way as the preventing goroutine leaks pattern, but it is applied when we are working with a channel that is separate from our system and we do not know if the channel has been canceled.
*/
func theOrDoneChannel() {
	orDone := func(done, c <-chan any) <-chan any {
		value_stream := make(chan any)

		go func() {
			defer close(value_stream)

			for {
				select {
				case <-done:
					return
				case value, ok := <-c:
					if ok == false {
						return
					}

					select {
					case value_stream <- value:
					case <-done:
					}
				}
			}
		}()

		return value_stream
	}

	done := make(chan any)
	defer close(done)

	my_chan := func(done <-chan any) <-chan any {
		my_stream := make(chan any)
		close(my_stream)

		return my_stream
	}(done)

	for val := range orDone(done, my_chan) {
		fmt.Println(val)
	}
}

func theTeeChannel() {
	repeat := func(
		done <-chan any,
		values ...any,
	) <-chan any {
		value_stream := make(chan any)

		go func() {
			defer close(value_stream)

			for {
				for _, v := range values {
					select {
					case <-done:
						return
					case value_stream <- v:
						fmt.Printf("Generated: %v\n", v)
					}
				}
			}
		}()

		return value_stream
	}

	take := func(done <-chan any, value_stream <-chan any, num int) <-chan any {
		take_stream := make(chan any)

		go func() {
			defer close(take_stream)

			for i := 0; i < num; i++ {
				select {
				case <-done:
					return
				case take_stream <- <-value_stream:
				}
			}
		}()

		return take_stream
	}

	orDone := func(done, c <-chan any) <-chan any {
		value_stream := make(chan any)

		go func() {
			defer close(value_stream)

			for {
				select {
				case <-done:
					return
				case value, ok := <-c:
					if ok == false {
						return
					}

					select {
					case value_stream <- value:
					case <-done:
					}
				}
			}
		}()

		return value_stream
	}

	tee := func(
		done <-chan any,
		in <-chan any,
	) (<-chan any, <-chan any) {
		out1 := make(chan any)
		out2 := make(chan any)

		go func() {
			defer close(out1)
			defer close(out2)

			for val := range orDone(done, in) {
				var out1, out2 = out1, out2

				for i := 0; i < 2; i++ {
					select {
					case <-done:
						/*
							kenapa masing2 case ini ketika casenya occured langsung diset jadi nil?
							karena agar channel yang sudah pernah dikirim value dari channel in, tidak dapat value lagi sehingga bisa bergantian ke case channel lainnya yang belum pernah dapat value dari channel in.
						*/
					case out1 <- val:
						out1 = nil
					case out2 <- val:
						out2 = nil
					}
				}
			}
		}()

		return out1, out2
	}

	done := make(chan any)
	defer close(done)

	out1, out2 := tee(done, take(done, repeat(done, 1, 2), 4))

	for val1 := range out1 {
		fmt.Printf("out1: %v, out2: %v\n", val1, <-out2)
	}
}

func theBridgeChannel() {
	orDone := func(done <-chan bool, c <-chan any) <-chan any {
		value_stream := make(chan any)

		go func() {
			defer close(value_stream)

			for {
				select {
				case <-done:
					return
				case value, ok := <-c:
					if !ok {
						return
					}

					select {
					case <-done:
					case value_stream <- value:
					}
				}
			}
		}()

		return value_stream
	}

	bridge := func(done <-chan bool, chan_stream <-chan (<-chan any)) <-chan any {
		value_stream := make(chan any)

		go func() {
			defer close(value_stream)

			for {
				var stream <-chan any

				select {
				case maybe_stream, ok := <-chan_stream:
					if !ok {
						return
					}

					stream = maybe_stream
				case <-done:
					return
				}

				for value := range orDone(done, stream) {
					select {
					case value_stream <- value:
					case <-done:
					}
				}
			}
		}()

		return value_stream
	}

	genVals := func() <-chan <-chan any {
		chan_stream := make(chan (<-chan any))

		go func() {
			defer close(chan_stream)

			for i := 0; i < 10; i++ {
				stream := make(chan any, 1)
				stream <- i
				close(stream)
				chan_stream <- stream
			}
		}()

		return chan_stream
	}

	for v := range bridge(nil, genVals()) {
		fmt.Printf("%v ", v)
	}
}
