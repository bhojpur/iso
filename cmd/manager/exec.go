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

	b64 "encoding/base64"

	"github.com/bhojpur/iso/cmd/manager/util"
	"github.com/bhojpur/iso/pkg/manager/box"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec --rootfs /path [command]",
	Short: "Execute a command in the rootfs context",
	Long:  `Uses unshare technique and pivot root to execute a command inside a folder containing a valid rootfs`,
	PreRun: func(cmd *cobra.Command, args []string) {
	},
	// If you change this, look at pkg/box/exec that runs this command and adapt
	Run: func(cmd *cobra.Command, args []string) {

		stdin, _ := cmd.Flags().GetBool("stdin")
		stdout, _ := cmd.Flags().GetBool("stdout")
		stderr, _ := cmd.Flags().GetBool("stderr")
		rootfs, _ := cmd.Flags().GetString("rootfs")
		base, _ := cmd.Flags().GetBool("decode")

		entrypoint, _ := cmd.Flags().GetString("entrypoint")
		envs, _ := cmd.Flags().GetStringArray("env")
		mounts, _ := cmd.Flags().GetStringArray("mount")

		if base {
			var ss []string
			for _, a := range args {
				sDec, _ := b64.StdEncoding.DecodeString(a)
				ss = append(ss, string(sDec))
			}
			//If the command to run is complex,using base64 to avoid bad input

			args = ss
		}
		util.DefaultContext.Info("Executing", args, "in", rootfs)

		b := box.NewBox(entrypoint, args, mounts, envs, rootfs, stdin, stdout, stderr)
		err := b.Exec()
		if err != nil {
			util.DefaultContext.Fatal(errors.Wrap(err, fmt.Sprintf("entrypoint: %s rootfs: %s", entrypoint, rootfs)))
		}
	},
}

func init() {
	path, err := os.Getwd()
	if err != nil {
		util.DefaultContext.Fatal(err)
	}
	execCmd.Hidden = true
	execCmd.Flags().String("rootfs", path, "Rootfs path")
	execCmd.Flags().Bool("stdin", false, "Attach to stdin")
	execCmd.Flags().Bool("stdout", false, "Attach to stdout")
	execCmd.Flags().Bool("stderr", false, "Attach to stderr")
	execCmd.Flags().Bool("decode", false, "Base64 decode")

	execCmd.Flags().StringArrayP("env", "e", []string{}, "Environment settings")
	execCmd.Flags().StringArrayP("mount", "m", []string{}, "List of paths to bind-mount from the host")

	execCmd.Flags().String("entrypoint", "/bin/sh", "Entrypoint command (/bin/sh)")

	RootCmd.AddCommand(execCmd)
}
