package cmd_tree

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

	helpers "github.com/bhojpur/iso/cmd/manager/helpers"
	"github.com/bhojpur/iso/cmd/manager/util"
	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	"github.com/bhojpur/iso/pkg/manager/compiler"
	"github.com/bhojpur/iso/pkg/manager/compiler/backend"
	"github.com/bhojpur/iso/pkg/manager/compiler/types/options"
	"github.com/bhojpur/iso/pkg/manager/installer"
	"github.com/ghodss/yaml"

	pkg "github.com/bhojpur/iso/pkg/manager/database"
	tree "github.com/bhojpur/iso/pkg/manager/tree"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewTreeImageCommand() *cobra.Command {

	var ans = &cobra.Command{
		Use: "images [OPTIONS]",
		// Skip processing output
		Annotations: map[string]string{
			util.CommandProcessOutput: "",
		},
		Short: "List of the images of a package",
		PreRun: func(cmd *cobra.Command, args []string) {
			t, _ := cmd.Flags().GetStringArray("tree")
			if len(t) == 0 {
				util.DefaultContext.Fatal("Mandatory tree param missing.")
			}

			if len(args) != 1 {
				util.DefaultContext.Fatal("Expects one package as parameter")
			}
			util.BindValuesFlags(cmd)
			viper.BindPFlag("image-repository", cmd.Flags().Lookup("image-repository"))

		},
		Run: func(cmd *cobra.Command, args []string) {
			var results TreeResults

			treePath, _ := cmd.Flags().GetStringArray("tree")
			imageRepository := viper.GetString("image-repository")
			pullRepo, _ := cmd.Flags().GetStringArray("pull-repository")
			values := util.ValuesFlags()
			out, _ := cmd.Flags().GetString("output")
			reciper := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))

			for _, t := range treePath {
				err := reciper.Load(t)
				if err != nil {
					util.DefaultContext.Fatal("Error on load tree ", err)
				}
			}
			compilerBackend := backend.NewSimpleDockerBackend(util.DefaultContext)

			opts := util.DefaultContext.Config.Solver
			opts.SolverOptions = types.SolverOptions{Type: types.SolverSingleCoreSimple, Concurrency: 1}
			bhojpurCompiler := compiler.NewBhojpurCompiler(
				compilerBackend,
				reciper.GetDatabase(),
				options.WithBuildValues(values),
				options.WithContext(util.DefaultContext),
				options.WithPushRepository(imageRepository),
				options.WithPullRepositories(pullRepo),
				options.WithTemplateFolder(util.TemplateFolders(util.DefaultContext, installer.BuildTreeResult{}, treePath)),
				options.WithSolverOptions(opts),
			)

			a := args[0]

			pack, err := helpers.ParsePackageStr(a)
			if err != nil {
				util.DefaultContext.Fatal("Invalid package string ", a, ": ", err.Error())
			}

			spec, err := bhojpurCompiler.FromPackage(pack)
			if err != nil {
				util.DefaultContext.Fatal("Error: " + err.Error())
			}

			ht := compiler.NewHashTree(reciper.GetDatabase())

			copy, err := compiler.CompilerFinalImages(bhojpurCompiler)
			if err != nil {
				util.DefaultContext.Fatal("Error: " + err.Error())
			}
			hashtree, err := ht.Query(copy, spec)
			if err != nil {
				util.DefaultContext.Fatal("Error: " + err.Error())
			}

			for _, assertion := range hashtree.Solution { //highly dependent on the order

				//buildImageHash := imageRepository + ":" + assertion.Hash.BuildHash
				currentPackageImageHash := imageRepository + ":" + assertion.Hash.PackageHash

				results.Packages = append(results.Packages, TreePackageResult{
					Name:     assertion.Package.GetName(),
					Version:  assertion.Package.GetVersion(),
					Category: assertion.Package.GetCategory(),
					Image:    currentPackageImageHash,
				})
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
			default:
				for _, p := range results.Packages {
					fmt.Println(fmt.Sprintf("%s/%s-%s: %s", p.Category, p.Name, p.Version, p.Image))
				}
			}
		},
	}
	path, err := os.Getwd()
	if err != nil {
		util.DefaultContext.Fatal(err)
	}
	ans.Flags().StringP("output", "o", "terminal", "Output format ( Defaults: terminal, available: json,yaml )")
	ans.Flags().StringArrayP("tree", "t", []string{path}, "Path of the tree to use.")
	ans.Flags().String("image-repository", "bhojpur/cache", "Default base image string for generated image")
	ans.Flags().StringArrayP("pull-repository", "p", []string{}, "A list of repositories to pull the cache from")

	return ans
}
