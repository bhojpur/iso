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
	"os"
	"strings"

	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	installer "github.com/bhojpur/iso/pkg/manager/installer"

	"github.com/bhojpur/iso/cmd/manager/util"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var osCheckCmd = &cobra.Command{
	Use:   "oscheck",
	Short: "Checks packages integrity",
	Long: `List packages that are installed in the system which files are missing in the system.

	$ isomgr oscheck
	
To reinstall packages in the list:
	
	$ isomgr oscheck --reinstall
`,
	Aliases: []string{"i"},
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("onlydeps", cmd.Flags().Lookup("onlydeps"))
		viper.BindPFlag("nodeps", cmd.Flags().Lookup("nodeps"))
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		viper.BindPFlag("yes", cmd.Flags().Lookup("yes"))
	},
	Run: func(cmd *cobra.Command, args []string) {

		force := viper.GetBool("force")
		onlydeps := viper.GetBool("onlydeps")
		yes := viper.GetBool("yes")

		downloadOnly, _ := cmd.Flags().GetBool("download-only")

		system := &installer.System{
			Database: util.SystemDB(util.DefaultContext.Config),
			Target:   util.DefaultContext.Config.System.Rootfs,
		}
		packs := system.OSCheck(util.DefaultContext)
		if !util.DefaultContext.Config.General.Quiet {
			if len(packs) == 0 {
				util.DefaultContext.Success("All good!")
				os.Exit(0)
			} else {
				util.DefaultContext.Info("Following packages are missing files or are incomplete:")
				for _, p := range packs {
					util.DefaultContext.Info(p.HumanReadableString())
				}
			}
		} else {
			var s []string
			for _, p := range packs {
				s = append(s, p.HumanReadableString())
			}
			fmt.Println(strings.Join(s, " "))
		}

		reinstall, _ := cmd.Flags().GetBool("reinstall")
		if reinstall {

			// Strip version for reinstall
			toInstall := types.Packages{}
			for _, p := range packs {
				new := p.Clone()
				new.SetVersion(">=0")
				toInstall = append(toInstall, new)
			}

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

			err := inst.Swap(packs, toInstall, system)
			if err != nil {
				util.DefaultContext.Fatal("Error: " + err.Error())
			}
		}
	},
}

func init() {

	osCheckCmd.Flags().Bool("reinstall", false, "reinstall")

	osCheckCmd.Flags().Bool("onlydeps", false, "Consider **only** package dependencies")
	osCheckCmd.Flags().Bool("force", false, "Skip errors and keep going (potentially harmful)")
	osCheckCmd.Flags().BoolP("yes", "y", false, "Don't ask questions")
	osCheckCmd.Flags().Bool("download-only", false, "Download only")

	RootCmd.AddCommand(osCheckCmd)
}
