package benchmarks

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"code.cloudfoundry.org/lager"
)

// TODO: Benchmark against a real file (maybe /dev/null) to demonstrate
// the impact of calling Write() twice.

type nopWriter struct{}

func (nopWriter) Write(b []byte) (int, error) { return len(b), nil }

func benchmarkWriterSink(b *testing.B, log lager.LogFormat) {
	sink := lager.NewWriterSink(nopWriter{}, lager.DEBUG)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink.Log(log)
	}
}

func BenchmarkWriterSink_Small(b *testing.B) {
	log := lager.LogFormat{
		Timestamp: currentTimestamp(),
		Source:    "log-compenent",
		Message:   "log-message",
		LogLevel:  lager.DEBUG,
		Data: lager.Data{
			"url":    "https://golang.org/pkg/",
			"method": "GET",
		},
	}
	benchmarkWriterSink(b, log)
}

func BenchmarkWriterSink_Large(b *testing.B) {
	var m runtime.MemStats // big struct
	runtime.ReadMemStats(&m)

	log := lager.LogFormat{
		Timestamp: currentTimestamp(),
		Source:    "log-compenent",
		Message:   "log-message",
		LogLevel:  lager.DEBUG,
		Data: lager.Data{
			"mem_stats": m,
			"nested_data": lager.Data{
				"id":     123456,
				"method": "GET",
				"url":    "https://golang.org/pkg/",
			},
		},
	}
	benchmarkWriterSink(b, log)
}

func BenchmarkWriterSink_Parallel(b *testing.B) {
	log := lager.LogFormat{
		Timestamp: currentTimestamp(),
		Source:    "log-compenent",
		Message:   "log-message",
		LogLevel:  lager.DEBUG,
		Data: lager.Data{
			"url":    "https://golang.org/pkg/",
			"method": "GET",
		},
	}
	sink := lager.NewWriterSink(nopWriter{}, lager.DEBUG)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sink.Log(log)
		}
	})
}

func currentTimestamp() string {
	return fmt.Sprintf("%.9f", float64(time.Now().UnixNano())/1e9)
}
