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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bhojpur/iso/cmd/manager/util"
	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"

	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean packages cache.",
	Long:  `remove downloaded packages tarballs and clean cache directory`,

	Run: func(cmd *cobra.Command, args []string) {
		var cleaned int = 0
		// Check if cache dir exists
		if fileHelper.Exists(util.DefaultContext.Config.System.PkgsCachePath) {

			files, err := ioutil.ReadDir(util.DefaultContext.Config.System.PkgsCachePath)
			if err != nil {
				util.DefaultContext.Fatal("Error on read cachedir ", err.Error())
			}

			for _, file := range files {

				util.DefaultContext.Debug("Removing ", file.Name())

				err := os.RemoveAll(
					filepath.Join(util.DefaultContext.Config.System.PkgsCachePath, file.Name()))
				if err != nil {
					util.DefaultContext.Fatal("Error on removing", file.Name())
				}
				cleaned++
			}
		}

		util.DefaultContext.Info(fmt.Sprintf("Cleaned: %d files from %s", cleaned, util.DefaultContext.Config.System.PkgsCachePath))

	},
}

func init() {
	RootCmd.AddCommand(cleanupCmd)
}
