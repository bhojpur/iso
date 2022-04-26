package burner

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

	"github.com/bhojpur/iso/pkg/schema"
	"github.com/twpayne/go-vfs"
)

func copyConfig(config, rootfsWanted string, fs vfs.FS, s *schema.SystemSpec) error {
	//	cfg, _ := fs.RawPath(s.Bhojpur.Config)
	repos, err := s.Bhojpur.Repositories.Marshal()
	if err != nil {
		return err
	}
	rootfs, _ := fs.RawPath(rootfsWanted)

	// input, err := ioutil.ReadFile(cfg)
	// if err != nil {
	// 	return err
	// }
	input := []byte(repos + "\n" +
		`
system:
  rootfs: ` + rootfs + `
  database_path: "/isodb"
  database_engine: "boltdb"
repos_confdir:
  - ` + rootfs + `/etc/bhojpur/repos.conf.d
` + "\n")
	err = fs.WriteFile(config, input, 0644)
	if err != nil {
		return err
	}

	return nil
}

func BhojpurImageUnpack(image, destination string) error {
	return runEnv(fmt.Sprintf("isomgr util unpack %s %s", image, destination))
}

func BhojpurInstall(rootfs string, packages []string, repositories []string, keepDB bool, fs vfs.FS, spec *schema.SystemSpec) error {
	cfgFile := filepath.Join(rootfs, "iso.yaml")
	cfgRaw, _ := fs.RawPath(cfgFile)

	if err := copyConfig(cfgFile, rootfs, fs, spec); err != nil {
		return err
	}

	if len(repositories) > 0 {
		if err := run(fmt.Sprintf("isomgr install --no-spinner --config %s %s", cfgRaw, strings.Join(repositories, " "))); err != nil {
			return err
		}
	}

	if len(packages) > 0 {
		if err := run(fmt.Sprintf("isomgr install --no-spinner --config %s %s", cfgRaw, strings.Join(packages, " "))); err != nil {
			return err
		}
	}

	if err := run(fmt.Sprintf("isomgr --config %s cleanup", cfgRaw)); err != nil {
		return err
	}

	if keepDB {
		if err := vfs.MkdirAll(fs, filepath.Join(rootfs, "var", "bhojpur"), os.ModePerm); err != nil {
			return err
		}
		if _, err := fs.Stat(filepath.Join(rootfs, "var", "bhojpur", "db")); err == nil {
			fs.RemoveAll(filepath.Join(rootfs, "var", "bhojpur", "db"))
		}
		fs.Rename(filepath.Join(rootfs, "isodb"), filepath.Join(rootfs, "var", "bhojpur", "db"))
	} else {
		fs.RemoveAll(filepath.Join(rootfs, "isodb"))
	}
	fs.Remove(cfgFile)
	fs.Remove(filepath.Join(rootfs, "bhojpur", "repos.conf.d"))
	return nil
}
