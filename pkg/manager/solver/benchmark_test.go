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
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	types "github.com/bhojpur/iso/pkg/manager/api/core/types"
	pkg "github.com/bhojpur/iso/pkg/manager/database"
	"github.com/bhojpur/iso/tests/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/bhojpur/iso/pkg/manager/solver"
)

var _ = Describe("Solver Benchmarks", func() {
	db := pkg.NewInMemoryDatabase(false)
	dbInstalled := pkg.NewInMemoryDatabase(false)
	dbDefinitions := pkg.NewInMemoryDatabase(false)
	var s types.PackageSolver

	Context("Complex data sets", func() {
		BeforeEach(func() {
			db = pkg.NewInMemoryDatabase(false)
			dbInstalled = pkg.NewInMemoryDatabase(false)
			dbDefinitions = pkg.NewInMemoryDatabase(false)
			s = NewSolver(types.SolverOptions{Type: types.SolverSingleCoreSimple}, dbInstalled, dbDefinitions, db)
			if os.Getenv("BENCHMARK_TESTS") != "true" {
				Skip("BENCHMARK_TESTS not enabled")
			}
		})
		Measure("it should be fast in resolution from a 50000 dataset", func(b Benchmarker) {

			runtime := b.Time("runtime", func() {
				for i := 0; i < 50000; i++ {
					C := types.NewPackage("C"+strconv.Itoa(i), "", []*types.Package{}, []*types.Package{})
					E := types.NewPackage("E"+strconv.Itoa(i), "", []*types.Package{}, []*types.Package{})
					F := types.NewPackage("F"+strconv.Itoa(i), "", []*types.Package{}, []*types.Package{})
					G := types.NewPackage("G"+strconv.Itoa(i), "", []*types.Package{}, []*types.Package{})
					H := types.NewPackage("H"+strconv.Itoa(i), "", []*types.Package{G}, []*types.Package{})
					D := types.NewPackage("D"+strconv.Itoa(i), "", []*types.Package{H}, []*types.Package{})
					B := types.NewPackage("B"+strconv.Itoa(i), "", []*types.Package{D}, []*types.Package{})
					A := types.NewPackage("A"+strconv.Itoa(i), "", []*types.Package{B}, []*types.Package{})
					for _, p := range []*types.Package{A, B, C, D, E, F, G} {
						_, err := dbDefinitions.CreatePackage(p)
						Expect(err).ToNot(HaveOccurred())
					}
					_, err := dbInstalled.CreatePackage(C)
					Expect(err).ToNot(HaveOccurred())
				}

				for i := 0; i < 1; i++ {
					C := types.NewPackage("C"+strconv.Itoa(i), "", []*types.Package{}, []*types.Package{})
					G := types.NewPackage("G"+strconv.Itoa(i), "", []*types.Package{}, []*types.Package{})
					H := types.NewPackage("H"+strconv.Itoa(i), "", []*types.Package{G}, []*types.Package{})
					D := types.NewPackage("D"+strconv.Itoa(i), "", []*types.Package{H}, []*types.Package{})
					B := types.NewPackage("B"+strconv.Itoa(i), "", []*types.Package{D}, []*types.Package{})
					A := types.NewPackage("A"+strconv.Itoa(i), "", []*types.Package{B}, []*types.Package{})

					solution, err := s.Install([]*types.Package{A})
					Expect(err).ToNot(HaveOccurred())

					Expect(solution).To(ContainElement(types.PackageAssert{Package: A, Value: true}))
					Expect(solution).To(ContainElement(types.PackageAssert{Package: B, Value: true}))
					Expect(solution).To(ContainElement(types.PackageAssert{Package: D, Value: true}))
					Expect(solution).To(ContainElement(types.PackageAssert{Package: C, Value: true}))
					Expect(solution).To(ContainElement(types.PackageAssert{Package: H, Value: true}))
					Expect(solution).To(ContainElement(types.PackageAssert{Package: G, Value: true}))
				}
			})

			Ω(runtime.Seconds()).Should(BeNumerically("<", 120), "Install() shouldn't take too long.")
		}, 1)
	})

	Context("Complex data sets - Parallel", func() {
		BeforeEach(func() {
			db = pkg.NewInMemoryDatabase(false)
			dbInstalled = pkg.NewInMemoryDatabase(false)
			dbDefinitions = pkg.NewInMemoryDatabase(false)
			s = NewSolver(types.SolverOptions{Type: types.SolverSingleCoreSimple, Concurrency: 10}, dbInstalled, dbDefinitions, db)
			if os.Getenv("BENCHMARK_TESTS") != "true" {
				Skip("BENCHMARK_TESTS not enabled")
			}
		})
		Measure("it should be fast in resolution from a 50000 dataset", func(b Benchmarker) {
			runtime := b.Time("runtime", func() {
				for i := 0; i < 50000; i++ {
					C := types.NewPackage("C"+strconv.Itoa(i), "", []*types.Package{}, []*types.Package{})
					E := types.NewPackage("E"+strconv.Itoa(i), "", []*types.Package{}, []*types.Package{})
					F := types.NewPackage("F"+strconv.Itoa(i), "", []*types.Package{}, []*types.Package{})
					G := types.NewPackage("G"+strconv.Itoa(i), "", []*types.Package{}, []*types.Package{})
					H := types.NewPackage("H"+strconv.Itoa(i), "", []*types.Package{G}, []*types.Package{})
					D := types.NewPackage("D"+strconv.Itoa(i), "", []*types.Package{H}, []*types.Package{})
					B := types.NewPackage("B"+strconv.Itoa(i), "", []*types.Package{D}, []*types.Package{})
					A := types.NewPackage("A"+strconv.Itoa(i), "", []*types.Package{B}, []*types.Package{})
					for _, p := range []*types.Package{A, B, C, D, E, F, G} {
						_, err := dbDefinitions.CreatePackage(p)
						Expect(err).ToNot(HaveOccurred())
					}
					_, err := dbInstalled.CreatePackage(C)
					Expect(err).ToNot(HaveOccurred())
				}
				for i := 0; i < 1; i++ {
					C := types.NewPackage("C"+strconv.Itoa(i), "", []*types.Package{}, []*types.Package{})
					G := types.NewPackage("G"+strconv.Itoa(i), "", []*types.Package{}, []*types.Package{})
					H := types.NewPackage("H"+strconv.Itoa(i), "", []*types.Package{G}, []*types.Package{})
					D := types.NewPackage("D"+strconv.Itoa(i), "", []*types.Package{H}, []*types.Package{})
					B := types.NewPackage("B"+strconv.Itoa(i), "", []*types.Package{D}, []*types.Package{})
					A := types.NewPackage("A"+strconv.Itoa(i), "", []*types.Package{B}, []*types.Package{})

					solution, err := s.Install([]*types.Package{A})
					Expect(err).ToNot(HaveOccurred())

					Expect(solution).To(ContainElement(types.PackageAssert{Package: A, Value: true}))
					Expect(solution).To(ContainElement(types.PackageAssert{Package: B, Value: true}))
					Expect(solution).To(ContainElement(types.PackageAssert{Package: D, Value: true}))
					Expect(solution).To(ContainElement(types.PackageAssert{Package: C, Value: true}))
					Expect(solution).To(ContainElement(types.PackageAssert{Package: H, Value: true}))
					Expect(solution).To(ContainElement(types.PackageAssert{Package: G, Value: true}))

					//	Expect(len(solution)).To(Equal(6))
				}
			})

			Ω(runtime.Seconds()).Should(BeNumerically("<", 120), "Install() shouldn't take too long.")
		}, 1)
	})

	Context("Complex data sets - Parallel Upgrades", func() {
		BeforeEach(func() {
			db = pkg.NewInMemoryDatabase(false)
			tmpfile, _ := ioutil.TempFile(os.TempDir(), "tests")
			defer os.Remove(tmpfile.Name())              // clean up
			dbInstalled = pkg.NewInMemoryDatabase(false) // pkg.NewBoltDatabase(tmpfile.Name())

			//	dbInstalled = pkg.NewInMemoryDatabase(false)
			dbDefinitions = pkg.NewInMemoryDatabase(false)
			s = NewSolver(types.SolverOptions{Type: types.SolverSingleCoreSimple, Concurrency: 100}, dbInstalled, dbDefinitions, db)
			if os.Getenv("BENCHMARK_TESTS") != "true" {
				Skip("BENCHMARK_TESTS not enabled")
			}
		})

		Measure("it should be fast in resolution from a 10000*8 dataset", func(b Benchmarker) {
			runtime := b.Time("runtime", func() {
				for i := 2; i < 10000; i++ {
					C := types.NewPackage("C", strconv.Itoa(i), []*types.Package{}, []*types.Package{})
					E := types.NewPackage("E", strconv.Itoa(i), []*types.Package{}, []*types.Package{})
					F := types.NewPackage("F", strconv.Itoa(i), []*types.Package{}, []*types.Package{})
					G := types.NewPackage("G", strconv.Itoa(i), []*types.Package{}, []*types.Package{})
					H := types.NewPackage("H", strconv.Itoa(i), []*types.Package{G}, []*types.Package{})
					D := types.NewPackage("D", strconv.Itoa(i), []*types.Package{H}, []*types.Package{})
					B := types.NewPackage("B", strconv.Itoa(i), []*types.Package{D}, []*types.Package{})
					A := types.NewPackage("A", strconv.Itoa(i), []*types.Package{B}, []*types.Package{})
					for _, p := range []*types.Package{A, B, C, D, E, F, G, H} {
						_, err := dbDefinitions.CreatePackage(p)
						Expect(err).ToNot(HaveOccurred())
					}
				}

				//C := types.NewPackage("C", "1", []*types.Package{}, []*types.Package{})
				G := types.NewPackage("G", "1", []*types.Package{}, []*types.Package{})
				H := types.NewPackage("H", "1", []*types.Package{G}, []*types.Package{})
				D := types.NewPackage("D", "1", []*types.Package{H}, []*types.Package{})
				B := types.NewPackage("B", "1", []*types.Package{D}, []*types.Package{})
				A := types.NewPackage("A", "1", []*types.Package{B}, []*types.Package{})
				_, err := dbInstalled.CreatePackage(A)
				Expect(err).ToNot(HaveOccurred())
				_, err = dbInstalled.CreatePackage(B)
				Expect(err).ToNot(HaveOccurred())
				_, err = dbInstalled.CreatePackage(D)
				Expect(err).ToNot(HaveOccurred())
				_, err = dbInstalled.CreatePackage(H)
				Expect(err).ToNot(HaveOccurred())
				_, err = dbInstalled.CreatePackage(G)
				Expect(err).ToNot(HaveOccurred())
				fmt.Println("Upgrade starts")

				packages, ass, err := s.Upgrade(false, false)
				Expect(err).ToNot(HaveOccurred())

				Expect(packages).To(ContainElement(A))

				G = types.NewPackage("G", "9999", []*types.Package{}, []*types.Package{})
				H = types.NewPackage("H", "9999", []*types.Package{G}, []*types.Package{})
				D = types.NewPackage("D", "9999", []*types.Package{H}, []*types.Package{})
				B = types.NewPackage("B", "9999", []*types.Package{D}, []*types.Package{})
				A = types.NewPackage("A", "9999", []*types.Package{B}, []*types.Package{})
				Expect(ass).To(ContainElement(types.PackageAssert{Package: A, Value: true}))

				Expect(len(packages)).To(Equal(5))
				//	Expect(len(solution)).To(Equal(6))

			})

			Ω(runtime.Seconds()).Should(BeNumerically("<", 120), "Install() shouldn't take too long.")
		}, 1)

		Measure("it should be fast in installation with 12000 packages installed and 2000*8 available", func(b Benchmarker) {
			runtime := b.Time("runtime", func() {
				for i := 0; i < 2000; i++ {
					C := types.NewPackage("C", strconv.Itoa(i), []*types.Package{}, []*types.Package{})
					E := types.NewPackage("E", strconv.Itoa(i), []*types.Package{}, []*types.Package{})
					F := types.NewPackage("F", strconv.Itoa(i), []*types.Package{}, []*types.Package{})
					G := types.NewPackage("G", strconv.Itoa(i), []*types.Package{}, []*types.Package{})
					H := types.NewPackage("H", strconv.Itoa(i), []*types.Package{G}, []*types.Package{})
					D := types.NewPackage("D", strconv.Itoa(i), []*types.Package{H}, []*types.Package{})
					B := types.NewPackage("B", strconv.Itoa(i), []*types.Package{D}, []*types.Package{})
					A := types.NewPackage("A", strconv.Itoa(i), []*types.Package{B}, []*types.Package{})
					for _, p := range []*types.Package{A, B, C, D, E, F, G} {
						_, err := dbDefinitions.CreatePackage(p)
						Expect(err).ToNot(HaveOccurred())
					}
					fmt.Println("Creating package, run", i)
				}

				for i := 0; i < 12000; i++ {
					x := helpers.RandomPackage()
					_, err := dbInstalled.CreatePackage(x)
					Expect(err).ToNot(HaveOccurred())
				}

				G := types.NewPackage("G", strconv.Itoa(50000), []*types.Package{}, []*types.Package{})
				H := types.NewPackage("H", strconv.Itoa(50000), []*types.Package{G}, []*types.Package{})
				D := types.NewPackage("D", strconv.Itoa(50000), []*types.Package{H}, []*types.Package{})
				B := types.NewPackage("B", strconv.Itoa(50000), []*types.Package{D}, []*types.Package{})
				A := types.NewPackage("A", strconv.Itoa(50000), []*types.Package{B}, []*types.Package{})

				ass, err := s.Install([]*types.Package{A})
				Expect(err).ToNot(HaveOccurred())

				Expect(ass).To(ContainElement(types.PackageAssert{Package: types.NewPackage("A", "50000", []*types.Package{B}, []*types.Package{}), Value: true}))
				//Expect(ass).To(Equal(5))
				//	Expect(len(solution)).To(Equal(6))

			})

			Ω(runtime.Seconds()).Should(BeNumerically("<", 120), "Install() shouldn't take too long.")
		}, 1)

		Measure("it should be fast in resolution from a 50000 dataset with upgrade universe", func(b Benchmarker) {
			runtime := b.Time("runtime", func() {
				for i := 0; i < 2; i++ {
					C := types.NewPackage("C", strconv.Itoa(i), []*types.Package{}, []*types.Package{})
					E := types.NewPackage("E", strconv.Itoa(i), []*types.Package{}, []*types.Package{})
					F := types.NewPackage("F", strconv.Itoa(i), []*types.Package{}, []*types.Package{})
					G := types.NewPackage("G", strconv.Itoa(i), []*types.Package{}, []*types.Package{})
					H := types.NewPackage("H", strconv.Itoa(i), []*types.Package{G}, []*types.Package{})
					D := types.NewPackage("D", strconv.Itoa(i), []*types.Package{H}, []*types.Package{})
					B := types.NewPackage("B", strconv.Itoa(i), []*types.Package{D}, []*types.Package{})
					A := types.NewPackage("A", strconv.Itoa(i), []*types.Package{B}, []*types.Package{})
					for _, p := range []*types.Package{A, B, C, D, E, F, G} {
						_, err := dbDefinitions.CreatePackage(p)
						Expect(err).ToNot(HaveOccurred())
					}
					fmt.Println("Creating package, run", i)
				}

				G := types.NewPackage("G", "1", []*types.Package{}, []*types.Package{})
				H := types.NewPackage("H", "1", []*types.Package{G}, []*types.Package{})
				D := types.NewPackage("D", "1", []*types.Package{H}, []*types.Package{})
				B := types.NewPackage("B", "1", []*types.Package{D}, []*types.Package{})
				A := types.NewPackage("A", "1", []*types.Package{B}, []*types.Package{})
				_, err := dbInstalled.CreatePackage(A)
				Expect(err).ToNot(HaveOccurred())
				fmt.Println("Upgrade starts")

				packages, ass, err := s.UpgradeUniverse(true)
				Expect(err).ToNot(HaveOccurred())

				Expect(ass).To(ContainElement(types.PackageAssert{Package: types.NewPackage("A", "50000", []*types.Package{B}, []*types.Package{}), Value: true}))
				Expect(packages).To(ContainElement(types.NewPackage("A", "50000", []*types.Package{B}, []*types.Package{})))
				Expect(packages).To(Equal(5))
				//	Expect(len(solution)).To(Equal(6))

			})

			Ω(runtime.Seconds()).Should(BeNumerically("<", 120), "Install() shouldn't take too long.")
		}, 1)
	})

})
