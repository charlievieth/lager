package lager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// A Sink represents a write destination for a Logger. It provides
// a thread-safe interface for writing logs
type Sink interface {
	//Log to the sink.  Best effort -- no need to worry about errors.
	Log(log LogFormat)
}

type writerSink struct {
	writer      io.Writer
	minLogLevel LogLevel
	writeL      sync.Mutex
}

func NewWriterSink(writer io.Writer, minLogLevel LogLevel) Sink {
	return &writerSink{
		writer:      writer,
		minLogLevel: minLogLevel,
	}
}

func (w *writerSink) handleErr(log *LogFormat, err error) {
	switch err.(type) {
	case *json.UnsupportedTypeError, *json.MarshalerError:
		log.Data = Data{
			"lager serialisation error": err.Error(),
			"data_dump":                 fmt.Sprintf("%#v", log.Data),
		}
		if e := w.encode(log); e != nil {
			panic(e) // panic with new error
		}
	default:
		panic(err) // unhandled error
	}
}

func (w *writerSink) encode(log *LogFormat) error {
	buf := newBuffer()
	if err := json.NewEncoder(buf).Encode(log); err != nil {
		bufferPool.Put(buf)
		return err
	}
	w.writeL.Lock()
	_, err := buf.WriteTo(w.writer)
	w.writeL.Unlock()
	bufferPool.Put(buf)
	return err
}

func (sink *writerSink) Log(log LogFormat) {
	if log.LogLevel >= sink.minLogLevel {
		if err := sink.encode(&log); err != nil {
			sink.handleErr(&log, err)
		}
	}
}

var bufferPool sync.Pool

func newBuffer() *bytes.Buffer {
	if v := bufferPool.Get(); v != nil {
		b := v.(*bytes.Buffer)
		b.Reset()
		return b
	}
	return new(bytes.Buffer)
}
