package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func locale(done <-chan any) (string, error) {
	select {
	case <-done:
		return "", fmt.Errorf("cancelled")
	case <-time.After(3 * time.Second):
	}

	return "ID/ID", nil
}

func genGreeting(done <-chan any) (string, error) {
	switch locale, err := locale(done); {
	case err != nil:
		return "", err
	case locale == "ID/ID":
		return "halo", nil
	}

	return "", fmt.Errorf("unsupported locale")
}

func genFarewell(done <-chan any) (string, error) {
	switch locale, err := locale(done); {
	case err != nil:
		return "", err
	case locale == "ID/ID":
		return "sampai jumpa", nil
	}

	return "", fmt.Errorf("unsupported locale")
}

func printGreeting(done <-chan any) error {
	greeting, err := genGreeting(done)

	if err != nil {
		return err
	}

	fmt.Printf("%s, dunia!\n", greeting)

	return nil
}

func printFarewell(done <-chan any) error {
	farewell, err := genFarewell(done)
	if err != nil {
		return err
	}

	fmt.Printf("%s, dunia!\n", farewell)

	return nil
}

func TestDoneChannelPattern(t *testing.T) {
	var wg sync.WaitGroup
	done := make(chan any)
	defer close(done)

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := printGreeting(done); err != nil {
			fmt.Printf("%v", err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := printFarewell(done); err != nil {
			fmt.Printf("%v", err)
			return
		}
	}()

	defer wg.Wait()
}

func localeWithCtx(ctx context.Context) (string, error) {
	/*
		Pemeriksaan deadline di bawah ini dapat digunakan jika fungsi ini ingin menerapkan early return. Menagapa early return? karena bisa saja algoritma keseluruhan fungsi ini costnya lumayan mahal.
	*/
	// if deadline, ok := ctx.Deadline(); ok {
	// 	if deadline.Sub(time.Now().Add(1*time.Minute)) <= 0 {
	// 		return "", context.DeadlineExceeded
	// 	}
	// }

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(1 * time.Minute):
	}

	return "ID/ID", nil
}

func genGreetingUsingCtx(ctx context.Context) (string, error) {
	switch locale, err := localeWithCtx(ctx); {
	case err != nil:
		return "", err
	case locale == "ID/ID":
		return "halo", nil
	}

	return "", fmt.Errorf("unsupported locale")
}

func genFarewellUsingCtx(ctx context.Context) (string, error) {
	switch locale, err := localeWithCtx(ctx); {
	case err != nil:
		return "", err
	case locale == "ID/ID":
		return "selamat tinggal", nil
	}

	return "", fmt.Errorf("unsupported locale")
}

func printGreetingUsingCtx(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	greeting, err := genGreetingUsingCtx(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("%s, dunia!\n", greeting)

	return nil
}

func printFarewellUsingCtx(ctx context.Context) error {
	greeting, err := genFarewellUsingCtx(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("%s, dunia!\n", greeting)

	return nil
}

func TestContextWithTimeOut(t *testing.T) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printGreetingUsingCtx(ctx); err != nil {
			fmt.Printf("cannot print greeting: %v\n", err)
			cancel()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printFarewellUsingCtx(ctx); err != nil {
			fmt.Printf("cannot print farewell: %v\n", err)
		}
	}()

	wg.Wait()
}
