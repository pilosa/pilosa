// Copyright 2017 Pilosa Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/pilosa/pilosa"
)

// Field represents a test wrapper for pilosa.Field.
type Field struct {
	*pilosa.Field
}

// NewField returns a new instance of Field d/0.
func NewField(opt ...pilosa.FieldOption) *Field {
	path, err := ioutil.TempDir("", "pilosa-field-")
	if err != nil {
		panic(err)
	}
	field, err := pilosa.NewField(path, "i", "f", opt...)
	if err != nil {
		panic(err)
	}
	return &Field{Field: field}
}

// MustOpenField returns a new, opened field at a temporary path. Panic on error.
func MustOpenField(opt ...pilosa.FieldOption) *Field {
	f := NewField(opt...)
	if err := f.Open(); err != nil {
		panic(err)
	}
	return f
}

// Close closes the field and removes the underlying data.
func (f *Field) Close() error {
	defer os.RemoveAll(f.Path())
	return f.Field.Close()
}

// Reopen closes the index and reopens it.
func (f *Field) Reopen() error {
	var err error
	if err := f.Field.Close(); err != nil {
		return err
	}

	path, index, name := f.Path(), f.Index(), f.Name()
	f.Field, err = pilosa.NewField(path, index, name)
	if err != nil {
		return err
	}

	if err := f.Open(); err != nil {
		return err
	}
	return nil
}

// MustSetBit sets a bit on the field. Panic on error.
func (f *Field) MustSetBit(view string, rowID, columnID uint64, t *time.Time) (changed bool) {
	changed, err := f.SetBit(view, rowID, columnID, t)
	if err != nil {
		panic(err)
	}
	return changed
}
