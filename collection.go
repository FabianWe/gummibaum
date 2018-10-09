// Copyright 2018 Fabian Wenzelmann
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gummibaum

import (
	"fmt"
	"reflect"
)

const (
	// NoColEntry is returned by several methods of Column to indicate that a
	// key was not found.
	NoColEntry = "NO VALUE"
)

// Column represents a column in a collection.
// It has the head of the entries (that comes from the collection and all data
// entries as strings). Map is a mapping from head name (row names) to the
// data entry and is created in NewColumn.
//
// Head and Entries should have the same size, but are allowed to have different
// sizes. In this case the map contains an entry for each row name in
// min(len(Head), mint(Entries)).
type Column struct {
	Head    []string
	Entries []string
	Map     map[string]string
}

// NewColumn returns a new column and initializes the map m.
func NewColumn(head, entries []string) *Column {
	n := IntMin(len(entries), len(head))
	m := make(map[string]string, n)
	for i := 0; i < n; i++ {
		m[head[i]] = entries[i]
	}
	return &Column{
		Head:    head,
		Entries: entries,
		Map:     m,
	}
}

// GetPos returns the item on position i in Entries. If i is not a valid
// position in entries NoColEntry is returned.
// At is similar to GetPos but returns an error if i is invalid.
func (c *Column) GetPos(i int) string {
	if i < 0 || i >= len(c.Entries) {
		return NoColEntry
	}
	return c.Entries[i]
}

// GetKey returns the item with the given key where key is row name.
// If the key is not found NoColEntry is returned.
// Value is similar to GetKey but returns an error if key is not found.
func (c *Column) GetKey(key string) string {
	if val, has := c.Map[key]; has {
		return val
	}
	return NoColEntry
}

// Get returns either the element on position key if key is an int or the
// mapping at key if key is a string. If it is neither NoColEntry is returned.
// If the position / key is invalid NoColEntry is returned.
// Element is similar to Get but returns an error if key is invalid.
func (c *Column) Get(key interface{}) string {
	switch v := key.(type) {
	case int:
		return c.GetPos(v)
	case string:
		return c.GetKey(v)
	default:
		return NoColEntry
	}
}

// At returns the i-th entry. If i is not a valid position in entries an
// error is returned.
// GetPos is similar to At but does not return an error.
func (c *Column) At(i int) (string, error) {
	if i < 0 || i >= len(c.Entries) {
		return "", fmt.Errorf("Invalid index: %d in %v", i, c.Entries)
	}
	return c.Entries[i], nil
}

// Value returns the item with the given key where key is row name.
// If the key is not found an error is returned.
// GetKey is similar to Value but does not return an error if key is invalid.
func (c *Column) Value(key string) (string, error) {
	if val, has := c.Map[key]; has {
		return val, nil
	}
	return "", fmt.Errorf("Invalid key: %s in %v", key, c.Map)
}

// Element returns either the element on position key if key is an int or the
// mapping at key if key is a string. If it is neither an error is returned.
// If the position / key is invalid an error is returned.
// Get is similar to Element but does not return an error if key is invalid.
func (c *Column) Element(key interface{}) (string, error) {
	switch v := key.(type) {
	case int:
		return c.At(v)
	case string:
		return c.Value(v)
	default:
		return "", fmt.Errorf("Invalid key type for column: Expect int or string, got %v", reflect.TypeOf(key))
	}
}

// CollectionSource is everything that returns entries in the form of a column
// based data modell.
//
// Head describes the row names and can be nil in which case no names are given.
// Entries returns all columns.
//
// For example Head might return ["first-name", "last-name"] and Entries
// might return [["John", "Doe"], ["Jane"]]. The second column does not have a
// field "last-name".
type CollectionSource interface {
	Head() ([]string, error)
	Entries() ([][]string, error)
}

// Collection groups together several columns with the same head (row names).
type Collection struct {
	Head    []string
	Columns []*Column
}

// NewCollection returns a new collection initialized with all entries from the
// source.
func NewCollection(source CollectionSource) (Collection, error) {
	head, headErr := source.Head()
	if headErr != nil {
		return Collection{}, headErr
	}
	entries, entriesErr := source.Entries()
	if entriesErr != nil {
		return Collection{}, entriesErr
	}
	cols := make([]*Column, len(entries))
	for i, strCol := range entries {
		col := NewColumn(head, strCol)
		cols[i] = col
	}
	return Collection{head, cols}, nil
}

// MemoryCollection implements CollectionSource with a predefined set of
// content.
type MemoryCollection struct {
	HeadContent    []string
	ColumnsContent [][]string
}

// NewMemoryCollection returns a new MemoryCollection given the data.
func NewMemoryCollection(head []string, columns [][]string) *MemoryCollection {
	return &MemoryCollection{head, columns}
}

// Head returns the head.
func (c *MemoryCollection) Head() ([]string, error) {
	return c.HeadContent, nil
}

// Entries returns all columns.
func (c *MemoryCollection) Entries() ([][]string, error) {
	return c.ColumnsContent, nil
}
