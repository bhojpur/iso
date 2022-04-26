package logger_test

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
	"io"
	"io/ioutil"
	"os"

	"github.com/bhojpur/iso/pkg/manager/api/core/logger"
	. "github.com/bhojpur/iso/pkg/manager/api/core/logger"
	"github.com/gookit/color"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func captureStdout(f func(w io.Writer)) string {
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	color.SetOutput(w)
	f(w)

	_ = w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = originalStdout
	color.SetOutput(os.Stdout)

	_ = r.Close()

	return string(out)
}

var _ = Describe("Context and logging", func() {

	Context("Context", func() {
		It("detect if is a terminal", func() {
			Expect(captureStdout(func(w io.Writer) {
				_, _, err := GetTerminalSize()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("size not detectable"))
				os.Stdout.Write([]byte(err.Error()))
			})).To(ContainSubstring("size not detectable"))
		})

		It("respects loglevel", func() {

			l, err := New(WithLevel("info"))
			Expect(err).ToNot(HaveOccurred())

			Expect(captureStdout(func(w io.Writer) {
				l.Debug("")
			})).To(Equal(""))

			l, err = New(WithLevel("debug"))
			Expect(err).ToNot(HaveOccurred())

			Expect(captureStdout(func(w io.Writer) {
				l.Debug("foo")
			})).To(ContainSubstring("foo"))
		})

		It("logs with context", func() {
			l, err := New(WithLevel("debug"), WithContext("foo"))
			Expect(err).ToNot(HaveOccurred())

			Expect(captureStdout(func(w io.Writer) {
				l.Debug("bar")
			})).To(ContainSubstring("(foo)  bar"))
		})

		It("returns copies with logged context", func() {
			l, err := New(WithLevel("debug"))
			l, _ = l.Copy(logger.WithContext("bazzz"))
			Expect(err).ToNot(HaveOccurred())

			Expect(captureStdout(func(w io.Writer) {
				l.Debug("bar")
			})).To(ContainSubstring("(bazzz)  bar"))
		})

		It("logs to file", func() {

			t, err := ioutil.TempFile("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(t.Name()) // clean up

			l, err := New(WithLevel("debug"), WithFileLogging(t.Name(), ""))
			Expect(err).ToNot(HaveOccurred())

			//	ctx.Init()

			Expect(captureStdout(func(w io.Writer) {
				l.Info("foot")
			})).To(And(ContainSubstring("INFO"), ContainSubstring("foot")))

			Expect(captureStdout(func(w io.Writer) {
				l.Success("test")
			})).To(And(ContainSubstring("SUCCESS"), ContainSubstring("test")))

			Expect(captureStdout(func(w io.Writer) {
				l.Error("foobar")
			})).To(And(ContainSubstring("ERROR"), ContainSubstring("foobar")))

			Expect(captureStdout(func(w io.Writer) {
				l.Warning("foowarn")
			})).To(And(ContainSubstring("WARNING"), ContainSubstring("foowarn")))

			ll, err := ioutil.ReadFile(t.Name())
			Expect(err).ToNot(HaveOccurred())
			logs := string(ll)
			Expect(logs).To(ContainSubstring("foot"))
			Expect(logs).To(ContainSubstring("test"))
			Expect(logs).To(ContainSubstring("foowarn"))
			Expect(logs).To(ContainSubstring("foobar"))
		})
	})
})
