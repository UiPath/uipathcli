package utils

import (
	"io"
	"time"
)

// The ProgressReader is a wrapper over the io.Reader interface which computes statistics
// while the data is read. This is used to show progress for file uploads and downloads.
//
// The ProgressReader emits a Progress event whenever data is read from the reader.
// The event is debounced to avoid too many events.
type ProgressReader struct {
	io.Reader
	progressFunc func(progress Progress)
	startTime    time.Time
	bytesRead    int64
	lastProgress time.Time
}

const debounceInterval = 100 * time.Nanosecond

func (r *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	bytesRead, bytesPerSecond := r.calculateStats(n)
	if err == io.EOF || time.Since(r.lastProgress).Nanoseconds() > debounceInterval.Nanoseconds() {
		r.lastProgress = time.Now()
		progress := NewProgress(bytesRead, bytesPerSecond, err == io.EOF)
		r.progressFunc(*progress)
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
