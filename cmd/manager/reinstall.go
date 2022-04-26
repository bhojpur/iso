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
	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	installer "github.com/bhojpur/iso/pkg/manager/installer"

	helpers "github.com/bhojpur/iso/cmd/manager/helpers"
	"github.com/bhojpur/iso/cmd/manager/util"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var reinstallCmd = &cobra.Command{
	Use:   "reinstall <pkg1> <pkg2> <pkg3>",
	Short: "reinstall a set of packages",
	Long: `Reinstall a group of packages in the system:

	$ isomgr reinstall -y system/busybox shells/bash system/coreutils ...
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("onlydeps", cmd.Flags().Lookup("onlydeps"))
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		viper.BindPFlag("for", cmd.Flags().Lookup("for"))

		viper.BindPFlag("yes", cmd.Flags().Lookup("yes"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		var toUninstall types.Packages
		var toAdd types.Packages

		force := viper.GetBool("force")
		onlydeps := viper.GetBool("onlydeps")
		yes := viper.GetBool("yes")

		downloadOnly, _ := cmd.Flags().GetBool("download-only")
		installed, _ := cmd.Flags().GetBool("installed")

		util.DefaultContext.Debug("Solver", util.DefaultContext.Config.Solver.CompactString())

		inst := installer.NewBhojpurInstaller(installer.BhojpurInstallerOptions{
			Concurrency:                 util.DefaultContext.Config.General.Concurrency,
			SolverOptions:               util.DefaultContext.Config.Solver,
			NoDeps:                      true,
			Force:                       force,
			OnlyDeps:                    onlydeps,
			PreserveSystemEssentialData: true,
			Ask:                         !yes,
			DownloadOnly:                downloadOnly,
			Context:                     util.DefaultContext,
			PackageRepositories:         util.DefaultContext.Config.SystemRepositories,
		})

		system := &installer.System{Database: util.SystemDB(util.DefaultContext.Config), Target: util.DefaultContext.Config.System.Rootfs}

		if installed {
			for _, p := range system.Database.World() {
				toUninstall = append(toUninstall, p)
				c := p.Clone()
				c.SetVersion(">=0")
				toAdd = append(toAdd, c)
			}
		} else {
			for _, a := range args {
				pack, err := helpers.ParsePackageStr(a)
				if err != nil {
					util.DefaultContext.Fatal("Invalid package string ", a, ": ", err.Error())
				}
				toUninstall = append(toUninstall, pack)
				toAdd = append(toAdd, pack)
			}
		}

		err := inst.Swap(toUninstall, toAdd, system)
		if err != nil {
			util.DefaultContext.Fatal("Error: " + err.Error())
		}
	},
}

func init() {
	reinstallCmd.Flags().Bool("onlydeps", false, "Consider **only** package dependencies")
	reinstallCmd.Flags().Bool("force", false, "Skip errors and keep going (potentially harmful)")
	reinstallCmd.Flags().Bool("installed", false, "Reinstall installed packages")
	reinstallCmd.Flags().BoolP("yes", "y", false, "Don't ask questions")
	reinstallCmd.Flags().Bool("download-only", false, "Download only")

	RootCmd.AddCommand(reinstallCmd)
}
