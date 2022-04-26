package artifact_test

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

	"github.com/bhojpur/iso/pkg/manager/api/core/types"

	"github.com/bhojpur/iso/pkg/manager/api/core/context"
	"github.com/bhojpur/iso/pkg/manager/api/core/image"
	. "github.com/bhojpur/iso/pkg/manager/api/core/types/artifact"
	backend "github.com/bhojpur/iso/pkg/manager/compiler/backend"
	compression "github.com/bhojpur/iso/pkg/manager/compiler/types/compression"
	"github.com/bhojpur/iso/pkg/manager/compiler/types/options"
	compilerspec "github.com/bhojpur/iso/pkg/manager/compiler/types/spec"

	. "github.com/bhojpur/iso/pkg/manager/compiler"
	pkg "github.com/bhojpur/iso/pkg/manager/database"
	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"
	"github.com/bhojpur/iso/pkg/manager/tree"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Artifact", func() {
	Context("Simple package build definition", func() {
		ctx := context.NewContext()
		It("Generates a verified delta", func() {

			generalRecipe := tree.NewGeneralRecipe(pkg.NewInMemoryDatabase(false))

			err := generalRecipe.Load("../../../../../tests/fixtures/buildtree")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe.GetDatabase().GetPackages())).To(Equal(1))

			cc := NewBhojpurCompiler(nil, generalRecipe.GetDatabase(), options.WithContext(context.NewContext()))
			lspec, err := cc.FromPackage(&types.Package{Name: "enman", Category: "app-admin", Version: "1.4.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(lspec.Steps).To(Equal([]string{"echo foo > /test", "echo bar > /test2"}))
			Expect(lspec.Image).To(Equal("bhojpur/base"))
			Expect(lspec.Seed).To(Equal("alpine"))
			tmpdir, err := ioutil.TempDir(os.TempDir(), "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			tmpdir2, err := ioutil.TempDir(os.TempDir(), "tree2")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir2) // clean up

			unpacked, err := ioutil.TempDir(os.TempDir(), "unpacked")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(unpacked) // clean up

			rootfs, err := ioutil.TempDir(os.TempDir(), "rootfs")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(rootfs) // clean up

			err = lspec.WriteBuildImageDefinition(filepath.Join(tmpdir, "Dockerfile"))
			Expect(err).ToNot(HaveOccurred())
			dockerfile, err := fileHelper.Read(filepath.Join(tmpdir, "Dockerfile"))
			Expect(err).ToNot(HaveOccurred())
			Expect(dockerfile).To(Equal(`
FROM alpine
COPY . /isobuild
WORKDIR /isobuild
ENV PACKAGE_NAME=enman
ENV PACKAGE_VERSION=1.4.0
ENV PACKAGE_CATEGORY=app-admin`))
			b := backend.NewSimpleDockerBackend(ctx)
			opts := backend.Options{
				ImageName:      "bhojpur/base",
				SourcePath:     tmpdir,
				DockerFileName: "Dockerfile",
				Destination:    filepath.Join(tmpdir2, "output1.tar"),
			}
			Expect(b.BuildImage(opts)).ToNot(HaveOccurred())
			Expect(b.ExportImage(opts)).ToNot(HaveOccurred())
			Expect(fileHelper.Exists(filepath.Join(tmpdir2, "output1.tar"))).To(BeTrue())
			Expect(b.BuildImage(opts)).ToNot(HaveOccurred())

			err = lspec.WriteStepImageDefinition(lspec.Image, filepath.Join(tmpdir, "BhojpurDockerfile"))
			Expect(err).ToNot(HaveOccurred())
			dockerfile, err = fileHelper.Read(filepath.Join(tmpdir, "BhojpurDockerfile"))
			Expect(err).ToNot(HaveOccurred())
			Expect(dockerfile).To(Equal(`
FROM bhojpur/base
COPY . /isobuild
WORKDIR /isobuild
ENV PACKAGE_NAME=enman
ENV PACKAGE_VERSION=1.4.0
ENV PACKAGE_CATEGORY=app-admin
RUN echo foo > /test
RUN echo bar > /test2`))
			opts2 := backend.Options{
				ImageName:      "test",
				SourcePath:     tmpdir,
				DockerFileName: "BhojpurDockerfile",
				Destination:    filepath.Join(tmpdir, "output2.tar"),
			}
			Expect(b.BuildImage(opts2)).ToNot(HaveOccurred())
			Expect(b.ExportImage(opts2)).ToNot(HaveOccurred())

		})

		It("Generates packages images", func() {
			b := backend.NewSimpleDockerBackend(ctx)
			imageprefix := "foo/"
			testString := []byte(`funky test data`)

			tmpdir, err := ioutil.TempDir(os.TempDir(), "artifact")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			tmpWork, err := ioutil.TempDir(os.TempDir(), "artifact2")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpWork) // clean up

			Expect(os.MkdirAll(filepath.Join(tmpdir, "foo", "bar"), os.ModePerm)).ToNot(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(tmpdir, "test"), testString, 0644)
			Expect(err).ToNot(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(tmpdir, "foo", "bar", "test"), testString, 0644)
			Expect(err).ToNot(HaveOccurred())

			a := NewPackageArtifact(filepath.Join(tmpWork, "fake.tar"))
			a.CompileSpec = &compilerspec.BhojpurCompilationSpec{Package: &types.Package{Name: "foo", Version: "1.0"}}

			err = a.Compress(tmpdir, 1)
			Expect(err).ToNot(HaveOccurred())
			resultingImage := imageprefix + "foo--1.0"
			err = a.GenerateFinalImage(ctx, resultingImage, b, false)
			Expect(err).ToNot(HaveOccurred())

			Expect(b.ImageExists(resultingImage)).To(BeTrue())

			result, err := ioutil.TempDir(os.TempDir(), "result")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(result) // clean up

			img, err := b.ImageReference(resultingImage, true)
			Expect(err).ToNot(HaveOccurred())
			_, _, err = image.ExtractTo(
				ctx,
				img,
				result,
				nil,
			)
			Expect(err).ToNot(HaveOccurred())

			content, err := ioutil.ReadFile(filepath.Join(result, "test"))
			Expect(err).ToNot(HaveOccurred())

			Expect(content).To(Equal(testString))

			content, err = ioutil.ReadFile(filepath.Join(result, "foo", "bar", "test"))
			Expect(err).ToNot(HaveOccurred())

			Expect(content).To(Equal(testString))
		})

		It("Generates empty packages images", func() {
			b := backend.NewSimpleDockerBackend(ctx)
			imageprefix := "foo/"

			tmpdir, err := ioutil.TempDir(os.TempDir(), "artifact")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			tmpWork, err := ioutil.TempDir(os.TempDir(), "artifact2")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpWork) // clean up

			a := NewPackageArtifact(filepath.Join(tmpWork, "fake.tar"))
			a.CompileSpec = &compilerspec.BhojpurCompilationSpec{Package: &types.Package{Name: "foo", Version: "1.0"}}

			err = a.Compress(tmpdir, 1)
			Expect(err).ToNot(HaveOccurred())
			resultingImage := imageprefix + "foo--1.0"
			err = a.GenerateFinalImage(ctx, resultingImage, b, false)
			Expect(err).ToNot(HaveOccurred())

			Expect(b.ImageExists(resultingImage)).To(BeTrue())

			result, err := ioutil.TempDir(os.TempDir(), "result")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(result) // clean up

			img, err := b.ImageReference(resultingImage, false)
			Expect(err).ToNot(HaveOccurred())
			_, _, err = image.ExtractTo(
				ctx,
				img,
				result,
				nil,
			)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.DirectoryIsEmpty(result)).To(BeTrue())
		})

		It("Retrieves uncompressed name", func() {
			a := NewPackageArtifact("foo.tar.gz")
			a.CompressionType = (compression.GZip)
			Expect(a.GetUncompressedName()).To(Equal("foo.tar"))

			a = NewPackageArtifact("foo.tar.zst")
			a.CompressionType = compression.Zstandard
			Expect(a.GetUncompressedName()).To(Equal("foo.tar"))

			a = NewPackageArtifact("foo.tar")
			a.CompressionType = compression.None
			Expect(a.GetUncompressedName()).To(Equal("foo.tar"))
		})
	})
})
