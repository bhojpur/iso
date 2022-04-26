package helpers

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
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/moby/moby/pkg/archive"
)

func Tar(src, dest string) error {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	fs, err := archive.Tar(src, archive.Uncompressed)
	if err != nil {
		return err
	}
	defer fs.Close()

	_, err = io.Copy(out, fs)
	if err != nil {
		return err
	}

	err = out.Sync()
	if err != nil {
		return err
	}
	return err
}

type TarModifierWrapperFunc func(path, dst string, header *tar.Header, content io.Reader) (*tar.Header, []byte, error)
type TarModifierWrapper struct {
	DestinationPath string
	Modifier        TarModifierWrapperFunc
}

func NewTarModifierWrapper(dst string, modifier TarModifierWrapperFunc) *TarModifierWrapper {
	return &TarModifierWrapper{
		DestinationPath: dst,
		Modifier:        modifier,
	}
}

func (m *TarModifierWrapper) GetModifier() archive.TarModifierFunc {
	return func(path string, header *tar.Header, content io.Reader) (*tar.Header, []byte, error) {
		return m.Modifier(m.DestinationPath, path, header, content)
	}
}

func UntarProtect(src, dst string, sameOwner bool, protectedFiles []string, modifier *TarModifierWrapper) error {
	var ans error

	if len(protectedFiles) <= 0 {
		return Untar(src, dst, sameOwner)
	}

	// POST: we have files to protect. I create a ReplaceFileTarWrapper
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// Create modifier map
	mods := make(map[string]archive.TarModifierFunc)
	for _, file := range protectedFiles {
		mods[file] = modifier.GetModifier()
	}

	if sameOwner {
		// we do have root permissions, so we can extract keeping the same permissions.
		replacerArchive := archive.ReplaceFileTarWrapper(in, mods)

		opts := &archive.TarOptions{
			NoLchown:        false,
			ExcludePatterns: []string{"dev/"}, // prevent 'operation not permitted'
			//ContinueOnError: true,
		}

		ans = archive.Untar(replacerArchive, dst, opts)
	} else {
		ans = unTarIgnoreOwner(dst, in, mods)
	}

	return ans
}

func unTarIgnoreOwner(dest string, in io.ReadCloser, mods map[string]archive.TarModifierFunc) error {
	tr := tar.NewReader(in)
	for {
		header, err := tr.Next()

		var data []byte
		var headerReplaced = false

		switch {
		case err == io.EOF:
			goto tarEof
		case err != nil:
			return err
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dest, header.Name)
		if mods != nil {
			modifier, ok := mods[header.Name]
			if ok {
				header, data, err = modifier(header.Name, header, tr)
				if err != nil {
					return err
				}

				// Override target path
				target = filepath.Join(dest, header.Name)
				headerReplaced = true
			}

		}

		// Check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

			// handle creation of file
		case tar.TypeReg:

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if headerReplaced {
				_, err = io.Copy(f, bytes.NewReader(data))
			} else {
				_, err = io.Copy(f, tr)
			}
			if err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each
			// file close to wait until all operations have completed.
			f.Close()

		case tar.TypeSymlink:
			source := header.Linkname
			err := os.Symlink(source, target)
			if err != nil {
				return err
			}
		}
	}
tarEof:

	return nil
}

// Untar just a wrapper around the docker functions
func Untar(src, dest string, sameOwner bool) error {
	var ans error

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if sameOwner {
		opts := &archive.TarOptions{
			NoLchown:        false,
			ExcludePatterns: []string{"dev/"}, // prevent 'operation not permitted'
			//ContinueOnError: true,
		}

		ans = archive.Untar(in, dest, opts)
	} else {
		ans = unTarIgnoreOwner(dest, in, nil)
	}

	return ans
}
