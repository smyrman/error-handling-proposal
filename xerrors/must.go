package xerrors

type mustError struct {
	error
}

func (err mustError) Unwrap() error {
	return err.error
}

// Catch recovers from panics raised by Must or Must2 error handler returns
// only. Other panics are passed through. The error from Must or Must2 is
// passed through all handlers, if any. If the error is not set to nil by
// any of the handlers, then target will be set with the final error value.
// If target is nill, and the final error is not nil, Catch will panic instead.
//
// Likely not exposed in the final implementation. The final implementation may
// or may not use panics for it's control flow.
func Catch(target *error, handlers ...func(error) error) {
	r := recover()
	switch rt := r.(type) {
	case nil:
	case mustError:
		nextErr := rt.error
		for _, h := range handlers {
			nextErr = h(nextErr)
			if nextErr == nil {
				return
			}
		}
		if target == nil {
			panic(nextErr)
		}
		*target = nextErr
	default:
		panic(r)
	}
}

// Must implements '?' for wrapping functions with one return parameter when
// combined with a deferred Catch. Handlers are called in order given the input
// from the previous handler. If a handler returns nil, then that value is
// returned immediately. If the final handler returns an error, we raise a panic
// that is recovered by Catch. If there are no handlers, then Must will panic
// with the original error if it is not nil.
//
// Likely not exposed in the final implementation. The final implementation may
// or may not use panics for it's control flow.
func Must(err error) func(handlers ...func(error) error) {
	if err == nil {
		return func(_ ...func(error) error) {}
	}
	return func(handlers ...func(error) error) {
		for _, h := range handlers {
			err = h(err)
			if err == nil {
				break
			}
		}
		if err != nil {
			panic(mustError{error: err})
		}
	}
}

// Must2 implements '?' semantics for wrapping functions with two return
// parameter when combined with Catch. Handlers hare called in order given the
// input from the previous handler. If a handler returns nil, then that value is
// returned immediately. If the final handler returns an error, we raise a panic
// that is recovered by Catch. If there are no handlers, then Must2 will panic
// with the original error if it is not nil.
//
// Likely not exposed in the final implementation. The final implementation may
// or may not use panics for it's control flow.
func Must2[T any](v T, err error) func(handlers ...func(error) error) T {
	if err == nil {
		return func(_ ...func(error) error) T {
			return v
		}
	}
	return func(handlers ...func(error) error) T {
		for _, h := range handlers {
			err = h(err)
			if err == nil {
				break
			}
		}
		if err != nil {
			panic(mustError{error: err})
		}
		return v
	}
}
