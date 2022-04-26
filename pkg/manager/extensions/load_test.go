package extensions_test

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
	"os"
	"path/filepath"

	. "github.com/bhojpur/iso/pkg/manager/extensions"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Extension loader", func() {
	Context("a single extensions", func() {

		var extensions []ExtensionInterface
		BeforeEach(func() {
			extensions = Discover("project", "tests/ext-path")
		})

		It("Discovers it", func() {
			Expect(extensions).To(HaveLen(1))
		})

		It("Detect ShortName", func() {
			ext := extensions[0]
			Expect(ext.Short()).To(Equal("hello"), ext.String())
		})

		It("Detect correct Path", func() {
			ext := extensions[0]

			dir, err := os.Getwd()
			Expect(err).ToNot(HaveOccurred())

			Expect(ext.Path()).To(Equal(filepath.Join(dir, "tests", "ext-path", "project-hello")))
		})
	})
})
