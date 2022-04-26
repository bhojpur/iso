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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chuckpreslar/emission"
	"github.com/pkg/errors"
)

// Manager describes a set of Plugins and
// a set of Event types which are subscribed to a message bus
type Manager struct {
	Plugins []Plugin
	Events  []EventType
	Bus     *emission.Emitter
}

// NewManager returns a manager instance with a new bus and
func NewManager(events []EventType) *Manager {
	return &Manager{
		Events: events,
		Bus:    emission.NewEmitter(),
	}
}

// Register subscribes the plugin to its internal bus
func (m *Manager) Register() *Manager {
	m.Subscribe(m.Bus)
	return m
}

// Publish is a wrapper around NewEvent and the Manager internal Bus publishing system
// It accepts optionally a list of functions that are called with the plugin result (only once)
func (m *Manager) Publish(event EventType, obj interface{}) (*Manager, error) {
	ev, err := NewEvent(event, obj)
	if err == nil && ev != nil {
		m.Bus.Emit(string(ev.Name), ev)
	}
	return m, err
}

// Response binds a set of listeners to an event type. The listeners are called for each result from
// every plugin when Publish is called.
func (m *Manager) Response(event EventType, listener ...func(p *Plugin, r *EventResponse)) *Manager {
	ev, _ := NewEvent(event, nil)
	for _, l := range listener {
		m.Bus.On(string(ev.ResponseEventName("results")), l)
	}
	return m
}

func (m *Manager) propagateEvent(p Plugin) func(e *Event) {
	return func(e *Event) {
		resp, err := p.Run(*e)
		r := &resp
		if err != nil && !resp.Errored() {
			resp.Error = err.Error()
		}
		m.Bus.Emit(string(e.ResponseEventName("results")), &p, r)
	}
}

// Subscribe subscribes the plugin to the events in the given bus
func (m *Manager) Subscribe(b *emission.Emitter) *Manager {
	for _, p := range m.Plugins {
		for _, e := range m.Events {
			b.On(string(e), m.propagateEvent(p))
		}
	}
	return m
}

func relativeToCwd(p string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "failed getting current work directory")
	}

	return filepath.Join(cwd, p), nil
}

func (m *Manager) insertPlugin(p Plugin) {
	for _, i := range m.Plugins {
		// We don't want any ambiguity here.
		// Binary plugins must be unique in PATH and Name
		if i.Executable == p.Executable || i.Name == p.Name {
			return
		}
	}
	m.Plugins = append(m.Plugins, p)
}

// Autoload automatically loads plugins binaries prefixed by 'prefix' in the current path
// optionally takes a list of paths to look also into
func (m *Manager) Autoload(prefix string, extensionpath ...string) *Manager {
	projPrefix := fmt.Sprintf("%s-", prefix)
	paths := strings.Split(os.Getenv("PATH"), ":")

	for _, path := range extensionpath {
		if filepath.IsAbs(path) {
			paths = append(paths, path)
			continue
		}

		rel, err := relativeToCwd(path)
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
		for _, ma := range matches {
			short := strings.TrimPrefix(filepath.Base(ma), projPrefix)
			m.insertPlugin(Plugin{Name: short, Executable: ma})
		}
	}
	return m
}

// Load finds the binaries given as parameter (without path) and scan the system $PATH to retrieve those automatically
func (m *Manager) Load(extname ...string) *Manager {
	paths := strings.Split(os.Getenv("PATH"), ":")

	for _, p := range paths {
		for _, n := range extname {
			path := filepath.Join(p, n)
			_, err := os.Lstat(path)
			if err != nil {
				continue
			}
			m.insertPlugin(Plugin{Name: n, Executable: path})
		}
	}
	return m
}
