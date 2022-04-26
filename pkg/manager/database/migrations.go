package database

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
	storm "github.com/asdine/storm"
	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

type schemaMigration func(*storm.DB) error

var migrations = []schemaMigration{migrateDefaultPackage}

var migrateDefaultPackage schemaMigration = func(bs *storm.DB) error {
	packs := []types.Package{}

	bs.Bolt.View(
		func(tx *bbolt.Tx) error {
			// previously we had pkg.DefaultPackage
			// IF it's there, collect packages to add to the new schema
			b := tx.Bucket([]byte("DefaultPackage"))
			if b != nil {
				b.ForEach(func(k, v []byte) error {
					p, err := types.PackageFromYaml(v)
					if err == nil && p.ID != 0 {
						packs = append(packs, p)
					}
					return nil
				})
			}
			return nil
		},
	)

	for k := range packs {
		d := &packs[k]
		d.ID = 0
		err := bs.Save(d)
		if err != nil {
			return errors.Wrap(err, "Error saving package to "+d.Path)
		}
	}

	// Be sure to delete old only if everything was migrated without any error
	bs.Bolt.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("DefaultPackage"))
		if b != nil {
			return tx.DeleteBucket([]byte("DefaultPackage"))
		}
		return nil
	})

	return nil
}
