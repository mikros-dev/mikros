package logger

import (
	"runtime"
	"strconv"
	"strings"
)

type stacktrace struct {
	pcs     []uintptr
	frames  *runtime.Frames
	storage []uintptr
}

func newStacktrace() *stacktrace {
	return &stacktrace{
		pcs: make([]uintptr, 64),
	}
}

func (s *stacktrace) free() {
	s.pcs = nil
	s.frames = nil
	s.storage = nil
}

func (s *stacktrace) next() (runtime.Frame, bool) {
	return s.frames.Next()
}

func captureStacktrace(skip int) *stacktrace {
	s := newStacktrace()
	numFrames := runtime.Callers(skip+2, s.pcs)
	pcs := s.pcs

	for numFrames == len(pcs) {
		pcs = make([]uintptr, len(pcs)*2)
		numFrames = runtime.Callers(skip+2, pcs)
	}

	s.storage = pcs
	s.pcs = pcs[:numFrames]

	s.frames = runtime.CallersFrames(s.pcs)

	return s
}

func takeStacktrace(skip int) string {
	stack := captureStacktrace(skip)
	defer stack.free()

	stackfmt := newStackFormatter()
	stackfmt.FormatStack(stack)

	return stackfmt.s.String()
}

// stackFormatter formats a stack trace into a readable string representation.
type stackFormatter struct {
	s strings.Builder
}

// newStackFormatter builds a new stackFormatter.
func newStackFormatter() stackFormatter {
	return stackFormatter{}
}

// FormatStack formats all remaining frames in the provided stacktrace -- minus
// the final runtime.main/runtime.goexit frame.
func (sf *stackFormatter) FormatStack(stack *stacktrace) {
	// Note: On the last iteration, frames.Next() returns false, with a valid
	// frame, but we ignore this frame. The last frame is a runtime frame that
	// adds noise, since it's only either runtime.main or runtime.goexit.
	for frame, more := stack.next(); more; frame, more = stack.next() {
		sf.FormatFrame(frame)
	}
}

// FormatFrame formats the given frame.
func (sf *stackFormatter) FormatFrame(frame runtime.Frame) {
	sf.s.WriteString(frame.Function)
	sf.s.WriteString("\n")
	sf.s.WriteString("\t")
	sf.s.WriteString(frame.File)
	sf.s.WriteString(":")
	sf.s.WriteString(strconv.Itoa(frame.Line))
	sf.s.WriteString("\n")
}
