# Error handling proposal

**This should not be considered a formal Go proposal yet, but the pre-work needed in order to create one. When all requirements are in place, this may transition into a proposal.**

Contribution guide:

- Discuss changes first; either, either on the [Go issue][issue], or by raising an issue at the [GitHub repo][repo].
- Commits should be atomic; rebase your commits to match this.
- Examples with third-party dependencies should get it's own go.mod file.
- Include both working Go 1.24 code (with .go suffix), and a variant using the proposed ? syntax (with .go2 suffix). -
- Note that only files that are affected by the proposal syntax, needs a .go2 file.

Contributions that may be of particular interest for now:

- Contributions demonstrating how this change would help improve application code.
- Pointing out potential issues.

Links:

- Pre-work repository: [smyrman/error-handling-proposal][repo]
- Go issue: [#72149][issue]

[issue]: https://github.com/golang/go/issues/72149
[repo]: https://github.com/smyrman/error-handling-proposal

Remaining content is aligned with the issue text.

### Go Programming Experience

(Contributors may add their own replies)

@smyrman: Experienced

### Other Languages Experience

(Contributors may add their own replies)

@smyrman: C++, C, Python

### Related Idea

- [x] Has this idea, or one like it, been proposed before?
- [x] Does this affect error handling?
- [x] Is this about generics?
- [x] Is this change backward compatible? Breaking the Go 1 compatibility guarantee is a large cost and requires a large benefit

### Has this idea, or one like it, been proposed before?

The proposal is inspired by Go discussion [#71460][discussion]. Compared to the discussed proposal, this is similar in syntax, but different in semantics.

Key differences to #71460:

- Instead of the proposed syntax `?{...}`, we use syntax `?(...)`.
- Instead of acting as a control statement (like if), `?` in the proposal acts more like a normal function call.
- This proposal allows usage within a struct and chain statements.
- Instead of allowing N return arguments, this proposal allows a maximum of two return arguments.
- The proposal is paired with a standard library addition to make the language change useful.

Key similarities to #71460:

- Both proposals use the `?` character.
- Both proposals only aim at handling error types (not bool or other return types).

Semantically, this proposal is somewhat similar to [try-catch][try-catch] proposal, but simpler. The syntax and ergonomics are different.

[discussion]: https://github.com/golang/go/discussions/71460#discussioncomment-12365387
[try-catch]: https://github.com/golang/go/issues/32437

### Does this affect error handling?

Yes

This proposal includes:
1. An addition to the `errors` standard library package.
2. A new syntax for handling errors.

### Is this about generics?

It's not about generics, but the proto-type is using generics for it's implementation.

### Cases

Before discussing the proposal, we will demonstrate a few use-cases that could benefit from it. The cases will be relatively simple. Real use-cases may be more complex, and could therefore expect to result in saving more lines.

#### Return directly

The direct return of an error is a commonly used case for error handling when adding additional context is not necessary.

Old syntax:
```go
pipeline, err := A()
if err != nil {
	return err
}
pipeline, err = pipeline.B()
if err != nil {
	return err
}
```
New syntax:
```go
pipeline := A()?.B()?
```

#### Return wrapped error

To wrap an error before return is a commonly used case for error handling when adding additional context is useful.

Old syntax:
```go
pipeline, err := A()
if err != nil {
	return fmt.Errorf("a: %w", err)
}
pipeline = pipeline.B()
if err != nil {
	return fmt.Errorf("a: %w (pipeline ID: %s)", err, id)
}
```

New Syntax:
```go
pipeline :=
	A() ?(errors.Wrap("a: %w")).
	B() ?(errors.Wrap("b: %[2]w (pipeline ID: %[1]s)", id))
```

#### Collect errors

The case for collecting errors is likely not common in library code. However, it is likely useful for application code. Possible use-cases include form validation or JSON APIs.

Old syntax:
```go
func ParseMyStruct(in transportModel) (BusinessModel, error) {
	var errs []error
	a, err := ParseA(in.A)
	if err != nil {
		errs = append(fmt.Errorf("a: %w", err))
	}
	b, err := ParseB(in.A)
	if err != nil {
		errs = append(fmt.Errorf("b: %w", err))
	}
	if err := errors.Join(errs...); err != nil {
		return BusinessModel{}, err
	}

	return BusinessModel{
		A: a,
		B: b,
	}, nil
}
```

New Syntax:
```go
func ParseMyStruct(in transportModel) (BusinessModel, error) {
	var c errors.Collector
	out := BusinessModel{
		A: ParseA(in.A) ?(errors.Wrap("a: %w"), c.Collect),
		B: ParseB(in.B) ?(errors.Wrap("b: %w"), c.Collect),
	}
	if err := c.Err(); err != nil {
		return BusinessModel{}, err
	}
	return out, nil
}
```

#### Custom error wrapping

Custom error type:
```go
type PathError struct{
	Path string
	Err  error
}

func (err PathError) Error() string {
	return fmt.Sprintf("%s: %v",err.Path, err.Err)
}
```

Old syntax:
```go
func ParseMyStruct(in transportModel) (BusinessModel, error) {
	var errs []error
	a, err := ParseA(in.A)
	if err != nil {
		errs = append(PathError{Path:"a", Err: err))
	}
	b, err := ParseB(in.A)
	if err != nil {
		errs =  append(PathError{Path:"b", Err: err))
	}
	if err := errors.Join(errs...); err != nil {
		return BusinessModel{}, err
	}

	return BusinessModel{
		A: a,
		B: b,
	}, nil
}
```

New Syntax (inline handler):
```go
func ParseMyStruct(in transportModel) (BusinessModel, error) {
	var errs []error
	out := BusinessModel{
		A: ParseA(in.A) ?(func(err error) error{
			errs = append(PathError{Path:"a", Err: err))
		}),
		B: ParseB(in.B) ?(func(err error) error{
			errs = append(PathError{Path:"b", Err: err))
		}),
    }
    if err := errors.Join(errs...) {
     		return BusinessModel{}, err
     }
     return out, nil
 }
 ```

### Proposal

The proposal has two parts:
- An addition to the Go syntax, using `?()` /`?` to catch errors.
- Helper functions in the `errors` package.

The proposal follows the principal of the now implemented range-over-func proposal in making sure that the solution can be described as valid Go code using the current language syntax. As of the time of writing, this is the syntax of Go 1.24.

The `?`/`?()` syntax can be used to _move handling_ of errors from the left of the expression to the right. The default handler (no parenthesis), is to return on error. When parenthesis are provided, errors pass though handlers of format `func(error) error`. If any handler return nill, the code continuous along the happy path. If the final handler returns an error, the function with the `?` syntax returns.

It's not yet clear if the `?` syntax should be allowed inside functions that does _not_ return an error. If it's allowed, the suggestion is that the `?` syntax would result in a panic. See options for more details.

The standard library changes involve adding handlers for the most common cases for error handling.

### Standard library changes

The following exposed additions to the standard library `errors` package is suggested:

```go
// Wrap returns an error handler that returns:
//
//	fmt.Errorf(format, slices.Concat(args, []error{err})...)
func Wrap(format string, args ... any) func(error) error {
	return func(error) error {
		nextArgs := make([]any, 0, len(args)+1)
		nextArgs = append(nextArgs, args...)
		nextArgs = append(nextArgs, err)
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
	return Join(c.errs...)
}
```

### Language Spec Changes

_No response_

### Informal Change

The proposal introduce a  new `?` operator, which can be used after calls to functions that has any of the following signatures:

```go
func f1(...) error             // One return parameter, which must be an error
func f2[T any](...) (T, error) // Two return parameters, where the last one is an error
```

The syntax of `?` is similar to that of a function call, except the parenthesis `()` are optional. That is `?` and `?()` are equivalent. The signature of the operator can be described as:

```go
func ?(handlers ...func(error) error)
```

When using the `?` syntax, the last return parameter of the function is passed to the `?` operator to the right, instead of to the left as normal.

```go
func F(...) (..., error) {
	f1()?              // One return parameter without handlers; equivalent to ?()
	f1()?(h1,h2..)     // One return parameter with handlers
	v := f2()?(h1,...) // Two return parameters with
	...
}
```

The processing rules for error handlers is as follows:

If the `?` operator receives a `nil` error value, execution continues along the "happy path."

If the `?` operator receives an error, the error is passed to each handler in order. The output from each handler becomes the input to the next, as long as the output is not `nil`. If any handler return `nil`, the handler chain is aborted, and execution continues along the "happy path."

If after all handlers are called, the final return value is an error, then the flow of the current _statement_ is aborted similar to how a panic works. If `?` is used within a function where the final return statement is an error, then this panic is _recovered_ and the error value is populated with that error value and the function returns at once.


### Is this change backward compatible?

Yes

This work leans on the work done for [%71460][discussion], that highlights that the `?` operator is invalid to use in any existing code. Thus it's expected that no existing code will be able to break due to the introduction of the new syntax.

### Orthogonality: How does this change interact or overlap with existing features?

_No response_

### Would this change make Go easier or harder to learn, and why?

Any addition to the Go syntax, including this one, will make it harder to learn Go. However, people coming from an exception handling paradigm may find the new syntax less intrusive then the explicit return.

### Cost Description

The highest cost of this proposal is likely that there will now be multiple patterns for handling errors. There could be discrepancies and disagreement between different projects about which style to use.

### Changes to Go ToolChain

vet, gopls, gofmt

### Performance Costs

_No response_

### Prototype

The proto-type code is found in the [pre-work repo][repo].

Following the example of range-over-func, the implementation of the `?` semantics is not magic. A tool could be written to generate go code that rewrites the `?` syntax to valid go 1.24 syntax.

With proposed syntax:
```go
func AB() (Pipeline, error) {
	id := "test"
	result :=
		A() ?(errors.Wrap("a: %w")).
		B() ?(errors.Wrap("b: %[2]w (pipeline ID: %[1]s)", id))
	return result, nil
}
````

Can be written using the proto-type library as:

```go
func AB() (_ Pipeline, _err error) {
	defer	errors.Catch(&_err) // Added to the top of all function bodies that contain a `?` operator.

	id := "test"
	result :=
		xerrors.Must2(A())(xerrors.Wrap("a: %w")).                             // function syntax for ?
		xerrors.Must2(B())(xerrors.Wrap("b: %[2]w (pipeline ID: %[1]s)", id))  // function syntax for ?
	return result, nil
}
```

We defined the following functions in the `xerrors` package for our proto-type. This is proto-type code only. The final implementation will likely be handled by the compiler directly:

```go
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
```

### Options

Options could be applied to change the proposal in various ways.

#### 1: Disallow usage within non-error functions

We could choose to disallow the `?` syntax inside functions that doesn't return errors. This included the `main` function.

This would ensure that the `?` syntax _can not_ lead to panics.

#### 2: Allow explicit catch

An option could be to expose the `Catch` function from the proto-type, and allow supplying a set of error handlers that run on all errors.

When an explicit Catch is added, then an implicit Catch is not added.

If the Catch is called with a nil pointer, then any error that isn't fully handled (replaced by `nil`), results in a panic.

#### 3: Chain via ? syntax

Use syntax `? handler1 ? handler2` as shown in [this comment][comment-by-733amir] by @733amir.

[comment-by-733amir]: https://github.com/golang/go/discussions/71460#discussioncomment-12418576

### Why not...

#### Why not allow more than two return values?

```go
a, b, err := A()
if err != nil {
	return err
}
```
```go
a, b := A()?  // Not allowed
```
Most functions that return an error, return either a single parameter, or two parameters. So it wouldn't be many cases where it's useful. It's also assumed that error handling syntax is mostly useful if it allows to continue the flow of our programs. That is, we allow are allowed to chain functions `A()?.B()?`, or assign to struct fields from functions that return errors. Cases with two or more return values typically can not be chained.

Allowing for more return values risks complicating the implementation, and is likely offer little value in return.

#### Why require the final return parameter to be an error?

```go
a := os.Getenv("VARIABLE")? // not allowed
a := os.Getenv("VARIABLE")?(func(bool) error, ...func(error) error) // not allowed
a, bc := strings.Cut("a.b.c", ".")? // not allowed
```

If we allowed for other return values for the naked syntax, it's not clear what the error return value should be.

If we allow for explicit handlers, then we need a conversion from bool to error before we can pass it to the handlers. Thus the argument list changes.
