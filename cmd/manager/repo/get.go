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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bhojpur/iso/cmd/manager/util"
	"github.com/ghodss/yaml"

	"github.com/spf13/cobra"
)

func NewRepoGetCommand() *cobra.Command {
	var ans = &cobra.Command{
		Use:   "get [OPTIONS] name",
		Short: "get repository in the system",
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
		},
		Run: func(cmd *cobra.Command, args []string) {
			o, _ := cmd.Flags().GetString("output")

			for _, repo := range util.DefaultContext.Config.SystemRepositories {
				if repo.Name != args[0] {
					continue
				}

				switch strings.ToLower(o) {
				case "json":
					b, _ := json.Marshal(repo)
					fmt.Println(string(b))
				case "yaml":
					b, _ := yaml.Marshal(repo)
					fmt.Println(string(b))
				default:
					fmt.Println(repo)
				}
				break
			}
		},
	}

	ans.Flags().StringP("output", "o", "", "output format (json, yaml, text)")

	return ans
}
