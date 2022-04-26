package cmd_repo

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
	"path/filepath"
	"strconv"
	"time"

	"github.com/bhojpur/iso/cmd/manager/util"
	installer "github.com/bhojpur/iso/pkg/manager/installer"
	"github.com/pterm/pterm"

	"github.com/spf13/cobra"
)

func NewRepoListCommand() *cobra.Command {
	var ans = &cobra.Command{
		Use:   "list [OPTIONS]",
		Short: "List of the configured repositories.",
		Args:  cobra.OnlyValidArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
		},
		Run: func(cmd *cobra.Command, args []string) {
			var repoColor, repoText, repoRevision string

			enable, _ := cmd.Flags().GetBool("enabled")
			quiet, _ := cmd.Flags().GetBool("quiet")
			repoType, _ := cmd.Flags().GetString("type")

			for _, repo := range util.DefaultContext.Config.SystemRepositories {
				if enable && !repo.Enable {
					continue
				}

				if repoType != "" && repo.Type != repoType {
					continue
				}

				repoRevision = ""

				if quiet {
					fmt.Println(repo.Name)
				} else {
					if repo.Enable {
						repoColor = pterm.LightGreen(repo.Name)
					} else {
						repoColor = pterm.LightRed(repo.Name)
					}
					if repo.Description != "" {
						repoText = pterm.LightYellow(repo.Description)
					} else {
						repoText = pterm.LightYellow(repo.Urls[0])
					}

					repobasedir := util.DefaultContext.Config.System.GetRepoDatabaseDirPath(repo.Name)
					if repo.Cached {

						r := installer.NewSystemRepository(repo)
						localRepo, _ := r.ReadSpecFile(filepath.Join(repobasedir,
							installer.REPOSITORY_SPECFILE))
						if localRepo != nil {
							tsec, _ := strconv.ParseInt(localRepo.GetLastUpdate(), 10, 64)
							repoRevision = pterm.LightRed(localRepo.GetRevision()) +
								" - " + pterm.LightGreen(time.Unix(tsec, 0).String())
						}
					}

					if repoRevision != "" {
						fmt.Println(
							fmt.Sprintf("%s\n  %s\n  Revision %s", repoColor, repoText, repoRevision))
					} else {
						fmt.Println(
							fmt.Sprintf("%s\n  %s", repoColor, repoText))
					}
				}
			}
		},
	}

	ans.Flags().Bool("enabled", false, "Show only enable repositories.")
	ans.Flags().BoolP("quiet", "q", false, "Show only name of the repositories.")
	ans.Flags().StringP("type", "t", "", "Filter repositories of a specific type")

	return ans
}
