package benchmarks

import (
	"fmt"
	"io"
	"runtime"
	"testing"
	"time"

	"code.cloudfoundry.org/lager"
)

// TODO: Benchmark against a real file (maybe /dev/null) to demonstrate
// the impact of calling Write() twice.

type nopWriter struct{}

func (nopWriter) Write(b []byte) (int, error) { return len(b), nil }

// This should be a no-op
func BenchmarkLogger_NoSink(b *testing.B) {
	l := lager.NewLogger("benchmark")
	for i := 0; i < b.N; i++ {
		l.Debug("debug")
	}
}

func BenchmarkLogger(b *testing.B) {
	l := lager.NewLogger("benchmark")
	l.RegisterSink(lager.NewWriterSink(nopWriter{}, lager.DEBUG))
	for i := 0; i < b.N; i++ {
		l.Debug("debug")
	}
}

// When the log level is insufficient to trigger a write this should be a no-op
func BenchmarkLogger_Noop(b *testing.B) {
	l := lager.NewLogger("benchmark")
	l.RegisterSink(lager.NewWriterSink(nopWriter{}, lager.ERROR))
	for i := 0; i < b.N; i++ {
		l.Debug("debug")
	}
}

func BenchmarkLogger_Data(b *testing.B) {
	data := lager.Data{
		"id":     123456,
		"method": "GET",
		"url":    "https://golang.org/pkg/",
	}
	l := lager.NewLogger("benchmark")
	l.RegisterSink(lager.NewWriterSink(nopWriter{}, lager.DEBUG))
	for i := 0; i < b.N; i++ {
		l.Debug("debug", data)
	}
}

// Highlight the impact of adding another value to the map after its
// been constructed.
func BenchmarkLogger_Data_Error(b *testing.B) {
	data := lager.Data{
		"id":     123456,
		"method": "GET",
		"url":    "https://golang.org/pkg/",
	}
	l := lager.NewLogger("benchmark")
	l.RegisterSink(lager.NewWriterSink(nopWriter{}, lager.DEBUG))
	for i := 0; i < b.N; i++ {
		l.Error("error", io.EOF, data)
	}
}

func BenchmarkLogger_WithData(b *testing.B) {
	data := lager.Data{
		"id":     123456,
		"method": "GET",
		"url":    "https://golang.org/pkg/",
	}
	l := lager.NewLogger("benchmark").WithData(data)
	l.RegisterSink(lager.NewWriterSink(nopWriter{}, lager.DEBUG))
	for i := 0; i < b.N; i++ {
		l.Debug("debug")
	}
}

func BenchmarkLogger_CombineData(b *testing.B) {
	data := lager.Data{
		"id":     123456,
		"method": "GET",
		"url":    "https://golang.org/pkg/",
	}
	l := lager.NewLogger("benchmark").WithData(data)
	l.RegisterSink(lager.NewWriterSink(nopWriter{}, lager.DEBUG))
	for i := 0; i < b.N; i++ {
		l.Debug("debug", data)
	}
}

func BenchmarkLogger_CombineData_Error(b *testing.B) {
	data := lager.Data{
		"id":     123456,
		"method": "GET",
		"url":    "https://golang.org/pkg/",
	}
	l := lager.NewLogger("benchmark").WithData(data)
	l.RegisterSink(lager.NewWriterSink(nopWriter{}, lager.DEBUG))
	for i := 0; i < b.N; i++ {
		l.Error("debug", io.EOF, data)
	}
}

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
