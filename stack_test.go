package cerrors

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

var initpc = caller()

func caller() frame {
	var pcs [3]uintptr
	n := runtime.Callers(2, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	f, _ := frames.Next()
	return frame(f.PC)
}

func Test_frame_Format(t *testing.T) {
	tests := []struct {
		f      frame
		format string
		want   string
	}{
		{0, "%s", "unknown"},
		{0, "%d", "0"},
		{0, "%n", "unknown"},
		{0, "%v", "unknown:0"},
		{0, "%+s", "unknown\n\tunknown"},
		{0, "%+v", "unknown\n\tunknown:0"},
		{initpc, "%s", "stack_test.go"},
		{initpc, "%d", "11"},
		{initpc, "%n", "init"},
		{initpc, "%v", "stack_test.go:11"},
		{initpc, "%+s", "github.com/lllllan02/xerrors.init\n\t/home/lllllan/github/lllllan02/xerrors/stack_test.go"},
		{initpc, "%+v", "github.com/lllllan02/xerrors.init\n\t/home/lllllan/github/lllllan02/xerrors/stack_test.go:11"},
	}

	t.Parallel()
	is := assert.New(t)
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			is.Equal(fmt.Sprintf(tt.format, tt.f), tt.want)
		})
	}
}
