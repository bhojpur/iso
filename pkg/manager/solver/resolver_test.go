package solver_test

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
	"github.com/bhojpur/iso/pkg/manager/api/core/types"

	pkg "github.com/bhojpur/iso/pkg/manager/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/bhojpur/iso/pkg/manager/solver"
)

var _ = Describe("Resolver", func() {

	db := pkg.NewInMemoryDatabase(false)
	dbInstalled := pkg.NewInMemoryDatabase(false)
	dbDefinitions := pkg.NewInMemoryDatabase(false)
	s := NewSolver(types.SolverOptions{Type: types.SolverSingleCoreSimple}, dbInstalled, dbDefinitions, db)

	BeforeEach(func() {
		db = pkg.NewInMemoryDatabase(false)
		dbInstalled = pkg.NewInMemoryDatabase(false)
		dbDefinitions = pkg.NewInMemoryDatabase(false)
		s = NewSolver(types.SolverOptions{Type: types.SolverSingleCoreSimple}, dbInstalled, dbDefinitions, db)
	})

	Context("Conflict set", func() {
		Context("Explainer", func() {
			It("is unsolvable - as we something we ask to install conflict with system stuff", func() {
				C := types.NewPackage("C", "", []*types.Package{}, []*types.Package{})
				B := types.NewPackage("B", "", []*types.Package{}, []*types.Package{C})
				A := types.NewPackage("A", "", []*types.Package{B}, []*types.Package{})

				for _, p := range []*types.Package{A, B, C} {
					_, err := dbDefinitions.CreatePackage(p)
					Expect(err).ToNot(HaveOccurred())
				}

				for _, p := range []*types.Package{C} {
					_, err := dbInstalled.CreatePackage(p)
					Expect(err).ToNot(HaveOccurred())
				}

				solution, err := s.Install([]*types.Package{A})
				Expect(len(solution)).To(Equal(0))
				Expect(err).To(HaveOccurred())
			})
			It("succeeds to install D and F if explictly requested", func() {
				C := types.NewPackage("C", "", []*types.Package{}, []*types.Package{})
				B := types.NewPackage("B", "", []*types.Package{}, []*types.Package{C})
				A := types.NewPackage("A", "", []*types.Package{B}, []*types.Package{})
				D := types.NewPackage("D", "", []*types.Package{}, []*types.Package{})
				E := types.NewPackage("E", "", []*types.Package{B}, []*types.Package{})
				F := types.NewPackage("F", "", []*types.Package{}, []*types.Package{})

				for _, p := range []*types.Package{A, B, C, D, E, F} {
					_, err := dbDefinitions.CreatePackage(p)
					Expect(err).ToNot(HaveOccurred())
				}

				for _, p := range []*types.Package{C} {
					_, err := dbInstalled.CreatePackage(p)
					Expect(err).ToNot(HaveOccurred())
				}

				solution, err := s.Install([]*types.Package{D, F}) // D and F should go as they have no deps. A/E should be filtered by QLearn
				Expect(err).ToNot(HaveOccurred())

				Expect(len(solution)).To(Equal(6))

				Expect(solution).ToNot(ContainElement(types.PackageAssert{Package: A, Value: true}))
				Expect(solution).ToNot(ContainElement(types.PackageAssert{Package: B, Value: true}))
				Expect(solution).To(ContainElement(types.PackageAssert{Package: C, Value: true}))
				Expect(solution).To(ContainElement(types.PackageAssert{Package: D, Value: true}))
				Expect(solution).ToNot(ContainElement(types.PackageAssert{Package: E, Value: true}))
				Expect(solution).To(ContainElement(types.PackageAssert{Package: F, Value: true}))

			})

		})
		Context("QLearningResolver", func() {
			It("will find out that we can install D by ignoring A", func() {
				s.SetResolver(SimpleQLearningSolver())
				C := types.NewPackage("C", "", []*types.Package{}, []*types.Package{})
				B := types.NewPackage("B", "", []*types.Package{}, []*types.Package{C})
				A := types.NewPackage("A", "", []*types.Package{B}, []*types.Package{})
				D := types.NewPackage("D", "", []*types.Package{}, []*types.Package{})

				for _, p := range []*types.Package{A, B, C, D} {
					_, err := dbDefinitions.CreatePackage(p)
					Expect(err).ToNot(HaveOccurred())
				}

				for _, p := range []*types.Package{C} {
					_, err := dbInstalled.CreatePackage(p)
					Expect(err).ToNot(HaveOccurred())
				}

				solution, err := s.Install([]*types.Package{A, D})
				Expect(err).ToNot(HaveOccurred())

				Expect(solution).ToNot(ContainElement(types.PackageAssert{Package: A, Value: true}))
				Expect(solution).ToNot(ContainElement(types.PackageAssert{Package: B, Value: true}))
				Expect(solution).To(ContainElement(types.PackageAssert{Package: C, Value: true}))
				Expect(solution).To(ContainElement(types.PackageAssert{Package: D, Value: true}))

				Expect(len(solution)).To(Equal(4))
			})

			It("will find out that we can install D and F by ignoring E and A", func() {
				s.SetResolver(SimpleQLearningSolver())
				C := types.NewPackage("C", "", []*types.Package{}, []*types.Package{})
				B := types.NewPackage("B", "", []*types.Package{}, []*types.Package{C})
				A := types.NewPackage("A", "", []*types.Package{B}, []*types.Package{})
				D := types.NewPackage("D", "", []*types.Package{}, []*types.Package{})
				E := types.NewPackage("E", "", []*types.Package{B}, []*types.Package{})
				F := types.NewPackage("F", "", []*types.Package{}, []*types.Package{})

				for _, p := range []*types.Package{A, B, C, D, E, F} {
					_, err := dbDefinitions.CreatePackage(p)
					Expect(err).ToNot(HaveOccurred())
				}

				for _, p := range []*types.Package{C} {
					_, err := dbInstalled.CreatePackage(p)
					Expect(err).ToNot(HaveOccurred())
				}

				solution, err := s.Install([]*types.Package{A, D, E, F}) // D and F should go as they have no deps. A/E should be filtered by QLearn
				Expect(err).ToNot(HaveOccurred())

				Expect(solution).ToNot(ContainElement(types.PackageAssert{Package: A, Value: true}))
				Expect(solution).ToNot(ContainElement(types.PackageAssert{Package: B, Value: true}))
				Expect(solution).To(ContainElement(types.PackageAssert{Package: C, Value: true})) // Was already installed
				Expect(solution).To(ContainElement(types.PackageAssert{Package: D, Value: true}))
				Expect(solution).ToNot(ContainElement(types.PackageAssert{Package: E, Value: true}))
				Expect(solution).To(ContainElement(types.PackageAssert{Package: F, Value: true}))
				Expect(len(solution)).To(Equal(6))
			})
		})

		Context("Explainer", func() {
			It("cannot find a solution", func() {
				C := types.NewPackage("C", "", []*types.Package{}, []*types.Package{})
				B := types.NewPackage("B", "", []*types.Package{}, []*types.Package{C})
				A := types.NewPackage("A", "", []*types.Package{B}, []*types.Package{})
				D := types.NewPackage("D", "", []*types.Package{}, []*types.Package{})

				for _, p := range []*types.Package{A, B, C, D} {
					_, err := dbDefinitions.CreatePackage(p)
					Expect(err).ToNot(HaveOccurred())
				}

				for _, p := range []*types.Package{C} {
					_, err := dbInstalled.CreatePackage(p)
					Expect(err).ToNot(HaveOccurred())
				}

				solution, err := s.Install([]*types.Package{A, D})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(`could not satisfy the constraints: 
A-- and 
C-- and 
!(A--) or B-- and 
!(B--) or !(C--)`))

				Expect(len(solution)).To(Equal(0))
			})

		})
	})

})
