package cerrors

import (
	"fmt"
	"io"
	"net/http"

	"google.golang.org/grpc/status"
)

// New 使用格式化说明创建一个错误。
func New(format string, args ...any) error {
	return &withCode{
		stack:   callers(),
		message: fmt.Sprintf(format, args...),
	}
}

// WithCode 创建一个拥有错误码和注释的错误。
func WithCode(code int, format string, args ...any) error {
	return &withCode{
		code:    code,
		stack:   callers(),
		message: fmt.Sprintf(format, args...),
	}
}

// Wrap 使用格式化说明对错误进行包装，如果有错误码则沿用。
func Wrap(err error, format string, args ...any) error {
	return WrapC(err, Code(err), format, args...)
}

// Wrap 使用错误码和格式化说明对错误进行包装，如果原错误有错误码则覆盖。
func WrapC(err error, code int, format string, args ...any) error {
	if err == nil {
		return nil
	}

	return &withCode{
		code:    code,
		cause:   err,
		stack:   callers(),
		message: fmt.Sprintf(format, args...),
	}
}

// Code 提取错误中携带的错误码。
func Code(err error) int {
	if err == nil {
		return 0
	}

	if e, ok := err.(interface{ Code() int }); ok {
		return e.Code()
	}

	if e, ok := status.FromError(err); ok {
		return int(e.Code())
	}

	// 如果本身不携带错误码，则返回 500
	return http.StatusInternalServerError
}

// withCode 携带有错误码和注释的错误类型。
type withCode struct {
	*stack
	code    int
	cause   error
	message string
}

func (w *withCode) Error() string {
	if w.cause != nil {
		return w.message + ": " + w.cause.Error()
	}
	return w.message
}

func (w *withCode) Unwrap() error { return w.cause }

func (w *withCode) Cause() error { return w.cause }

func (w *withCode) Code() int { return w.code }

func (w *withCode) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			if w.Cause() != nil {
				fmt.Fprintf(s, "%+v\n", w.Cause())
			}
			io.WriteString(s, w.message)
			w.stack.Format(s, verb)

		default:
			w.Format(s, 's')
		}
	case 's', 'q':
		io.WriteString(s, w.Error())
	}
}
