package cerrors

import "github.com/pkg/errors"

func Cause(err error) error {
	return errors.Cause(err)
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target any) bool {
	return errors.As(err, target)
}
