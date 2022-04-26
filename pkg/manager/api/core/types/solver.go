package types

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
	"github.com/crillab/gophersat/bf"
)

type SolverType int

const (
	SolverSingleCoreSimple SolverType = 0
)

// PackageSolver is an interface to a generic package solving algorithm
type PackageSolver interface {
	SetDefinitionDatabase(PackageDatabase)
	Install(p Packages) (PackagesAssertions, error)
	RelaxedInstall(p Packages) (PackagesAssertions, error)

	Uninstall(checkconflicts, full bool, candidate ...*Package) (Packages, error)
	ConflictsWithInstalled(p *Package) (bool, error)
	ConflictsWith(p *Package, ls Packages) (bool, error)
	Conflicts(pack *Package, lsp Packages) (bool, error)

	World() Packages
	Upgrade(checkconflicts, full bool) (Packages, PackagesAssertions, error)

	UpgradeUniverse(dropremoved bool) (Packages, PackagesAssertions, error)
	UninstallUniverse(toremove Packages) (Packages, error)

	SetResolver(PackageResolver)

	Solve() (PackagesAssertions, error)
	//	BestInstall(c Packages) (PackagesAssertions, error)
}

type SolverOptions struct {
	Type        SolverType `yaml:"type,omitempty"`
	Concurrency int        `yaml:"concurrency,omitempty"`
}

// PackageResolver assists PackageSolver on unsat cases
type PackageResolver interface {
	Solve(bf.Formula, PackageSolver) (PackagesAssertions, error)
}

type PackagesAssertions []PackageAssert

type PackageHash struct {
	BuildHash   string
	PackageHash string
}
