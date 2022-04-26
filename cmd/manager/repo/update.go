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
	"github.com/bhojpur/iso/cmd/manager/util"
	installer "github.com/bhojpur/iso/pkg/manager/installer"

	"github.com/spf13/cobra"
)

func NewRepoUpdateCommand() *cobra.Command {
	var repoUpdate = &cobra.Command{
		Use:   "update [repo1] [repo2] [OPTIONS]",
		Short: "Update a specific cached repository or all cached repositories.",
		Example: `
# Update all cached repositories:
$> isomgr repo update

# Update only repo1 and repo2
$> isomgr repo update repo1 repo2
`,
		Aliases: []string{"up"},
		PreRun: func(cmd *cobra.Command, args []string) {
		},
		Run: func(cmd *cobra.Command, args []string) {

			ignore, _ := cmd.Flags().GetBool("ignore-errors")
			force, _ := cmd.Flags().GetBool("force")

			if len(args) > 0 {
				for _, rname := range args {
					repo, err := util.DefaultContext.Config.GetSystemRepository(rname)
					if err != nil && !ignore {
						util.DefaultContext.Fatal(err.Error())
					} else if err != nil {
						continue
					}

					r := installer.NewSystemRepository(*repo)
					_, err = r.Sync(util.DefaultContext, force)
					if err != nil && !ignore {
						util.DefaultContext.Fatal("Error on sync repository " + rname + ": " + err.Error())
					}
				}

			} else {
				for _, repo := range util.DefaultContext.Config.SystemRepositories {
					if repo.Cached && repo.Enable {
						r := installer.NewSystemRepository(repo)
						_, err := r.Sync(util.DefaultContext, force)
						if err != nil && !ignore {
							util.DefaultContext.Fatal("Error on sync repository " + r.GetName() + ": " + err.Error())
						}
					}
				}
			}
		},
	}

	repoUpdate.Flags().BoolP("ignore-errors", "i", false, "Ignore errors on sync repositories.")
	repoUpdate.Flags().BoolP("force", "f", true, "Force resync.")

	return repoUpdate
}
