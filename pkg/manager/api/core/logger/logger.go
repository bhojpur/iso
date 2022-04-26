package logger

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"fmt"
	"path"
	"regexp"
	"runtime"
	"strings"
	"sync"

	log "github.com/ipfs/go-log/v2"
	"github.com/kyokomi/emoji"
	"github.com/pterm/pterm"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the default logger
type Logger struct {
	level       log.LogLevel
	emoji       bool
	logToFile   bool
	noSpinner   bool
	fileLogger  *zap.Logger
	context     string
	spinnerLock sync.Mutex
	s           *pterm.SpinnerPrinter
}

// LogLevel represents a log severity level. Use the package variables as an
// enum.
type LogLevel zapcore.Level

type LoggerOptions func(*Logger) error

var NoSpinner LoggerOptions = func(l *Logger) error {
	l.noSpinner = true
	return nil
}

func WithLevel(level string) LoggerOptions {
	return func(l *Logger) error {
		lvl, _ := log.LevelFromString(level) // Defaults to Info
		l.level = lvl
		if l.level == log.LevelDebug {
			pterm.EnableDebugMessages()
		}
		return nil
	}
}

func WithContext(c string) LoggerOptions {
	return func(l *Logger) error {
		l.context = c
		return nil
	}
}

func WithFileLogging(p, encoding string) LoggerOptions {
	return func(l *Logger) error {
		if encoding == "" {
			encoding = "console"
		}
		l.logToFile = true
		var err error
		cfg := zap.NewProductionConfig()
		cfg.OutputPaths = []string{p}
		cfg.Level = zap.NewAtomicLevelAt(zapcore.Level(l.level))
		cfg.ErrorOutputPaths = []string{}
		cfg.Encoding = encoding
		cfg.DisableCaller = true
		cfg.DisableStacktrace = true
		cfg.EncoderConfig.TimeKey = "time"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		l.fileLogger, err = cfg.Build()
		return err
	}
}

var EnableEmoji = func() LoggerOptions {
	return func(l *Logger) error {
		l.emoji = true
		return nil
	}
}

func New(opts ...LoggerOptions) (*Logger, error) {
	l := &Logger{
		level: log.LevelDebug,
		s:     pterm.DefaultSpinner.WithShowTimer(false).WithRemoveWhenDone(true),
	}
	for _, o := range opts {
		if err := o(l); err != nil {
			return nil, err
		}
	}

	return l, nil
}

func (l *Logger) Copy(opts ...LoggerOptions) (*Logger, error) {
	c := *l
	copy := &c
	for _, o := range opts {
		if err := o(copy); err != nil {
			return nil, err
		}
	}

	return copy, nil
}

func joinMsg(args ...interface{}) (message string) {
	for _, m := range args {
		message += " " + fmt.Sprintf("%v", m)
	}
	return
}

func (l *Logger) enabled(lvl log.LogLevel) bool {
	return lvl >= l.level
}

var emojiStrip = regexp.MustCompile(`[:][\w]+[:]`)

func (l *Logger) transform(args ...interface{}) (sanitized []interface{}) {
	for _, a := range args {
		var aString string

		// Strip emoji if needed
		if l.emoji {
			aString = emoji.Sprint(a)
		} else {
			aString = emojiStrip.ReplaceAllString(joinMsg(a), "")
		}

		sanitized = append(sanitized, aString)
	}

	if l.context != "" {
		sanitized = append([]interface{}{fmt.Sprintf("(%s)", l.context)}, sanitized...)
	}
	return
}

func prefixCodeLine(args ...interface{}) []interface{} {
	pc, file, line, ok := runtime.Caller(3)
	if ok {
		args = append([]interface{}{fmt.Sprintf("(%s:#%d:%v)",
			path.Base(file), line, runtime.FuncForPC(pc).Name())}, args...)
	}
	return args
}

func (l *Logger) send(ll log.LogLevel, f string, args ...interface{}) {
	if !l.enabled(ll) {
		return
	}

	sanitizedArgs := joinMsg(l.transform(args...)...)
	sanitizedF := joinMsg(l.transform(f)...)
	formatDefined := f != ""

	switch {
	case log.LevelDebug == ll && !formatDefined:
		pterm.Debug.Println(prefixCodeLine(sanitizedArgs)...)
		if l.logToFile {
			l.fileLogger.Debug(joinMsg(prefixCodeLine(sanitizedArgs)...))
		}
	case log.LevelDebug == ll && formatDefined:
		pterm.Debug.Printfln(joinMsg(prefixCodeLine(sanitizedF)...), args...)
		if l.logToFile {
			l.fileLogger.Sugar().Debugf(joinMsg(prefixCodeLine(sanitizedF)...), args...)
		}
	case log.LevelError == ll && !formatDefined:
		pterm.Error.Println(pterm.LightRed(sanitizedArgs))
		if l.logToFile {
			l.fileLogger.Error(sanitizedArgs)
		}
	case log.LevelError == ll && formatDefined:
		pterm.Error.Printfln(pterm.LightRed(sanitizedF), args...)
		if l.logToFile {
			l.fileLogger.Sugar().Errorf(sanitizedF, args...)
		}

	case log.LevelFatal == ll && !formatDefined:
		pterm.Error.Println(sanitizedArgs)
		if l.logToFile {
			l.fileLogger.Error(sanitizedArgs)
		}
	case log.LevelFatal == ll && formatDefined:
		pterm.Error.Printfln(sanitizedF, args...)
		if l.logToFile {
			l.fileLogger.Sugar().Errorf(sanitizedF, args...)
		}
		//INFO
	case log.LevelInfo == ll && !formatDefined:
		pterm.Info.Println(sanitizedArgs)
		if l.logToFile {
			l.fileLogger.Info(sanitizedArgs)
		}
	case log.LevelInfo == ll && formatDefined:
		pterm.Info.Printfln(sanitizedF, args...)
		if l.logToFile {
			l.fileLogger.Sugar().Infof(sanitizedF, args...)
		}
		//WARN
	case log.LevelWarn == ll && !formatDefined:
		pterm.Warning.Println(sanitizedArgs)
		if l.logToFile {
			l.fileLogger.Warn(sanitizedArgs)
		}
	case log.LevelWarn == ll && formatDefined:
		pterm.Warning.Printfln(sanitizedF, args...)
		if l.logToFile {
			l.fileLogger.Sugar().Warnf(sanitizedF, args...)
		}
	}
}

func (l *Logger) Debug(args ...interface{}) {
	l.send(log.LevelDebug, "", args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.send(log.LevelError, "", args...)
}

func (l *Logger) Trace(args ...interface{}) {
	l.send(log.LevelDebug, "", args...)
}

func (l *Logger) Tracef(t string, args ...interface{}) {
	l.send(log.LevelDebug, t, args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.send(log.LevelFatal, "", args...)
	panic("fatal error")
}

func (l *Logger) Info(args ...interface{}) {
	l.send(log.LevelInfo, "", args...)
}

func (l *Logger) Success(args ...interface{}) {
	l.Info(append([]interface{}{"SUCCESS"}, args...)...)
}

func (l *Logger) Panic(args ...interface{}) {
	l.Fatal(args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.send(log.LevelWarn, "", args...)
}

func (l *Logger) Warning(args ...interface{}) {
	l.Warn(args...)
}

func (l *Logger) Debugf(f string, args ...interface{}) {
	l.send(log.LevelDebug, f, args...)
}

func (l *Logger) Errorf(f string, args ...interface{}) {
	l.send(log.LevelError, f, args...)
}

func (l *Logger) Fatalf(f string, args ...interface{}) {
	l.send(log.LevelFatal, f, args...)
}

func (l *Logger) Infof(f string, args ...interface{}) {
	l.send(log.LevelInfo, f, args...)
}

func (l *Logger) Panicf(f string, args ...interface{}) {
	l.Fatalf(joinMsg(f), args...)
}

func (l *Logger) Warnf(f string, args ...interface{}) {
	l.send(log.LevelWarn, f, args...)
}

func (l *Logger) Warningf(f string, args ...interface{}) {
	l.Warnf(f, args...)
}

func (l *Logger) Ask() bool {
	var input string

	l.Info("Do you want to continue with this operation? [y/N]: ")
	_, err := fmt.Scanln(&input)
	if err != nil {
		return false
	}
	input = strings.ToLower(input)

	if input == "y" || input == "yes" {
		return true
	}
	return false
}

// Spinner starts the spinner
func (l *Logger) Spinner() {
	if !IsTerminal() || l.noSpinner {
		return
	}

	l.spinnerLock.Lock()
	defer l.spinnerLock.Unlock()

	if l.s != nil && !l.s.IsActive {
		l.s, _ = l.s.Start()
	}
}

func (l *Logger) Screen(text string) {
	l.Infof(":::> %s", text)
}

func (l *Logger) SpinnerText(suffix, prefix string) {
	if !IsTerminal() || l.noSpinner {
		return
	}
	l.spinnerLock.Lock()
	defer l.spinnerLock.Unlock()

	if l.level == log.LevelDebug {
		fmt.Printf("%s %s\n",
			suffix, prefix,
		)
	} else {
		l.s.UpdateText(suffix + prefix)
	}
}

func (l *Logger) SpinnerStop() {
	if !IsTerminal() || l.noSpinner {
		return
	}
	l.spinnerLock.Lock()
	defer l.spinnerLock.Unlock()

	if l.s != nil {
		l.s.Success()
	}
}
