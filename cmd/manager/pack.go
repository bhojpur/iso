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
	"time"

	helpers "github.com/bhojpur/iso/cmd/manager/helpers"
	"github.com/bhojpur/iso/cmd/manager/util"
	"github.com/bhojpur/iso/pkg/manager/api/core/types/artifact"
	"github.com/bhojpur/iso/pkg/manager/compiler/types/compression"
	compilerspec "github.com/bhojpur/iso/pkg/manager/compiler/types/spec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var packCmd = &cobra.Command{
	Use:   "pack <package name>",
	Short: "pack a custom package",
	Long: `Pack creates a package from a directory, generating the metadata required from a tree to generate a repository.

Pack can be used to manually replace what "isomgr build" does automatically by reading the packages build.yaml files.

	$ mkdir -p output/etc/foo
	$ echo "my config" > output/etc/foo
	$ isomgr pack foo/bar@1.1 --source output

Afterwards, you can use the content generated and associate it with a tree and a corresponding definition.yaml file with "isomgr create-repo".
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("destination", cmd.Flags().Lookup("destination"))
		viper.BindPFlag("compression", cmd.Flags().Lookup("compression"))
		viper.BindPFlag("source", cmd.Flags().Lookup("source"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		sourcePath := viper.GetString("source")

		dst := viper.GetString("destination")
		compressionType := viper.GetString("compression")
		concurrency := util.DefaultContext.Config.General.Concurrency

		if len(args) != 1 {
			util.DefaultContext.Fatal("You must specify a package name")
		}

		packageName := args[0]

		p, err := helpers.ParsePackageStr(packageName)
		if err != nil {
			util.DefaultContext.Fatal("Invalid package string ", packageName, ": ", err.Error())
		}

		spec := &compilerspec.BhojpurCompilationSpec{Package: p}
		a := artifact.NewPackageArtifact(filepath.Join(dst, p.GetFingerPrint()+".package.tar"))
		a.CompressionType = compression.Implementation(compressionType)
		err = a.Compress(sourcePath, concurrency)
		if err != nil {
			util.DefaultContext.Fatal("failed compressing ", packageName, ": ", err.Error())
		}
		a.CompileSpec = spec
		filelist, err := a.FileList()
		if err != nil {
			util.DefaultContext.Fatal("failed generating file list for ", packageName, ": ", err.Error())
		}
		a.Files = filelist
		a.CompileSpec.GetPackage().SetBuildTimestamp(time.Now().String())
		err = a.WriteYAML(dst)
		if err != nil {
			util.DefaultContext.Fatal("failed writing metadata yaml file for ", packageName, ": ", err.Error())
		}
	},
}

func init() {
	path, err := os.Getwd()
	if err != nil {
		util.DefaultContext.Fatal(err)
	}
	packCmd.Flags().String("source", path, "Source folder")
	packCmd.Flags().String("destination", path, "Destination folder")
	packCmd.Flags().String("compression", "gzip", "Compression alg: none, gzip")

	RootCmd.AddCommand(packCmd)
}
