package test

import (
	"bufio"
	"io"
	"log"
	"os"
	"testing"
)

func BenchmarkUnbufferedWrite(b *testing.B) {
	performWrite(b, tmpFileOrFatal())
}

func BenchmarkBufferedWrite(b *testing.B) {
	buffered_file := bufio.NewWriter(tmpFileOrFatal())
	performWrite(b, bufio.NewWriter(buffered_file))
}

func tmpFileOrFatal() *os.File {
	file, err := os.CreateTemp("", "tmp")
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return file
}

func performWrite(b *testing.B, writer io.Writer) {
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

	done := make(chan any)
	defer close(done)

	b.ResetTimer()
	for bt := range take(done, repeat(done, byte(0)), b.N) {
		writer.Write([]byte{bt.(byte)})
	}
}
