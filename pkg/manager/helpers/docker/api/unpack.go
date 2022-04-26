package api

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
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/docker/docker/pkg/archive"
	docker "github.com/fsouza/go-dockerclient"
	jww "github.com/spf13/jwalterweatherman"
)

// SEPARATOR contains system-specific separator
const SEPARATOR = string(filepath.Separator)

// ROOTFS is our temporary rootfs path
const ROOTFS = "." + SEPARATOR + "rootfs_overlay"

// Unpack unpacks a docker image into a path
func Unpack(client *docker.Client, image string, dirname string, fatal bool) error {
	var err error
	r, w := io.Pipe()

	if dirname == "" {
		dirname = ROOTFS
	}

	os.MkdirAll(dirname, 0777)

	filename, err := ioutil.TempFile(os.TempDir(), "artemide")
	if err != nil {
		return fmt.Errorf("Couldn't create the temporary file")
	}
	os.Remove(filename.Name())

	jww.INFO.Println("Creating container")

	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: image,
			Cmd:   []string{"true"},
		},
	})
	if err != nil {
		jww.FATAL.Fatalln("Couldn't export container, sorry", err)
	}
	defer func(*docker.Container) {
		client.RemoveContainer(docker.RemoveContainerOptions{
			ID:    container.ID,
			Force: true,
		})
	}(container)

	signalchan := make(chan os.Signal, 1)
	signal.Notify(signalchan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		for {
			s := <-signalchan
			switch s {

			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				jww.WARN.Println("SIGTERM/SIGINT/SIGQUIT detected, removing pending containers")
				client.RemoveContainer(docker.RemoveContainerOptions{
					ID:    container.ID,
					Force: true,
				})
			}
		}
	}()

	// writing without a reader will deadlock so write in a goroutine
	go func() {
		// it is important to close the writer or reading from the other end of the
		// pipe will never finish
		defer w.Close()
		err := client.ExportContainer(docker.ExportContainerOptions{ID: container.ID, OutputStream: w})
		if err != nil {
			jww.FATAL.Fatalln("Couldn't export container, sorry", err)
		}

	}()

	jww.INFO.Println("Extracting to", dirname)

	err = Untar(r, dirname, true)
	if err != nil {
		return fmt.Errorf("could not unpack to " + dirname)
	}
	err = prepareRootfs(dirname, fatal)

	return err
}

func prepareRootfs(dirname string, fatal bool) error {

	_, err := os.Stat(dirname + SEPARATOR + ".dockerenv")
	if err == nil {
		err = os.Remove(dirname + SEPARATOR + ".dockerenv")
		if err != nil {
			if fatal == true {
				return fmt.Errorf("could not remove docker env file")
			} else {
				jww.WARN.Println("error on remove .dockerenv, extracting anyway")
			}
		}
	}

	_, err = os.Stat(dirname + SEPARATOR + ".dockerinit")
	if err == nil {
		err = os.Remove(dirname + SEPARATOR + ".dockerinit")
		if err != nil {
			if fatal == true {
				return fmt.Errorf("could not remove docker init file")
			} else {
				jww.WARN.Println("error on remove .dockerinit, extracting anyway")
			}
		}
	}

	err = os.MkdirAll(dirname+SEPARATOR+"dev", 0751)
	if err != nil {
		if fatal == true {
			return fmt.Errorf("could not create dev folder")
		} else {
			jww.WARN.Println("could not create dev folder")
		}
	}

	// Google DNS as default
	d1 := []byte("nameserver 8.8.8.8\nnameserver 8.8.4.4\n")
	err = ioutil.WriteFile(dirname+SEPARATOR+"etc"+SEPARATOR+"resolv.conf", d1, 0644)
	if err != nil {
		if fatal == true {
			return fmt.Errorf("could not write resolv.conf file")
		} else {
			jww.WARN.Println("could not create resolv.conf file")
		}
	}

	return nil
}

// Untar just a wrapper around the docker functions
func Untar(in io.Reader, dest string, sameOwner bool) error {
	return archive.Untar(in, dest, &archive.TarOptions{
		NoLchown:        !sameOwner,
		ExcludePatterns: []string{"dev/"}, // prevent 'operation not permitted'
	})
}