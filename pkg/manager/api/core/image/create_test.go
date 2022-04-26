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
	"os"
	"path/filepath"
	"runtime"

	"github.com/bhojpur/iso/pkg/manager/api/core/context"
	. "github.com/bhojpur/iso/pkg/manager/api/core/image"
	"github.com/bhojpur/iso/pkg/manager/api/core/types/artifact"
	"github.com/bhojpur/iso/pkg/manager/compiler/backend"
	"github.com/bhojpur/iso/pkg/manager/helpers/file"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create", func() {
	Context("Creates an OCI image from a standard tar", func() {
		It("creates an image which is loadable", func() {
			ctx := context.NewContext()

			dst, err := ctx.TempFile("dst")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(dst.Name())
			srcTar, err := ctx.TempFile("srcTar")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(srcTar.Name())

			b := backend.NewSimpleDockerBackend(ctx)

			b.DownloadImage(backend.Options{ImageName: "alpine"})
			img, err := b.ImageReference("alpine", false)
			Expect(err).ToNot(HaveOccurred())

			_, dir, err := Extract(ctx, img, nil)
			Expect(err).ToNot(HaveOccurred())

			defer os.RemoveAll(dir)

			Expect(file.Touch(filepath.Join(dir, "test"))).ToNot(HaveOccurred())
			Expect(file.Exists(filepath.Join(dir, "bin"))).To(BeTrue())

			a := artifact.NewPackageArtifact(srcTar.Name())
			a.Compress(dir, 1)

			// Unfortunately there is no other easy way to test this
			err = CreateTar(srcTar.Name(), dst.Name(), "testimage", runtime.GOARCH, runtime.GOOS)
			Expect(err).ToNot(HaveOccurred())

			b.LoadImage(dst.Name())

			Expect(b.ImageExists("testimage")).To(BeTrue())

			img, err = b.ImageReference("testimage", false)
			Expect(err).ToNot(HaveOccurred())

			_, dir, err = Extract(ctx, img, nil)
			Expect(err).ToNot(HaveOccurred())

			defer os.RemoveAll(dir)
			Expect(file.Exists(filepath.Join(dir, "bin"))).To(BeTrue())
			Expect(file.Exists(filepath.Join(dir, "test"))).To(BeTrue())
		})
	})
})
