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

	"github.com/bhojpur/iso/cmd/manager/util"
	spectooling "github.com/bhojpur/iso/pkg/manager/spectooling"
	tree "github.com/bhojpur/iso/pkg/manager/tree"
	version "github.com/bhojpur/iso/pkg/manager/versioner"

	"github.com/spf13/cobra"
)

func NewTreeBumpCommand() *cobra.Command {

	var ans = &cobra.Command{
		Use:   "bump [OPTIONS]",
		Short: "Bump a new package build version.",
		Args:  cobra.OnlyValidArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			df, _ := cmd.Flags().GetString("definition-file")
			if df == "" {
				util.DefaultContext.Fatal("Mandatory definition.yaml path missing.")
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			spec, _ := cmd.Flags().GetString("definition-file")
			toStdout, _ := cmd.Flags().GetBool("to-stdout")
			pkgVersion, _ := cmd.Flags().GetString("pkg-version")
			pack, err := tree.ReadDefinitionFile(spec)
			if err != nil {
				util.DefaultContext.Fatal(err.Error())
			}

			if pkgVersion != "" {
				validator := &version.WrappedVersioner{}
				err := validator.Validate(pkgVersion)
				if err != nil {
					util.DefaultContext.Fatal("Invalid version string: " + err.Error())
				}
				pack.SetVersion(pkgVersion)
			} else {
				// Retrieve version build section with Gentoo parser
				err = pack.BumpBuildVersion()
				if err != nil {
					util.DefaultContext.Fatal("Error on increment build version: " + err.Error())
				}
			}
			if toStdout {
				data, err := spectooling.NewDefaultPackageSanitized(&pack).Yaml()
				if err != nil {
					util.DefaultContext.Fatal("Error on yaml conversion: " + err.Error())
				}
				fmt.Println(string(data))
			} else {

				err = tree.WriteDefinitionFile(&pack, spec)
				if err != nil {
					util.DefaultContext.Fatal("Error on write definition file: " + err.Error())
				}

				fmt.Printf("Bumped package %s/%s-%s.\n", pack.Category, pack.Name, pack.Version)
			}
		},
	}

	ans.Flags().StringP("pkg-version", "p", "", "Set a specific package version")
	ans.Flags().StringP("definition-file", "f", "", "Path of the definition to bump.")
	ans.Flags().BoolP("to-stdout", "o", false, "Bump package to output.")

	return ans
}
