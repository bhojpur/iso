package docker

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
	"github.com/bhojpur/iso/pkg/manager/helpers/docker/api"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/urfave/cli"
)

func SquashImage(c *cli.Context) error {

	var sourceImage string
	var outputImage string

	client, err := api.NewDocker()
	if err != nil {
		return cli.NewExitError("could not connect to the Docker daemon", 87)
	}

	if c.NArg() == 2 {
		sourceImage = c.Args()[0]
		outputImage = c.Args()[1]
		jww.DEBUG.Println("sourceImage " + sourceImage + " outputImage: " + outputImage)
	} else if c.NArg() == 1 {
		sourceImage = c.Args()[0]
		outputImage = sourceImage
		jww.WARN.Println("You didn't specified a second image, i'll squash the one you supplied.")
		if c.Bool("remove") == false {
			jww.WARN.Println("!!! Be careful, docker will leave an image tagged as <none> which is your old one. Use the --remove option to remove it automatically")
		}
		jww.DEBUG.Println("sourceImage " + sourceImage + " outputImage: " + outputImage)
		oldImage, err := client.InspectImage(sourceImage)
		if c.Bool("remove") == true && err == nil {
			defer func(id string) {
				jww.INFO.Println("Removing the untagged image left by the overwrite ID: " + id)
				client.RemoveImage(id)
			}(oldImage.ID)
		}
	} else {
		return cli.NewExitError("This command requires two arguments: squash source-image output-image", 86)
	}

	if c.GlobalBool("pull") == true {
		api.PullImage(client, sourceImage)
	}
	jww.INFO.Println("Squashing " + sourceImage + " in " + outputImage)

	err = api.Squash(client, sourceImage, outputImage)
	if err == nil {
		jww.INFO.Println("Done")
	}
	return err
}
