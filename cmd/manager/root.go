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

	"github.com/bhojpur/iso/cmd/manager/util"
	bus "github.com/bhojpur/iso/pkg/manager/api/core/bus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var Verbose bool

const (
	BhojpurCLIVersion = "0.0.1"
	BhojpurEnvPrefix  = "BHOJPUR"
)

var license = []string{
	"Copyright (c) 2018 Bhojpur Consulting Private Limited, India.",
	"For documentation, visit https://docs.bhojpur.net",
}

// Build time and commit information.
//
// ⚠️ WARNING: should only be set by "-ldflags".
var (
	BuildTime   string
	BuildCommit string
)

func version() string {
	return fmt.Sprintf("%s-g%s %s", BhojpurCLIVersion, BuildCommit, BuildTime)
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "isomgr",
	Short: "Container based package manager",
	Long: `It is a single-binary package manager based on containers to build packages. 

For documentation, visit https://docs.bhojpur.net.
	
To install a package:

	$ isomgr install package

To search for a package in the repositories:

$ isomgr search package

To list all packages installed in the system:

	$ isomgr search --installed .

To show hidden packages:

	$ isomgr search --hidden package

To build a package, from a tree definition:

	$ isomgr build --tree tree/path package
	
`,
	Version: version(),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		ctx, err := util.InitContext(cmd)
		if err != nil {
			fmt.Println("failed to load configuration:", err.Error())
			os.Exit(1)
		}

		util.DefaultContext = ctx

		util.DisplayVersionBanner(util.DefaultContext, version, license)

		viper.BindPFlag("plugin", cmd.Flags().Lookup("plugin"))

		plugin := viper.GetStringSlice("plugin")

		bus.Manager.Initialize(util.DefaultContext, plugin...)
		if len(bus.Manager.Plugins) != 0 {
			util.DefaultContext.Info(":lollipop:Enabled plugins:")
			for _, p := range bus.Manager.Plugins {
				util.DefaultContext.Info(fmt.Sprintf("\t:arrow_right: %s (at %s)", p.Name, p.Executable))
			}
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Cleanup all tmp directories used by Bhojpur ISO
		err := util.DefaultContext.Clean()
		if err != nil {
			util.DefaultContext.Warning("failed on cleanup tmpdir:", err.Error())
		}
	},
	SilenceErrors: true,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	util.HandleLock()

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	util.InitViper(RootCmd)
}
