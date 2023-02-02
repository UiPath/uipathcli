package executor

import (
	"io"
	"time"
)

type ProgressReader struct {
	io.Reader
	ProgressFunc func(progress Progress)
	startTime    time.Time
	bytesRead    int64
	lastProgress time.Time
}

func (r *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	bytesRead, bytesPerSecond := r.calculateStats(n)
	if err == io.EOF || time.Since(r.lastProgress).Nanoseconds() > (100*time.Nanosecond).Nanoseconds() {
		r.lastProgress = time.Now()
		progress := NewProgress(bytesRead, bytesPerSecond, err == io.EOF)
		r.ProgressFunc(*progress)
	}
	return n, err
}

func (r *ProgressReader) calculateStats(n int) (int64, int64) {
	bytesRead := r.bytesRead + int64(n)
	seconds := time.Since(r.startTime).Seconds()
	bytesPerSecond := int64(0)
	if seconds > 0 {
		bytesPerSecond = int64(float64(bytesRead) / seconds)
	}
	r.bytesRead = bytesRead
	return bytesRead, bytesPerSecond
}

func NewProgressReader(reader io.Reader, progressFunc func(progress Progress)) *ProgressReader {
	return &ProgressReader{reader, progressFunc, time.Now(), 0, time.Time{}}
}
