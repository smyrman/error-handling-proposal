package xerrors

import (
	"errors"
	"fmt"
)

// Wrap returns an error handler that returns:
//
//	fmt.Errorf(format, err, args...)
func Wrap(format string, args ...any) func(error) error {
	return func(err error) error {
		nextArgs := make([]any, 0, len(args)+1)
		nextArgs = append(nextArgs, err)
		nextArgs = append(nextArgs, args...)
		return fmt.Errorf(format, nextArgs...)
	}
}

// Collector expose an error handler function [Collect] for collecting
// errors into a slice. After the collection is complete, A joined error
// can be retrieved from [Err].
type Collector struct {
	errs []error
}

// Collect is an error handler that appends err to c.
func (c *Collector) Collect(err error) error {
	if err != nil {
		c.errs = append(c.errs, err)
	}
	return nil
}

// Err returns an joined
func (c *Collector) Err() error {
	return errors.Join(c.errs...)
}
