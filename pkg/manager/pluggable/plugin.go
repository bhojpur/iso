package pluggable

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
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

// Plugin describes binaries to be hooked on events, with common js input, and common js output
type Plugin struct {
	Name       string
	Executable string
}

// A safe threshold to avoid unpleasant exec buffer fill for argv too big. Seems 128K is the limit on Linux.
const maxMessageSize = 1 << 13

// Run runs the Event on the plugin, and returns an EventResponse
func (p Plugin) Run(e Event) (EventResponse, error) {
	r := EventResponse{}

	eventToprocess := &e

	if len(e.Data) > maxMessageSize {
		copy := e.Copy()
		copy.Data = ""
		f, err := ioutil.TempFile(os.TempDir(), "pluggable")
		if err != nil {
			return r, errors.Wrap(err, "while creating temporary file")
		}
		if err := ioutil.WriteFile(f.Name(), []byte(e.Data), os.ModePerm); err != nil {
			return r, errors.Wrap(err, "while writing to temporary file")
		}
		copy.File = f.Name()
		eventToprocess = copy
		defer os.RemoveAll(f.Name())
	}

	k, err := eventToprocess.JSON()
	if err != nil {
		return r, errors.Wrap(err, "while marshalling event")
	}
	cmd := exec.Command(p.Executable, string(e.Name), k)
	cmd.Env = os.Environ()
	var b bytes.Buffer
	cmd.Stderr = &b
	out, err := cmd.Output()
	if err != nil {
		r.Error = "error while executing plugin: " + err.Error() + string(b.String())
		return r, errors.Wrap(err, "while executing plugin: "+string(b.String()))
	}

	if err := json.Unmarshal(out, &r); err != nil {
		r.Error = err.Error()
		return r, errors.Wrap(err, "while unmarshalling response")
	}
	return r, nil
}
