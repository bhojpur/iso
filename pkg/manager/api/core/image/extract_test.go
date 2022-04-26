package image_test

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

	"github.com/bhojpur/iso/pkg/manager/api/core/context"
	. "github.com/bhojpur/iso/pkg/manager/api/core/image"
	"github.com/bhojpur/iso/pkg/manager/helpers/file"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	daemon "github.com/google/go-containerregistry/pkg/v1/daemon"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Extract", func() {

	Context("extract files from images", func() {
		Context("ExtractFiles", func() {
			ctx := context.NewContext()
			var tmpfile *os.File
			var ref name.Reference
			var img v1.Image
			var err error

			BeforeEach(func() {
				ctx = context.NewContext()

				tmpfile, err = ioutil.TempFile("", "extract")
				Expect(err).ToNot(HaveOccurred())
				defer os.RemoveAll(tmpfile.Name()) // clean up

				ref, err = name.ParseReference("alpine")
				Expect(err).ToNot(HaveOccurred())

				img, err = daemon.Image(ref)
				Expect(err).ToNot(HaveOccurred())
			})

			It("Extract all files", func() {
				_, tmpdir, err := Extract(
					ctx,
					img,
					ExtractFiles(ctx, "", []string{}, []string{}),
				)
				Expect(err).ToNot(HaveOccurred())
				defer os.RemoveAll(tmpdir) // clean up

				Expect(file.Exists(filepath.Join(tmpdir, "usr", "bin"))).To(BeTrue())
				Expect(file.Exists(filepath.Join(tmpdir, "bin", "sh"))).To(BeTrue())
			})

			It("Extract specific dir", func() {
				_, tmpdir, err := Extract(
					ctx,
					img,
					ExtractFiles(ctx, "/usr", []string{}, []string{}),
				)
				Expect(err).ToNot(HaveOccurred())
				defer os.RemoveAll(tmpdir) // clean up
				Expect(file.Exists(filepath.Join(tmpdir, "usr", "sbin"))).To(BeTrue())
				Expect(file.Exists(filepath.Join(tmpdir, "usr", "bin"))).To(BeTrue())
				Expect(file.Exists(filepath.Join(tmpdir, "bin", "sh"))).To(BeFalse())
			})

			It("Extract a dir with includes/excludes", func() {
				_, tmpdir, err := Extract(
					ctx,
					img,
					ExtractFiles(ctx, "/usr", []string{"bin"}, []string{"sbin"}),
				)
				Expect(err).ToNot(HaveOccurred())
				defer os.RemoveAll(tmpdir) // clean up

				Expect(file.Exists(filepath.Join(tmpdir, "usr", "bin"))).To(BeTrue())
				Expect(file.Exists(filepath.Join(tmpdir, "bin", "sh"))).To(BeFalse())
				Expect(file.Exists(filepath.Join(tmpdir, "usr", "sbin"))).To(BeFalse())
			})

			It("Extract with includes/excludes", func() {
				_, tmpdir, err := Extract(
					ctx,
					img,
					ExtractFiles(ctx, "", []string{"/usr|/usr/bin"}, []string{"^/bin"}),
				)
				Expect(err).ToNot(HaveOccurred())
				defer os.RemoveAll(tmpdir) // clean up

				Expect(file.Exists(filepath.Join(tmpdir, "usr", "bin"))).To(BeTrue())
				Expect(file.Exists(filepath.Join(tmpdir, "bin", "sh"))).To(BeFalse())
			})
		})
	})
})
