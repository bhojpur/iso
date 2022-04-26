package installer_test

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
	//	. "github.com/bhojpur/iso/pkg/manager/installer"

	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bhojpur/iso/pkg/manager/api/core/context"
	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	artifact "github.com/bhojpur/iso/pkg/manager/api/core/types/artifact"
	"github.com/bhojpur/iso/pkg/manager/compiler"
	backend "github.com/bhojpur/iso/pkg/manager/compiler/backend"
	"github.com/bhojpur/iso/pkg/manager/compiler/types/options"
	compilerspec "github.com/bhojpur/iso/pkg/manager/compiler/types/spec"
	pkg "github.com/bhojpur/iso/pkg/manager/database"
	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"
	. "github.com/bhojpur/iso/pkg/manager/installer"
	"github.com/bhojpur/iso/pkg/manager/tree"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func dockerStubRepo(tmpdir, tree, image string, push, force bool) (*BhojpurSystemRepository, error) {
	return GenerateRepository(
		WithName("test"),
		WithDescription("description"),
		WithType("docker"),
		WithUrls(image),
		WithPriority(1),
		WithSource(tmpdir),
		WithTree(tree),
		WithDatabase(pkg.NewInMemoryDatabase(false)),
		WithCompilerBackend(backend.NewSimpleDockerBackend(context.NewContext())),
		WithImagePrefix(image),
		WithPushImages(push),
		WithContext(context.NewContext()),
		WithForce(force))
}

var _ = Describe("Repository", func() {
	Context("Generation", func() {
		ctx := context.NewContext()
		It("Generate repository metadata", func() {

			tmpdir, err := ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			generalRecipe := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe.Load("../../tests/fixtures/buildable")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe.GetDatabase().GetPackages())).To(Equal(3))

			compiler := compiler.NewBhojpurCompiler(backend.NewSimpleDockerBackend(ctx), generalRecipe.GetDatabase())

			spec, err := compiler.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(spec.GetPackage().GetPath()).ToNot(Equal(""))

			tmpdir, err = ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			Expect(spec.BuildSteps()).To(Equal([]string{"echo artifact5 > /test5", "echo artifact6 > /test6", "chmod +x generate.sh", "./generate.sh"}))
			Expect(spec.GetPreBuildSteps()).To(Equal([]string{"echo foo > /test", "echo bar > /test2"}))

			spec.SetOutputPath(tmpdir)

			artifact, err := compiler.Compile(false, spec)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileHelper.Exists(artifact.Path)).To(BeTrue())
			Expect(artifact.Unpack(ctx, tmpdir, false)).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(spec.Rel("test5"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel("test6"))).To(BeTrue())

			content1, err := fileHelper.Read(spec.Rel("test5"))
			Expect(err).ToNot(HaveOccurred())
			content2, err := fileHelper.Read(spec.Rel("test6"))
			Expect(err).ToNot(HaveOccurred())
			Expect(content1).To(Equal("artifact5\n"))
			Expect(content2).To(Equal("artifact6\n"))

			Expect(fileHelper.Exists(spec.Rel("b-test-1.0.package.tar"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel("b-test-1.0.metadata.yaml"))).To(BeTrue())

			repo, err := stubRepo(tmpdir, "../../tests/fixtures/buildable")
			Expect(err).ToNot(HaveOccurred())
			Expect(repo.GetName()).To(Equal("test"))
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_SPECFILE))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).ToNot(BeTrue())
			err = repo.Write(ctx, tmpdir, false, true)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_SPECFILE))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).To(BeTrue())
		})

		It("Generate repository metadata of files ONLY referenced in a tree", func() {

			tmpdir, err := ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			generalRecipe := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe.Load("../../tests/fixtures/buildable")
			Expect(err).ToNot(HaveOccurred())

			generalRecipe2 := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe2.Load("../../tests/fixtures/finalizers")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe2.GetDatabase().GetPackages())).To(Equal(1))
			Expect(len(generalRecipe.GetDatabase().GetPackages())).To(Equal(3))

			compiler2 := compiler.NewBhojpurCompiler(backend.NewSimpleDockerBackend(ctx), generalRecipe2.GetDatabase(), options.WithContext(context.NewContext()))
			spec2, err := compiler2.FromPackage(&types.Package{Name: "alpine", Category: "seed", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			compiler := compiler.NewBhojpurCompiler(backend.NewSimpleDockerBackend(ctx), generalRecipe.GetDatabase(), options.WithContext(context.NewContext()))

			spec, err := compiler.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(spec.GetPackage().GetPath()).ToNot(Equal(""))
			Expect(spec2.GetPackage().GetPath()).ToNot(Equal(""))

			tmpdir, err = ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			Expect(spec.BuildSteps()).To(Equal([]string{"echo artifact5 > /test5", "echo artifact6 > /test6", "chmod +x generate.sh", "./generate.sh"}))
			Expect(spec.GetPreBuildSteps()).To(Equal([]string{"echo foo > /test", "echo bar > /test2"}))

			spec.SetOutputPath(tmpdir)
			spec2.SetOutputPath(tmpdir)

			artifact, err := compiler.Compile(false, spec)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileHelper.Exists(artifact.Path)).To(BeTrue())
			Expect(artifact.Unpack(ctx, tmpdir, false)).ToNot(HaveOccurred())

			artifact2, err := compiler2.Compile(false, spec2)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileHelper.Exists(artifact2.Path)).To(BeTrue())
			Expect(artifact2.Unpack(ctx, tmpdir, false)).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(spec.Rel("test5"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel("test6"))).To(BeTrue())

			content1, err := fileHelper.Read(spec.Rel("test5"))
			Expect(err).ToNot(HaveOccurred())
			content2, err := fileHelper.Read(spec.Rel("test6"))
			Expect(err).ToNot(HaveOccurred())
			Expect(content1).To(Equal("artifact5\n"))
			Expect(content2).To(Equal("artifact6\n"))

			Expect(fileHelper.Exists(spec.Rel("b-test-1.0.package.tar"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel("b-test-1.0.metadata.yaml"))).To(BeTrue())
			Expect(fileHelper.Exists(spec2.Rel("alpine-seed-1.0.package.tar"))).To(BeTrue())
			Expect(fileHelper.Exists(spec2.Rel("alpine-seed-1.0.metadata.yaml"))).To(BeTrue())

			repo, err := stubRepo(tmpdir, "../../tests/fixtures/buildable")
			Expect(err).ToNot(HaveOccurred())
			Expect(repo.GetName()).To(Equal("test"))
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_SPECFILE))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).ToNot(BeTrue())
			err = repo.Write(ctx, tmpdir, false, true)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_SPECFILE))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).To(BeTrue())

			// We check now that the artifact not referenced in the tree
			// (spec2) is not indexed in the repository
			repository, err := NewBhojpurSystemRepositoryFromYaml([]byte(`
name: "test"
type: "disk"
urls:
  - "`+tmpdir+`"
`), pkg.NewInMemoryDatabase(false))
			Expect(err).ToNot(HaveOccurred())
			repos, err := repository.Sync(ctx, true)
			Expect(err).ToNot(HaveOccurred())

			_, err = repos.GetTree().GetDatabase().FindPackage(spec.GetPackage())
			Expect(err).ToNot(HaveOccurred())
			_, err = repos.GetTree().GetDatabase().FindPackage(spec2.GetPackage())
			Expect(err).To(HaveOccurred()) // should throw error
		})

		It("Generates snapshots", func() {

			tmpdir, err := ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			generalRecipe := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe.Load("../../tests/fixtures/buildable")
			Expect(err).ToNot(HaveOccurred())

			generalRecipe2 := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe2.Load("../../tests/fixtures/finalizers")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe2.GetDatabase().GetPackages())).To(Equal(1))
			Expect(len(generalRecipe.GetDatabase().GetPackages())).To(Equal(3))

			compiler2 := compiler.NewBhojpurCompiler(backend.NewSimpleDockerBackend(ctx), generalRecipe2.GetDatabase(), options.WithContext(ctx))
			spec2, err := compiler2.FromPackage(&types.Package{Name: "alpine", Category: "seed", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			compiler := compiler.NewBhojpurCompiler(backend.NewSimpleDockerBackend(ctx), generalRecipe.GetDatabase(), options.WithContext(ctx))

			spec, err := compiler.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(spec.GetPackage().GetPath()).ToNot(Equal(""))
			Expect(spec2.GetPackage().GetPath()).ToNot(Equal(""))

			tmpdir, err = ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			Expect(spec.BuildSteps()).To(Equal([]string{"echo artifact5 > /test5", "echo artifact6 > /test6", "chmod +x generate.sh", "./generate.sh"}))
			Expect(spec.GetPreBuildSteps()).To(Equal([]string{"echo foo > /test", "echo bar > /test2"}))

			spec.SetOutputPath(tmpdir)
			spec2.SetOutputPath(tmpdir)

			artifact, err := compiler.Compile(false, spec)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileHelper.Exists(artifact.Path)).To(BeTrue())
			Expect(artifact.Unpack(ctx, tmpdir, false)).ToNot(HaveOccurred())

			artifact2, err := compiler2.Compile(false, spec2)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileHelper.Exists(artifact2.Path)).To(BeTrue())
			Expect(artifact2.Unpack(ctx, tmpdir, false)).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(spec.Rel("test5"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel("test6"))).To(BeTrue())

			content1, err := fileHelper.Read(spec.Rel("test5"))
			Expect(err).ToNot(HaveOccurred())
			content2, err := fileHelper.Read(spec.Rel("test6"))
			Expect(err).ToNot(HaveOccurred())
			Expect(content1).To(Equal("artifact5\n"))
			Expect(content2).To(Equal("artifact6\n"))

			// will contain both
			Expect(fileHelper.Exists(spec.Rel("b-test-1.0.package.tar"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel("b-test-1.0.metadata.yaml"))).To(BeTrue())
			Expect(fileHelper.Exists(spec2.Rel("alpine-seed-1.0.package.tar"))).To(BeTrue())
			Expect(fileHelper.Exists(spec2.Rel("alpine-seed-1.0.metadata.yaml"))).To(BeTrue())

			repo, err := GenerateRepository(
				WithName("test"),
				WithDescription("description"),
				WithType("disk"),
				WithUrls(tmpdir),
				WithPriority(1),
				WithSource(tmpdir),
				FromMetadata(true), // Enabling from metadata makes the package visible
				WithTree("../../tests/fixtures/buildable"),
				WithContext(ctx),
				WithDatabase(pkg.NewInMemoryDatabase(false)),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(repo.GetName()).To(Equal("test"))
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_SPECFILE))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).ToNot(BeTrue())
			err = repo.Write(ctx, tmpdir, false, true)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_SPECFILE))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).To(BeTrue())

			artifacts, index, err := repo.Snapshot("foo", tmpdir)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(artifacts)).To(Equal(3))
			Expect(index).To(ContainSubstring("foo-repository.yaml"))
			r := &BhojpurSystemRepository{}

			r, err = r.ReadSpecFile(index)
			Expect(err).ToNot(HaveOccurred())

			Expect(err).ToNot(HaveOccurred())

			Expect(len(r.RepositoryFiles)).To(Equal(3))

			for k, v := range r.RepositoryFiles {
				_, err := os.Stat(filepath.Join(tmpdir, "foo-compilertree.tar.gz"))
				Expect(err).ToNot(HaveOccurred())
				switch k {
				case REPOFILE_COMPILER_TREE_KEY:
					Expect(v.FileName).To(Equal("foo-compilertree.tar.gz"))
				case REPOFILE_META_KEY:
					Expect(v.FileName).To(Equal("foo-repository.meta.yaml.tar"))
				case REPOFILE_TREE_KEY:
					Expect(v.FileName).To(Equal("foo-tree.tar.gz"))
				}
			}
		})

		It("Generate repository metadata of files referenced in a tree and from packages", func() {

			tmpdir, err := ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			generalRecipe := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe.Load("../../tests/fixtures/buildable")
			Expect(err).ToNot(HaveOccurred())

			generalRecipe2 := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe2.Load("../../tests/fixtures/finalizers")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe2.GetDatabase().GetPackages())).To(Equal(1))
			Expect(len(generalRecipe.GetDatabase().GetPackages())).To(Equal(3))

			compiler2 := compiler.NewBhojpurCompiler(backend.NewSimpleDockerBackend(ctx), generalRecipe2.GetDatabase(), options.WithContext(ctx))
			spec2, err := compiler2.FromPackage(&types.Package{Name: "alpine", Category: "seed", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			compiler := compiler.NewBhojpurCompiler(backend.NewSimpleDockerBackend(ctx), generalRecipe.GetDatabase(), options.WithContext(ctx))

			spec, err := compiler.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(spec.GetPackage().GetPath()).ToNot(Equal(""))
			Expect(spec2.GetPackage().GetPath()).ToNot(Equal(""))

			tmpdir, err = ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			Expect(spec.BuildSteps()).To(Equal([]string{"echo artifact5 > /test5", "echo artifact6 > /test6", "chmod +x generate.sh", "./generate.sh"}))
			Expect(spec.GetPreBuildSteps()).To(Equal([]string{"echo foo > /test", "echo bar > /test2"}))

			spec.SetOutputPath(tmpdir)
			spec2.SetOutputPath(tmpdir)

			artifact, err := compiler.Compile(false, spec)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileHelper.Exists(artifact.Path)).To(BeTrue())
			Expect(artifact.Unpack(ctx, tmpdir, false)).ToNot(HaveOccurred())

			artifact2, err := compiler2.Compile(false, spec2)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileHelper.Exists(artifact2.Path)).To(BeTrue())
			Expect(artifact2.Unpack(ctx, tmpdir, false)).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(spec.Rel("test5"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel("test6"))).To(BeTrue())

			content1, err := fileHelper.Read(spec.Rel("test5"))
			Expect(err).ToNot(HaveOccurred())
			content2, err := fileHelper.Read(spec.Rel("test6"))
			Expect(err).ToNot(HaveOccurred())
			Expect(content1).To(Equal("artifact5\n"))
			Expect(content2).To(Equal("artifact6\n"))

			// will contain both
			Expect(fileHelper.Exists(spec.Rel("b-test-1.0.package.tar"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel("b-test-1.0.metadata.yaml"))).To(BeTrue())
			Expect(fileHelper.Exists(spec2.Rel("alpine-seed-1.0.package.tar"))).To(BeTrue())
			Expect(fileHelper.Exists(spec2.Rel("alpine-seed-1.0.metadata.yaml"))).To(BeTrue())

			repo, err := GenerateRepository(
				WithName("test"),
				WithDescription("description"),
				WithType("disk"),
				WithUrls(tmpdir),
				WithPriority(1),
				WithSource(tmpdir),
				FromMetadata(true), // Enabling from metadata makes the package visible
				WithTree("../../tests/fixtures/buildable"),
				WithContext(ctx),
				WithDatabase(pkg.NewInMemoryDatabase(false)),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(repo.GetName()).To(Equal("test"))
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_SPECFILE))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).ToNot(BeTrue())
			err = repo.Write(ctx, tmpdir, false, true)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_SPECFILE))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).To(BeTrue())

			// We check now that the artifact not referenced in the tree
			// (spec2) is not indexed in the repository
			repository, err := NewBhojpurSystemRepositoryFromYaml([]byte(`
name: "test"
type: "disk"
urls:
  - "`+tmpdir+`"
`), pkg.NewInMemoryDatabase(false))
			Expect(err).ToNot(HaveOccurred())
			repos, err := repository.Sync(ctx, true)
			Expect(err).ToNot(HaveOccurred())

			_, err = repos.GetTree().GetDatabase().FindPackage(spec.GetPackage())
			Expect(err).ToNot(HaveOccurred())
			_, err = repos.GetTree().GetDatabase().FindPackage(spec2.GetPackage())
			Expect(err).ToNot(HaveOccurred()) // should NOT throw error
		})
	})
	Context("Matching packages", func() {
		It("Matches packages in different repositories by priority", func() {
			package1 := &types.Package{Name: "Test"}
			package2 := &types.Package{Name: "Test2"}
			builder1 := tree.NewInstallerRecipe(pkg.NewInMemoryDatabase(false))
			builder2 := tree.NewInstallerRecipe(pkg.NewInMemoryDatabase(false))

			_, err := builder1.GetDatabase().CreatePackage(package1)
			Expect(err).ToNot(HaveOccurred())

			_, err = builder2.GetDatabase().CreatePackage(package2)
			Expect(err).ToNot(HaveOccurred())
			repo1 := &BhojpurSystemRepository{BhojpurRepository: &types.BhojpurRepository{Name: "test1"}, Tree: builder1}
			repo2 := &BhojpurSystemRepository{BhojpurRepository: &types.BhojpurRepository{Name: "test2"}, Tree: builder2}
			repositories := Repositories{repo1, repo2}
			matches := repositories.PackageMatches([]*types.Package{package1})
			Expect(matches).To(Equal([]PackageMatch{{Repo: repo1, Package: package1}}))

		})
	})
	Context("Docker repository", func() {
		repoImage := os.Getenv("UNIT_TEST_DOCKER_IMAGE_REPOSITORY")
		ctx := context.NewContext()
		BeforeEach(func() {
			if repoImage == "" {
				Skip("UNIT_TEST_DOCKER_IMAGE_REPOSITORY not specified")
			}
			ctx = context.NewContext()
		})

		It("generates images", func() {
			b := backend.NewSimpleDockerBackend(ctx)
			tmpdir, err := ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			generalRecipe := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe.Load("../../tests/fixtures/buildable")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe.GetDatabase().GetPackages())).To(Equal(3))

			localcompiler := compiler.NewBhojpurCompiler(
				backend.NewSimpleDockerBackend(ctx), generalRecipe.GetDatabase(), options.WithContext(ctx))

			spec, err := localcompiler.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(spec.GetPackage().GetPath()).ToNot(Equal(""))

			tmpdir, err = ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			spec.SetOutputPath(tmpdir)

			a, err := localcompiler.Compile(false, spec)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileHelper.Exists(a.Path)).To(BeTrue())
			Expect(a.Unpack(ctx, tmpdir, false)).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(spec.Rel("test5"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel("test6"))).To(BeTrue())

			repo, err := dockerStubRepo(tmpdir, "../../tests/fixtures/buildable", repoImage, true, true)
			Expect(err).ToNot(HaveOccurred())
			Expect(repo.GetName()).To(Equal("test"))
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_SPECFILE))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).ToNot(BeTrue())
			err = repo.Write(ctx, repoImage, false, true)
			Expect(err).ToNot(HaveOccurred())

			Expect(b.ImageAvailable(fmt.Sprintf("%s:%s", repoImage, "tree.tar.gz"))).To(BeTrue())
			Expect(b.ImageAvailable(fmt.Sprintf("%s:%s", repoImage, "repository.meta.yaml.tar"))).To(BeTrue())
			Expect(b.ImageAvailable(fmt.Sprintf("%s:%s", repoImage, "repository.yaml"))).To(BeTrue())
			Expect(b.ImageAvailable(fmt.Sprintf("%s:%s", repoImage, "b-test-1.0"))).To(BeTrue())

			extracted, err := ioutil.TempDir("", "extracted")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(extracted) // clean up

			c := repo.Client(ctx)

			f, err := c.DownloadFile("repository.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(fileHelper.Read(f)).To(ContainSubstring("name: test"))

			a, err = c.DownloadArtifact(&artifact.PackageArtifact{
				Path: "test.tar",
				CompileSpec: &compilerspec.BhojpurCompilationSpec{
					Package: &types.Package{
						Name:     "b",
						Category: "test",
						Version:  "1.0",
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(a.Unpack(ctx, extracted, false)).ToNot(HaveOccurred())
			Expect(fileHelper.Read(filepath.Join(extracted, "test6"))).To(Equal("artifact6\n"))
		})

		It("generates images of virtual packages", func() {
			b := backend.NewSimpleDockerBackend(ctx)
			tmpdir, err := ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			generalRecipe := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe.Load("../../tests/fixtures/virtuals")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe.GetDatabase().GetPackages())).To(Equal(5))

			localcompiler := compiler.NewBhojpurCompiler(
				backend.NewSimpleDockerBackend(ctx), generalRecipe.GetDatabase(), options.WithContext(ctx))

			spec, err := localcompiler.FromPackage(&types.Package{Name: "a", Category: "test", Version: "1.99"})
			Expect(err).ToNot(HaveOccurred())

			Expect(spec.GetPackage().GetPath()).ToNot(Equal(""))

			tmpdir, err = ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			spec.SetOutputPath(tmpdir)

			a, err := localcompiler.Compile(false, spec)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileHelper.Exists(a.Path)).To(BeTrue())
			Expect(a.Unpack(ctx, tmpdir, false)).ToNot(HaveOccurred())

			repo, err := dockerStubRepo(tmpdir, "../../tests/fixtures/virtuals", repoImage, true, true)
			Expect(err).ToNot(HaveOccurred())
			Expect(repo.GetName()).To(Equal("test"))
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_SPECFILE))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).ToNot(BeTrue())
			err = repo.Write(ctx, repoImage, false, true)
			Expect(err).ToNot(HaveOccurred())

			Expect(b.ImageAvailable(fmt.Sprintf("%s:%s", repoImage, "tree.tar.gz"))).To(BeTrue())
			Expect(b.ImageAvailable(fmt.Sprintf("%s:%s", repoImage, "repository.meta.yaml.tar"))).To(BeTrue())
			Expect(b.ImageAvailable(fmt.Sprintf("%s:%s", repoImage, "repository.yaml"))).To(BeTrue())
			Expect(b.ImageAvailable(fmt.Sprintf("%s:%s", repoImage, "a-test-1.99"))).To(BeTrue())

			extracted, err := ioutil.TempDir("", "extracted")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(extracted) // clean up

			c := repo.Client(ctx)

			f, err := c.DownloadFile("repository.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(fileHelper.Read(f)).To(ContainSubstring("name: test"))

			a, err = c.DownloadArtifact(&artifact.PackageArtifact{
				Path: "test.tar",
				CompileSpec: &compilerspec.BhojpurCompilationSpec{
					Package: &types.Package{
						Name:     "a",
						Category: "test",
						Version:  "1.99",
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(a.Unpack(ctx, extracted, false)).ToNot(HaveOccurred())

			Expect(fileHelper.DirectoryIsEmpty(extracted)).To(BeTrue())

		})

		It("Searches files", func() {
			repos := Repositories{
				&BhojpurSystemRepository{
					Index: compiler.ArtifactIndex{
						&artifact.PackageArtifact{
							CompileSpec: &compilerspec.BhojpurCompilationSpec{
								Package: &types.Package{},
							},
							Path:  "bar",
							Files: []string{"boo"},
						},
						&artifact.PackageArtifact{
							Path:  "d",
							Files: []string{"baz"},
						},
					},
				},
			}

			matches := repos.SearchPackages("bo", FileSearch)
			Expect(len(matches)).To(Equal(1))
			Expect(matches[0].Artifact.Path).To(Equal("bar"))
		})

		It("Searches packages", func() {
			repo := &BhojpurSystemRepository{
				Index: compiler.ArtifactIndex{
					&artifact.PackageArtifact{
						Path: "foo",
						CompileSpec: &compilerspec.BhojpurCompilationSpec{
							Package: &types.Package{
								Name:     "foo",
								Category: "bar",
								Version:  "1.0",
							},
						},
					},
					&artifact.PackageArtifact{
						Path: "baz",
						CompileSpec: &compilerspec.BhojpurCompilationSpec{
							Package: &types.Package{
								Name:     "foo",
								Category: "baz",
								Version:  "1.0",
							},
						},
					},
				},
			}

			a, err := repo.SearchArtefact(&types.Package{
				Name:     "foo",
				Category: "baz",
				Version:  "1.0",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(a.Path).To(Equal("baz"))

			a, err = repo.SearchArtefact(&types.Package{
				Name:     "foo",
				Category: "bar",
				Version:  "1.0",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(a.Path).To(Equal("foo"))

			// Doesn't exist. so must fail
			_, err = repo.SearchArtefact(&types.Package{
				Name:     "foo",
				Category: "bar",
				Version:  "1.1",
			})
			Expect(err).To(HaveOccurred())
		})
	})
})
