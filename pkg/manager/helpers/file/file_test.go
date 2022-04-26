package file_test

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
	"io/ioutil"
	"os"
	"path/filepath"

	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helpers", func() {
	Context("Exists", func() {
		It("Detect existing and not-existing files", func() {
			Expect(fileHelper.Exists("../../tests/fixtures/buildtree/app-admin/enman/1.4.0/build.yaml")).To(BeTrue())
			Expect(fileHelper.Exists("../../tests/fixtures/buildtree/app-admin/enman/1.4.0/build.yaml.not.exists")).To(BeFalse())
		})
	})

	Context("DirectoryIsEmpty", func() {
		It("Detects empty directory", func() {
			testDir, err := ioutil.TempDir(os.TempDir(), "test")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(testDir)
			Expect(fileHelper.DirectoryIsEmpty(testDir)).To(BeTrue())
		})
		It("Detects directory with files", func() {
			testDir, err := ioutil.TempDir(os.TempDir(), "test")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(testDir)
			err = fileHelper.Touch(filepath.Join(testDir, "foo"))
			Expect(err).ToNot(HaveOccurred())
			Expect(fileHelper.DirectoryIsEmpty(testDir)).To(BeFalse())
		})
	})

	Context("Orders dir and files correctly", func() {
		It("puts files first and folders at end", func() {
			testDir, err := ioutil.TempDir(os.TempDir(), "test")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(testDir)

			err = ioutil.WriteFile(filepath.Join(testDir, "foo"), []byte("test\n"), 0644)
			Expect(err).ToNot(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(testDir, "baz"), []byte("test\n"), 0644)
			Expect(err).ToNot(HaveOccurred())

			err = os.MkdirAll(filepath.Join(testDir, "bar"), 0755)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(filepath.Join(testDir, "bar", "foo"), []byte("test\n"), 0644)
			Expect(err).ToNot(HaveOccurred())

			err = os.MkdirAll(filepath.Join(testDir, "baz2"), 0755)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(filepath.Join(testDir, "baz2", "foo"), []byte("test\n"), 0644)
			Expect(err).ToNot(HaveOccurred())

			ordered, notExisting := fileHelper.OrderFiles(testDir, []string{"bar", "baz", "bar/foo", "baz2", "foo", "baz2/foo", "notexisting"})

			Expect(ordered).To(Equal([]string{"baz", "bar/foo", "foo", "baz2/foo", "bar", "baz2"}))
			Expect(notExisting).To(Equal([]string{"notexisting"}))
		})

		It("orders correctly when there are folders with folders", func() {
			testDir, err := ioutil.TempDir(os.TempDir(), "test")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(testDir)

			err = os.MkdirAll(filepath.Join(testDir, "bar"), os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			err = os.MkdirAll(filepath.Join(testDir, "foo"), os.ModePerm)
			Expect(err).ToNot(HaveOccurred())

			err = os.MkdirAll(filepath.Join(testDir, "foo", "bar"), os.ModePerm)
			Expect(err).ToNot(HaveOccurred())

			err = os.MkdirAll(filepath.Join(testDir, "foo", "baz"), os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			err = os.MkdirAll(filepath.Join(testDir, "foo", "baz", "fa"), os.ModePerm)
			Expect(err).ToNot(HaveOccurred())

			ordered, _ := fileHelper.OrderFiles(testDir, []string{"foo", "foo/bar", "bar", "foo/baz/fa", "foo/baz"})
			Expect(ordered).To(Equal([]string{"foo/baz/fa", "foo/bar", "foo/baz", "foo", "bar"}))
		})
	})
})
