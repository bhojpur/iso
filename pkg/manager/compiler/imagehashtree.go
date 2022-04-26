package compiler

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

	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	compilerspec "github.com/bhojpur/iso/pkg/manager/compiler/types/spec"
	"github.com/pkg/errors"
)

// ImageHashTree is holding the Database
// and the options to resolve PackageImageHashTrees
// for a given specfile
// It is responsible of returning a concrete result
// which identifies a Package in a HashTree
type ImageHashTree struct {
	Database      types.PackageDatabase
	SolverOptions types.BhojpurSolverOptions
}

// PackageImageHashTree represent the Package into a given image hash tree
// The hash tree is constructed by a set of images representing
// the package during its build stage. A Hash is assigned to each image
// from the package fingerprint, plus the SAT solver assertion result (which is hashed as well)
// and the specfile signatures. This guarantees that each image of the build stage
// is unique and can be identified later on.
type PackageImageHashTree struct {
	Target                       *types.PackageAssert
	Dependencies                 types.PackagesAssertions
	Solution                     types.PackagesAssertions
	dependencyBuilderImageHashes map[string]string
	SourceHash                   string
	BuilderImageHash             string
}

func NewHashTree(db types.PackageDatabase) *ImageHashTree {
	return &ImageHashTree{
		Database: db,
	}
}

func (ht *PackageImageHashTree) DependencyBuildImage(p *types.Package) (string, error) {
	found, ok := ht.dependencyBuilderImageHashes[p.GetFingerPrint()]
	if !ok {
		return "", errors.New("package hash not found")
	}
	return found, nil
}

func (ht *PackageImageHashTree) String() string {
	return fmt.Sprintf(
		"Target buildhash: %s\nTarget packagehash: %s\nBuilder Imagehash: %s\nSource Imagehash: %s\n",
		ht.Target.Hash.BuildHash,
		ht.Target.Hash.PackageHash,
		ht.BuilderImageHash,
		ht.SourceHash,
	)
}

// Query takes a compiler and a compilation spec and returns a PackageImageHashTree tied to it.
// PackageImageHashTree contains all the informations to resolve the spec build images in order to
// reproducibly re-build images from packages
func (ht *ImageHashTree) Query(cs *BhojpurCompiler, p *compilerspec.BhojpurCompilationSpec) (*PackageImageHashTree, error) {
	assertions, err := ht.resolve(cs, p)
	if err != nil {
		return nil, err
	}
	targetAssertion := assertions.Search(p.GetPackage().GetFingerPrint())

	dependencies := assertions.Drop(p.GetPackage())
	var sourceHash string
	imageHashes := map[string]string{}
	for _, assertion := range dependencies {
		var depbuildImageTag string
		compileSpec, err := cs.FromPackage(assertion.Package)
		if err != nil {
			return nil, errors.Wrap(err, "Error while generating compilespec for "+assertion.Package.GetName())
		}
		if compileSpec.GetImage() != "" {
			depbuildImageTag = assertion.Hash.BuildHash
		} else {
			depbuildImageTag = ht.genBuilderImageTag(compileSpec, targetAssertion.Hash.PackageHash)
		}
		imageHashes[assertion.Package.GetFingerPrint()] = depbuildImageTag
		sourceHash = assertion.Hash.PackageHash
	}

	return &PackageImageHashTree{
		Dependencies:                 dependencies,
		Target:                       targetAssertion,
		SourceHash:                   sourceHash,
		BuilderImageHash:             ht.genBuilderImageTag(p, targetAssertion.Hash.PackageHash),
		dependencyBuilderImageHashes: imageHashes,
		Solution:                     assertions,
	}, nil
}

func (ht *ImageHashTree) genBuilderImageTag(p *compilerspec.BhojpurCompilationSpec, packageImage string) string {
	// Use packageImage as salt into the fp being used
	// so the hash is unique also in cases where
	// some package deps does have completely different
	// depgraphs
	return fmt.Sprintf("builder-%s", p.GetPackage().HashFingerprint(packageImage))
}

// resolve computes the dependency tree of a compilation spec and returns solver assertions
// in order to be able to compile the spec.
func (ht *ImageHashTree) resolve(cs *BhojpurCompiler, p *compilerspec.BhojpurCompilationSpec) (types.PackagesAssertions, error) {
	dependencies, err := cs.ComputeDepTree(p, cs.Database)
	if err != nil {
		return nil, errors.Wrap(err, "While computing a solution for "+p.GetPackage().HumanReadableString())
	}

	// Get hash from buildpsecs
	salts := map[string]string{}
	for _, assertion := range dependencies { //highly dependent on the order
		if assertion.Value {
			spec, err := cs.FromPackage(assertion.Package)
			if err != nil {
				return nil, errors.Wrap(err, "while computing hash buildspecs")
			}
			hash, err := spec.Hash()
			if err != nil {
				return nil, errors.Wrap(err, "failed computing hash")
			}
			salts[assertion.Package.GetFingerPrint()] = hash
		}
	}

	assertions := types.PackagesAssertions{}
	for _, assertion := range dependencies { //highly dependent on the order
		if assertion.Value {
			nthsolution := dependencies.Cut(assertion.Package)
			assertion.Hash = types.PackageHash{
				BuildHash:   nthsolution.SaltedHashFrom(assertion.Package, salts),
				PackageHash: nthsolution.SaltedAssertionHash(salts),
			}
			assertion.Package.SetTreeDir(p.Package.GetTreeDir())
			assertions = append(assertions, assertion)
		}
	}

	return assertions, nil
}
