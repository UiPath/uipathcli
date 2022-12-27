package plugin

type FileParameter struct {
	Filename string
	Data     []byte
}

func NewFileParameter(filename string, data []byte) *FileParameter {
	return &FileParameter{filename, data}
}
