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
	"runtime"

	"github.com/bhojpur/iso/pkg/manager/api/core/image"
	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"
	"github.com/docker/docker/api/types"
	"github.com/docker/go-units"
	"github.com/pkg/errors"

	"github.com/bhojpur/iso/cmd/manager/util"
	"github.com/bhojpur/iso/pkg/manager/api/core/context"
	"github.com/bhojpur/iso/pkg/manager/helpers/docker"

	"github.com/spf13/cobra"
)

func pack(ctx *context.Context, p, dst, imageName, arch, OS string) error {

	tempimage, err := ctx.TempFile("tempimage")
	if err != nil {
		return errors.Wrap(err, "error met while creating tempdir for "+p)
	}
	defer os.RemoveAll(tempimage.Name()) // clean up

	if err := image.CreateTar(p, tempimage.Name(), imageName, arch, OS); err != nil {
		return errors.Wrap(err, "could not create image from tar")
	}

	return fileHelper.CopyFile(tempimage.Name(), dst)
}

func NewPackCommand() *cobra.Command {

	c := &cobra.Command{
		Use:   "pack image src.tar dst.tar",
		Short: "Pack a standard tar archive as a container image",
		Long: `Pack creates a tar which can be loaded as an image from a standard flat tar archive, for e.g. with docker load. 
It doesn't need the docker daemon to run, and allows to override default os/arch:
		
	isomgr util pack --os arm64 image:tag src.tar dst.tar
`,
		Args: cobra.MinimumNArgs(3),
		Run: func(cmd *cobra.Command, args []string) {

			image := args[0]
			src := args[1]
			dst := args[2]

			arch, _ := cmd.Flags().GetString("arch")
			os, _ := cmd.Flags().GetString("os")

			err := pack(util.DefaultContext, src, dst, image, arch, os)
			if err != nil {
				util.DefaultContext.Fatal(err.Error())
			}
			util.DefaultContext.Info("Image packed as", image)
		},
	}

	c.Flags().String("arch", runtime.GOARCH, "Image architecture")
	c.Flags().String("os", runtime.GOOS, "Image OS")

	return c
}

func NewUnpackCommand() *cobra.Command {

	c := &cobra.Command{
		Use:   "unpack image path",
		Short: "Unpack a docker image natively",
		Long: `unpack doesn't need the docker daemon to run, and unpacks a docker image in the specified directory:

	isomgr util unpack golang:alpine /alpine
`,
		PreRun: func(cmd *cobra.Command, args []string) {

			if len(args) != 2 {
				util.DefaultContext.Fatal("Expects an image and a path")
			}

		},
		Run: func(cmd *cobra.Command, args []string) {

			image := args[0]
			destination, err := filepath.Abs(args[1])
			if err != nil {
				util.DefaultContext.Error("Invalid path %s", destination)
				os.Exit(1)
			}
			local, _ := cmd.Flags().GetBool("local")
			verify, _ := cmd.Flags().GetBool("verify")
			user, _ := cmd.Flags().GetString("auth-username")
			pass, _ := cmd.Flags().GetString("auth-password")
			authType, _ := cmd.Flags().GetString("auth-type")
			server, _ := cmd.Flags().GetString("auth-server-address")
			identity, _ := cmd.Flags().GetString("auth-identity-token")
			registryToken, _ := cmd.Flags().GetString("auth-registry-token")

			util.DefaultContext.Info("Downloading", image, "to", destination)
			auth := &types.AuthConfig{
				Username:      user,
				Password:      pass,
				ServerAddress: server,
				Auth:          authType,
				IdentityToken: identity,
				RegistryToken: registryToken,
			}

			if !local {
				info, err := docker.DownloadAndExtractDockerImage(util.DefaultContext, image, destination, auth, verify)
				if err != nil {
					util.DefaultContext.Error(err.Error())
					os.Exit(1)
				}
				util.DefaultContext.Info(fmt.Sprintf("Pulled: %s %s", info.Target.Digest, info.Name))
				util.DefaultContext.Info(fmt.Sprintf("Size: %s", units.BytesSize(float64(info.Target.Size))))
			} else {
				info, err := docker.ExtractDockerImage(util.DefaultContext, image, destination)
				if err != nil {
					util.DefaultContext.Error(err.Error())
					os.Exit(1)
				}
				util.DefaultContext.Info(fmt.Sprintf("Size: %s", units.BytesSize(float64(info.Target.Size))))
			}
		},
	}

	c.Flags().String("auth-username", "", "Username to authenticate to registry/notary")
	c.Flags().String("auth-password", "", "Password to authenticate to registry")
	c.Flags().String("auth-type", "", "Auth type")
	c.Flags().String("auth-server-address", "", "Authentication server address")
	c.Flags().String("auth-identity-token", "", "Authentication identity token")
	c.Flags().String("auth-registry-token", "", "Authentication registry token")
	c.Flags().Bool("verify", false, "Verify signed images to notary before to pull")
	c.Flags().Bool("local", false, "Unpack local image")
	return c
}

func NewExistCommand() *cobra.Command {

	c := &cobra.Command{
		Use:   "image-exist image path",
		Short: "Check if an image exist",
		Long:  `Exits 0 if the image exist, otherwise exits with 1`,
		PreRun: func(cmd *cobra.Command, args []string) {

			if len(args) != 1 {
				util.DefaultContext.Fatal("Expects an image")
			}

		},
		Run: func(cmd *cobra.Command, args []string) {
			if image.Available(args[0]) {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		},
	}

	return c
}

var utilGroup = &cobra.Command{
	Use:   "util [command] [OPTIONS]",
	Short: "General Bhojpur ISO internal utilities exposed",
}

func init() {
	RootCmd.AddCommand(utilGroup)

	utilGroup.AddCommand(
		NewUnpackCommand(),
		NewPackCommand(),
		NewExistCommand(),
	)
}
