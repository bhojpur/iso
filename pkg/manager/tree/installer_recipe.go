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

// InstallerRecipe is a builder imeplementation.

// It reads a Tree and spit it in human readable form (YAML), called recipe,
// It also loads a tree (recipe) from a YAML (to a db, e.g. BoltDB), allowing to query it
// with the solver, using the package object.

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"

	"github.com/pkg/errors"
)

const (
	FinalizerFile = "finalize.yaml"
)

func NewInstallerRecipe(db types.PackageDatabase) Builder {
	return &InstallerRecipe{Database: db}
}

// InstallerRecipe is the "general" reciper for Trees
type InstallerRecipe struct {
	SourcePath []string
	Database   types.PackageDatabase
}

func (r *InstallerRecipe) Save(path string) error {

	for _, p := range r.Database.World() {

		dir := filepath.Join(path, p.GetCategory(), p.GetName(), p.GetVersion())
		os.MkdirAll(dir, os.ModePerm)
		data, err := p.Yaml()
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(filepath.Join(dir, types.PackageDefinitionFile), data, 0644)
		if err != nil {
			return err
		}
		// Instead of rdeps, have a different tree for build deps.
		finalizerPath := p.Rel(FinalizerFile)
		if fileHelper.Exists(finalizerPath) { // copy finalizer file from the source tree
			fileHelper.CopyFile(finalizerPath, filepath.Join(dir, FinalizerFile))
		}

	}
	return nil
}

func (r *InstallerRecipe) Load(path string) error {

	if !fileHelper.Exists(path) {
		return errors.New(fmt.Sprintf(
			"Path %s doesn't exit.", path,
		))
	}

	r.SourcePath = append(r.SourcePath, path)

	//r.Tree().SetPackageSet(pkg.NewBoltDatabase(tmpfile.Name()))
	// TODO: Handle cleaning after? Cleanup implemented in GetPackageSet().Clean()

	// the function that handles each file or dir
	var ff = func(currentpath string, info os.FileInfo, err error) error {

		if info.Name() != types.PackageDefinitionFile && info.Name() != types.PackageCollectionFile {
			return nil // Skip with no errors
		}

		dat, err := ioutil.ReadFile(currentpath)
		if err != nil {
			return errors.Wrap(err, "Error reading file "+currentpath)
		}

		switch info.Name() {
		case types.PackageDefinitionFile:
			pack, err := types.PackageFromYaml(dat)
			if err != nil {
				return errors.Wrap(err, "Error reading yaml "+currentpath)
			}

			// Path is set only internally when tree is loaded from disk
			pack.SetPath(filepath.Dir(currentpath))
			_, err = r.Database.CreatePackage(&pack)
			if err != nil {
				return errors.Wrap(err, "Error creating package "+pack.GetName())
			}

		case types.PackageCollectionFile:
			packs, err := types.PackagesFromYAML(dat)
			if err != nil {
				return errors.Wrap(err, "Error reading yaml "+currentpath)
			}
			for _, p := range packs {
				// Path is set only internally when tree is loaded from disk
				p.SetPath(filepath.Dir(currentpath))
				_, err = r.Database.CreatePackage(&p)
				if err != nil {
					return errors.Wrap(err, "Error creating package "+p.GetName())
				}
			}

		}

		return nil
	}

	err := filepath.Walk(path, ff)
	if err != nil {
		return err
	}
	return nil
}

func (r *InstallerRecipe) GetDatabase() types.PackageDatabase   { return r.Database }
func (r *InstallerRecipe) WithDatabase(d types.PackageDatabase) { r.Database = d }
func (r *InstallerRecipe) GetSourcePath() []string              { return r.SourcePath }
