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

var installCmd = &cobra.Command{
	Use:   "install <pkg1> <pkg2> ...",
	Short: "Install a package",
	Long: `Installs one or more packages without asking questions:

	$ isomgr install -y utils/busybox utils/yq ...
	
To install only deps of a package:
	
	$ isomgr install --onlydeps utils/busybox ...
	
To not install deps of a package:
	
	$ isomgr install --nodeps utils/busybox ...

To force install a package:
	
	$ isomgr install --force utils/busybox ...
`,
	Aliases: []string{"i"},
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("onlydeps", cmd.Flags().Lookup("onlydeps"))
		viper.BindPFlag("nodeps", cmd.Flags().Lookup("nodeps"))
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		viper.BindPFlag("yes", cmd.Flags().Lookup("yes"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		var toInstall types.Packages

		for _, a := range args {
			pack, err := helpers.ParsePackageStr(a)
			if err != nil {
				util.DefaultContext.Fatal("Invalid package string ", a, ": ", err.Error())
			}
			toInstall = append(toInstall, pack)
		}

		force := viper.GetBool("force")
		nodeps := viper.GetBool("nodeps")
		onlydeps := viper.GetBool("onlydeps")
		yes := viper.GetBool("yes")
		downloadOnly, _ := cmd.Flags().GetBool("download-only")
		relax, _ := cmd.Flags().GetBool("relax")

		util.DefaultContext.Debug("Solver", util.DefaultContext.Config.Solver.CompactString())

		inst := installer.NewBhojpurInstaller(installer.BhojpurInstallerOptions{
			Concurrency:                 util.DefaultContext.Config.General.Concurrency,
			SolverOptions:               util.DefaultContext.Config.Solver,
			NoDeps:                      nodeps,
			Force:                       force,
			OnlyDeps:                    onlydeps,
			PreserveSystemEssentialData: true,
			DownloadOnly:                downloadOnly,
			Ask:                         !yes,
			Relaxed:                     relax,
			PackageRepositories:         util.DefaultContext.Config.SystemRepositories,
			Context:                     util.DefaultContext,
		})

		system := &installer.System{
			Database: util.SystemDB(util.DefaultContext.Config),
			Target:   util.DefaultContext.Config.System.Rootfs,
		}
		err := inst.Install(toInstall, system)
		if err != nil {
			util.DefaultContext.Fatal("Error: " + err.Error())
		}
	},
}

func init() {

	installCmd.Flags().Bool("nodeps", false, "Don't consider package dependencies (harmful!)")
	installCmd.Flags().Bool("relax", false, "Relax installation constraints")

	installCmd.Flags().Bool("onlydeps", false, "Consider **only** package dependencies")
	installCmd.Flags().Bool("force", false, "Skip errors and keep going (potentially harmful)")
	installCmd.Flags().Bool("solver-concurrent", false, "Use concurrent solver (experimental)")
	installCmd.Flags().BoolP("yes", "y", false, "Don't ask questions")
	installCmd.Flags().Bool("download-only", false, "Download only")
	installCmd.Flags().StringArray("finalizer-env", []string{},
		"Set finalizer environment in the format key=value.")

	RootCmd.AddCommand(installCmd)
}
