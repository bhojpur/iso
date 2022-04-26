package main

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
	"path/filepath"
	"regexp"

	"github.com/bhojpur/iso/pkg/manager/helpers/docker"
	"github.com/bhojpur/iso/pkg/version"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/urfave/cli"
)

var released = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)

var appHelpTemplate = `Usage: {{.Name}} {{if .Flags}}[OPTIONS] {{end}}COMMAND [arg...]
{{.Usage}}
Version: {{.Version}}{{if or .Author .Email}}
Author:{{if .Author}}
  {{.Author}}{{if .Email}} - <{{.Email}}>{{end}}{{else}}
  {{.Email}}{{end}}{{end}}
{{if .Flags}}
Options:
  {{range .Flags}}{{.}}
  {{end}}{{end}}
Commands:
  {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
  {{end}}
Run '{{.Name}} COMMAND --help' for more information on a command.
`

var commandHelpTemplate = `Usage: isocntr {{.Name}}{{if .Flags}} [OPTIONS]{{end}} [arg...]
{{.Usage}}{{if .Description}}
Description:
   {{.Description}}{{end}}{{if .Flags}}
Options:
   {{range .Flags}}
   {{.}}{{end}}{{ end }}
`

func main() {
	cli.AppHelpTemplate = appHelpTemplate
	cli.CommandHelpTemplate = commandHelpTemplate

	logrus.SetOutput(colorable.NewColorableStdout())

	if err := mainErr(); err != nil {
		logrus.Fatal(err)
	}
}

func mainErr() error {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Version = version.FullVersion()

	app.Author = "Bhojpur Consulting Private Limited, India"
	app.Email = "https://www.bhojpur-consulting.com"

	app.Usage = "Bhojpur CLI tool for Docker container download, unpack, squash"
	jww.SetStdoutThreshold(jww.LevelInfo)
	if os.Getenv("DEBUG") == "1" {
		jww.SetStdoutThreshold(jww.LevelDebug)
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "pull",
			Usage: "pull image before doing operations",
		},
		cli.BoolFlag{
			Name:  "fatal",
			Usage: "threat errors as fatal",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "download",
			Aliases: []string{"dl"},
			Usage:   "Download and unpacks an image without using Docker - Usage: download foo/barimage /foobar/folder",
			Action:  docker.DownloadImage,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "keep",
					Usage: "Keeps downloaded layers around (useful for debugging)",
				},
			},
		},
		{
			Name:    "unpack",
			Aliases: []string{"un"},
			Usage:   "unpack the specified Docker image content as-is (run as root!) in a folder - Usage: unpack foo/barimage /foobar/folder",
			Action:  docker.UnpackImage,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "squash",
					Usage: "squash image before doing operations",
				},
			},
		},
		{
			Name:    "squash",
			Aliases: []string{"s"},
			Usage:   "squash the Docker image (loosing metadata) into another - Usage: squash foo/bar foo/bar-squashed:latest. The second argument is optional",
			Action:  docker.SquashImage,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "remove, rm",
					Usage: "If you supplied just one image, remove the untagged image",
				},
			},
		},
	}

	app.Run(os.Args)
	return nil
}
