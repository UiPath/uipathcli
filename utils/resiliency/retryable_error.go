package resiliency

// RetryableError can be returned inside of the Retry() block to indicate
// that the function failed and should be retried.
type RetryableError struct {
	err error
}

func (e RetryableError) Error() string {
	return e.err.Error()
}

func Retryable(err error) *RetryableError {
	return &RetryableError{err}
}
