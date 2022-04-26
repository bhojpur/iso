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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bhojpur/iso/cmd/manager/util"
	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	"github.com/bhojpur/iso/pkg/manager/helpers"
	"github.com/ghodss/yaml"

	"github.com/spf13/cobra"
)

func NewRepoAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [OPTIONS] https://..../something.yaml /local/file.yaml",
		Short: "Add a repository to the system",
		Args:  cobra.ExactArgs(1),
		Long: `
Adds a repository to the system. URLs, local files or inline repo can be specified, examples:

# URL/File:

 isomgr repo add /path/to/file

 isomgr repo add https://....

 isomgr repo add ... --name "foo"

# Inline:

 isomgr repo add testfo --description "Bar" --url "FOZZ" --type "ff"

 `,
		Run: func(cmd *cobra.Command, args []string) {

			uri := args[0]
			d, _ := cmd.Flags().GetString("dir")
			yes, _ := cmd.Flags().GetBool("yes")

			desc, _ := cmd.Flags().GetString("description")
			t, _ := cmd.Flags().GetString("type")
			url, _ := cmd.Flags().GetString("url")
			ref, _ := cmd.Flags().GetString("reference")
			prio, _ := cmd.Flags().GetInt("priority")

			if len(util.DefaultContext.Config.RepositoriesConfDir) == 0 && d == "" {
				util.DefaultContext.Fatal("No repository dirs defined")
				return
			}
			if d == "" {
				d = util.DefaultContext.Config.RepositoriesConfDir[0]
			}

			var r *types.BhojpurRepository
			str, err := helpers.GetURI(uri)
			if err != nil {
				r = &types.BhojpurRepository{
					Enable:      true,
					Cached:      true,
					Name:        uri,
					Description: desc,
					ReferenceID: ref,
					Type:        t,
					Urls:        []string{url},
					Priority:    prio,
				}
			} else {
				r, err = types.LoadRepository([]byte(str))
				if err != nil {
					util.DefaultContext.Fatal(err)
				}
				if desc != "" {
					r.Description = desc
				}
				if ref != "" {
					r.ReferenceID = ref
				}
				if t != "" {
					r.Type = t
				}
				if url != "" {
					r.Urls = []string{url}
				}
				if prio != 0 {
					r.Priority = prio
				}
			}

			file := filepath.Join(util.DefaultContext.Config.System.Rootfs, d, fmt.Sprintf("%s.yaml", r.Name))

			b, err := yaml.Marshal(r)
			if err != nil {
				util.DefaultContext.Fatal(err)
			}

			util.DefaultContext.Infof("Adding repository to the sytem as %s", file)
			fmt.Println(string(b))
			util.DefaultContext.Info(r.String())

			if !yes && !util.DefaultContext.Ask() {
				util.DefaultContext.Info("Aborted by user")
				return
			}

			if err := ioutil.WriteFile(file, b, os.ModePerm); err != nil {
				util.DefaultContext.Fatal(err)
			}
		},
	}
	cmd.Flags().BoolP("yes", "y", false, "Assume yes to questions")
	cmd.Flags().StringP("dir", "o", "", "Folder to write to")
	cmd.Flags().String("description", "", "Repository description")
	cmd.Flags().String("type", "", "Repository type")
	cmd.Flags().String("url", "", "Repository URL")
	cmd.Flags().String("reference", "", "Repository Reference ID")
	cmd.Flags().IntP("priority", "p", 99, "repository prio")
	return cmd
}
