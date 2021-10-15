package errors

import (
	"fmt"
	"runtime"
)

// getFrame returns the stack frame at index i the moment it was called.
// It automatically discards its own frame.
func getFrame(skipFrames int) runtime.Frame {
	targetFrameIndex := skipFrames + 2
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}

// getStackTrace returns the stacktrace the moment it was called.
// The string format for every entry is the following: <index>: <file> <func>:<line>
func getStackTrace() []string {
	var temp []string
	// Init i to 2 in order to skip Error constructor and this function's calls
	for i := 2; ; i++ {
		b := getFrame(i)
		if b.Function == "unknown" && b.Line == 0 {
			break
		}
		s := fmt.Sprintf("%v: %v %v:%v", i-2, b.File, b.Function, b.Line)
		temp = append(temp, s)
	}
	return temp
}
