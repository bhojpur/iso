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

var _ = Describe("Delta", func() {
	Context("Generates deltas of images", func() {
		It("computes delta", func() {
			ref, err := name.ParseReference("alpine")
			Expect(err).ToNot(HaveOccurred())

			img, err := daemon.Image(ref)
			Expect(err).ToNot(HaveOccurred())

			layers, err := Delta(img, img)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(layers.Changes)).To(Equal(0))
			Expect(len(layers.Additions)).To(Equal(0))
			Expect(len(layers.Deletions)).To(Equal(0))
		})

		Context("ExtractDeltaFiles", func() {
			ctx := context.NewContext()
			var tmpfile *os.File
			var ref, ref2 name.Reference
			var img, img2 v1.Image
			var err error

			ref, _ = name.ParseReference("alpine")
			ref2, _ = name.ParseReference("golang:alpine")
			img, _ = daemon.Image(ref)
			img2, _ = daemon.Image(ref2)

			BeforeEach(func() {
				ctx = context.NewContext()

				tmpfile, err = ioutil.TempFile("", "delta")
				Expect(err).ToNot(HaveOccurred())
				defer os.RemoveAll(tmpfile.Name()) // clean up
			})

			It("Extract all deltas", func() {

				f, err := ExtractDeltaAdditionsFiles(ctx, img, []string{}, []string{})
				Expect(err).ToNot(HaveOccurred())

				_, tmpdir, err := Extract(
					ctx,
					img2,
					f,
				)
				Expect(err).ToNot(HaveOccurred())
				defer os.RemoveAll(tmpdir) // clean up

				// No extra dirs are present
				Expect(file.Exists(filepath.Join(tmpdir, "home"))).To(BeFalse())
				// Cache from go
				Expect(file.Exists(filepath.Join(tmpdir, "root", ".cache"))).To(BeTrue())
				// sh is present from alpine, hence not in the result
				Expect(file.Exists(filepath.Join(tmpdir, "bin", "sh"))).To(BeFalse())
				// /usr/local/go is part of golang:alpine
				Expect(file.Exists(filepath.Join(tmpdir, "usr", "local", "go"))).To(BeTrue())
				Expect(file.Exists(filepath.Join(tmpdir, "usr", "local", "go", "bin"))).To(BeTrue())
			})

			It("Extract deltas and excludes /usr/local/go", func() {
				f, err := ExtractDeltaAdditionsFiles(ctx, img, []string{}, []string{"usr/local/go"})
				Expect(err).ToNot(HaveOccurred())

				Expect(err).ToNot(HaveOccurred())
				_, tmpdir, err := Extract(
					ctx,
					img2,
					f,
				)
				Expect(err).ToNot(HaveOccurred())
				defer os.RemoveAll(tmpdir) // clean up
				Expect(file.Exists(filepath.Join(tmpdir, "usr", "local", "go"))).To(BeFalse())
			})

			It("Extract deltas and excludes /usr/local/go/bin, but includes /usr/local/go", func() {
				f, err := ExtractDeltaAdditionsFiles(ctx, img, []string{"usr/local/go"}, []string{"usr/local/go/bin"})
				Expect(err).ToNot(HaveOccurred())

				_, tmpdir, err := Extract(
					ctx,
					img2,
					f,
				)
				Expect(err).ToNot(HaveOccurred())
				defer os.RemoveAll(tmpdir) // clean up
				Expect(file.Exists(filepath.Join(tmpdir, "usr", "local", "go"))).To(BeTrue())
				Expect(file.Exists(filepath.Join(tmpdir, "usr", "local", "go", "bin"))).To(BeFalse())
			})

			It("Extract deltas and includes /usr/local/go", func() {
				f, err := ExtractDeltaAdditionsFiles(ctx, img, []string{"usr/local/go"}, []string{})
				Expect(err).ToNot(HaveOccurred())
				_, tmpdir, err := Extract(
					ctx,
					img2,
					f,
				)
				Expect(err).ToNot(HaveOccurred())
				defer os.RemoveAll(tmpdir) // clean up

				Expect(file.Exists(filepath.Join(tmpdir, "usr", "local", "go"))).To(BeTrue())
				Expect(file.Exists(filepath.Join(tmpdir, "root", ".cache"))).To(BeFalse())
			})
		})
	})
})
