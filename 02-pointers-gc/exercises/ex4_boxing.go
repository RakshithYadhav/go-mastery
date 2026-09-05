package exercises

// Exercise 4 — FIX THE BUG: a logging helper that allocates before it decides
// it has nothing to do.
//
// This is the most common performance bug in production Go, and it is almost
// never noticed, because the code reads perfectly:
//
//	log.Debug("frame received", "machine", machine, "seq", seq)
//
// The service runs at info level. Debug does nothing. And yet this line costs
// three allocations every time it executes, 40,000 times a second, to build
// arguments for a function that immediately discards them.
//
// The cause is in NOTES section 4 and in demo 04's first row: a variadic
// ...any parameter boxes every argument at the CALL SITE, before the function
// is entered. The level check inside Debug happens far too late — the work is
// already done by the time it runs.
//
// YOUR TASKS
//
//  1. Prove it before you fix it. Run the benchmark and read allocs/op:
//
//         go test -run '^$' -bench BenchmarkLogger -benchmem ./02-pointers-gc/exercises
//
//     Confirm that logging at a level that is switched OFF still allocates.
//     Then explain to yourself why moving the level check to the first line of
//     Debug does not help at all.
//
//  2. Fix it so TestLogger_DisabledLevelDoesNotAllocate passes, while
//     TestLogger_StillLogs and TestLogger_Format keep passing.
//
// THE CONSTRAINTS
//
//   - Enabled-level logging may still allocate. It has to: it produces a
//     string. Only the DISABLED path must reach zero.
//   - The output format must not change. TestLogger_Format pins it.
//   - You may add methods. You may not delete Debug, Info, or SetLevel.
//
// THE HINT YOU ARE ALLOWED, because it is a design decision rather than a
// solution: there are two established fixes and they are different trades.
//
//	(a) Give the caller a way to ask first, so the arguments are never built:
//	    `if log.Enabled(LevelDebug) { log.Debug(...) }`. Cheap, ugly at every
//	    call site, and it is what the standard library's log/slog does with
//	    its Enabled method.
//	(b) Replace ...any with typed arguments so nothing needs boxing:
//	    `log.DebugKV("frame received", "machine", machine, "seq", seq)` where
//	    the parameters are typed string and int rather than any.
//
// Pick one, implement it, and write two sentences in MISTAKES.md on why you
// picked it. Both are defensible. The reasoning is the exercise.
//
// Do not change the signatures of SetLevel, Debug, or Info.

import (
	"fmt"
	"strings"
)

// Level is a logging level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
)

// Logger writes levelled log lines into an in-memory buffer.
type Logger struct {
	level Level
	lines []string

	// last holds the key/value pairs from the most recent written line, for a
	// "what were the last fields we saw" debugging endpoint. It is the reason
	// the compiler cannot keep the caller's arguments on the stack: escape
	// analysis works per function, and because SOME path through write stores
	// kv, every caller has to box its arguments on the heap before calling.
	last []any
}

// NewLogger returns a Logger at the given level.
func NewLogger(level Level) *Logger {
	return &Logger{level: level, lines: make([]string, 0, 16)}
}

// LastFields returns the key/value pairs from the most recently written line.
func (l *Logger) LastFields() []any { return l.last }

// SetLevel changes the minimum level that will be written.
func (l *Logger) SetLevel(level Level) { l.level = level }

// Lines returns everything written so far.
func (l *Logger) Lines() []string { return l.lines }

// Debug writes a debug line. BUG: costs three allocations per call even when
// the debug level is switched off.
func (l *Logger) Debug(msg string, kv ...any) {
	l.write(LevelDebug, msg, kv...)
}

// Info writes an info line.
func (l *Logger) Info(msg string, kv ...any) {
	l.write(LevelInfo, msg, kv...)
}

func (l *Logger) write(level Level, msg string, kv ...any) {
	if level < l.level {
		return
	}

	l.last = kv

	var b strings.Builder
	b.WriteString(levelName(level))
	b.WriteByte(' ')
	b.WriteString(msg)
	for i := 0; i+1 < len(kv); i += 2 {
		b.WriteByte(' ')
		b.WriteString(fmt.Sprint(kv[i]))
		b.WriteByte('=')
		b.WriteString(fmt.Sprint(kv[i+1]))
	}
	l.lines = append(l.lines, b.String())
}

func levelName(l Level) string {
	if l == LevelDebug {
		return "DEBUG"
	}
	return "INFO"
}
