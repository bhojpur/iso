package util

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

	"github.com/marcsauter/single"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bhojpur/iso/pkg/manager/api/core/context"
	"github.com/bhojpur/iso/pkg/manager/api/core/template"
	"github.com/bhojpur/iso/pkg/manager/installer"
)

var lockedCommands = []string{"install", "uninstall", "upgrade"}
var bannerCommands = []string{"install", "build", "uninstall", "upgrade"}

func BindValuesFlags(cmd *cobra.Command) {
	viper.BindPFlag("values", cmd.Flags().Lookup("values"))
}

func ValuesFlags() []string {
	return viper.GetStringSlice("values")
}

// TemplateFolders returns the default folders which holds shared template between packages in a given tree path
func TemplateFolders(ctx *context.Context, i installer.BuildTreeResult, treePaths []string) []string {
	templateFolders := []string{}
	for _, t := range treePaths {
		templateFolders = append(templateFolders, template.FindPossibleTemplatesDir(t)...)
	}
	for _, r := range i.TemplatesDir {
		templateFolders = append(templateFolders, r...)
	}

	return templateFolders
}

func HandleLock() {
	if os.Getenv("BHOJPUR_ISO_NOLOCK") == "true" {
		return
	}

	if len(os.Args) == 0 {
		return
	}

	for _, lockedCmd := range lockedCommands {
		if os.Args[1] == lockedCmd {
			s := single.New("isomgr")
			if err := s.CheckLock(); err != nil && err == single.ErrAlreadyRunning {
				fmt.Println("another instance of the app is already running, exiting")
				os.Exit(1)
			} else if err != nil {
				// Another error occurred, might be worth handling it as well
				fmt.Println("failed to acquire exclusive app lock:", err.Error())
				os.Exit(1)
			}
			defer s.TryUnlock()
			break
		}
	}
}

func DisplayVersionBanner(c *context.Context, version func() string, license []string) {
	display := false
	if len(os.Args) > 1 {
		for _, c := range bannerCommands {
			if os.Args[1] == c {
				display = true
			}
		}
	}
	if display {
		pterm.Info.Printf("Bhojpur ISO manager %s\n", version())
		pterm.Info.Println(strings.Join(license, "\n"))
	}
}
