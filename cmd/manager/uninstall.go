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
	helpers "github.com/bhojpur/iso/cmd/manager/helpers"
	"github.com/bhojpur/iso/cmd/manager/util"
	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	installer "github.com/bhojpur/iso/pkg/manager/installer"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var uninstallCmd = &cobra.Command{
	Use:     "uninstall <pkg> <pkg2> ...",
	Short:   "Uninstall a package or a list of packages",
	Long:    `Uninstall packages`,
	Aliases: []string{"rm", "un"},
	PreRun: func(cmd *cobra.Command, args []string) {

		viper.BindPFlag("nodeps", cmd.Flags().Lookup("nodeps"))
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		viper.BindPFlag("yes", cmd.Flags().Lookup("yes"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		toRemove := []*types.Package{}
		for _, a := range args {

			pack, err := helpers.ParsePackageStr(a)
			if err != nil {
				util.DefaultContext.Fatal("Invalid package string ", a, ": ", err.Error())
			}
			toRemove = append(toRemove, pack)
		}

		force := viper.GetBool("force")
		nodeps, _ := cmd.Flags().GetBool("nodeps")
		full, _ := cmd.Flags().GetBool("full")
		checkconflicts, _ := cmd.Flags().GetBool("conflictscheck")
		fullClean, _ := cmd.Flags().GetBool("full-clean")
		yes := viper.GetBool("yes")
		keepProtected, _ := cmd.Flags().GetBool("keep-protected-files")

		util.DefaultContext.Config.ConfigProtectSkip = !keepProtected

		util.DefaultContext.Config.Solver.Implementation = types.SolverSingleCoreSimple

		util.DefaultContext.Debug("Solver", util.DefaultContext.Config.Solver.CompactString())

		inst := installer.NewBhojpurInstaller(installer.BhojpurInstallerOptions{
			Concurrency:                 util.DefaultContext.Config.General.Concurrency,
			SolverOptions:               util.DefaultContext.Config.Solver,
			NoDeps:                      nodeps,
			Force:                       force,
			FullUninstall:               full,
			FullCleanUninstall:          fullClean,
			CheckConflicts:              checkconflicts,
			Ask:                         !yes,
			PreserveSystemEssentialData: true,
			Context:                     util.DefaultContext,
		})

		system := &installer.System{Database: util.SystemDB(util.DefaultContext.Config), Target: util.DefaultContext.Config.System.Rootfs}

		if err := inst.Uninstall(system, toRemove...); err != nil {
			util.DefaultContext.Fatal("Error: " + err.Error())
		}
	},
}

func init() {

	uninstallCmd.Flags().Bool("nodeps", false, "Don't consider package dependencies (harmful! overrides checkconflicts and full!)")
	uninstallCmd.Flags().Bool("force", false, "Force uninstall")
	uninstallCmd.Flags().Bool("full", false, "Attempts to remove as much packages as possible which aren't required (slow)")
	uninstallCmd.Flags().Bool("conflictscheck", true, "Check if the package marked for deletion is required by other packages")
	uninstallCmd.Flags().Bool("full-clean", false, "(experimental) Uninstall packages and all the other deps/revdeps of it.")
	uninstallCmd.Flags().Bool("solver-concurrent", false, "Use concurrent solver (experimental)")
	uninstallCmd.Flags().BoolP("yes", "y", false, "Don't ask questions")
	uninstallCmd.Flags().BoolP("keep-protected-files", "k", false, "Keep package protected files around")

	RootCmd.AddCommand(uninstallCmd)
}
