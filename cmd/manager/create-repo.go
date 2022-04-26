package cmd

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
	"os"
	"path/filepath"

	helpers "github.com/bhojpur/iso/cmd/manager/helpers"
	"github.com/bhojpur/iso/cmd/manager/util"
	"github.com/bhojpur/iso/pkg/manager/compiler"
	"github.com/bhojpur/iso/pkg/manager/compiler/types/compression"
	installer "github.com/bhojpur/iso/pkg/manager/installer"

	//	. "github.com/bhojpur/iso/pkg/manager/logger"
	pkg "github.com/bhojpur/iso/pkg/manager/database"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var createrepoCmd = &cobra.Command{
	Use:   "create-repo",
	Short: "Create a Bhojpur ISO repository from a build",
	Long: `Builds tree metadata from a set of packages and a tree definition:

	$ isomgr create-repo

Provide specific paths for packages, tree, and metadata output which is generated:

	$ isomgr create-repo --packages my/packages/path --tree my/tree/path --output my/packages/path ...

Provide name and description of the repository:

	$ isomgr create-repo --name "foo" --description "bar" ...

Change compression method:
	
	$ isomgr create-repo --tree-compression gzip --meta-compression gzip

Create a repository from the metadata description defined in the iso.yaml config file:

	$ isomgr create-repo --repo repository1
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("packages", cmd.Flags().Lookup("packages"))
		viper.BindPFlag("tree", cmd.Flags().Lookup("tree"))
		viper.BindPFlag("output", cmd.Flags().Lookup("output"))
		viper.BindPFlag("backend", cmd.Flags().Lookup("backend"))
		viper.BindPFlag("name", cmd.Flags().Lookup("name"))
		viper.BindPFlag("descr", cmd.Flags().Lookup("descr"))
		viper.BindPFlag("urls", cmd.Flags().Lookup("urls"))
		viper.BindPFlag("type", cmd.Flags().Lookup("type"))
		viper.BindPFlag("tree-compression", cmd.Flags().Lookup("tree-compression"))
		viper.BindPFlag("tree-filename", cmd.Flags().Lookup("tree-filename"))
		viper.BindPFlag("meta-compression", cmd.Flags().Lookup("meta-compression"))
		viper.BindPFlag("meta-filename", cmd.Flags().Lookup("meta-filename"))
		viper.BindPFlag("reset-revision", cmd.Flags().Lookup("reset-revision"))
		viper.BindPFlag("repo", cmd.Flags().Lookup("repo"))
		viper.BindPFlag("from-metadata", cmd.Flags().Lookup("from-metadata"))
		viper.BindPFlag("force-push", cmd.Flags().Lookup("force-push"))
		viper.BindPFlag("push-images", cmd.Flags().Lookup("push-images"))

	},
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var repo *installer.BhojpurSystemRepository

		treePaths := viper.GetStringSlice("tree")
		dst := viper.GetString("output")

		name := viper.GetString("name")
		descr := viper.GetString("descr")
		urls := viper.GetStringSlice("urls")
		t := viper.GetString("type")
		reset := viper.GetBool("reset-revision")
		treetype := viper.GetString("tree-compression")
		treeName := viper.GetString("tree-filename")
		metatype := viper.GetString("meta-compression")
		metaName := viper.GetString("meta-filename")
		source_repo := viper.GetString("repo")
		backendType := viper.GetString("backend")
		fromRepo, _ := cmd.Flags().GetBool("from-repositories")

		treeFile := installer.NewDefaultTreeRepositoryFile()
		metaFile := installer.NewDefaultMetaRepositoryFile()
		compilerBackend, err := compiler.NewBackend(util.DefaultContext, backendType)
		helpers.CheckErr(err)
		force := viper.GetBool("force-push")
		imagePush := viper.GetBool("push-images")
		snapshotID, _ := cmd.Flags().GetString("snapshot-id")

		opts := []installer.RepositoryOption{
			installer.WithSource(viper.GetString("packages")),
			installer.WithPushImages(imagePush),
			installer.WithForce(force),
			installer.FromRepository(fromRepo),
			installer.WithImagePrefix(dst),
			installer.WithDatabase(pkg.NewInMemoryDatabase(false)),
			installer.WithCompilerBackend(compilerBackend),
			installer.FromMetadata(viper.GetBool("from-metadata")),
			installer.WithContext(util.DefaultContext),
		}

		if source_repo != "" {
			// Search for system repository
			lrepo, err := util.DefaultContext.Config.GetSystemRepository(source_repo)
			helpers.CheckErr(err)

			if len(treePaths) <= 0 {
				treePaths = []string{lrepo.TreePath}
			}

			if t == "" {
				t = lrepo.Type
			}

			opts = append(opts,
				installer.WithName(lrepo.Name),
				installer.WithDescription(lrepo.Description),
				installer.WithType(t),
				installer.WithUrls(lrepo.Urls...),
				installer.WithPriority(lrepo.Priority),
				installer.WithTree(treePaths...),
			)

		} else {
			opts = append(opts,
				installer.WithName(name),
				installer.WithDescription(descr),
				installer.WithType(t),
				installer.WithUrls(urls...),
				installer.WithTree(treePaths...),
			)
		}

		repo, err = installer.GenerateRepository(opts...)
		helpers.CheckErr(err)

		if treetype != "" {
			treeFile.SetCompressionType(compression.Implementation(treetype))
		}

		if treeName != "" {
			treeFile.SetFileName(treeName)
		}

		if metatype != "" {
			metaFile.SetCompressionType(compression.Implementation(metatype))
		}

		if metaName != "" {
			metaFile.SetFileName(metaName)
		}
		repo.SetSnapshotID(snapshotID)
		repo.SetRepositoryFile(installer.REPOFILE_TREE_KEY, treeFile)
		repo.SetRepositoryFile(installer.REPOFILE_META_KEY, metaFile)

		err = repo.Write(util.DefaultContext, dst, reset, true)
		helpers.CheckErr(err)

	},
}

func init() {
	path, err := os.Getwd()
	helpers.CheckErr(err)

	createrepoCmd.Flags().String("packages", filepath.Join(path, "build"), "Packages folder (output from build)")
	createrepoCmd.Flags().StringSliceP("tree", "t", []string{path}, "Path of the source trees to use.")
	createrepoCmd.Flags().String("output", filepath.Join(path, "build"), "Destination for generated archives. With 'docker' repository type, it should be an image reference (e.g 'foo/bar')")
	createrepoCmd.Flags().String("name", "bhojpur", "Repository name")
	createrepoCmd.Flags().String("descr", "bhojpur", "Repository description")
	createrepoCmd.Flags().StringSlice("urls", []string{}, "Repository URLs")
	createrepoCmd.Flags().String("type", "disk", "Repository type (disk, http, docker)")
	createrepoCmd.Flags().Bool("reset-revision", false, "Reset repository revision.")
	createrepoCmd.Flags().String("repo", "", "Use repository defined in configuration.")
	createrepoCmd.Flags().String("backend", "docker", "backend used (docker,img)")

	createrepoCmd.Flags().Bool("force-push", false, "Force overwrite of docker images if already present online")
	createrepoCmd.Flags().Bool("push-images", false, "Enable/Disable docker image push for docker repositories")
	createrepoCmd.Flags().Bool("from-metadata", false, "Consider metadata files from the packages folder while indexing the new tree")

	createrepoCmd.Flags().String("tree-compression", "gzip", "Compression alg: none, gzip, zstd")
	createrepoCmd.Flags().String("tree-filename", installer.TREE_TARBALL, "Repository tree filename")
	createrepoCmd.Flags().String("meta-compression", "none", "Compression alg: none, gzip, zstd")
	createrepoCmd.Flags().String("meta-filename", installer.REPOSITORY_METAFILE+".tar", "Repository metadata filename")
	createrepoCmd.Flags().Bool("from-repositories", false, "Consume the user-defined repositories to pull specfiles from")
	createrepoCmd.Flags().String("snapshot-id", "", "Unique ID to use when creating repository snapshots")

	RootCmd.AddCommand(createrepoCmd)
}
