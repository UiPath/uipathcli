package utils

// The Progress structure contains statistics about how many bytes
// have been read from the underlying reader.
type Progress struct {
	BytesRead      int64
	BytesPerSecond int64
	Completed      bool
}

func NewProgress(bytesRead int64, bytesPerSecond int64, completed bool) *Progress {
	return &Progress{bytesRead, bytesPerSecond, completed}
}
