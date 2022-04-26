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
	"github.com/bhojpur/iso/cmd/manager/util"
	installer "github.com/bhojpur/iso/pkg/manager/installer"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var reclaimCmd = &cobra.Command{
	Use:   "reclaim",
	Short: "Reclaim packages to Bhojpur ISO database from available repositories",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
	},
	Long: `Reclaim tries to find association between packages in the online repositories and the system one.

	$ isomgr reclaim

It scans the target file system, and if finds a match with a package available in the repositories, it marks as installed in the system database.
`,
	Run: func(cmd *cobra.Command, args []string) {

		force := viper.GetBool("force")

		util.DefaultContext.Debug("Solver", util.DefaultContext.Config.Solver.CompactString())

		inst := installer.NewBhojpurInstaller(installer.BhojpurInstallerOptions{
			Concurrency:                 util.DefaultContext.Config.General.Concurrency,
			Force:                       force,
			PreserveSystemEssentialData: true,
			PackageRepositories:         util.DefaultContext.Config.SystemRepositories,
			Context:                     util.DefaultContext,
		})

		system := &installer.System{
			Database: util.SystemDB(util.DefaultContext.Config),
			Target:   util.DefaultContext.Config.System.Rootfs,
		}
		err := inst.Reclaim(system)
		if err != nil {
			util.DefaultContext.Fatal("Error: " + err.Error())
		}
	},
}

func init() {

	reclaimCmd.Flags().Bool("force", false, "Skip errors and keep going (potentially harmful)")

	RootCmd.AddCommand(reclaimCmd)
}
