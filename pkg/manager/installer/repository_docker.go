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

	"github.com/bhojpur/iso/pkg/manager/api/core/bus"
	"github.com/bhojpur/iso/pkg/manager/api/core/image"
	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	artifact "github.com/bhojpur/iso/pkg/manager/api/core/types/artifact"
	compiler "github.com/bhojpur/iso/pkg/manager/compiler"
	"github.com/bhojpur/iso/pkg/manager/compiler/backend"
	"github.com/bhojpur/iso/pkg/manager/helpers"

	"github.com/pkg/errors"
)

type dockerRepositoryGenerator struct {
	b                       compiler.CompilerBackend
	imagePrefix, snapshotID string
	imagePush, force        bool
	context                 types.Context
}

func (l *dockerRepositoryGenerator) Initialize(path string, db types.PackageDatabase) ([]*artifact.PackageArtifact, error) {
	l.context.Info("Generating docker images for packages in", l.imagePrefix)
	var art []*artifact.PackageArtifact
	var ff = func(currentpath string, info os.FileInfo, err error) error {
		if err != nil {
			l.context.Debug("Skipping", info.Name(), err.Error())
			return nil
		}
		if info.IsDir() {
			l.context.Debug("Skipping directories")
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".metadata.yaml") {
			return nil
		}

		if err := l.pushImageFromArtifact(artifact.NewPackageArtifact(currentpath), l.b, true); err != nil {
			return errors.Wrap(err, "while pushing metadata file associated to the artifact")
		}

		dat, err := ioutil.ReadFile(currentpath)
		if err != nil {
			return errors.Wrap(err, "Error reading file "+currentpath)
		}

		a, err := artifact.NewPackageArtifactFromYaml(dat)
		if err != nil {
			return errors.Wrap(err, "Error reading yaml "+currentpath)
		}
		// Set the path relative to the file.
		// The metadata contains the full path where the file was located during buildtime.
		a.Path = filepath.Join(filepath.Dir(currentpath), filepath.Base(a.Path))

		// We want to include packages that are ONLY referenced in the tree.
		// the ones which aren't should be deleted. (TODO: by another cli command?)
		if _, notfound := db.FindPackage(a.CompileSpec.Package); notfound != nil {
			l.context.Debug(fmt.Sprintf("Package %s not found in tree. Ignoring it.",
				a.CompileSpec.Package.HumanReadableString()))
			return nil
		}

		packageImage := fmt.Sprintf("%s:%s", l.imagePrefix, a.CompileSpec.GetPackage().ImageID())

		if l.imagePush && l.b.ImageAvailable(packageImage) && !l.force {
			l.context.Info("Image", packageImage, "already present, skipping. use --force-push to override")
		} else {
			l.context.Info("Generating final image", packageImage,
				"for package ", a.CompileSpec.GetPackage().HumanReadableString())
			if err := a.GenerateFinalImage(l.context, packageImage, l.b, true); err != nil {
				return errors.Wrap(err, "Failed generating metadata tree"+packageImage)
			}
		}
		if l.imagePush {
			if err := pushImage(l.context, l.b, packageImage, l.force); err != nil {
				return errors.Wrapf(err, "Failed while pushing image: '%s'", packageImage)
			}
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

func pushImage(ctx types.Context, b compiler.CompilerBackend, image string, force bool) error {
	if b.ImageAvailable(image) && !force {
		ctx.Debug("Image", image, "already present, skipping")
		return nil
	}
	return b.Push(backend.Options{ImageName: image})
}

func (d *dockerRepositoryGenerator) pushFileFromArtifact(a *artifact.PackageArtifact, imageTree string) error {
	d.context.Debug("Generating image", imageTree)
	if err := a.GenerateFinalImage(d.context, imageTree, d.b, false); err != nil {
		return errors.Wrap(err, "Failed generating metadata tree "+imageTree)
	}
	if d.imagePush {
		if err := pushImage(d.context, d.b, imageTree, true); err != nil {
			return errors.Wrapf(err, "Failed while pushing image: '%s'", imageTree)
		}
	}
	return nil
}

func (d *dockerRepositoryGenerator) pushRepoMetadata(repospec, tag string, r *BhojpurSystemRepository) error {
	// create temp dir for metafile
	metaDir, err := d.context.TempDir("metadata")
	if err != nil {
		return errors.Wrap(err, "Error met while creating tempdir for metadata")
	}
	defer os.RemoveAll(metaDir) // clean up

	tempRepoFile := filepath.Join(metaDir, tag+".tar")
	if err := helpers.Tar(repospec, tempRepoFile); err != nil {
		return errors.Wrap(err, "Error met while archiving repository file")
	}

	a := artifact.NewPackageArtifact(tempRepoFile)
	imageRepo := fmt.Sprintf("%s:%s", d.imagePrefix, tag)

	if err := d.pushFileFromArtifact(a, imageRepo); err != nil {
		return errors.Wrap(err, "while pushing file from artifact")
	}
	return nil
}

func (d *dockerRepositoryGenerator) pushImageFromArtifact(a *artifact.PackageArtifact, b compiler.CompilerBackend, checkIfExists bool) error {
	// we generate a new archive containing the required compressed file.
	// TODO: Bundle all the extra files in 1 docker image only, instead of an image for each file
	treeArchive, err := artifact.CreateArtifactForFile(d.context, a.Path)
	if err != nil {
		return errors.Wrap(err, "failed generating checksums for tree")
	}
	imageTree := fmt.Sprintf("%s:%s", d.imagePrefix, helpers.SanitizeImageString(a.GetFileName()))
	if checkIfExists && d.imagePush && d.b.ImageAvailable(imageTree) && !d.force {
		d.context.Info("Image", imageTree, "already present, skipping. use --force-push to override")
		return nil
	} else {
		return d.pushFileFromArtifact(treeArchive, imageTree)
	}
}

// Generate creates a Docker Bhojpur ISO repository
func (d *dockerRepositoryGenerator) Generate(r *BhojpurSystemRepository, imagePrefix string, resetRevision bool) error {
	// - Iterate over meta, build final images, push them if necessary
	//   - while pushing, check if image already exists, and if exist push them only if --force is supplied
	// - Generate final images for metadata and push

	imageRepository := fmt.Sprintf("%s:%s", imagePrefix, REPOSITORY_SPECFILE)

	r.LastUpdate = strconv.FormatInt(time.Now().Unix(), 10)

	repoTemp, err := d.context.TempDir("repo")
	if err != nil {
		return errors.Wrap(err, "error met while creating tempdir for repository")
	}
	defer os.RemoveAll(repoTemp) // clean up

	if r.GetBackend().ImageAvailable(imageRepository) {

		err := r.GetBackend().DownloadImage(backend.Options{ImageName: imageRepository})
		if err != nil {
			return errors.Wrapf(err, "while downloading '%s'", imageRepository)
		}
		img, err := r.GetBackend().ImageReference(imageRepository, true)
		if err != nil {
			return errors.Wrapf(err, "while downloading '%s'", imageRepository)
		}
		_, _, err = image.ExtractTo(
			d.context,
			img,
			repoTemp,
			nil,
		)
		if err != nil {
			return errors.Wrapf(err, "while extracting '%s'", imageRepository)
		}

	}

	repospec := filepath.Join(repoTemp, REPOSITORY_SPECFILE)

	// Increment the internal revision version by reading the one which is already available (if any)
	if err := r.BumpRevision(repospec, resetRevision); err != nil {
		return err
	}

	d.context.Info(fmt.Sprintf(
		"For repository %s creating revision %d and last update %s...",
		r.Name, r.Revision, r.LastUpdate,
	))

	bus.Manager.Publish(bus.EventRepositoryPreBuild, struct {
		Repo BhojpurSystemRepository
		Path string
	}{
		Repo: *r,
		Path: imageRepository,
	})

	// Create tree and repository file
	a, err := r.AddTree(d.context, r.GetTree(), repoTemp, REPOFILE_TREE_KEY, NewDefaultTreeRepositoryFile())
	if err != nil {
		return errors.Wrap(err, "error met while adding runtime tree to repository")
	}

	// we generate a new archive containing the required compressed file.
	// TODO: Bundle all the extra files in 1 docker image only, instead of an image for each file
	if err := d.pushImageFromArtifact(a, d.b, false); err != nil {
		return errors.Wrap(err, "error met while pushing runtime tree")
	}

	a, err = r.AddTree(d.context, r.BuildTree, repoTemp, REPOFILE_COMPILER_TREE_KEY, NewDefaultCompilerTreeRepositoryFile())
	if err != nil {
		return errors.Wrap(err, "error met while adding compiler tree to repository")
	}
	// we generate a new archive containing the required compressed file.
	// TODO: Bundle all the extra files in 1 docker image only, instead of an image for each file
	if err := d.pushImageFromArtifact(a, d.b, false); err != nil {
		return errors.Wrap(err, "error met while pushing compiler tree")
	}

	a, err = r.AddMetadata(d.context, repospec, repoTemp)
	if err != nil {
		return errors.Wrap(err, "failed adding Metadata file to repository")
	}

	if err := d.pushImageFromArtifact(a, d.b, false); err != nil {
		return errors.Wrap(err, "error met while pushing docker image from artifact")
	}

	if err := d.pushRepoMetadata(repospec, REPOSITORY_SPECFILE, r); err != nil {
		return errors.Wrap(err, "while pushing repository metadata tree")
	}

	// Create a named snapshot and push it.
	// It edits the metadata pointing at the repository files associated with the snapshot
	// And copies the new files
	artifacts, snapshotRepoFile, err := r.Snapshot(d.snapshotID, repoTemp)
	if err != nil {
		return errors.Wrap(err, "while creating snapshot")
	}
	if err := d.pushRepoMetadata(snapshotRepoFile, filepath.Base(snapshotRepoFile), r); err != nil {
		return errors.Wrap(err, "while pushing repository snapshot metadata tree")
	}

	for _, a := range artifacts {
		if err := d.pushImageFromArtifact(a, d.b, false); err != nil {
			return errors.Wrap(err, "error met while pushing docker image from artifact")
		}
	}

	bus.Manager.Publish(bus.EventRepositoryPostBuild, struct {
		Repo BhojpurSystemRepository
		Path string
	}{
		Repo: *r,
		Path: imagePrefix,
	})
	return nil
}
