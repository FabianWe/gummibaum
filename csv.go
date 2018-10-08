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
	"encoding/csv"
	"errors"
	"io"
	"os"
)

// CSVReader implements CollectionSource by reading content as csv.
type CSVReader struct {
	HeadContent    []string
	ColumnsContent [][]string
}

// NewCSVReader returns a new csv reader given the reader source and the separator
// (usually comma). If head is true the first column is assumed to be the head column
// and must be present.
//
// This function exhaustively reads all data from the reader in memory, thus it
// can be closed after the function returns
func NewCSVReader(r io.Reader, sep rune, head bool) (CSVReader, error) {
	// try to parse csv
	csvReader := csv.NewReader(r)
	csvReader.Comma = sep
	allEntries, entriesErr := csvReader.ReadAll()
	if entriesErr != nil {
		return CSVReader{}, entriesErr
	}
	var headContent []string
	if head {
		// head must be the first entry
		if len(allEntries) == 0 {
			return CSVReader{}, errors.New("Can't read head from csv, does not contain any row")
		}
		headContent = allEntries[0]
		allEntries = allEntries[1:]
	}
	return CSVReader{
			HeadContent:    headContent,
			ColumnsContent: allEntries,
		},
		nil
}

// NewCSVFileReader returns a new csv reader given a file path.
func NewCSVFileReader(file string, sep rune, head bool) (CSVReader, error) {
	f, err := os.Open(file)
	if err != nil {
		return CSVReader{}, err
	}
	defer f.Close()
	return NewCSVReader(f, sep, head)
}

// Head returns the head.
func (r CSVReader) Head() ([]string, error) {
	return r.HeadContent, nil
}

// Entries returns all columns.
func (r CSVReader) Entries() ([][]string, error) {
	return r.ColumnsContent, nil
}
