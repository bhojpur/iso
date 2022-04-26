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
	"path/filepath"
	"strings"

	"github.com/bhojpur/iso/pkg/burner"
	"github.com/bhojpur/iso/pkg/schema"
	"github.com/bhojpur/iso/pkg/version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/twpayne/go-vfs"
)

func fail(s string) {
	log.Error(s)
	os.Exit(1)
}
func checkErr(err error) {
	if err != nil {
		fail("fatal error: " + err.Error())
	}
}

func init() {
	switch strings.ToLower(os.Getenv("LOGLEVEL")) {
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warning":
		log.SetLevel(log.WarnLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "isomake",
	Short:   "generate ISO images using Bhojpur ISO manager tools",
	Version: fmt.Sprintf("%s-g%s %s", version.Version, version.BuildCommit, version.BuildTime),
	Long: `It reads specifications to generate ISO image files from Bhojpur ISO repositories or trees.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Error("One argument (i.e. ISO file specification) required")
			os.Exit(1)
		}

		localPath, _ := cmd.Flags().GetString("local")
		image, _ := cmd.Flags().GetString("image")
		output, _ := cmd.Flags().GetString("output")

		if localPath != "" && !filepath.IsAbs(localPath) {
			var err error
			localPath, err = filepath.Abs(localPath)
			checkErr(err)
		}

		for _, a := range args {
			spec, err := schema.LoadFromFile(a, vfs.OSFS)
			checkErr(err)

			if image != "" {
				spec.RootfsImage = image
			}
			if output != "" {
				spec.ImageName = output
				spec.Date = false
				spec.ImagePrefix = ""
			}

			if localPath != "" {
				spec.Bhojpur.Repositories = append(spec.Bhojpur.Repositories, schema.NewLocalRepo("local", localPath))
			}
			checkErr(burner.Burn(spec, vfs.OSFS))
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	checkErr(err)
}

func init() {
	rootCmd.Flags().StringP("local", "l", "", "A path to a local Bhojpur ISO repository to use during ISO build")
	rootCmd.Flags().StringP("image", "i", "", "An image reference to use as a rootfs for the ISO")
	rootCmd.Flags().StringP("output", "o", "", "Name of the output ISO file (overrides yaml config)")
}
