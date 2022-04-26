package tree

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
	"path/filepath"

	"github.com/bhojpur/iso/pkg/manager/api/core/template"
	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"
	"github.com/pkg/errors"
)

const (
	CompilerDefinitionFile = "build.yaml"
)

func NewCompilerRecipe(d types.PackageDatabase) Builder {
	return &CompilerRecipe{Recipe: Recipe{Database: d}}
}

func ReadDefinitionFile(path string) (types.Package, error) {
	empty := types.Package{}
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return empty, errors.Wrap(err, "Error reading file "+path)
	}
	pack, err := types.PackageFromYaml(dat)
	if err != nil {
		return empty, errors.Wrap(err, "Error reading yaml "+path)
	}

	return pack, nil
}

// Recipe is the "general" reciper for Trees
type CompilerRecipe struct {
	Recipe
}

// CompilerRecipes copies tree 1:1 as they contain the specs
// and the build context required for reproducible builds
func (r *CompilerRecipe) Save(path string) error {
	for _, p := range r.SourcePath {
		if err := fileHelper.CopyDir(p, filepath.Join(path, filepath.Base(p))); err != nil {
			return errors.Wrap(err, "while copying source tree")
		}
	}
	return nil
}

func (r *CompilerRecipe) Load(path string) error {

	r.SourcePath = append(r.SourcePath, path)

	c, err := template.FilesInDir(template.FindPossibleTemplatesDir(path))
	if err != nil {
		return err
	}

	var ff = func(currentpath string, info os.FileInfo, err error) error {

		if err != nil {
			return errors.Wrap(err, "Error on walk path "+currentpath)
		}

		if info.Name() != types.PackageDefinitionFile && info.Name() != types.PackageCollectionFile {
			return nil // Skip with no errors
		}

		switch info.Name() {
		case types.PackageDefinitionFile:

			pack, err := ReadDefinitionFile(currentpath)
			if err != nil {
				return err
			}
			// Path is set only internally when tree is loaded from disk
			pack.SetPath(filepath.Dir(currentpath))
			pack.SetTreeDir(path)

			// Instead of rdeps, have a different tree for build deps.
			compileDefPath := pack.Rel(CompilerDefinitionFile)
			if fileHelper.Exists(compileDefPath) {
				dat, err := template.RenderWithValues(append(c, compileDefPath), currentpath)
				if err != nil {
					return errors.Wrap(err,
						"Error templating file "+CompilerDefinitionFile+" from "+
							filepath.Dir(currentpath))
				}

				packbuild, err := types.PackageFromYaml([]byte(dat))
				if err != nil {
					return errors.Wrap(err,
						"Error reading yaml "+CompilerDefinitionFile+" from "+
							filepath.Dir(currentpath))
				}
				pack.Requires(packbuild.GetRequires())
				pack.Conflicts(packbuild.GetConflicts())
			}

			_, err = r.Database.CreatePackage(&pack)
			if err != nil {
				return errors.Wrap(err, "Error creating package "+pack.GetName())
			}

		case types.PackageCollectionFile:

			dat, err := ioutil.ReadFile(currentpath)
			if err != nil {
				return errors.Wrap(err, "Error reading file "+currentpath)
			}

			packs, err := types.PackagesFromYAML(dat)
			if err != nil {
				return errors.Wrap(err, "Error reading yaml "+currentpath)
			}

			packsRaw, err := types.GetRawPackages(dat)
			if err != nil {
				return errors.Wrap(err, "Error reading raw packages from "+currentpath)
			}

			for _, pack := range packs {
				pack.SetPath(filepath.Dir(currentpath))
				pack.SetTreeDir(path)

				// Instead of rdeps, have a different tree for build deps.
				compileDefPath := pack.Rel(CompilerDefinitionFile)
				if fileHelper.Exists(compileDefPath) {

					raw := packsRaw.Find(pack.GetName(), pack.GetCategory(), pack.GetVersion())
					buildyaml, err := ioutil.ReadFile(compileDefPath)
					if err != nil {
						return errors.Wrap(err, "Error reading file "+currentpath)
					}
					dat, err := template.Render(append(template.ReadFiles(c...), string(buildyaml)), raw, map[string]interface{}{})
					if err != nil {
						return errors.Wrap(err,
							"Error templating file "+CompilerDefinitionFile+" from "+
								filepath.Dir(currentpath))
					}

					packbuild, err := types.PackageFromYaml([]byte(dat))
					if err != nil {
						return errors.Wrap(err,
							"Error reading yaml "+CompilerDefinitionFile+" from "+
								filepath.Dir(currentpath))
					}
					pack.Requires(packbuild.GetRequires())

					pack.Conflicts(packbuild.GetConflicts())
				}

				_, err = r.Database.CreatePackage(&pack)
				if err != nil {
					return errors.Wrap(err, "Error creating package "+pack.GetName())
				}
			}
		}
		return nil
	}

	err = filepath.Walk(path, ff)
	if err != nil {
		return err
	}
	return nil
}

func (r *CompilerRecipe) GetDatabase() types.PackageDatabase   { return r.Database }
func (r *CompilerRecipe) WithDatabase(d types.PackageDatabase) { r.Database = d }
func (r *CompilerRecipe) GetSourcePath() []string              { return r.SourcePath }
