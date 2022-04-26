package installer

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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	artifact "github.com/bhojpur/iso/pkg/manager/api/core/types/artifact"

	"github.com/bhojpur/iso/pkg/manager/api/core/bus"
	"github.com/pkg/errors"
)

type localRepositoryGenerator struct {
	context    types.Context
	snapshotID string
}

func (l *localRepositoryGenerator) Initialize(path string, db types.PackageDatabase) ([]*artifact.PackageArtifact, error) {
	return buildPackageIndex(l.context, path, db)
}

func buildPackageIndex(ctx types.Context, path string, db types.PackageDatabase) ([]*artifact.PackageArtifact, error) {

	var art []*artifact.PackageArtifact
	var ff = func(currentpath string, info os.FileInfo, err error) error {
		if err != nil {
			ctx.Debug("Failed walking", err.Error())
			return err
		}

		if !strings.HasSuffix(info.Name(), ".metadata.yaml") {
			return nil // Skip with no errors
		}

		dat, err := ioutil.ReadFile(currentpath)
		if err != nil {
			return errors.Wrap(err, "Error reading file "+currentpath)
		}

		a, err := artifact.NewPackageArtifactFromYaml(dat)
		if err != nil {
			return errors.Wrap(err, "Error reading yaml "+currentpath)
		}

		// We want to include packages that are ONLY referenced in the tree.
		// the ones which aren't should be deleted. (TODO: by another cli command?)
		if _, notfound := db.FindPackage(a.CompileSpec.GetPackage()); notfound != nil {
			ctx.Debug(fmt.Sprintf("Package %s not found in tree. Ignoring it.",
				a.CompileSpec.GetPackage().HumanReadableString()))
			return nil
		}

		art = append(art, a)

		return nil
	}

	err := filepath.Walk(path, ff)
	if err != nil {
		return nil, err

	}
	return art, nil
}

// Generate creates a Local Bhojpur ISO repository
func (g *localRepositoryGenerator) Generate(r *BhojpurSystemRepository, dst string, resetRevision bool) error {
	err := os.MkdirAll(dst, os.ModePerm)
	if err != nil {
		return err
	}
	r.LastUpdate = strconv.FormatInt(time.Now().Unix(), 10)

	repospec := filepath.Join(dst, REPOSITORY_SPECFILE)
	// Increment the internal revision version by reading the one which is already available (if any)
	if err := r.BumpRevision(repospec, resetRevision); err != nil {
		return err
	}

	g.context.Info(fmt.Sprintf(
		"Repository %s: creating revision %d and last update %s...",
		r.Name, r.Revision, r.LastUpdate,
	))

	bus.Manager.Publish(bus.EventRepositoryPreBuild, struct {
		Repo BhojpurSystemRepository
		Path string
	}{
		Repo: *r,
		Path: dst,
	})

	if _, err := r.AddTree(g.context, r.GetTree(), dst, REPOFILE_TREE_KEY, NewDefaultTreeRepositoryFile()); err != nil {
		return errors.Wrap(err, "error met while adding runtime tree to repository")
	}

	if _, err := r.AddTree(g.context, r.BuildTree, dst, REPOFILE_COMPILER_TREE_KEY, NewDefaultCompilerTreeRepositoryFile()); err != nil {
		return errors.Wrap(err, "error met while adding compiler tree to repository")
	}

	if _, err := r.AddMetadata(g.context, repospec, dst); err != nil {
		return errors.Wrap(err, "failed adding Metadata file to repository")
	}

	// Create named snapshot.
	// It edits the metadata pointing at the repository files associated with the snapshot
	// And copies the new files
	if _, _, err := r.Snapshot(g.snapshotID, dst); err != nil {
		return errors.Wrap(err, "while creating snapshot")
	}

	bus.Manager.Publish(bus.EventRepositoryPostBuild, struct {
		Repo BhojpurSystemRepository
		Path string
	}{
		Repo: *r,
		Path: dst,
	})
	return nil
}
