package compilerspec_test

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

	options "github.com/bhojpur/iso/pkg/manager/compiler/types/options"
	compilerspec "github.com/bhojpur/iso/pkg/manager/compiler/types/spec"
	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"

	. "github.com/bhojpur/iso/pkg/manager/compiler"
	pkg "github.com/bhojpur/iso/pkg/manager/database"
	"github.com/bhojpur/iso/pkg/manager/tree"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Spec", func() {
	Context("Bhojpur specs", func() {
		It("Allows normal operations", func() {
			testSpec := &compilerspec.BhojpurCompilationSpec{Package: &types.Package{Name: "foo", Category: "a", Version: "0"}}
			testSpec2 := &compilerspec.BhojpurCompilationSpec{Package: &types.Package{Name: "bar", Category: "a", Version: "0"}}
			testSpec3 := &compilerspec.BhojpurCompilationSpec{Package: &types.Package{Name: "baz", Category: "a", Version: "0"}}
			testSpec4 := &compilerspec.BhojpurCompilationSpec{Package: &types.Package{Name: "foo", Category: "a", Version: "0"}}

			specs := compilerspec.NewBhojpurCompilationspecs(testSpec, testSpec2)
			Expect(specs.Len()).To(Equal(2))
			Expect(specs.All()).To(Equal([]*compilerspec.BhojpurCompilationSpec{testSpec, testSpec2}))
			specs.Add(testSpec3)
			Expect(specs.All()).To(Equal([]*compilerspec.BhojpurCompilationSpec{testSpec, testSpec2, testSpec3}))
			specs.Add(testSpec4)
			Expect(specs.All()).To(Equal([]*compilerspec.BhojpurCompilationSpec{testSpec, testSpec2, testSpec3, testSpec4}))
			newSpec := specs.Unique()
			Expect(newSpec.All()).To(Equal([]*compilerspec.BhojpurCompilationSpec{testSpec, testSpec2, testSpec3}))

			newSpec2 := specs.Remove(compilerspec.NewBhojpurCompilationspecs(testSpec, testSpec2))
			Expect(newSpec2.All()).To(Equal([]*compilerspec.BhojpurCompilationSpec{testSpec3}))

		})
		Context("virtuals", func() {
			When("is empty", func() {
				It("is virtual", func() {
					spec := &compilerspec.BhojpurCompilationSpec{}
					Expect(spec.IsVirtual()).To(BeTrue())
				})
			})
			When("has defined steps", func() {
				It("is not a virtual", func() {
					spec := &compilerspec.BhojpurCompilationSpec{Steps: []string{"foo"}}
					Expect(spec.IsVirtual()).To(BeFalse())
				})
			})
			When("has defined image", func() {
				It("is not a virtual", func() {
					spec := &compilerspec.BhojpurCompilationSpec{Image: "foo"}
					Expect(spec.IsVirtual()).To(BeFalse())
				})
			})
		})
	})

	Context("Image hashing", func() {
		It("is stable", func() {
			spec1 := &compilerspec.BhojpurCompilationSpec{
				Image:        "foo",
				BuildOptions: &options.Compiler{BuildValues: []map[string]interface{}{{"foo": "bar", "baz": true}}},

				Package: &types.Package{
					Name:     "foo",
					Category: "Bar",
					Labels: map[string]string{
						"foo": "bar",
						"baz": "foo",
					},
				},
			}
			spec2 := &compilerspec.BhojpurCompilationSpec{
				Image:        "foo",
				BuildOptions: &options.Compiler{BuildValues: []map[string]interface{}{{"foo": "bar", "baz": true}}},
				Package: &types.Package{
					Name:     "foo",
					Category: "Bar",
					Labels: map[string]string{
						"foo": "bar",
						"baz": "foo",
					},
				},
			}
			spec3 := &compilerspec.BhojpurCompilationSpec{
				Image: "foo",
				Steps: []string{"foo"},
				Package: &types.Package{
					Name:     "foo",
					Category: "Bar",
					Labels: map[string]string{
						"foo": "bar",
						"baz": "foo",
					},
				},
			}
			hash, err := spec1.Hash()
			Expect(err).ToNot(HaveOccurred())

			hash2, err := spec2.Hash()
			Expect(err).ToNot(HaveOccurred())

			hash3, err := spec3.Hash()
			Expect(err).ToNot(HaveOccurred())

			Expect(hash).To(Equal(hash2))
			hashagain, err := spec2.Hash()
			Expect(err).ToNot(HaveOccurred())
			Expect(hash).ToNot(Equal(hash3))
			Expect(hash).To(Equal(hashagain))
		})
	})

	Context("Simple package build definition", func() {
		It("Loads it correctly", func() {
			generalRecipe := tree.NewGeneralRecipe(pkg.NewInMemoryDatabase(false))

			err := generalRecipe.Load("../../../../tests/fixtures/buildtree")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe.GetDatabase().GetPackages())).To(Equal(1))

			compiler := NewBhojpurCompiler(nil, generalRecipe.GetDatabase())
			lspec, err := compiler.FromPackage(&types.Package{Name: "enman", Category: "app-admin", Version: "1.4.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(lspec.Steps).To(Equal([]string{"echo foo > /test", "echo bar > /test2"}))
			Expect(lspec.Image).To(Equal("bhojpur/base"))
			Expect(lspec.Seed).To(Equal("alpine"))
			tmpdir, err := ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			lspec.Env = []string{"test=1"}
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
ENV PACKAGE_CATEGORY=app-admin
ENV test=1`))

			err = lspec.WriteStepImageDefinition(lspec.Image, filepath.Join(tmpdir, "Dockerfile"))
			Expect(err).ToNot(HaveOccurred())
			dockerfile, err = fileHelper.Read(filepath.Join(tmpdir, "Dockerfile"))
			Expect(err).ToNot(HaveOccurred())
			Expect(dockerfile).To(Equal(`
FROM bhojpur/base
COPY . /isobuild
WORKDIR /isobuild
ENV PACKAGE_NAME=enman
ENV PACKAGE_VERSION=1.4.0
ENV PACKAGE_CATEGORY=app-admin
ENV test=1
RUN echo foo > /test
RUN echo bar > /test2`))

		})

	})

	It("Renders retrieve and env fields", func() {
		generalRecipe := tree.NewGeneralRecipe(pkg.NewInMemoryDatabase(false))

		err := generalRecipe.Load("../../../../tests/fixtures/retrieve")
		Expect(err).ToNot(HaveOccurred())

		Expect(len(generalRecipe.GetDatabase().GetPackages())).To(Equal(1))

		compiler := NewBhojpurCompiler(nil, generalRecipe.GetDatabase())
		lspec, err := compiler.FromPackage(&types.Package{Name: "a", Category: "test", Version: "1.0"})
		Expect(err).ToNot(HaveOccurred())

		Expect(lspec.Steps).To(Equal([]string{"echo foo > /test", "echo bar > /test2"}))
		Expect(lspec.Image).To(Equal("bhojpur/base"))
		Expect(lspec.Seed).To(Equal("alpine"))
		tmpdir, err := ioutil.TempDir("", "tree")
		Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(tmpdir) // clean up

		err = lspec.WriteBuildImageDefinition(filepath.Join(tmpdir, "Dockerfile"))
		Expect(err).ToNot(HaveOccurred())
		dockerfile, err := fileHelper.Read(filepath.Join(tmpdir, "Dockerfile"))
		Expect(err).ToNot(HaveOccurred())
		Expect(dockerfile).To(Equal(`
FROM alpine
COPY . /isobuild
WORKDIR /isobuild
ENV PACKAGE_NAME=a
ENV PACKAGE_VERSION=1.0
ENV PACKAGE_CATEGORY=test
ADD test /isobuild/
ADD http://www.google.com /isobuild/
ENV test=1`))

		lspec.SetOutputPath("/foo/bar")

		err = lspec.WriteBuildImageDefinition(filepath.Join(tmpdir, "Dockerfile"))
		Expect(err).ToNot(HaveOccurred())
		dockerfile, err = fileHelper.Read(filepath.Join(tmpdir, "Dockerfile"))
		Expect(err).ToNot(HaveOccurred())
		Expect(dockerfile).To(Equal(`
FROM alpine
COPY . /isobuild
WORKDIR /isobuild
ENV PACKAGE_NAME=a
ENV PACKAGE_VERSION=1.0
ENV PACKAGE_CATEGORY=test
ADD test /isobuild/
ADD http://www.google.com /isobuild/
ENV test=1`))

		err = lspec.WriteStepImageDefinition(lspec.Image, filepath.Join(tmpdir, "Dockerfile"))
		Expect(err).ToNot(HaveOccurred())
		dockerfile, err = fileHelper.Read(filepath.Join(tmpdir, "Dockerfile"))
		Expect(err).ToNot(HaveOccurred())

		Expect(dockerfile).To(Equal(`
FROM bhojpur/base
COPY . /isobuild
WORKDIR /isobuild
ENV PACKAGE_NAME=a
ENV PACKAGE_VERSION=1.0
ENV PACKAGE_CATEGORY=test
ADD test /isobuild/
ADD http://www.google.com /isobuild/
ENV test=1
RUN echo foo > /test
RUN echo bar > /test2`))

	})
})
