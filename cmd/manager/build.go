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
	"fmt"
	"os"
	"path/filepath"

	helpers "github.com/bhojpur/iso/cmd/manager/helpers"
	"github.com/bhojpur/iso/cmd/manager/util"
	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	"github.com/bhojpur/iso/pkg/manager/api/core/types/artifact"
	"github.com/bhojpur/iso/pkg/manager/compiler"
	compilerspec "github.com/bhojpur/iso/pkg/manager/compiler/types/spec"
	"github.com/bhojpur/iso/pkg/manager/installer"
	"github.com/ghodss/yaml"

	"github.com/bhojpur/iso/pkg/manager/compiler/types/compression"
	"github.com/bhojpur/iso/pkg/manager/compiler/types/options"
	pkg "github.com/bhojpur/iso/pkg/manager/database"
	fileHelpers "github.com/bhojpur/iso/pkg/manager/helpers/file"
	tree "github.com/bhojpur/iso/pkg/manager/tree"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var buildCmd = &cobra.Command{
	// Skip processing output
	Annotations: map[string]string{
		util.CommandProcessOutput: "",
	},
	Use:   "build <package name> <package name> <package name> ...",
	Short: "build a package or a tree",
	Long: `Builds one or more packages from a tree (current directory is implied):

	$ isomgr build utils/busybox utils/yq ...

Builds all packages

	$ isomgr build --all

Builds only the leaf packages:

	$ isomgr build --full

Build package revdeps:

	$ isomgr build --revdeps utils/yq

Build package without dependencies (needs the images already in the host, or either need to be available online):

	$ isomgr build --nodeps utils/yq ...

Build packages specifying multiple definition trees:

	$ isomgr build --tree overlay/path --tree overlay/path2 utils/yq ...
`, PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("tree", cmd.Flags().Lookup("tree"))
		viper.BindPFlag("destination", cmd.Flags().Lookup("destination"))
		viper.BindPFlag("backend", cmd.Flags().Lookup("backend"))
		viper.BindPFlag("privileged", cmd.Flags().Lookup("privileged"))
		viper.BindPFlag("revdeps", cmd.Flags().Lookup("revdeps"))
		viper.BindPFlag("all", cmd.Flags().Lookup("all"))
		viper.BindPFlag("compression", cmd.Flags().Lookup("compression"))
		viper.BindPFlag("nodeps", cmd.Flags().Lookup("nodeps"))
		viper.BindPFlag("onlydeps", cmd.Flags().Lookup("onlydeps"))
		util.BindValuesFlags(cmd)
		viper.BindPFlag("backend-args", cmd.Flags().Lookup("backend-args"))

		viper.BindPFlag("image-repository", cmd.Flags().Lookup("image-repository"))
		viper.BindPFlag("push", cmd.Flags().Lookup("push"))
		viper.BindPFlag("pull", cmd.Flags().Lookup("pull"))
		viper.BindPFlag("wait", cmd.Flags().Lookup("wait"))
		viper.BindPFlag("keep-images", cmd.Flags().Lookup("keep-images"))

		viper.BindPFlag("backend-args", cmd.Flags().Lookup("backend-args"))

	},
	Run: func(cmd *cobra.Command, args []string) {

		treePaths := viper.GetStringSlice("tree")
		dst := viper.GetString("destination")
		concurrency := util.DefaultContext.Config.General.Concurrency
		backendType := viper.GetString("backend")
		privileged := viper.GetBool("privileged")
		revdeps := viper.GetBool("revdeps")
		all := viper.GetBool("all")
		compressionType := viper.GetString("compression")
		imageRepository := viper.GetString("image-repository")
		values := util.ValuesFlags()
		wait := viper.GetBool("wait")
		push := viper.GetBool("push")
		pull := viper.GetBool("pull")
		keepImages := viper.GetBool("keep-images")
		nodeps := viper.GetBool("nodeps")
		onlydeps := viper.GetBool("onlydeps")
		onlyTarget, _ := cmd.Flags().GetBool("only-target-package")
		full, _ := cmd.Flags().GetBool("full")
		rebuild, _ := cmd.Flags().GetBool("rebuild")
		pushFinalImages, _ := cmd.Flags().GetBool("push-final-images")
		pushFinalImagesRepository, _ := cmd.Flags().GetString("push-final-images-repository")
		pushFinalImagesForce, _ := cmd.Flags().GetBool("push-final-images-force")
		generateImages, _ := cmd.Flags().GetBool("generate-final-images")

		backendArgs := viper.GetStringSlice("backend-args")
		out, _ := cmd.Flags().GetString("output")
		pretend, _ := cmd.Flags().GetBool("pretend")
		fromRepo, _ := cmd.Flags().GetBool("from-repositories")

		compilerSpecs := compilerspec.NewBhojpurCompilationspecs()

		var db types.PackageDatabase
		var results Results
		var templateFolders []string

		compilerBackend, err := compiler.NewBackend(util.DefaultContext, backendType)
		helpers.CheckErr(err)

		db = pkg.NewInMemoryDatabase(false)
		defer db.Clean()

		runtimeDB := pkg.NewInMemoryDatabase(false)
		defer runtimeDB.Clean()

		installerRecipe := tree.NewInstallerRecipe(runtimeDB)
		generalRecipe := tree.NewCompilerRecipe(db)

		for _, src := range treePaths {
			util.DefaultContext.Info("Loading tree", src)
			helpers.CheckErr(generalRecipe.Load(src))
			helpers.CheckErr(installerRecipe.Load(src))
		}

		if fromRepo {
			bt, err := installer.LoadBuildTree(generalRecipe, db, util.DefaultContext)
			if err != nil {
				util.DefaultContext.Warning("errors while loading trees from repositories", err.Error())
			}

			for _, r := range bt.RepoDir {
				helpers.CheckErr(installerRecipe.Load(r))
			}

			templateFolders = util.TemplateFolders(util.DefaultContext, bt, treePaths)
		} else {
			templateFolders = util.TemplateFolders(util.DefaultContext, installer.BuildTreeResult{}, treePaths)
		}

		util.DefaultContext.Info("Building in", dst)

		if !fileHelpers.Exists(dst) {
			os.MkdirAll(dst, 0600)
			util.DefaultContext.Debug("Creating destination folder", dst)
		}

		opts := util.DefaultContext.GetConfig().Solver
		pullRepo, _ := cmd.Flags().GetStringArray("pull-repository")

		util.DefaultContext.Debug("Solver", opts.CompactString())

		compileropts := []options.Option{options.NoDeps(nodeps),
			options.WithBackendType(backendType),
			options.PushImages(push),
			options.WithBuildValues(values),
			options.WithPullRepositories(pullRepo),
			options.WithPushRepository(imageRepository),
			options.Rebuild(rebuild),
			options.WithTemplateFolder(templateFolders),
			options.WithSolverOptions(opts),
			options.Wait(wait),
			options.WithRuntimeDatabase(installerRecipe.GetDatabase()),
			options.OnlyTarget(onlyTarget),
			options.PullFirst(pull),
			options.KeepImg(keepImages),
			options.OnlyDeps(onlydeps),
			options.WithContext(util.DefaultContext),
			options.BackendArgs(backendArgs),
			options.Concurrency(concurrency),
			options.WithCompressionType(compression.Implementation(compressionType))}

		if pushFinalImages {
			compileropts = append(compileropts, options.EnablePushFinalImages)
			if pushFinalImagesForce {
				compileropts = append(compileropts, options.ForcePushFinalImages)
			}
			if pushFinalImagesRepository != "" {
				compileropts = append(compileropts, options.WithFinalRepository(pushFinalImagesRepository))
			} else if imageRepository != "" {
				compileropts = append(compileropts, options.WithFinalRepository(imageRepository))
			}
		}

		if generateImages {
			compileropts = append(compileropts, options.EnableGenerateFinalImages)
		}

		bhojpurCompiler := compiler.NewBhojpurCompiler(compilerBackend, generalRecipe.GetDatabase(), compileropts...)

		if full {
			specs, err := bhojpurCompiler.FromDatabase(generalRecipe.GetDatabase(), true, dst)
			if err != nil {
				util.DefaultContext.Fatal(err.Error())
			}
			for _, spec := range specs {
				util.DefaultContext.Info(":package: Selecting ", spec.GetPackage().GetName(), spec.GetPackage().GetVersion())

				compilerSpecs.Add(spec)
			}
		} else if !all {
			for _, a := range args {
				pack, err := helpers.ParsePackageStr(a)
				if err != nil {
					util.DefaultContext.Fatal("Invalid package string ", a, ": ", err.Error())
				}

				spec, err := bhojpurCompiler.FromPackage(pack)
				if err != nil {
					util.DefaultContext.Fatal("Error: " + err.Error())
				}

				spec.SetOutputPath(dst)
				compilerSpecs.Add(spec)
			}
		} else {
			w := generalRecipe.GetDatabase().World()

			for _, p := range w {
				spec, err := bhojpurCompiler.FromPackage(p)
				if err != nil {
					util.DefaultContext.Fatal("Error: " + err.Error())
				}
				util.DefaultContext.Info(":package: Selecting ", p.GetName(), p.GetVersion())
				spec.SetOutputPath(dst)
				compilerSpecs.Add(spec)
			}
		}

		var artifact []*artifact.PackageArtifact
		var errs []error
		if revdeps {
			artifact, errs = bhojpurCompiler.CompileWithReverseDeps(privileged, compilerSpecs)

		} else if pretend {
			var toCalculate []*compilerspec.BhojpurCompilationSpec
			if full {
				var err error
				toCalculate, err = bhojpurCompiler.ComputeMinimumCompilableSet(compilerSpecs.All()...)
				if err != nil {
					errs = append(errs, err)
				}
			} else {
				toCalculate = compilerSpecs.All()
			}

			for _, sp := range toCalculate {
				ht := compiler.NewHashTree(generalRecipe.GetDatabase())
				hashTree, err := ht.Query(bhojpurCompiler, sp)
				if err != nil {
					errs = append(errs, err)
				}
				for _, p := range hashTree.Dependencies {
					results.Packages = append(results.Packages,
						PackageResult{
							Name:       p.Package.GetName(),
							Version:    p.Package.GetVersion(),
							Category:   p.Package.GetCategory(),
							Repository: "",
							Hidden:     p.Package.IsHidden(),
							Target:     sp.GetPackage().HumanReadableString(),
						})
				}
			}

			y, err := yaml.Marshal(results)
			if err != nil {
				fmt.Printf("err: %v\n", err)
				return
			}
			switch out {
			case "yaml":
				fmt.Println(string(y))
			case "json":
				j2, err := yaml.YAMLToJSON(y)
				if err != nil {
					fmt.Printf("err: %v\n", err)
					return
				}
				fmt.Println(string(j2))
			case "terminal":
				for _, p := range results.Packages {
					util.DefaultContext.Info(p.String())
				}
			}
		} else {

			artifact, errs = bhojpurCompiler.CompileParallel(privileged, compilerSpecs)
		}
		if len(errs) != 0 {
			for _, e := range errs {
				util.DefaultContext.Error("Error: " + e.Error())
			}
			util.DefaultContext.Fatal("Bailing out")
		}
		for _, a := range artifact {
			util.DefaultContext.Info("Artifact generated:", a.Path)
		}
	},
}

func init() {
	path, err := os.Getwd()
	if err != nil {
		util.DefaultContext.Fatal(err)
	}

	buildCmd.Flags().StringSliceP("tree", "t", []string{path}, "Path of the tree to use.")
	buildCmd.Flags().String("backend", "docker", "backend used (docker,img)")
	buildCmd.Flags().Bool("privileged", true, "Privileged (Keep permissions)")
	buildCmd.Flags().Bool("revdeps", false, "Build with revdeps")
	buildCmd.Flags().Bool("all", false, "Build all specfiles in the tree")

	buildCmd.Flags().Bool("generate-final-images", false, "Generate final images while building")
	buildCmd.Flags().Bool("push-final-images", false, "Push final images while building")
	buildCmd.Flags().Bool("push-final-images-force", false, "Override existing images")
	buildCmd.Flags().String("push-final-images-repository", "", "Repository where to push final images to")

	buildCmd.Flags().Bool("full", false, "Build all packages (optimized)")
	buildCmd.Flags().StringSlice("values", []string{}, "Build values file to interpolate with each package")
	buildCmd.Flags().StringSliceP("backend-args", "a", []string{}, "Backend args")

	buildCmd.Flags().String("destination", filepath.Join(path, "build"), "Destination folder")
	buildCmd.Flags().String("compression", "none", "Compression alg: none, gzip, zstd")
	buildCmd.Flags().String("image-repository", "bhojpur/cache", "Default base image string for generated image")
	buildCmd.Flags().Bool("push", false, "Push images to a hub")
	buildCmd.Flags().Bool("pull", false, "Pull images from a hub")
	buildCmd.Flags().Bool("wait", false, "Don't build all intermediate images, but wait for them until they are available")
	buildCmd.Flags().Bool("keep-images", true, "Keep built docker images in the host")
	buildCmd.Flags().Bool("nodeps", false, "Build only the target packages, skipping deps (it works only if you already built the deps locally, or by using --pull) ")
	buildCmd.Flags().Bool("onlydeps", false, "Build only package dependencies")
	buildCmd.Flags().Bool("only-target-package", false, "Build packages of only the required target. Otherwise builds all the necessary ones not present in the destination")
	buildCmd.Flags().String("solver-type", "", "Solver strategy")
	buildCmd.Flags().Float32("solver-rate", 0.7, "Solver learning rate")
	buildCmd.Flags().Float32("solver-discount", 1.0, "Solver discount rate")
	buildCmd.Flags().Int("solver-attempts", 9000, "Solver maximum attempts")
	buildCmd.Flags().Bool("solver-concurrent", false, "Use concurrent solver (experimental)")
	buildCmd.Flags().Bool("from-repositories", false, "Consume the user-defined repositories to pull specfiles from")
	buildCmd.Flags().Bool("rebuild", false, "To combine with --pull. Allows to rebuild the target package even if an image is available, against a local values file")
	buildCmd.Flags().Bool("pretend", false, "Just print what packages will be compiled")
	buildCmd.Flags().StringArrayP("pull-repository", "p", []string{}, "A list of repositories to pull the cache from")

	buildCmd.Flags().StringP("output", "o", "terminal", "Output format ( Defaults: terminal, available: json,yaml )")

	RootCmd.AddCommand(buildCmd)
}
