package main

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"golang.org/x/time/rate"
)

type reader struct {
	r       io.Reader
	limiter *rate.Limiter
}

// Reader returns a reader that is rate limited by
// the given token bucket. Each token in the bucket
// represents one byte.
func NewReader(r io.Reader, l *rate.Limiter) io.Reader {
	return &reader{
		r:       r,
		limiter: l,
	}
}

func (r *reader) Read(buf []byte) (int, error) {
	n, err := r.r.Read(buf)
	if n <= 0 {
		return n, err
	}

	now := time.Now()
	rv := r.limiter.ReserveN(now, n)
	if !rv.OK() {
		return 0, fmt.Errorf("Exceeds limiter's burst")
	}
	delay := rv.DelayFrom(now)
	//fmt.Printf("Read %d bytes, delay %d\n", n, delay)
	time.Sleep(delay)
	return n, err
}

func main() {
	// Source holding 1MB
	src := bytes.NewReader(make([]byte, 1024*1024))
	// Destination
	dst := &bytes.Buffer{}

	// Bucket adding 100KB every second, holding max 100KB
	limit := rate.NewLimiter(100*1024, 100*1024)

	start := time.Now()

	buf := make([]byte, 10*1024)
	// Copy source to destination, but wrap our reader with rate limited one
	//io.CopyBuffer(dst, NewReader(src, limit), buf)
	r := NewReader(src, limit)
	for {
		if n, err := r.Read(buf); err == nil {
			dst.Write(buf[0:n])
		} else {
			break
		}
	}

	fmt.Printf("Copied %d bytes in %s\n", dst.Len(), time.Since(start))
}
