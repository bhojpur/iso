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
	helpers "github.com/bhojpur/iso/cmd/manager/helpers"
	"github.com/bhojpur/iso/cmd/manager/util"

	"github.com/spf13/cobra"
)

func NewDatabaseRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove [package1] [package2] ...",
		Short: "Remove a package from the system DB (forcefully - you normally don't want to do that)",
		Long: `Removes a package in the system database without actually uninstalling it:

		$ isomgr database remove foo/bar

This commands takes multiple packages as arguments and prunes their entries from the system database.
`,
		Args: cobra.OnlyValidArgs,

		Run: func(cmd *cobra.Command, args []string) {

			systemDB := util.SystemDB(util.DefaultContext.Config)

			for _, a := range args {
				pack, err := helpers.ParsePackageStr(a)
				if err != nil {
					util.DefaultContext.Fatal("Invalid package string ", a, ": ", err.Error())
				}

				if err := systemDB.RemovePackage(pack); err != nil {
					util.DefaultContext.Fatal("Failed removing ", a, ": ", err.Error())
				}

				if err := systemDB.RemovePackageFiles(pack); err != nil {
					util.DefaultContext.Fatal("Failed removing files for ", a, ": ", err.Error())
				}
			}

		},
	}

}
