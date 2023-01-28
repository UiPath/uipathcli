package output

type OutputWriter interface {
	WriteResponse(response ResponseInfo) error
}
