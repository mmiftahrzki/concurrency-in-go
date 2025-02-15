package test

import "testing"

func BenchmarkGeneric(b *testing.B) {
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

	toString := func(done <-chan any, value_stream <-chan any) <-chan string {
		string_stream := make(chan string)

		go func() {
			defer close(string_stream)

			for v := range value_stream {
				select {
				case <-done:
					return
				case string_stream <- v.(string):
				}
			}
		}()

		return string_stream
	}

	done := make(chan any)
	defer close(done)

	b.ResetTimer()
	for range toString(done, take(done, repeat(done, "a"), b.N)) {
	}
}

func BenchmarkTyped(b *testing.B) {
	repeat := func(done <-chan any, values ...string) <-chan string {
		value_stream := make(chan string)

		go func() {
			defer close(value_stream)
			for _, v := range values {
				select {
				case <-done:
					return
				case value_stream <- v:
				}
			}
		}()

		return value_stream
	}

	take := func(done <-chan any, value_stream <-chan string, num int) <-chan string {
		take_stream := make(chan string)

		go func() {
			defer close(take_stream)

			for i := num; i > 0 || i == -1; {
				if i != -1 {
					i--
				}

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
	for range take(done, repeat(done, "a"), b.N) {
	}
}
