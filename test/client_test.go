package main

import "testing"

func BenchmarkHttpClient(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HttpClient()
	}
}
