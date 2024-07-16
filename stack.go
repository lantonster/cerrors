package cerrors

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"
)

const unknown = "unknown"

type frame uintptr

// pc program counter 返回该 frame 的程序计数器
func (f frame) pc() uintptr { return uintptr(f) - 1 }

// file 返回包含此 frame 的函数的文件的完整路径。
func (f frame) file() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return unknown
	}
	file, _ := fn.FileLine(f.pc())
	return file
}

// line 返回包含此 framge 的函数源码在文件中的行数。
func (f frame) line() int {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.pc())
	return line
}

// name 返回包含该 frame 的函数的名称。
func (f frame) name() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return unknown
	}
	return fn.Name()
}

// Format of frame formats the frame according to the fmt.Formatter interface.
//
//	%s    文件名
//	%d    行号
//	%n    函数名
//	%v    等价于 %s:%d
//
// Format accepts flags that alter the printing of some verbs, as follows:
//
//	%+s   源文件的函数名和路径相对于编译时 GOPATH 以 \n\t 分隔(<funcname>\n\t<path>)
//	%+v   等价于 to %+s:%d
func (f frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		switch {
		case s.Flag('+'):
			io.WriteString(s, f.name())
			io.WriteString(s, "\n\t")
			io.WriteString(s, f.file())
		default:
			io.WriteString(s, path.Base(f.file()))
		}

	case 'd':
		io.WriteString(s, strconv.Itoa(f.line()))
	case 'n':
		io.WriteString(s, funcname(f.name()))
	case 'v':
		f.Format(s, 's')
		io.WriteString(s, ":")
		f.Format(s, 'd')
	}
}

// funcname 删除 func.Name() 返回的函数名称的路径前缀组件。
func funcname(name string) string {
	i := strings.LastIndex(name, "/")
	name = name[i+1:]

	i = strings.Index(name, ".")
	return name[i+1:]
}

// MarshalText 将堆栈跟踪帧格式化为文本字符串。
// 输出与 fmt.Sprintf("%+v",f) 相同，但没有换行符或制表符。
func (f frame) MarshalText() ([]byte, error) {
	name := f.name()
	if name == unknown {
		return []byte(name), nil
	}
	return []byte(fmt.Sprintf("%s %s:%d", name, f.file(), f.line())), nil
}

type stack []uintptr

func (s *stack) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case st.Flag('+'):
			for _, pc := range *s {
				f := frame(pc)
				fmt.Fprintf(st, "\n%+v", f)
			}
		}
	}
}

func (s *stack) StackTrace() stackTrace {
	frames := make([]frame, len(*s))
	for i := 0; i < len(frames); i++ {
		frames[i] = frame((*s)[i])
	}
	return frames
}

func callers() *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(4, pcs[:])
	var st stack = pcs[0:n]
	return &st
}

type stackTrace []frame

// Format formats the stack of Frames according to the fmt.Formatter interface.
//
//	%s	列出堆栈中每个 frame 的源文件
//	%v	列出堆栈中每个 frame 的源文件和行号
//
// Format accepts flags that alter the printing of some verbs, as follows:
//
//	%+v   打印堆栈中每个 frame 的文件名、函数和行号。
func (st stackTrace) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			for _, f := range st {
				io.WriteString(s, "\n")
				f.Format(s, verb)
			}

		case s.Flag('#'):
			fmt.Fprintf(s, "%#v", []frame(st))
		default:
			st.formatSlice(s, verb)
		}

	case 's':
		st.formatSlice(s, verb)
	}
}

// formatSlice 将此 stackTrace 格式化为给定缓冲区中的帧切片，
// 仅在使用 "%s" 或 "%v" 调用时有效。
func (st stackTrace) formatSlice(s fmt.State, verb rune) {
	_, _ = io.WriteString(s, "[")
	for i, f := range st {
		if i > 0 {
			io.WriteString(s, " ")
		}
		f.Format(s, verb)
	}
	io.WriteString(s, "]")
}
