package executor

type FileReference struct {
	Filename string
	Data     []byte
}

func NewFileReference(filename string, data []byte) *FileReference {
	return &FileReference{filename, data}
}
