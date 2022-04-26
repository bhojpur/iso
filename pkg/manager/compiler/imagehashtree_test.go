package compiler_test

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
	"github.com/bhojpur/iso/pkg/manager/api/core/context"
	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	. "github.com/bhojpur/iso/pkg/manager/compiler"
	sd "github.com/bhojpur/iso/pkg/manager/compiler/backend"
	"github.com/bhojpur/iso/pkg/manager/compiler/types/options"
	pkg "github.com/bhojpur/iso/pkg/manager/database"
	"github.com/bhojpur/iso/pkg/manager/tree"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ImageHashTree", func() {
	ctx := context.NewContext()
	generalRecipe := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))
	compiler := NewBhojpurCompiler(sd.NewSimpleDockerBackend(ctx), generalRecipe.GetDatabase(), options.Concurrency(2))
	hashtree := NewHashTree(generalRecipe.GetDatabase())
	Context("Simple package definition", func() {
		BeforeEach(func() {
			generalRecipe = tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))
			err := generalRecipe.Load("../../tests/fixtures/buildable")
			Expect(err).ToNot(HaveOccurred())
			compiler = NewBhojpurCompiler(sd.NewSimpleDockerBackend(ctx), generalRecipe.GetDatabase(), options.Concurrency(2))
			hashtree = NewHashTree(generalRecipe.GetDatabase())

		})

		It("Calculates the hash correctly", func() {

			spec, err := compiler.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			packageHash, err := hashtree.Query(compiler, spec)
			Expect(err).ToNot(HaveOccurred())
			Expect(packageHash.Target.Hash.BuildHash).To(Equal("895697a8bb51b219b78ed081fa1b778801e81505bb03f56acafcf3c476620fc1"))
			Expect(packageHash.Target.Hash.PackageHash).To(Equal("2a6c3dc0dd7af2902fd8823a24402d89b2030cfbea6e63fe81afb34af8b1a005"))
			Expect(packageHash.BuilderImageHash).To(Equal("builder-3a28d240f505d69123735a567beaf80e"))
		})
	})

	expectedPackageHash := "4154ad4e5dfa2aea41292b3c49eeb04ef327456ecb6312f12d7b94d18ac8cb64"

	Context("complex package definition", func() {
		BeforeEach(func() {
			generalRecipe = tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err := generalRecipe.Load("../../tests/fixtures/upgrade_old_repo_revision")
			Expect(err).ToNot(HaveOccurred())
			compiler = NewBhojpurCompiler(sd.NewSimpleDockerBackend(ctx), generalRecipe.GetDatabase(), options.Concurrency(2))
			hashtree = NewHashTree(generalRecipe.GetDatabase())

		})
		It("Calculates the hash correctly", func() {
			spec, err := compiler.FromPackage(&types.Package{Name: "c", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			packageHash, err := hashtree.Query(compiler, spec)
			Expect(err).ToNot(HaveOccurred())

			expectedHash := "b4b61939260263582da1dfa5289182a0a7570ef8658f3b01b1997fe5d8a95e49"

			Expect(packageHash.Dependencies[len(packageHash.Dependencies)-1].Hash.PackageHash).To(Equal(expectedPackageHash))
			Expect(packageHash.SourceHash).To(Equal(expectedPackageHash))
			Expect(packageHash.BuilderImageHash).To(Equal("builder-381bd2ad9abe1ac6c3c26cba8f8cca0b"))

			//Expect(packageHash.Target.Hash.BuildHash).To(Equal("79d7107d13d578b362e6a7bf10ec850efce26316405b8d732ce8f9e004d64281"))
			Expect(packageHash.Target.Hash.PackageHash).To(Equal("3a372fcee17b2c7912eabb04b50f7d5a83e75402da0c96c102f7c2e836ebaa10"))
			a := &types.Package{Name: "a", Category: "test", Version: "1.1"}
			hash, err := packageHash.DependencyBuildImage(a)
			Expect(err).ToNot(HaveOccurred())

			Expect(hash).To(Equal(expectedHash))

			assertionA := packageHash.Dependencies.Search(a.GetFingerPrint())
			Expect(assertionA.Hash.PackageHash).To(Equal(expectedPackageHash))
			b := &types.Package{Name: "b", Category: "test", Version: "1.0"}
			assertionB := packageHash.Dependencies.Search(b.GetFingerPrint())

			Expect(assertionB.Hash.PackageHash).To(Equal(expectedHash))
			hashB, err := packageHash.DependencyBuildImage(b)
			Expect(err).ToNot(HaveOccurred())

			Expect(hashB).To(Equal("fc6fdd4bd62d51fc06c2c22e8bc56543727a2340220972594e28c623ea3a9c6c"))
		})
	})

	Context("complex package definition, with small change in build.yaml", func() {
		BeforeEach(func() {
			generalRecipe = tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			//Definition of A here is slightly changed in the steps build.yaml file (1 character only)
			err := generalRecipe.Load("../../tests/fixtures/upgrade_old_repo_revision_content_changed")
			Expect(err).ToNot(HaveOccurred())
			compiler = NewBhojpurCompiler(sd.NewSimpleDockerBackend(ctx), generalRecipe.GetDatabase(), options.Concurrency(2))
			hashtree = NewHashTree(generalRecipe.GetDatabase())

		})
		It("Calculates the hash correctly", func() {
			spec, err := compiler.FromPackage(&types.Package{Name: "c", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			packageHash, err := hashtree.Query(compiler, spec)
			Expect(err).ToNot(HaveOccurred())
			Expect(packageHash.Dependencies[len(packageHash.Dependencies)-1].Hash.PackageHash).ToNot(Equal(expectedPackageHash))
			sourceHash := "5534399abed19a3c93b0e638811a5ba6d07e68f6782e2b40aaf2b09c408a3154"
			Expect(packageHash.Dependencies[len(packageHash.Dependencies)-1].Hash.PackageHash).To(Equal(sourceHash))
			Expect(packageHash.SourceHash).To(Equal(sourceHash))

			Expect(packageHash.SourceHash).ToNot(Equal(expectedPackageHash))

			Expect(packageHash.BuilderImageHash).To(Equal("builder-2a3905cf55bdcd1e4cea6b128cbf5b3a"))

			//Expect(packageHash.Target.Hash.BuildHash).To(Equal("79d7107d13d578b362e6a7bf10ec850efce26316405b8d732ce8f9e004d64281"))
			Expect(packageHash.Target.Hash.PackageHash).To(Equal("4a13154de2e802fbd250236294562fad8c9f2c51ab8a3fc359323dd1ed064907"))
			a := &types.Package{Name: "a", Category: "test", Version: "1.1"}
			hash, err := packageHash.DependencyBuildImage(a)
			Expect(err).ToNot(HaveOccurred())

			Expect(hash).To(Equal("b4b61939260263582da1dfa5289182a0a7570ef8658f3b01b1997fe5d8a95e49"))

			assertionA := packageHash.Dependencies.Search(a.GetFingerPrint())

			Expect(assertionA.Hash.PackageHash).To(Equal("5534399abed19a3c93b0e638811a5ba6d07e68f6782e2b40aaf2b09c408a3154"))
			Expect(assertionA.Hash.PackageHash).ToNot(Equal(expectedPackageHash))

			b := &types.Package{Name: "b", Category: "test", Version: "1.0"}
			assertionB := packageHash.Dependencies.Search(b.GetFingerPrint())

			Expect(assertionB.Hash.PackageHash).To(Equal("b4b61939260263582da1dfa5289182a0a7570ef8658f3b01b1997fe5d8a95e49"))
			hashB, err := packageHash.DependencyBuildImage(b)
			Expect(err).ToNot(HaveOccurred())

			Expect(hashB).To(Equal("fc6fdd4bd62d51fc06c2c22e8bc56543727a2340220972594e28c623ea3a9c6c"))
		})
	})

})
