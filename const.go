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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// ConstMapper maps variable names to constant values.
type ConstMapper map[string]interface{}

// ParseConstPair parses a string of the form var=val. It returns var and val.
// For syntax errors an error != nil is returned.
func ParseConstPair(s string) (string, string, error) {
	i := strings.Index(s, "=")
	if i < 0 {
		return "", "", fmt.Errorf("Invalid variable / value pair %s: Must be var=val", s)
	}
	return s[:i], s[i+1:], nil
}

// ParseConstPairs parses a list of var=val pairs (each entry in pairs one such
// pair). The vars are mapped to the value. For syntax errors an error != nil is
// returned.
func ParseConstPairs(pairs []string) (ConstMapper, error) {
	res := make(ConstMapper, len(pairs))
	for _, pairStr := range pairs {
		// parse
		variable, val, err := ParseConstPair(pairStr)
		if err != nil {
			return nil, err
		}
		res[variable] = val
	}
	return res, nil
}

func ConstJSON(r io.Reader) (ConstMapper, error) {
	dec := json.NewDecoder(r)
	m := make(ConstMapper)
	err := dec.Decode(&m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func ConstJSONFromeFile(file string) (ConstMapper, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ConstJSON(f)
}
