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
	"fmt"
)

// EventType describes an event type
type EventType string

// Event describes the event structure.
// Contains a Name field and a Data field which
// is marshalled in JSON
type Event struct {
	Name EventType `json:"name"`
	Data string    `json:"data"`
	File string    `json:"file"` // If Data >> 10K write content to file instead
}

// EventResponse describes the event response structure
// It represent the JSON response from plugins
type EventResponse struct {
	State string `json:"state"`
	Data  string `json:"data"`
	Error string `json:"error"`
}

// JSON returns the stringified JSON of the Event
func (e Event) JSON() (string, error) {
	dat, err := json.Marshal(e)
	return string(dat), err
}

// Copy returns a copy of Event
func (e Event) Copy() *Event {
	copy := &e
	return copy
}

func (e Event) ResponseEventName(s string) EventType {
	return EventType(fmt.Sprintf("%s-%s", e.Name, s))
}

// Unmarshal decodes the json payload in the given parameteer
func (r EventResponse) Unmarshal(i interface{}) error {
	return json.Unmarshal([]byte(r.Data), i)
}

// Errored returns true if the response contains an error
func (r EventResponse) Errored() bool {
	return len(r.Error) != 0
}

// NewEvent returns a new event which can be used for publishing
// the obj gets automatically serialized in json.
func NewEvent(name EventType, obj interface{}) (*Event, error) {
	dat, err := json.Marshal(obj)
	return &Event{Name: name, Data: string(dat)}, err
}
