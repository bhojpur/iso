package cmd_database

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

	"github.com/bhojpur/iso/cmd/manager/util"
	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	artifact "github.com/bhojpur/iso/pkg/manager/api/core/types/artifact"

	"github.com/spf13/cobra"
)

func NewDatabaseCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create <artifact_metadata1.yaml> <artifact_metadata1.yaml>",
		Short: "Insert a package in the system DB",
		Long: `Inserts a package in the system database:

		$ isomgr database create foo.yaml

"isomgr database create" injects a package in the system database without actually installing it, use it with caution.

This commands takes multiple yaml input file representing package artifacts, that are usually generated while building packages.

The yaml must contain the package definition, and the file list at least.

For reference, inspect a "metadata.yaml" file generated while running "isomgr build"`,
		Args: cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {

			systemDB := util.SystemDB(util.DefaultContext.Config)

			for _, a := range args {
				dat, err := ioutil.ReadFile(a)
				if err != nil {
					util.DefaultContext.Fatal("Failed reading ", a, ": ", err.Error())
				}
				art, err := artifact.NewPackageArtifactFromYaml(dat)
				if err != nil {
					util.DefaultContext.Fatal("Failed reading yaml ", a, ": ", err.Error())
				}

				files := art.Files

				// Check if the package is already present
				if p, err := systemDB.FindPackage(art.CompileSpec.GetPackage()); err == nil && p.GetName() != "" {
					util.DefaultContext.Fatal("Package", art.CompileSpec.GetPackage().HumanReadableString(),
						" already present.")
				}

				if _, err := systemDB.CreatePackage(art.CompileSpec.GetPackage()); err != nil {
					util.DefaultContext.Fatal("Failed to create ", a, ": ", err.Error())
				}
				if err := systemDB.SetPackageFiles(&types.PackageFile{PackageFingerprint: art.CompileSpec.GetPackage().GetFingerPrint(), Files: files}); err != nil {
					util.DefaultContext.Fatal("Failed setting package files for ", a, ": ", err.Error())
				}

				util.DefaultContext.Info(art.CompileSpec.GetPackage().HumanReadableString(), " created")
			}

		},
	}

}
