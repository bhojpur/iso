package extensions

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

	"github.com/spf13/cobra"
)

type Extension struct {
	AbsPath   string
	ShortName string
}

func (e Extension) String() string {
	return e.ShortName
}

func (e Extension) Short() string {
	return e.ShortName
}

func (e Extension) Path() string {
	return e.AbsPath
}

func (e Extension) CobraCommand() *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s --help", e.Short()),
		Short: fmt.Sprintf("extension: %s (run to show the extension helper)", e.Short()),
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return e.Exec(args)
		}}
}

// Discover returns extensions found in the paths specified and in PATH
// Extensions must start with the project tag (e.g. 'myawesomecli-' )
func Discover(project string, extensionpath ...string) []ExtensionInterface {
	var result []ExtensionInterface

	// by convention, extensions paths must have a prefix with the name of the project
	// e.g. 'foo-ext1' 'foo-ext2'
	projPrefix := fmt.Sprintf("%s-", project)
	paths := strings.Split(os.Getenv("PATH"), ":")

	for _, path := range extensionpath {
		if filepath.IsAbs(path) {
			paths = append(paths, path)
			continue
		}

		rel, err := RelativeToCwd(path)
		if err != nil {
			continue
		}
		paths = append(paths, rel)
	}

	for _, p := range paths {
		matches, err := filepath.Glob(filepath.Join(p, fmt.Sprintf("%s*", projPrefix)))
		if err != nil {
			continue
		}
		for _, m := range matches {
			short := strings.TrimPrefix(filepath.Base(m), projPrefix)
			result = append(result, Extension{AbsPath: m, ShortName: short})
		}
	}
	return result
}
