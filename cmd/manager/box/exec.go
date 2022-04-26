package cmd_box

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
	"os"

	b64 "encoding/base64"

	"github.com/bhojpur/iso/cmd/manager/util"
	"github.com/bhojpur/iso/pkg/manager/box"

	"github.com/spf13/cobra"
)

func NewBoxExecCommand() *cobra.Command {
	var ans = &cobra.Command{
		Use:   "exec [OPTIONS]",
		Short: "Execute a binary in a box",
		Args:  cobra.OnlyValidArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
		},
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
			err := b.Run()
			if err != nil {
				util.DefaultContext.Fatal(err)
			}
		},
	}
	path, err := os.Getwd()
	if err != nil {
		util.DefaultContext.Fatal(err)
	}
	ans.Flags().String("rootfs", path, "Rootfs path")
	ans.Flags().Bool("stdin", false, "Attach to stdin")
	ans.Flags().Bool("stdout", true, "Attach to stdout")
	ans.Flags().Bool("stderr", true, "Attach to stderr")
	ans.Flags().Bool("decode", false, "Base64 decode")
	ans.Flags().StringArrayP("env", "e", []string{}, "Environment settings")
	ans.Flags().StringArrayP("mount", "m", []string{}, "List of paths to bind-mount from the host")

	ans.Flags().String("entrypoint", "/bin/sh", "Entrypoint command (/bin/sh)")

	return ans
}
