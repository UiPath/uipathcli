// Package resiliency contains functions which can be used to retry operations
// which can sporadically fail to add reresiliency to the project
package resiliency

import (
	"errors"
	"time"
)

const MaxAttempts = 3

// Retry the given function up to 3 times when it returns an RetryableError.
func Retry(f func(attempt int) error) error {
	return RetryN(MaxAttempts, f)
}

// RetryN retries the given function up to n times when it returns an RetryableError.
func RetryN(maxAttempts int, f func(attempt int) error) error {
	var err error
	for i := 1; ; i++ {
		err = f(i)
		var retryableError *RetryableError
		retryable := errors.As(err, &retryableError)
		if !retryable || i == maxAttempts {
			return err
		}
		time.Sleep(1 * time.Second)
	}
}
