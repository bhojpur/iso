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
	"io/ioutil"
	"os"
	"path/filepath"

	//	. "github.com/bhojpur/iso/pkg/manager/installer"
	"github.com/bhojpur/iso/pkg/manager/api/core/context"
	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	compiler "github.com/bhojpur/iso/pkg/manager/compiler"
	backend "github.com/bhojpur/iso/pkg/manager/compiler/backend"
	compression "github.com/bhojpur/iso/pkg/manager/compiler/types/compression"
	"github.com/bhojpur/iso/pkg/manager/compiler/types/options"
	compilerspec "github.com/bhojpur/iso/pkg/manager/compiler/types/spec"
	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"

	pkg "github.com/bhojpur/iso/pkg/manager/database"
	. "github.com/bhojpur/iso/pkg/manager/installer"
	"github.com/bhojpur/iso/pkg/manager/tree"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func stubRepo(tmpdir, tree string) (*BhojpurSystemRepository, error) {
	return GenerateRepository(
		WithName("test"),
		WithDescription("description"),
		WithType("disk"),
		WithUrls(tmpdir),
		WithPriority(1),
		WithSource(tmpdir),
		WithTree(tree),
		WithContext(context.NewContext()),
		WithDatabase(pkg.NewInMemoryDatabase(false)),
	)
}

var _ = Describe("Installer", func() {
	ctx := context.NewContext()

	BeforeEach(func() {
		ctx = context.NewContext()
	})

	Context("Writes a repository definition", func() {
		It("Writes a repo and can install packages from it", func() {
			//repo:=NewBhojpurSystemRepository()

			tmpdir, err := ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			generalRecipe := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe.Load("../../tests/fixtures/buildable")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe.GetDatabase().GetPackages())).To(Equal(3))

			c := compiler.NewBhojpurCompiler(backend.NewSimpleDockerBackend(ctx),
				generalRecipe.GetDatabase(),
				options.Concurrency(2))

			spec, err := c.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(spec.GetPackage().GetPath()).ToNot(Equal(""))

			tmpdir, err = ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			Expect(spec.BuildSteps()).To(Equal([]string{"echo artifact5 > /test5", "echo artifact6 > /test6", "chmod +x generate.sh", "./generate.sh"}))
			Expect(spec.GetPreBuildSteps()).To(Equal([]string{"echo foo > /test", "echo bar > /test2"}))

			spec.SetOutputPath(tmpdir)

			a, err := c.Compile(false, spec)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileHelper.Exists(a.Path)).To(BeTrue())
			Expect(a.Unpack(ctx, tmpdir, false)).ToNot(HaveOccurred())

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
			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).ToNot(BeTrue())
			err = repo.Write(ctx, tmpdir, false, false)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).To(BeTrue())
			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))

			fakeroot, err := ioutil.TempDir("", "fakeroot")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(fakeroot) // clean up

			repo2, err := NewBhojpurSystemRepositoryFromYaml([]byte(`
name: "test"
type: "disk"
enable: true
urls:
  - "`+tmpdir+`"
`), pkg.NewInMemoryDatabase(false))
			Expect(err).ToNot(HaveOccurred())
			inst := NewBhojpurInstaller(BhojpurInstallerOptions{
				Concurrency: 1, Context: ctx,
				PackageRepositories: types.BhojpurRepositories{*repo2.BhojpurRepository},
			})
			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))
			systemDB := pkg.NewInMemoryDatabase(false)
			system := &System{Database: systemDB, Target: fakeroot}
			err = inst.Install([]*types.Package{&types.Package{Name: "b", Category: "test", Version: "1.0"}}, system)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test5"))).To(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test6"))).To(BeTrue())
			_, err = systemDB.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			files, err := systemDB.GetPackageFiles(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(files).To(Equal([]string{"artifact42", "test5", "test6"}))
			Expect(err).ToNot(HaveOccurred())

			Expect(len(system.Database.GetPackages())).To(Equal(1))
			p, err := system.Database.GetPackage(system.Database.GetPackages()[0])
			Expect(err).ToNot(HaveOccurred())
			Expect(p.GetName()).To(Equal("b"))

			err = inst.Uninstall(system, &types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			// Nothing should be there anymore (files, packagedb entry)
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test5"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test6"))).ToNot(BeTrue())

			_, err = systemDB.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).To(HaveOccurred())
			_, err = systemDB.GetPackageFiles(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).To(HaveOccurred())
		})

	})

	Context("Writes a repository definition without compression", func() {
		It("Writes a repo and can install packages from it", func() {
			//repo:=NewBhojpurSystemRepository()

			tmpdir, err := ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			generalRecipe := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe.Load("../../tests/fixtures/buildable")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe.GetDatabase().GetPackages())).To(Equal(3))

			c := compiler.NewBhojpurCompiler(backend.NewSimpleDockerBackend(ctx),
				generalRecipe.GetDatabase(), options.Concurrency(2))

			spec, err := c.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(spec.GetPackage().GetPath()).ToNot(Equal(""))

			tmpdir, err = ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			Expect(spec.BuildSteps()).To(Equal([]string{"echo artifact5 > /test5", "echo artifact6 > /test6", "chmod +x generate.sh", "./generate.sh"}))
			Expect(spec.GetPreBuildSteps()).To(Equal([]string{"echo foo > /test", "echo bar > /test2"}))

			spec.SetOutputPath(tmpdir)

			artifact, err := c.Compile(false, spec)
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

			treeFile := NewDefaultTreeRepositoryFile()
			treeFile.SetCompressionType(compression.None)
			repo.SetRepositoryFile(REPOFILE_TREE_KEY, treeFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(repo.GetName()).To(Equal("test"))
			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).ToNot(BeTrue())
			err = repo.Write(ctx, tmpdir, false, false)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).To(BeTrue())
			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))

			fakeroot, err := ioutil.TempDir("", "fakeroot")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(fakeroot) // clean up

			repo2, err := NewBhojpurSystemRepositoryFromYaml([]byte(`
name: "test"
type: "disk"
enable: true
urls:
  - "`+tmpdir+`"
`), pkg.NewInMemoryDatabase(false))
			Expect(err).ToNot(HaveOccurred())

			inst := NewBhojpurInstaller(BhojpurInstallerOptions{
				Concurrency: 1, Context: ctx,
				PackageRepositories: types.BhojpurRepositories{*repo2.BhojpurRepository},
			})
			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))
			systemDB := pkg.NewInMemoryDatabase(false)
			system := &System{Database: systemDB, Target: fakeroot}
			err = inst.Install([]*types.Package{&types.Package{Name: "b", Category: "test", Version: "1.0"}}, system)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test5"))).To(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test6"))).To(BeTrue())
			_, err = systemDB.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			files, err := systemDB.GetPackageFiles(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(files).To(Equal([]string{"artifact42", "test5", "test6"}))
			Expect(err).ToNot(HaveOccurred())

			Expect(len(system.Database.GetPackages())).To(Equal(1))
			p, err := system.Database.GetPackage(system.Database.GetPackages()[0])
			Expect(err).ToNot(HaveOccurred())
			Expect(p.GetName()).To(Equal("b"))

			err = inst.Uninstall(system, &types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			// Nothing should be there anymore (files, packagedb entry)
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test5"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test6"))).ToNot(BeTrue())

			_, err = systemDB.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).To(HaveOccurred())
			_, err = systemDB.GetPackageFiles(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).To(HaveOccurred())
		})

	})

	Context("Installation", func() {
		It("Installs in a system with a persistent db", func() {
			//repo:=NewBhojpurSystemRepository()

			tmpdir, err := ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			generalRecipe := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe.Load("../../tests/fixtures/buildable")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe.GetDatabase().GetPackages())).To(Equal(3))

			c := compiler.NewBhojpurCompiler(backend.NewSimpleDockerBackend(ctx), generalRecipe.GetDatabase(),
				options.Concurrency(2))

			spec, err := c.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(spec.GetPackage().GetPath()).ToNot(Equal(""))

			tmpdir, err = ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			Expect(spec.BuildSteps()).To(Equal([]string{"echo artifact5 > /test5", "echo artifact6 > /test6", "chmod +x generate.sh", "./generate.sh"}))
			Expect(spec.GetPreBuildSteps()).To(Equal([]string{"echo foo > /test", "echo bar > /test2"}))

			spec.SetOutputPath(tmpdir)

			artifact, err := c.Compile(false, spec)
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

			repo, err := GenerateRepository(
				WithName("test"),
				WithDescription("description"),
				WithType("disk"),
				WithUrls(tmpdir),
				WithPriority(1),
				WithSource(tmpdir),
				WithTree("../../tests/fixtures/buildable"),
				WithDatabase(pkg.NewInMemoryDatabase(false)),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(repo.GetName()).To(Equal("test"))
			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).ToNot(BeTrue())
			err = repo.Write(ctx, tmpdir, false, false)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).To(BeTrue())
			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))

			fakeroot, err := ioutil.TempDir("", "fakeroot")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(fakeroot) // clean up

			repo2, err := NewBhojpurSystemRepositoryFromYaml([]byte(`
name: "test"
type: "disk"
enable: true
urls:
  - "`+tmpdir+`"
`), pkg.NewInMemoryDatabase(false))
			Expect(err).ToNot(HaveOccurred())
			inst := NewBhojpurInstaller(BhojpurInstallerOptions{
				Concurrency: 1, Context: ctx,
				PackageRepositories: types.BhojpurRepositories{*repo2.BhojpurRepository},
			})
			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))

			bolt, err := ioutil.TempDir("", "db")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(bolt) // clean up

			systemDB := pkg.NewBoltDatabase(filepath.Join(bolt, "db.db"))
			system := &System{Database: systemDB, Target: fakeroot}
			err = inst.Install([]*types.Package{&types.Package{Name: "b", Category: "test", Version: "1.0"}}, system)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test5"))).To(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test6"))).To(BeTrue())
			_, err = systemDB.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(len(system.Database.GetPackages())).To(Equal(1))
			p, err := system.Database.GetPackage(system.Database.GetPackages()[0])
			Expect(err).ToNot(HaveOccurred())
			Expect(p.GetName()).To(Equal("b"))

			files, err := systemDB.GetPackageFiles(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(files).To(Equal([]string{"artifact42", "test5", "test6"}))
			Expect(err).ToNot(HaveOccurred())

			err = inst.Uninstall(system, &types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			// Nothing should be there anymore (files, packagedb entry)
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test5"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test6"))).ToNot(BeTrue())

			_, err = system.Database.GetPackageFiles(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).To(HaveOccurred())
			_, err = system.Database.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).To(HaveOccurred())

		})

		It("Installs new packages from a syste with others installed", func() {
			//repo:=NewBhojpurSystemRepository()

			tmpdir, err := ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			generalRecipe := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe.Load("../../tests/fixtures/buildable")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe.GetDatabase().GetPackages())).To(Equal(3))

			c := compiler.NewBhojpurCompiler(backend.NewSimpleDockerBackend(ctx), generalRecipe.GetDatabase(),
				options.Concurrency(2))

			spec, err := c.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(spec.GetPackage().GetPath()).ToNot(Equal(""))

			tmpdir, err = ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			Expect(spec.BuildSteps()).To(Equal([]string{"echo artifact5 > /test5", "echo artifact6 > /test6", "chmod +x generate.sh", "./generate.sh"}))
			Expect(spec.GetPreBuildSteps()).To(Equal([]string{"echo foo > /test", "echo bar > /test2"}))

			spec.SetOutputPath(tmpdir)

			artifact, err := c.Compile(false, spec)
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

			repo, err := GenerateRepository(
				WithName("test"),
				WithDescription("description"),
				WithType("disk"),
				WithUrls(tmpdir),
				WithPriority(1),
				WithSource(tmpdir),
				WithTree("../../tests/fixtures/buildable"),
				WithDatabase(pkg.NewInMemoryDatabase(false)),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(repo.GetName()).To(Equal("test"))
			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).ToNot(BeTrue())
			err = repo.Write(ctx, tmpdir, false, false)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).To(BeTrue())
			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))

			fakeroot, err := ioutil.TempDir("", "fakeroot")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(fakeroot) // clean up

			repo2, err := NewBhojpurSystemRepositoryFromYaml([]byte(`
name: "test"
type: "disk"
enable: true
urls:
  - "`+tmpdir+`"
`), pkg.NewInMemoryDatabase(false))
			Expect(err).ToNot(HaveOccurred())

			inst := NewBhojpurInstaller(BhojpurInstallerOptions{
				Concurrency: 1, Context: ctx,
				PackageRepositories: types.BhojpurRepositories{*repo2.BhojpurRepository},
			})
			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))

			bolt, err := ioutil.TempDir("", "db")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(bolt) // clean up

			systemDB := pkg.NewBoltDatabase(filepath.Join(bolt, "db.db"))
			system := &System{Database: systemDB, Target: fakeroot}
			err = inst.Install([]*types.Package{&types.Package{Name: "b", Category: "test", Version: "1.0"}}, system)
			Expect(err).ToNot(HaveOccurred())

			tmpdir2, err := ioutil.TempDir("", "tree2")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			generalRecipe2 := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe2.Load("../../tests/fixtures/alpine")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe2.GetDatabase().GetPackages())).To(Equal(1))

			c = compiler.NewBhojpurCompiler(backend.NewSimpleDockerBackend(ctx), generalRecipe2.GetDatabase(), options.Concurrency(2))

			spec, err = c.FromPackage(&types.Package{Name: "alpine", Category: "seed", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(spec.GetPackage().GetPath()).ToNot(Equal(""))

			spec.SetOutputPath(tmpdir2)

			artifact, err = c.Compile(false, spec)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileHelper.Exists(artifact.Path)).To(BeTrue())

			repo, err = stubRepo(tmpdir2, "../../tests/fixtures/alpine")
			Expect(err).ToNot(HaveOccurred())
			err = repo.Write(ctx, tmpdir2, false, false)
			Expect(err).ToNot(HaveOccurred())

			fakeroot, err = ioutil.TempDir("", "fakeroot")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(fakeroot) // clean up

			repo2, err = NewBhojpurSystemRepositoryFromYaml([]byte(`
name: "test"
type: "disk"
enable: true
urls:
  - "`+tmpdir2+`"
`), pkg.NewInMemoryDatabase(false))
			Expect(err).ToNot(HaveOccurred())
			inst = NewBhojpurInstaller(BhojpurInstallerOptions{
				Concurrency: 1, Context: ctx,
				PackageRepositories: types.BhojpurRepositories{*repo2.BhojpurRepository},
			})
			Expect(repo.GetUrls()[0]).To(Equal(tmpdir2))
			Expect(repo.GetType()).To(Equal("disk"))
			system.Target = fakeroot
			err = inst.Install([]*types.Package{&types.Package{Name: "alpine", Category: "seed", Version: "1.0"}}, system)
			Expect(err).ToNot(HaveOccurred())
			_, err = system.Database.FindPackage(&types.Package{Name: "alpine", Category: "seed", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

		})

	})

	Context("Simple upgrades", func() {
		It("Installs packages and Upgrades a system with a persistent db", func() {
			//repo:=NewBhojpurSystemRepository()

			tmpdir, err := ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			generalRecipe := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe.Load("../../tests/fixtures/upgrade")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe.GetDatabase().GetPackages())).To(Equal(4))

			c := compiler.NewBhojpurCompiler(backend.NewSimpleDockerBackend(ctx), generalRecipe.GetDatabase(), options.Concurrency(2))

			spec, err := c.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())
			spec2, err := c.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.1"})
			Expect(err).ToNot(HaveOccurred())
			spec3, err := c.FromPackage(&types.Package{Name: "c", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(spec.GetPackage().GetPath()).ToNot(Equal(""))

			tmpdir, err = ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			spec.SetOutputPath(tmpdir)
			spec2.SetOutputPath(tmpdir)
			spec3.SetOutputPath(tmpdir)

			_, errs := c.CompileParallel(false, compilerspec.NewBhojpurCompilationspecs(spec, spec2, spec3))

			Expect(errs).To(BeEmpty())

			repo, err := stubRepo(tmpdir, "../../tests/fixtures/upgrade")
			Expect(err).ToNot(HaveOccurred())
			Expect(repo.GetName()).To(Equal("test"))
			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).ToNot(BeTrue())
			err = repo.Write(ctx, tmpdir, false, false)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).To(BeTrue())
			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))

			fakeroot, err := ioutil.TempDir("", "fakeroot")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(fakeroot) // clean up

			repo2, err := NewBhojpurSystemRepositoryFromYaml([]byte(`
name: "test"
type: "disk"
enable: true
urls:
  - "`+tmpdir+`"
`), pkg.NewInMemoryDatabase(false))
			Expect(err).ToNot(HaveOccurred())
			inst := NewBhojpurInstaller(BhojpurInstallerOptions{
				Concurrency: 1, Context: ctx,
				Relaxed:             true,
				PackageRepositories: types.BhojpurRepositories{*repo2.BhojpurRepository},
			})
			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))

			bolt, err := ioutil.TempDir("", "db")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(bolt) // clean up

			systemDB := pkg.NewBoltDatabase(filepath.Join(bolt, "db.db"))
			system := &System{Database: systemDB, Target: fakeroot}
			err = inst.Install([]*types.Package{&types.Package{Name: "b", Category: "test", Version: "1.0"}}, system)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test5"))).To(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test6"))).To(BeTrue())
			_, err = systemDB.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(len(system.Database.GetPackages())).To(Equal(1))
			p, err := system.Database.GetPackage(system.Database.GetPackages()[0])
			Expect(err).ToNot(HaveOccurred())
			Expect(p.GetName()).To(Equal("b"))

			files, err := systemDB.GetPackageFiles(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(files).To(Equal([]string{"artifact42", "test5", "test6"}))
			Expect(err).ToNot(HaveOccurred())

			err = inst.Upgrade(system)
			Expect(err).ToNot(HaveOccurred())

			// Nothing should be there anymore (files, packagedb entry)
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test5"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test6"))).ToNot(BeTrue())

			// New version - new files
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "newc"))).To(BeTrue())
			_, err = system.Database.GetPackageFiles(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).To(HaveOccurred())
			_, err = system.Database.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).To(HaveOccurred())

			// New package should be there
			_, err = system.Database.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.1"})
			Expect(err).ToNot(HaveOccurred())

		})

		It("Compute the correct upgrade order", func() {
			tmpdir, err := ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			generalRecipe := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe.Load("../../tests/fixtures/upgrade_complex")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe.GetDatabase().GetPackages())).To(Equal(4))

			c := compiler.NewBhojpurCompiler(backend.NewSimpleDockerBackend(ctx), generalRecipe.GetDatabase(), options.Concurrency(2))

			spec, err := c.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())
			spec2, err := c.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.1"})
			Expect(err).ToNot(HaveOccurred())

			spec4, err := c.FromPackage(&types.Package{Name: "a", Category: "test", Version: "1.1"})
			Expect(err).ToNot(HaveOccurred())
			spec5, err := c.FromPackage(&types.Package{Name: "a", Category: "test", Version: "1.2"})
			Expect(err).ToNot(HaveOccurred())
			Expect(spec.GetPackage().GetPath()).ToNot(Equal(""))

			tmpdir, err = ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			spec.SetOutputPath(tmpdir)
			spec2.SetOutputPath(tmpdir)
			spec4.SetOutputPath(tmpdir)
			spec5.SetOutputPath(tmpdir)

			_, errs := c.CompileParallel(false, compilerspec.NewBhojpurCompilationspecs(spec, spec2, spec4, spec5))

			Expect(errs).To(BeEmpty())

			repo, err := stubRepo(tmpdir, "../../tests/fixtures/upgrade_complex")
			Expect(err).ToNot(HaveOccurred())
			Expect(repo.GetName()).To(Equal("test"))
			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).ToNot(BeTrue())
			err = repo.Write(ctx, tmpdir, false, false)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).To(BeTrue())
			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))

			fakeroot, err := ioutil.TempDir("", "fakeroot")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(fakeroot) // clean up

			repo2, err := NewBhojpurSystemRepositoryFromYaml([]byte(`
name: "test"
type: "disk"
enable: true
urls:
  - "`+tmpdir+`"
`), pkg.NewInMemoryDatabase(false))
			Expect(err).ToNot(HaveOccurred())
			inst := NewBhojpurInstaller(BhojpurInstallerOptions{
				Concurrency: 1, Context: ctx,
				Relaxed:             true,
				PackageRepositories: types.BhojpurRepositories{*repo2.BhojpurRepository},
			})
			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))

			bolt, err := ioutil.TempDir("", "db")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(bolt) // clean up

			systemDB := pkg.NewBoltDatabase(filepath.Join(bolt, "db.db"))
			system := &System{Database: systemDB, Target: fakeroot}

			err = inst.Install([]*types.Package{&types.Package{Name: "b", Category: "test", Version: "1.0"}}, system)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test5"))).To(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test6"))).To(BeTrue())
			_, err = systemDB.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(len(system.Database.GetPackages())).To(Equal(1))
			p, err := system.Database.GetPackage(system.Database.GetPackages()[0])
			Expect(err).ToNot(HaveOccurred())
			Expect(p.GetName()).To(Equal("b"))

			files, err := systemDB.GetPackageFiles(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(files).To(Equal([]string{"test5", "test6"}))
			Expect(err).ToNot(HaveOccurred())

			err = inst.Install([]*types.Package{&types.Package{Name: "a", Category: "test", Version: "1.1"}}, system)
			Expect(err).ToNot(HaveOccurred())

			files, err = systemDB.GetPackageFiles(&types.Package{Name: "a", Category: "test", Version: "1.1"})
			Expect(files).To(Equal([]string{"test3", "test4"}))
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test3"))).To(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test4"))).To(BeTrue())

			err = inst.Upgrade(system)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test3"))).To(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test4"))).To(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test5"))).To(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test6"))).To(BeTrue())

		})

		It("Compute the correct upgrade order with a package replacing multiple ones", func() {
			tmpdir, err := ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			generalRecipe := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe.Load("../../tests/fixtures/upgrade_complex_multiple")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe.GetDatabase().GetPackages())).To(Equal(6))

			c := compiler.NewBhojpurCompiler(backend.NewSimpleDockerBackend(ctx), generalRecipe.GetDatabase(), options.Concurrency(2))

			spec, err := c.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())
			spec2, err := c.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.1"})
			Expect(err).ToNot(HaveOccurred())
			spec3, err := c.FromPackage(&types.Package{Name: "c", Category: "test", Version: "1.1"})
			Expect(err).ToNot(HaveOccurred())

			spec4, err := c.FromPackage(&types.Package{Name: "a", Category: "test", Version: "1.1"})
			Expect(err).ToNot(HaveOccurred())
			spec5, err := c.FromPackage(&types.Package{Name: "a", Category: "test", Version: "1.2"})
			Expect(err).ToNot(HaveOccurred())
			spec6, err := c.FromPackage(&types.Package{Name: "c", Category: "test", Version: "1.2"})
			Expect(err).ToNot(HaveOccurred())

			Expect(spec.GetPackage().GetPath()).ToNot(Equal(""))

			tmpdir, err = ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			spec.SetOutputPath(tmpdir)
			spec2.SetOutputPath(tmpdir)
			spec4.SetOutputPath(tmpdir)
			spec5.SetOutputPath(tmpdir)
			spec3.SetOutputPath(tmpdir)
			spec6.SetOutputPath(tmpdir)

			_, errs := c.CompileParallel(false, compilerspec.NewBhojpurCompilationspecs(spec, spec2, spec3, spec4, spec5, spec6))

			Expect(errs).To(BeEmpty())

			repo, err := stubRepo(tmpdir, "../../tests/fixtures/upgrade_complex_multiple")
			Expect(err).ToNot(HaveOccurred())
			Expect(repo.GetName()).To(Equal("test"))
			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).ToNot(BeTrue())
			err = repo.Write(ctx, tmpdir, false, false)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).To(BeTrue())
			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))

			fakeroot, err := ioutil.TempDir("", "fakeroot")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(fakeroot) // clean up

			repo2, err := NewBhojpurSystemRepositoryFromYaml([]byte(`
name: "test"
type: "disk"
enable: true
urls:
  - "`+tmpdir+`"
`), pkg.NewInMemoryDatabase(false))
			Expect(err).ToNot(HaveOccurred())
			inst := NewBhojpurInstaller(BhojpurInstallerOptions{
				Concurrency: 1, Context: ctx,
				Relaxed:             true,
				PackageRepositories: types.BhojpurRepositories{*repo2.BhojpurRepository},
			})
			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))

			bolt, err := ioutil.TempDir("", "db")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(bolt) // clean up

			systemDB := pkg.NewBoltDatabase(filepath.Join(bolt, "db.db"))
			system := &System{Database: systemDB, Target: fakeroot}

			err = inst.Install([]*types.Package{&types.Package{Name: "b", Category: "test", Version: "1.0"}}, system)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test5"))).To(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test6"))).To(BeTrue())
			_, err = systemDB.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(len(system.Database.GetPackages())).To(Equal(1))
			p, err := system.Database.GetPackage(system.Database.GetPackages()[0])
			Expect(err).ToNot(HaveOccurred())
			Expect(p.GetName()).To(Equal("b"))

			files, err := systemDB.GetPackageFiles(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(files).To(Equal([]string{"test5", "test6"}))
			Expect(err).ToNot(HaveOccurred())

			err = inst.Install([]*types.Package{&types.Package{Name: "c", Category: "test", Version: "1.1"}}, system)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test1"))).To(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test2"))).To(BeTrue())

			err = inst.Install([]*types.Package{&types.Package{Name: "a", Category: "test", Version: "1.1"}}, system)
			Expect(err).ToNot(HaveOccurred())

			files, err = systemDB.GetPackageFiles(&types.Package{Name: "a", Category: "test", Version: "1.1"})
			Expect(files).To(Equal([]string{"test3", "test4"}))
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test3"))).To(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test4"))).To(BeTrue())

			err = inst.Upgrade(system)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test1"))).To(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test2"))).To(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test3"))).To(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test4"))).To(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test5"))).To(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test6"))).To(BeTrue())

		})

		It("Handles package drops", func() {
			//repo:=NewBhojpurSystemRepository()

			tmpdir, err := ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			generalRecipe := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))
			generalRecipeNewRepo := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe.Load("../../tests/fixtures/upgrade_old_repo")
			Expect(err).ToNot(HaveOccurred())

			err = generalRecipeNewRepo.Load("../../tests/fixtures/upgrade_new_repo")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe.GetDatabase().GetPackages())).To(Equal(3))
			Expect(len(generalRecipeNewRepo.GetDatabase().GetPackages())).To(Equal(3))

			c := compiler.NewBhojpurCompiler(backend.NewSimpleDockerBackend(ctx), generalRecipe.GetDatabase(), options.Concurrency(2))
			c2 := compiler.NewBhojpurCompiler(backend.NewSimpleDockerBackend(ctx), generalRecipeNewRepo.GetDatabase())

			spec, err := c.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())
			spec3, err := c.FromPackage(&types.Package{Name: "c", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			spec2, err := c2.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.1"})
			Expect(err).ToNot(HaveOccurred())

			Expect(spec.GetPackage().GetPath()).ToNot(Equal(""))

			tmpdir, err = ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up
			tmpdirnewrepo, err := ioutil.TempDir("", "tree2")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdirnewrepo) // clean up

			spec.SetOutputPath(tmpdir)
			spec2.SetOutputPath(tmpdirnewrepo)
			spec3.SetOutputPath(tmpdir)

			_, errs := c.CompileParallel(false, compilerspec.NewBhojpurCompilationspecs(spec, spec3))

			Expect(errs).To(BeEmpty())

			_, errs = c2.CompileParallel(false, compilerspec.NewBhojpurCompilationspecs(spec2))
			Expect(errs).To(BeEmpty())

			repo, err := stubRepo(tmpdir, "../../tests/fixtures/upgrade_old_repo")
			Expect(err).ToNot(HaveOccurred())
			Expect(repo.GetName()).To(Equal("test"))
			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).ToNot(BeTrue())
			err = repo.Write(ctx, tmpdir, false, false)
			Expect(err).ToNot(HaveOccurred())

			repoupgrade, err := stubRepo(tmpdirnewrepo, "../../tests/fixtures/upgrade_new_repo")
			Expect(err).ToNot(HaveOccurred())
			err = repoupgrade.Write(ctx, tmpdirnewrepo, false, false)
			Expect(err).ToNot(HaveOccurred())

			fakeroot, err := ioutil.TempDir("", "fakeroot")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(fakeroot) // clean up

			repo2, err := NewBhojpurSystemRepositoryFromYaml([]byte(`
name: "test"
type: "disk"
enable: true
urls:
  - "`+tmpdir+`"
`), pkg.NewInMemoryDatabase(false))
			Expect(err).ToNot(HaveOccurred())

			repoupgrade2, err := NewBhojpurSystemRepositoryFromYaml([]byte(`
name: "test"
type: "disk"
enable: true
urls:
  - "`+tmpdirnewrepo+`"
`), pkg.NewInMemoryDatabase(false))
			Expect(err).ToNot(HaveOccurred())
			inst := NewBhojpurInstaller(BhojpurInstallerOptions{
				Concurrency: 1, Context: ctx,
				PackageRepositories: types.BhojpurRepositories{*repo2.BhojpurRepository},
			})
			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))

			bolt, err := ioutil.TempDir("", "db")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(bolt) // clean up

			systemDB := pkg.NewBoltDatabase(filepath.Join(bolt, "db.db"))
			system := &System{Database: systemDB, Target: fakeroot}
			err = inst.Install([]*types.Package{&types.Package{Name: "b", Category: "test", Version: "1.0"}}, system)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test5"))).To(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test6"))).To(BeTrue())
			_, err = systemDB.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(len(system.Database.GetPackages())).To(Equal(1))
			p, err := system.Database.GetPackage(system.Database.GetPackages()[0])
			Expect(err).ToNot(HaveOccurred())
			Expect(p.GetName()).To(Equal("b"))

			files, err := systemDB.GetPackageFiles(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(files).To(Equal([]string{"artifact42", "test5", "test6"}))
			Expect(err).ToNot(HaveOccurred())

			inst.Options.PackageRepositories = types.BhojpurRepositories{*repoupgrade2.BhojpurRepository}

			err = inst.Upgrade(system)
			Expect(err).ToNot(HaveOccurred())

			// Nothing should be there anymore (files, packagedb entry)
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test5"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test6"))).ToNot(BeTrue())

			// New version - new files
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "newc"))).To(BeTrue())
			_, err = system.Database.GetPackageFiles(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).To(HaveOccurred())
			_, err = system.Database.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).To(HaveOccurred())

			// New package should be there
			_, err = system.Database.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.1"})
			Expect(err).ToNot(HaveOccurred())

		})

	})

	Context("Compressed packages", func() {
		It("Installs", func() {
			//repo:=NewBhojpurSystemRepository()

			tmpdir, err := ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			generalRecipe := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe.Load("../../tests/fixtures/upgrade")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe.GetDatabase().GetPackages())).To(Equal(4))

			c := compiler.NewBhojpurCompiler(
				backend.NewSimpleDockerBackend(ctx),
				generalRecipe.GetDatabase(),
				options.Concurrency(2),
				options.WithCompressionType(compression.GZip),
			)

			spec, err := c.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())
			spec2, err := c.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.1"})
			Expect(err).ToNot(HaveOccurred())
			spec3, err := c.FromPackage(&types.Package{Name: "c", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(spec.GetPackage().GetPath()).ToNot(Equal(""))

			tmpdir, err = ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up
			spec.SetOutputPath(tmpdir)
			spec2.SetOutputPath(tmpdir)
			spec3.SetOutputPath(tmpdir)

			_, errs := c.CompileParallel(false, compilerspec.NewBhojpurCompilationspecs(spec, spec2, spec3))

			Expect(errs).To(BeEmpty())

			repo, err := stubRepo(tmpdir, "../../tests/fixtures/upgrade")
			Expect(err).ToNot(HaveOccurred())
			Expect(repo.GetName()).To(Equal("test"))
			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).ToNot(BeTrue())
			err = repo.Write(ctx, tmpdir, false, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileHelper.Exists(spec.Rel("b-test-1.1.package.tar.gz"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel("b-test-1.1.package.tar"))).ToNot(BeTrue())

			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).To(BeTrue())
			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))

			fakeroot, err := ioutil.TempDir("", "fakeroot")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(fakeroot) // clean up

			repo2, err := NewBhojpurSystemRepositoryFromYaml([]byte(`
name: "test"
type: "disk"
enable: true
urls:
  - "`+tmpdir+`"
`), pkg.NewInMemoryDatabase(false))
			Expect(err).ToNot(HaveOccurred())

			inst := NewBhojpurInstaller(BhojpurInstallerOptions{
				Concurrency: 1, Context: ctx,
				PackageRepositories: types.BhojpurRepositories{*repo2.BhojpurRepository},
			})

			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))

			bolt, err := ioutil.TempDir("", "db")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(bolt) // clean up

			systemDB := pkg.NewBoltDatabase(filepath.Join(bolt, "db.db"))
			system := &System{Database: systemDB, Target: fakeroot}
			err = inst.Install([]*types.Package{&types.Package{Name: "b", Category: "test", Version: "1.0"}}, system)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test5"))).To(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test6"))).To(BeTrue())
			_, err = systemDB.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(len(system.Database.GetPackages())).To(Equal(1))
			p, err := system.Database.GetPackage(system.Database.GetPackages()[0])
			Expect(err).ToNot(HaveOccurred())
			Expect(p.GetName()).To(Equal("b"))

			files, err := systemDB.GetPackageFiles(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(files).To(Equal([]string{"artifact42", "test5", "test6"}))
			Expect(err).ToNot(HaveOccurred())

			err = inst.Upgrade(system)
			Expect(err).ToNot(HaveOccurred())

			// Nothing should be there anymore (files, packagedb entry)
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test5"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test6"))).ToNot(BeTrue())

			// New version - new files
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "newc"))).To(BeTrue())
			_, err = system.Database.GetPackageFiles(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).To(HaveOccurred())
			_, err = system.Database.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).To(HaveOccurred())

			// New package should be there
			_, err = system.Database.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.1"})
			Expect(err).ToNot(HaveOccurred())

		})

	})

	Context("Uninstallation", func() {
		It("fails if package is required by others which are installed", func() {

			fakeroot, err := ioutil.TempDir("", "fakeroot")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(fakeroot) // clean up
			bolt, err := ioutil.TempDir("", "db")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(bolt) // clean up

			systemDB := pkg.NewBoltDatabase(filepath.Join(bolt, "db.db"))
			system := &System{Database: systemDB, Target: fakeroot}

			inst := NewBhojpurInstaller(BhojpurInstallerOptions{Concurrency: 1, Context: ctx, CheckConflicts: true})

			D := types.NewPackage("D", "", []*types.Package{}, []*types.Package{})
			B := types.NewPackage("calamares", "", []*types.Package{D}, []*types.Package{})
			C := types.NewPackage("kpmcore", "", []*types.Package{B}, []*types.Package{})
			A := types.NewPackage("A", "", []*types.Package{B}, []*types.Package{})
			Z := types.NewPackage("chromium", "", []*types.Package{A}, []*types.Package{})
			F := types.NewPackage("F", "", []*types.Package{Z, B}, []*types.Package{})

			Z.SetVersion("86.0.4240.193+2")
			Z.SetCategory("www-client")
			B.SetVersion("3.2.32.1+5")
			B.SetCategory("app-admin")
			C.SetVersion("4.2.0+2")
			C.SetCategory("sys-libs-5")
			D.SetVersion("5.19.5+9")
			D.SetCategory("layers")

			for _, p := range []*types.Package{A, B, C, D, Z, F} {
				_, err := systemDB.CreatePackage(p)
				Expect(err).ToNot(HaveOccurred())
			}

			err = inst.Uninstall(system, D)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Existing files", func() {
		It("Reclaims them", func() {
			//repo:=NewBhojpurSystemRepository()

			tmpdir, err := ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			generalRecipe := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe.Load("../../tests/fixtures/upgrade")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe.GetDatabase().GetPackages())).To(Equal(4))

			c := compiler.NewBhojpurCompiler(backend.NewSimpleDockerBackend(ctx), generalRecipe.GetDatabase(),
				options.Concurrency(2),
				options.WithCompressionType(compression.GZip))

			spec, err := c.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())
			spec2, err := c.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.1"})
			Expect(err).ToNot(HaveOccurred())
			spec3, err := c.FromPackage(&types.Package{Name: "c", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(spec.GetPackage().GetPath()).ToNot(Equal(""))

			tmpdir, err = ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up
			spec.SetOutputPath(tmpdir)
			spec2.SetOutputPath(tmpdir)
			spec3.SetOutputPath(tmpdir)
			_, errs := c.CompileParallel(false, compilerspec.NewBhojpurCompilationspecs(spec, spec2, spec3))

			Expect(errs).To(BeEmpty())

			repo, err := stubRepo(tmpdir, "../../tests/fixtures/upgrade")
			Expect(err).ToNot(HaveOccurred())
			Expect(repo.GetName()).To(Equal("test"))
			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).ToNot(BeTrue())
			err = repo.Write(ctx, tmpdir, false, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileHelper.Exists(spec.Rel("b-test-1.1.package.tar.gz"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel("b-test-1.1.package.tar"))).ToNot(BeTrue())

			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).To(BeTrue())
			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))

			fakeroot, err := ioutil.TempDir("", "fakeroot")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(fakeroot) // clean up

			repo2, err := NewBhojpurSystemRepositoryFromYaml([]byte(`
name: "test"
type: "disk"
enable: true
urls:
  - "`+tmpdir+`"
`), pkg.NewInMemoryDatabase(false))
			Expect(err).ToNot(HaveOccurred())

			inst := NewBhojpurInstaller(BhojpurInstallerOptions{
				Concurrency: 1, Context: ctx,
				PackageRepositories: types.BhojpurRepositories{*repo2.BhojpurRepository},
			})

			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))

			bolt, err := ioutil.TempDir("", "db")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(bolt) // clean up

			systemDB := pkg.NewBoltDatabase(filepath.Join(bolt, "db.db"))
			system := &System{Database: systemDB, Target: fakeroot}

			_, err = system.Database.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).To(HaveOccurred())

			_, err = system.Database.FindPackage(&types.Package{Name: "c", Category: "test", Version: "1.0"})
			Expect(err).To(HaveOccurred())

			Expect(len(system.Database.World())).To(Equal(0))
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test5"))).To(BeFalse())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test6"))).To(BeFalse())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "c"))).To(BeFalse())

			Expect(fileHelper.Touch(filepath.Join(fakeroot, "test5"))).ToNot(HaveOccurred())
			Expect(fileHelper.Touch(filepath.Join(fakeroot, "test6"))).ToNot(HaveOccurred())
			Expect(fileHelper.Touch(filepath.Join(fakeroot, "c"))).ToNot(HaveOccurred())

			err = inst.Reclaim(system)
			Expect(err).ToNot(HaveOccurred())

			_, err = system.Database.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			_, err = system.Database.FindPackage(&types.Package{Name: "c", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(len(system.Database.World())).To(Equal(2))
		})

		It("Upgrades reclaimed packages", func() {
			//repo:=NewBhojpurSystemRepository()

			tmpdir, err := ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			generalRecipe := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe.Load("../../tests/fixtures/upgrade_old_repo")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe.GetDatabase().GetPackages())).To(Equal(3))

			c := compiler.NewBhojpurCompiler(backend.NewSimpleDockerBackend(ctx), generalRecipe.GetDatabase(),
				options.WithCompressionType(compression.GZip))

			spec, err := c.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())
			spec3, err := c.FromPackage(&types.Package{Name: "c", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(spec.GetPackage().GetPath()).ToNot(Equal(""))

			tmpdir, err = ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up
			spec.SetOutputPath(tmpdir)
			spec3.SetOutputPath(tmpdir)
			_, errs := c.CompileParallel(false, compilerspec.NewBhojpurCompilationspecs(spec, spec3))

			Expect(errs).To(BeEmpty())

			repo, err := stubRepo(tmpdir, "../../tests/fixtures/upgrade_old_repo")
			Expect(err).ToNot(HaveOccurred())
			Expect(repo.GetName()).To(Equal("test"))
			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).ToNot(BeTrue())
			err = repo.Write(ctx, tmpdir, false, false)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(spec.Rel("repository.yaml"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(TREE_TARBALL + ".gz"))).To(BeTrue())
			Expect(fileHelper.Exists(spec.Rel(REPOSITORY_METAFILE + ".tar"))).To(BeTrue())
			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))

			fakeroot, err := ioutil.TempDir("", "fakeroot")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(fakeroot) // clean up

			repo2, err := NewBhojpurSystemRepositoryFromYaml([]byte(`
name: "test"
type: "disk"
enable: true
urls:
  - "`+tmpdir+`"
`), pkg.NewInMemoryDatabase(false))
			Expect(err).ToNot(HaveOccurred())

			inst := NewBhojpurInstaller(BhojpurInstallerOptions{
				Concurrency: 1, Context: ctx,
				PackageRepositories: types.BhojpurRepositories{*repo2.BhojpurRepository},
			})

			Expect(repo.GetUrls()[0]).To(Equal(tmpdir))
			Expect(repo.GetType()).To(Equal("disk"))

			bolt, err := ioutil.TempDir("", "db")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(bolt) // clean up

			systemDB := pkg.NewBoltDatabase(filepath.Join(bolt, "db.db"))
			system := &System{Database: systemDB, Target: fakeroot}

			_, err = system.Database.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).To(HaveOccurred())

			_, err = system.Database.FindPackage(&types.Package{Name: "c", Category: "test", Version: "1.0"})
			Expect(err).To(HaveOccurred())

			Expect(len(system.Database.World())).To(Equal(0))
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test5"))).To(BeFalse())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test6"))).To(BeFalse())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "c"))).To(BeFalse())

			Expect(fileHelper.Touch(filepath.Join(fakeroot, "test5"))).ToNot(HaveOccurred())
			Expect(fileHelper.Touch(filepath.Join(fakeroot, "test6"))).ToNot(HaveOccurred())
			Expect(fileHelper.Touch(filepath.Join(fakeroot, "c"))).ToNot(HaveOccurred())

			err = inst.Reclaim(system)
			Expect(err).ToNot(HaveOccurred())

			_, err = system.Database.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			_, err = system.Database.FindPackage(&types.Package{Name: "c", Category: "test", Version: "1.0"})
			Expect(err).ToNot(HaveOccurred())

			Expect(len(system.Database.World())).To(Equal(2))

			generalRecipe2 := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			err = generalRecipe2.Load("../../tests/fixtures/upgrade_new_repo")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(generalRecipe2.GetDatabase().GetPackages())).To(Equal(3))

			c = compiler.NewBhojpurCompiler(backend.NewSimpleDockerBackend(ctx), generalRecipe2.GetDatabase())

			spec, err = c.FromPackage(&types.Package{Name: "b", Category: "test", Version: "1.1"})
			Expect(err).ToNot(HaveOccurred())

			tmpdir2, err := ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir2) // clean up
			spec.SetOutputPath(tmpdir2)

			_, errs = c.CompileParallel(false, compilerspec.NewBhojpurCompilationspecs(spec))

			Expect(errs).To(BeEmpty())

			repo, err = stubRepo(tmpdir2, "../../tests/fixtures/upgrade_new_repo")
			Expect(err).ToNot(HaveOccurred())
			Expect(repo.GetName()).To(Equal("test"))
			err = repo.Write(ctx, tmpdir2, false, false)
			Expect(err).ToNot(HaveOccurred())

			repo2, err = NewBhojpurSystemRepositoryFromYaml([]byte(`
name: "test"
type: "disk"
enable: true
urls:
  - "`+tmpdir2+`"
`), pkg.NewInMemoryDatabase(false))
			Expect(err).ToNot(HaveOccurred())

			inst = NewBhojpurInstaller(BhojpurInstallerOptions{
				Concurrency: 1, Context: ctx,
				PackageRepositories: types.BhojpurRepositories{*repo2.BhojpurRepository},
			})

			err = inst.Upgrade(system)
			Expect(err).ToNot(HaveOccurred())

			// Nothing should be there anymore (files, packagedb entry)
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test5"))).ToNot(BeTrue())
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "test6"))).ToNot(BeTrue())

			// New version - new files
			Expect(fileHelper.Exists(filepath.Join(fakeroot, "newc"))).To(BeTrue())
			_, err = system.Database.GetPackageFiles(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).To(HaveOccurred())
			_, err = system.Database.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.0"})
			Expect(err).To(HaveOccurred())

			// New package should be there
			_, err = system.Database.FindPackage(&types.Package{Name: "b", Category: "test", Version: "1.1"})
			Expect(err).ToNot(HaveOccurred())

		})
	})

})
