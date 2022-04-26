package image

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
	"os"
	"strings"

	"github.com/peterbourgon/diskv"
)

// Cache represents a key-value store which is capable to upgrade to disk when it
// reaches a pre-defined threshold.
type Cache struct {
	store                      *diskv.Diskv
	memory                     map[string]string
	dir                        string
	onDisk                     bool
	maxmemorySize, maxItemSize int
}

// New creates a new key-value cache
// the cache acts in memory as long as the maxItemsize is not reached.
// Once the threshold is met the cache is offloaded to disk automatically,
// with a buffer of maxmemorySize into memory.
func NewCache(path string, maxmemorySize, maxItemsize int) *Cache {
	disk := diskv.New(diskv.Options{
		BasePath:     path,
		CacheSizeMax: uint64(maxmemorySize), // 500MB
	})

	return &Cache{
		memory:        make(map[string]string),
		store:         disk,
		dir:           path,
		maxmemorySize: maxmemorySize,
		maxItemSize:   maxItemsize,
	}
}

// This is needed as the disk cache is merely stored as separate files
// thus we don't want to conflict file names with the path separator.
// XXX: This is inconvenient as while we are looping result we can't rely
// anymore originally to the key name.
// We don't do any hashing to avoid any performance impact
func cleanKey(s string) string {
	return strings.ReplaceAll(s, string(os.PathSeparator), "_")
}

// Count returns the items in the cache.
// If it's a disk cache might be an expensive call.
func (c *Cache) Count() int {
	if !c.onDisk {
		return len(c.memory)
	}

	count := 0
	for range c.store.Keys(nil) {
		count++
	}
	return count
}

// Get attempts to retrieve a value for a key
func (c *Cache) Get(key string) (value string, found bool) {

	if !c.onDisk {
		v, ok := c.memory[key]
		return v, ok
	}
	v, err := c.store.Read(cleanKey(key))
	if err == nil {
		found = true
	}
	value = string(v)
	return
}

func (c *Cache) flushToDisk() {
	for k, v := range c.memory {
		c.store.Write(cleanKey(k), []byte(v))
	}
	c.memory = make(map[string]string)
	c.onDisk = true
}

// Set updates or inserts a new value
func (c *Cache) Set(key, value string) error {

	if !c.onDisk && c.Count() >= c.maxItemSize && c.maxItemSize != 0 {
		c.flushToDisk()
	}

	if c.onDisk {
		return c.store.Write(cleanKey(key), []byte(value))
	}

	c.memory[key] = value

	return nil
}

// SetValue updates or inserts a new value by marshalling it into JSON.
func (c *Cache) SetValue(key string, value interface{}) error {
	dat, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Set(cleanKey(key), string(dat))
}

// CacheResult represent the key value result when
// iterating over the cache
type CacheResult struct {
	key, value string
}

// Value returns the underlying value
func (c CacheResult) Value() string {
	return c.value
}

// Key returns the cache result key
func (c CacheResult) Key() string {
	return c.key
}

// Unmarshal the result into the interface. Use it to retrieve data
// set with SetValue
func (c CacheResult) Unmarshal(i interface{}) error {
	return json.Unmarshal([]byte(c.Value()), i)
}

// Iterates over cache by key
func (c *Cache) All(fn func(CacheResult)) {
	if !c.onDisk {
		for k, v := range c.memory {
			fn(CacheResult{key: k, value: v})
		}
		return
	}

	for key := range c.store.Keys(nil) {
		val, _ := c.store.Read(key)
		fn(CacheResult{key: key, value: string(val)})
	}
}

// Clean the cache
func (c *Cache) Clean() {
	c.memory = make(map[string]string)
	c.onDisk = false
	os.RemoveAll(c.dir)
}
