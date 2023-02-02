package executor

type Progress struct {
	BytesRead      int64
	BytesPerSecond int64
	Completed      bool
}

func NewProgress(bytesRead int64, bytesPerSecond int64, completed bool) *Progress {
	return &Progress{bytesRead, bytesPerSecond, completed}
}
