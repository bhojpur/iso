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
	"encoding/json"
	"io"
	"io/ioutil"
)

type FactoryPlugin struct {
	EventType     EventType
	PluginHandler PluginHandler
}

func NewPluginFactory(p ...FactoryPlugin) PluginFactory {
	f := make(PluginFactory)
	for _, pp := range p {
		f.Add(pp.EventType, pp.PluginHandler)
	}
	return f
}

// PluginHandler represent a generic plugin which
// talks go-pluggable API
// It receives an event, and is always expected to give a response
type PluginHandler func(*Event) EventResponse

// PluginFactory is a collection of handlers for a given event type.
// a plugin has to respond to multiple events and it always needs to return an
// Event response as result
type PluginFactory map[EventType]PluginHandler

// Run runs the PluginHandler given a event type and a payload
//
// The result is written to the writer provided
// as argument.
func (p PluginFactory) Run(name EventType, payload string, w io.Writer) error {
	ev := &Event{}

	if err := json.Unmarshal([]byte(payload), ev); err != nil {
		return err
	}

	if ev.File != "" {
		c, err := ioutil.ReadFile(ev.File)
		if err != nil {
			return err
		}

		ev.Data = string(c)
	}

	resp := EventResponse{}
	for e, r := range p {
		if name == e {
			resp = r(ev)
		}
	}

	dat, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	_, err = w.Write(dat)
	return err
}

// Add associates an handler to an event type
func (p PluginFactory) Add(ev EventType, ph PluginHandler) {
	p[ev] = ph
}
