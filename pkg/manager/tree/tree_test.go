package tree_test

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

// Recipe is a builder imeplementation.

// It reads a Tree and spit it in human readable form (YAML), called recipe,
// It also loads a tree (recipe) from a YAML (to a db, e.g. BoltDB), allowing to query it
// with the solver, using the package object.

import (
	"io/ioutil"
	"os"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	pkg "github.com/bhojpur/iso/pkg/manager/database"
	"github.com/bhojpur/iso/pkg/manager/solver"
	. "github.com/bhojpur/iso/pkg/manager/tree"
)

var _ = Describe("Tree", func() {

	Context("Simple solving with the fixture tree", func() {
		It("writes and reads back the same tree", func() {
			for index := 0; index < 300; index++ { // Just to make sure we don't have false positives
				db := pkg.NewInMemoryDatabase(false)
				generalRecipe := NewCompilerRecipe(db)
				tmpdir, err := ioutil.TempDir("", "package")
				Expect(err).ToNot(HaveOccurred())
				defer os.RemoveAll(tmpdir) // clean up

				err = generalRecipe.Load("../../tests/fixtures/buildableseed")
				Expect(err).ToNot(HaveOccurred())

				Expect(len(generalRecipe.GetDatabase().World())).To(Equal(4))

				D, err := generalRecipe.GetDatabase().FindPackage(&types.Package{Name: "d", Category: "test", Version: "1.0"})
				Expect(err).ToNot(HaveOccurred())

				Expect(D.GetRequires()[0].GetName()).To(Equal("c"))
				CfromD, err := generalRecipe.GetDatabase().FindPackage(D.GetRequires()[0])
				Expect(err).ToNot(HaveOccurred())

				Expect(len(CfromD.GetRequires()) != 0).To(BeTrue())
				Expect(CfromD.GetRequires()[0].GetName()).To(Equal("b"))

				s := solver.NewSolver(types.SolverOptions{Type: types.SolverSingleCoreSimple}, pkg.NewInMemoryDatabase(false), generalRecipe.GetDatabase(), db)
				pack, err := generalRecipe.GetDatabase().FindPackage(&types.Package{Name: "d", Category: "test", Version: "1.0"})
				Expect(err).ToNot(HaveOccurred())

				solution, err := s.Install([]*types.Package{pack})
				Expect(err).ToNot(HaveOccurred())

				solution, err = solution.Order(generalRecipe.GetDatabase(), pack.GetFingerPrint())
				Expect(err).ToNot(HaveOccurred())

				Expect(solution[0].Package.GetName()).To(Equal("b"))
				Expect(solution[0].Value).To(BeTrue())

				Expect(solution[1].Package.GetName()).To(Equal("c"))
				Expect(solution[1].Value).To(BeTrue())

				Expect(solution[2].Package.GetName()).To(Equal("d"))
				Expect(solution[2].Value).To(BeTrue())
				Expect(len(solution)).To(Equal(3))

				newsolution := solution.Drop(&types.Package{Name: "d", Category: "test", Version: "1.0"})
				Expect(len(newsolution)).To(Equal(2))

				Expect(newsolution[0].Package.GetName()).To(Equal("b"))
				Expect(newsolution[0].Value).To(BeTrue())

				Expect(newsolution[1].Package.GetName()).To(Equal("c"))
				Expect(newsolution[1].Value).To(BeTrue())

			}
		})
	})

	Context("Multiple trees", func() {
		It("Merges", func() {
			for index := 0; index < 300; index++ { // Just to make sure we don't have false positives
				db := pkg.NewInMemoryDatabase(false)
				generalRecipe := NewCompilerRecipe(db)
				tmpdir, err := ioutil.TempDir("", "package")
				Expect(err).ToNot(HaveOccurred())
				defer os.RemoveAll(tmpdir) // clean up

				err = generalRecipe.Load("../../tests/fixtures/buildableseed")
				Expect(err).ToNot(HaveOccurred())

				Expect(len(generalRecipe.GetDatabase().World())).To(Equal(4))

				err = generalRecipe.Load("../../tests/fixtures/layers")
				Expect(err).ToNot(HaveOccurred())

				Expect(len(generalRecipe.GetDatabase().World())).To(Equal(6))

				extra, err := generalRecipe.GetDatabase().FindPackage(&types.Package{Name: "extra", Category: "layer", Version: "1.0"})
				Expect(err).ToNot(HaveOccurred())
				Expect(extra).ToNot(BeNil())

				D, err := generalRecipe.GetDatabase().FindPackage(&types.Package{Name: "d", Category: "test", Version: "1.0"})
				Expect(err).ToNot(HaveOccurred())

				Expect(D.GetRequires()[0].GetName()).To(Equal("c"))
				CfromD, err := generalRecipe.GetDatabase().FindPackage(D.GetRequires()[0])
				Expect(err).ToNot(HaveOccurred())

				Expect(len(CfromD.GetRequires()) != 0).To(BeTrue())
				Expect(CfromD.GetRequires()[0].GetName()).To(Equal("b"))

				s := solver.NewSolver(types.SolverOptions{Type: types.SolverSingleCoreSimple}, pkg.NewInMemoryDatabase(false), generalRecipe.GetDatabase(), db)

				Dd, err := generalRecipe.GetDatabase().FindPackage(&types.Package{Name: "d", Category: "test", Version: "1.0"})
				Expect(err).ToNot(HaveOccurred())

				solution, err := s.Install([]*types.Package{Dd})
				Expect(err).ToNot(HaveOccurred())

				solution, err = solution.Order(generalRecipe.GetDatabase(), Dd.GetFingerPrint())
				Expect(err).ToNot(HaveOccurred())
				pack, err := generalRecipe.GetDatabase().FindPackage(&types.Package{Name: "a", Category: "test", Version: "1.0"})
				Expect(err).ToNot(HaveOccurred())

				base, err := generalRecipe.GetDatabase().FindPackage(&types.Package{Name: "base", Category: "layer", Version: "0.2"})
				Expect(err).ToNot(HaveOccurred())
				Expect(solution).ToNot(ContainElement(types.PackageAssert{Package: pack, Value: true}))
				Expect(solution).To(ContainElement(types.PackageAssert{Package: D, Value: true}))
				Expect(solution).ToNot(ContainElement(types.PackageAssert{Package: extra, Value: true}))
				Expect(solution).ToNot(ContainElement(types.PackageAssert{Package: base, Value: true}))
				Expect(len(solution)).To(Equal(3))
			}
		})
	})

	Context("Simple tree with labels", func() {
		It("Read tree with labels", func() {
			db := pkg.NewInMemoryDatabase(false)
			generalRecipe := NewCompilerRecipe(db)

			err := generalRecipe.Load("../../tests/fixtures/labels")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe.GetDatabase().World())).To(Equal(1))
			pack, err := generalRecipe.GetDatabase().FindPackage(&types.Package{Name: "pkgA", Category: "test", Version: "0.1"})
			Expect(err).ToNot(HaveOccurred())

			Expect(pack.HasLabel("label1")).To(Equal(true))
			Expect(pack.HasLabel("label3")).To(Equal(false))
		})
	})

	Context("Simple tree with annotations", func() {
		It("Read tree with annotations", func() {
			db := pkg.NewInMemoryDatabase(false)
			generalRecipe := NewCompilerRecipe(db)
			r := regexp.MustCompile("^label")

			err := generalRecipe.Load("../../tests/fixtures/annotations")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe.GetDatabase().World())).To(Equal(1))
			pack, err := generalRecipe.GetDatabase().FindPackage(&types.Package{Name: "pkgA", Category: "test", Version: "0.1"})
			Expect(err).ToNot(HaveOccurred())

			_, existsLabel1 := pack.Annotations["label1"]
			_, existsLabel2 := pack.Annotations["label3"]
			Expect(existsLabel1).To(Equal(true))
			Expect(existsLabel2).To(Equal(false))
			Expect(pack.MatchAnnotation(r)).To(Equal(true))
		})
	})

})
