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
	"fmt"

	helpers "github.com/bhojpur/iso/cmd/manager/helpers"
	"github.com/bhojpur/iso/cmd/manager/util"
	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"
)

func NewDatabaseGetCommand() *cobra.Command {
	var c = &cobra.Command{
		Use:   "get <package>",
		Short: "Get a package in the system DB as yaml",
		Long: `Get a package in the system database in the YAML format:

		$ isomgr database get system/foo

To return also files:
		$ isomgr database get --files system/foo`,
		Args: cobra.OnlyValidArgs,

		Run: func(cmd *cobra.Command, args []string) {
			showFiles, _ := cmd.Flags().GetBool("files")

			systemDB := util.SystemDB(util.DefaultContext.Config)

			for _, a := range args {
				pack, err := helpers.ParsePackageStr(a)
				if err != nil {
					continue
				}

				ps, err := systemDB.FindPackages(pack)
				if err != nil {
					continue
				}
				for _, p := range ps {
					y, err := p.Yaml()
					if err != nil {
						continue
					}
					fmt.Println(string(y))
					if showFiles {
						files, err := systemDB.GetPackageFiles(p)
						if err != nil {
							continue
						}
						b, err := yaml.Marshal(files)
						if err != nil {
							continue
						}
						fmt.Println("files:\n" + string(b))
					}
				}
			}
		},
	}
	c.Flags().Bool("files", false, "Show package files.")

	return c
}
