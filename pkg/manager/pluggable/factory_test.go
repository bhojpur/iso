package pluggable_test

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
	"fmt"

	. "github.com/bhojpur/iso/pkg/manager/pluggable"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PluginFactory", func() {
	Context("creating plugins", func() {
		factory := NewPluginFactory()

		BeforeEach(func() {
			factory = NewPluginFactory()
		})

		It("reacts to events", func() {
			b := bytes.NewBufferString("")
			payload := &Event{Name: "foo", Data: "bar"}

			payloadDat, err := json.Marshal(payload)
			Expect(err).ToNot(HaveOccurred())
			factory.Add("foo", func(e *Event) EventResponse { return EventResponse{State: "foo", Data: fmt.Sprint(e.Data == "bar")} })
			err = factory.Run("foo", string(payloadDat), b)
			Expect(err).ToNot(HaveOccurred())

			Expect(b.String()).ToNot(BeEmpty())
			resp := &EventResponse{}
			err = json.Unmarshal(b.Bytes(), resp)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Data).To(Equal("true"))
		})
	})
})
