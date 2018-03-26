package perf_test

import "testing"

const arraySize = 100

func BenchmarkRangeLoop(b *testing.B) {
	data := make([]int, arraySize)
	for i := range data {
		data[i] = i
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := range data {
			_ = data[i]
		}
	}
}

func BenchmarkForLoop(b *testing.B) {
	data := make([]int, arraySize)
	for i := 0; i < arraySize; i++ {
		data[i] = i
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := 0; i < arraySize; i++ {
			_ = data[i]
		}
	}
}

type ibox struct {
	int
}

func BenchmarkIndirectRangeLoop(b *testing.B) {
	data := make([]*ibox, arraySize)
	for i := range data {
		data[i] = &ibox{i}
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := range data {
			_ = data[i].int
		}
	}
}

func BenchmarkIndirectForLoop(b *testing.B) {
	data := make([]*ibox, arraySize)
	for i := 0; i < arraySize; i++ {
		data[i] = &ibox{i}
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := 0; i < arraySize; i++ {
			_ = data[i].int
		}
	}
}
