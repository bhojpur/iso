package compilerspec

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
	"path/filepath"

	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	options "github.com/bhojpur/iso/pkg/manager/compiler/types/options"
	"github.com/mitchellh/hashstructure/v2"

	"github.com/ghodss/yaml"
	"github.com/otiai10/copy"
	dirhash "golang.org/x/mod/sumdb/dirhash"
)

type BhojpurCompilationspecs []BhojpurCompilationSpec

func NewBhojpurCompilationspecs(s ...*BhojpurCompilationSpec) *BhojpurCompilationspecs {
	all := BhojpurCompilationspecs{}

	for _, spec := range s {
		all.Add(spec)
	}
	return &all
}

func (specs BhojpurCompilationspecs) Len() int {
	return len(specs)
}

func (specs *BhojpurCompilationspecs) Remove(s *BhojpurCompilationspecs) *BhojpurCompilationspecs {
	newSpecs := BhojpurCompilationspecs{}
SPECS:
	for _, spec := range specs.All() {
		for _, target := range s.All() {
			if target.GetPackage().Matches(spec.GetPackage()) {
				continue SPECS
			}
		}
		newSpecs.Add(spec)
	}
	return &newSpecs
}

func (specs *BhojpurCompilationspecs) Add(s *BhojpurCompilationSpec) {
	*specs = append(*specs, *s)
}

func (specs *BhojpurCompilationspecs) All() []*BhojpurCompilationSpec {
	var cspecs []*BhojpurCompilationSpec
	for i := range *specs {
		f := (*specs)[i]
		cspecs = append(cspecs, &f)
	}

	return cspecs
}

func (specs *BhojpurCompilationspecs) Unique() *BhojpurCompilationspecs {
	newSpecs := BhojpurCompilationspecs{}
	seen := map[string]bool{}

	for i := range *specs {
		j := (*specs)[i]
		_, ok := seen[j.GetPackage().GetFingerPrint()]
		if !ok {
			seen[j.GetPackage().GetFingerPrint()] = true
			newSpecs = append(newSpecs, j)
		}
	}
	return &newSpecs
}

type CopyField struct {
	Package     *types.Package `json:"package"`
	Image       string         `json:"image"`
	Source      string         `json:"source"`
	Destination string         `json:"destination"`
}

type BhojpurCompilationSpec struct {
	Steps           []string                 `json:"steps"` // Are run inside a container and the result layer diff is saved
	Env             []string                 `json:"env"`
	Prelude         []string                 `json:"prelude"` // Are run inside the image which will be our builder
	Image           string                   `json:"image"`
	Seed            string                   `json:"seed"`
	Package         *types.Package           `json:"package"`
	SourceAssertion types.PackagesAssertions `json:"-"`
	PackageDir      string                   `json:"package_dir" yaml:"package_dir"`

	Retrieve []string `json:"retrieve"`

	OutputPath string   `json:"-"` // Where the build processfiles go
	Unpack     bool     `json:"unpack"`
	Includes   []string `json:"includes"`
	Excludes   []string `json:"excludes"`

	BuildOptions *options.Compiler `json:"build_options"`

	Copy []CopyField `json:"copy"`

	RequiresFinalImages bool `json:"requires_final_images" yaml:"requires_final_images"`
}

// Signature is a portion of the spec that yields a signature for the hash
type Signature struct {
	Image               string
	Steps               []string
	PackageDir          string
	Prelude             []string
	Seed                string
	Env                 []string
	Retrieve            []string
	Unpack              bool
	Includes            []string
	Excludes            []string
	Copy                []CopyField
	Requires            types.Packages
	RequiresFinalImages bool
}

func (cs *BhojpurCompilationSpec) signature() Signature {
	return Signature{
		Image:               cs.Image,
		Steps:               cs.Steps,
		PackageDir:          cs.PackageDir,
		Prelude:             cs.Prelude,
		Seed:                cs.Seed,
		Env:                 cs.Env,
		Retrieve:            cs.Retrieve,
		Unpack:              cs.Unpack,
		Includes:            cs.Includes,
		Excludes:            cs.Excludes,
		Copy:                cs.Copy,
		Requires:            cs.Package.GetRequires(),
		RequiresFinalImages: cs.RequiresFinalImages,
	}
}

func NewBhojpurCompilationSpec(b []byte, p *types.Package) (*BhojpurCompilationSpec, error) {
	var spec BhojpurCompilationSpec
	var packageDefinition types.Package
	err := yaml.Unmarshal(b, &spec)
	if err != nil {
		return &spec, err
	}
	err = yaml.Unmarshal(b, &packageDefinition)
	if err != nil {
		return &spec, err
	}

	// Update requires/conflict/provides
	// When we have been passed a bytes slice, parse it as a package
	// and updates requires/conflicts/provides.
	// This is required in order to allow manipulation of such fields with templating
	copy := *p
	spec.Package = &copy
	if len(packageDefinition.GetRequires()) != 0 {
		spec.Package.Requires(packageDefinition.GetRequires())
	}
	if len(packageDefinition.GetConflicts()) != 0 {
		spec.Package.Conflicts(packageDefinition.GetConflicts())
	}
	if len(packageDefinition.GetProvides()) != 0 {
		spec.Package.SetProvides(packageDefinition.GetProvides())
	}
	return &spec, nil
}
func (cs *BhojpurCompilationSpec) GetSourceAssertion() types.PackagesAssertions {
	return cs.SourceAssertion
}

func (cs *BhojpurCompilationSpec) SetBuildOptions(b options.Compiler) {
	cs.BuildOptions = &b
}

func (cs *BhojpurCompilationSpec) SetSourceAssertion(as types.PackagesAssertions) {
	cs.SourceAssertion = as
}
func (cs *BhojpurCompilationSpec) GetPackage() *types.Package {
	return cs.Package
}

func (cs *BhojpurCompilationSpec) GetPackageDir() string {
	return cs.PackageDir
}

func (cs *BhojpurCompilationSpec) SetPackageDir(s string) {
	cs.PackageDir = s
}

func (cs *BhojpurCompilationSpec) BuildSteps() []string {
	return cs.Steps
}

func (cs *BhojpurCompilationSpec) ImageUnpack() bool {
	return cs.Unpack
}

func (cs *BhojpurCompilationSpec) GetPreBuildSteps() []string {
	return cs.Prelude
}

func (cs *BhojpurCompilationSpec) GetIncludes() []string {
	return cs.Includes
}

func (cs *BhojpurCompilationSpec) GetExcludes() []string {
	return cs.Excludes
}

func (cs *BhojpurCompilationSpec) GetRetrieve() []string {
	return cs.Retrieve
}

// IsVirtual returns true if the spec is virtual.
// A spec is virtual if the package is empty, and it has no image source to unpack from.
func (cs *BhojpurCompilationSpec) IsVirtual() bool {
	return cs.EmptyPackage() && !cs.HasImageSource()
}

func (cs *BhojpurCompilationSpec) GetSeedImage() string {
	return cs.Seed
}

func (cs *BhojpurCompilationSpec) GetImage() string {
	return cs.Image
}

func (cs *BhojpurCompilationSpec) GetOutputPath() string {
	return cs.OutputPath
}

func (p *BhojpurCompilationSpec) Rel(s string) string {
	return filepath.Join(p.GetOutputPath(), s)
}

func (cs *BhojpurCompilationSpec) SetImage(s string) {
	cs.Image = s
}

func (cs *BhojpurCompilationSpec) SetOutputPath(s string) {
	cs.OutputPath = s
}

func (cs *BhojpurCompilationSpec) SetSeedImage(s string) {
	cs.Seed = s
}

func (cs *BhojpurCompilationSpec) EmptyPackage() bool {
	return len(cs.BuildSteps()) == 0 && !cs.UnpackedPackage()
}

func (cs *BhojpurCompilationSpec) UnpackedPackage() bool {
	// If package_dir was specified in the spec, we want to treat the content of the directory
	// as the root of our archive.  ImageUnpack is implied to be true. override it
	unpack := cs.ImageUnpack()
	if cs.GetPackageDir() != "" {
		unpack = true
	}
	return unpack
}

// HasImageSource returns true when the compilation spec has an image source.
// a compilation spec has an image source when it depends on other packages or have a source image
// explictly supplied
func (cs *BhojpurCompilationSpec) HasImageSource() bool {
	return (cs.Package != nil && len(cs.GetPackage().GetRequires()) != 0) || cs.GetImage() != "" || (cs.RequiresFinalImages && len(cs.Package.GetRequires()) != 0)
}

func (cs *BhojpurCompilationSpec) Hash() (string, error) {
	// build a signature, we want to be part of the hash only the fields that are relevant for build purposes
	signature := cs.signature()
	h, err := hashstructure.Hash(signature, hashstructure.FormatV2, nil)
	if err != nil {
		return "", err
	}
	sum, err := dirhash.HashDir(cs.Package.Path, "", dirhash.DefaultHash)
	if err != nil {
		return fmt.Sprint(h), err
	}
	return fmt.Sprint(h, sum), err
}

func (cs *BhojpurCompilationSpec) CopyRetrieves(dest string) error {
	var err error
	if len(cs.Retrieve) > 0 {
		for _, s := range cs.Retrieve {
			matches, err := filepath.Glob(cs.Rel(s))

			if err != nil {
				continue
			}

			for _, m := range matches {
				err = copy.Copy(m, filepath.Join(dest, filepath.Base(m)))
			}
		}
	}
	return err
}

func (cs *BhojpurCompilationSpec) genDockerfile(image string, steps []string) string {
	spec := `
FROM ` + image + `
COPY . /isobuild
WORKDIR /isobuild
ENV PACKAGE_NAME=` + cs.Package.GetName() + `
ENV PACKAGE_VERSION=` + cs.Package.GetVersion() + `
ENV PACKAGE_CATEGORY=` + cs.Package.GetCategory()

	if len(cs.Retrieve) > 0 {
		for _, s := range cs.Retrieve {
			//var file string
			// if helpers.IsValidUrl(s) {
			// 	file = s
			// } else {
			// 	file = cs.Rel(s)
			// }
			spec = spec + `
ADD ` + s + ` /isobuild/`
		}
	}

	for _, c := range cs.Copy {
		if c.Image != "" {
			copyLine := fmt.Sprintf("\nCOPY --from=%s %s %s\n", c.Image, c.Source, c.Destination)
			spec = spec + copyLine
		}
	}

	for _, s := range cs.Env {
		spec = spec + `
ENV ` + s
	}

	for _, s := range steps {
		spec = spec + `
RUN ` + s
	}
	return spec
}

// RenderBuildImage renders the dockerfile of the image used as a pre-build step
func (cs *BhojpurCompilationSpec) RenderBuildImage() (string, error) {
	return cs.genDockerfile(cs.GetSeedImage(), cs.GetPreBuildSteps()), nil

}

// RenderStepImage renders the dockerfile used for the image used for building the package
func (cs *BhojpurCompilationSpec) RenderStepImage(image string) (string, error) {
	return cs.genDockerfile(image, cs.BuildSteps()), nil
}

func (cs *BhojpurCompilationSpec) WriteBuildImageDefinition(path string) error {
	data, err := cs.RenderBuildImage()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, []byte(data), 0644)
}

func (cs *BhojpurCompilationSpec) WriteStepImageDefinition(fromimage, path string) error {
	data, err := cs.RenderStepImage(fromimage)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, []byte(data), 0644)
}
