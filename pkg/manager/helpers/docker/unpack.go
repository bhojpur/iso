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
	"strconv"
	"time"

	"github.com/bhojpur/iso/pkg/manager/helpers/docker/api"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/urfave/cli"
)

func UnpackImage(c *cli.Context) error {

	var sourceImage string
	var output string
	if c.NArg() == 2 {
		sourceImage = c.Args()[0]
		output = c.Args()[1]
	} else {
		return cli.NewExitError("This command requires to argument: source-image output-folder(absolute)", 86)
	}
	client, err := api.NewDocker()
	if err != nil {
		return cli.NewExitError("could not connect to the Docker daemon", 87)
	}
	if c.GlobalBool("pull") == true {
		api.PullImage(client, sourceImage)
	}

	if c.Bool("squash") == true {
		jww.INFO.Println("Squashing and unpacking " + sourceImage + " in " + output)
		time := strconv.Itoa(int(makeTimestamp()))
		api.Squash(client, sourceImage, sourceImage+"-tmpsquashed"+time)
		sourceImage = sourceImage + "-tmpsquashed" + time
		defer func() {
			jww.INFO.Println("Removing squashed image " + sourceImage)
			client.RemoveImage(sourceImage)
		}()
	}

	jww.INFO.Println("Unpacking " + sourceImage + " in " + output)
	err = api.Unpack(client, sourceImage, output, c.GlobalBool("fatal"))
	if err == nil {
		jww.INFO.Println("Done")
	}
	return err
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
