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
	"net/http"
	"os"

	"github.com/bhojpur/iso/cmd/manager/util"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serverepoCmd = &cobra.Command{
	Use:   "serve-repo",
	Short: "Embedded micro-http server",
	Long:  `Embedded mini http server for serving local repositories`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("dir", cmd.Flags().Lookup("dir"))
		viper.BindPFlag("address", cmd.Flags().Lookup("address"))
		viper.BindPFlag("port", cmd.Flags().Lookup("port"))
	},
	Run: func(cmd *cobra.Command, args []string) {

		dir := viper.GetString("dir")
		port := viper.GetString("port")
		address := viper.GetString("address")

		http.Handle("/", http.FileServer(http.Dir(dir)))

		util.DefaultContext.Info("Serving ", dir, " on HTTP port: ", port)
		util.DefaultContext.Fatal(http.ListenAndServe(address+":"+port, nil))
	},
}

func init() {
	path, err := os.Getwd()
	if err != nil {
		util.DefaultContext.Fatal(err)
	}
	serverepoCmd.Flags().String("dir", path, "Packages folder (output from build)")
	serverepoCmd.Flags().String("port", "9090", "Listening port")
	serverepoCmd.Flags().String("address", "0.0.0.0", "Listening address")

	RootCmd.AddCommand(serverepoCmd)
}
