// Copyright 2018 - 2020 Fabian Wenzelmann
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
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// ParseVarValPair parses a string of the form var=val. It returns var and val.
// For syntax errors an error != nil is returned.
func ParseVarValPair(s string) (string, string, error) {
	i := strings.Index(s, "=")
	if i < 0 {
		return "", "", fmt.Errorf("invalid variable / value pair \"%s\": Must be var=val", s)
	}
	return s[:i], s[i+1:], nil
}

// ParseVarValList parses a list of var=val pairs (each entry in pairs one such
// pair). The vars are mapped to the value. For syntax errors an error != nil is
// returned.
func ParseVarValList(pairs []string) (map[string]string, error) {
	res := make(map[string]string, len(pairs))
	for _, pairStr := range pairs {
		// parse
		variable, val, err := ParseVarValPair(pairStr)
		if err != nil {
			return nil, err
		}
		res[variable] = val
	}
	return res, nil
}

// ExpandHandler is used for the expand mode to process a single line.
// A handler transforms a line and returns the new line.
type ExpandHandler interface {
	HandleLine(line string) string
}

// ApplyExpandHandlers applies the handlers one after the other to the original
// line.
func ApplyExpandHandlers(line string, handlers ...ExpandHandler) string {
	s := line
	for _, handler := range handlers {
		s = handler.HandleLine(s)
	}
	return s
}

// WriteExpandHandlers works as ApplyExpandHandlers but writes the result
// to a writer. It returns the number of bytes written and any error that
// occurred.
func WriteExpandHandlers(w io.Writer, line string, handlers ...ExpandHandler) (int, error) {
	return fmt.Fprintln(w, ApplyExpandHandlers(line, handlers...))
}

// ConstHandler replaces place holders with constant values.
type ConstHandler struct {
	replacer *strings.Replacer
}

// NewConstHandler returns a new ConstHandler give the mapping place holder
// to constant value (for example "NAME" mapped to "John").
// replaceFunc is a function to escape LaTeX special characters. If it is nil
// no replacements will take place.
func NewConstHandler(mapper map[string]string, replaceFunc LatexEscapeFunc) *ConstHandler {
	replaceMap := make([]string, 0, 2*len(mapper))
	for key, value := range mapper {
		valueStr := value
		if replaceFunc != nil {
			valueStr = replaceFunc(valueStr)
		}
		replaceMap = append(replaceMap, key, valueStr)
	}
	return &ConstHandler{strings.NewReplacer(replaceMap...)}
}

func (h *ConstHandler) HandleLine(line string) string {
	return h.replacer.Replace(line)
}

// RowHandler replaces placeholders with values from a given column. It is not
// save for concurrent use with different columns, use WithColumn to create
// new RewHandlers with a new column and then run replacement on those instances
// concurrently. WithColumn must be called before using HandleLine.
type RowHandler struct {
	replaceVarMap map[string]string
	replaceFunc   LatexEscapeFunc
	currentCol    *Column
}

// NewRowHandler returns a new RowHandler. replaceVarMap must be a mapping
// mapping replace names to row names, for example "REPL-TOKEN" --> "token".
// WithColumn must be called before HandleLine can be used.
func NewRowHandler(replaceVarMap map[string]string, replaceFunc LatexEscapeFunc) *RowHandler {
	return &RowHandler{replaceVarMap, replaceFunc, nil}
}

// WithColumn returns a new row handler with the column set.
func (h *RowHandler) WithColumn(c *Column) *RowHandler {
	return &RowHandler{h.replaceVarMap, h.replaceFunc, c}
}

// HandleLine applies the actual replacement by substituting values for the current column.
// If the current column is nil this method panics, a column must be set before with WithColumn.
func (h *RowHandler) HandleLine(line string) string {
	if h.currentCol == nil {
		panic("no column set, WithColumn must be called before using RowHandler")
	}
	// fast: if no variables are given that need replacing return line
	if len(h.replaceVarMap) == 0 {
		return line
	}
	// now create a replace and get each value from colMap
	replaceMap := make([]string, 0, len(h.replaceVarMap)*2)
	for replName, rowName := range h.replaceVarMap {
		// lookup in colMap, apply replace func if given
		val := h.currentCol.GetKey(rowName)
		if h.replaceFunc != nil {
			val = h.replaceFunc(val)
		}
		replaceMap = append(replaceMap, replName, val)
	}
	replacer := strings.NewReplacer(replaceMap...)
	return replacer.Replace(line)
}

type expandParseState int

const (
	inHeadState expandParseState = iota
	inBodyState
	inFootState
)

// ExpandParseTex splits the tex file into the three parts:
// Head, everything before the line "%begin gummibaum repeat", body
// everything between "%begin gummibaum repeat" and "%end gummibaum repeat"
// and foot everything after "%end gummibaum repeat".
func ExpandParseTex(r io.Reader) ([]string, []string, []string, error) {
	state := inHeadState
	scanner := bufio.NewScanner(r)
	var head, body, foot []string
	for scanner.Scan() {
		line := scanner.Text()
		switch state {
		case inHeadState:
			if strings.HasPrefix(line, "%begin gummibaum repeat") {
				// start reading the body
				state = inBodyState
			} else {
				// append to head
				head = append(head, line)
			}
		case inBodyState:
			if strings.HasPrefix(line, "%end gummibaum repeat") {
				// switch to foot state
				state = inFootState
			} else {
				body = append(body, line)
			}
		case inFootState:
			// in this state just append to foot
			foot = append(foot, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, nil, err
	}
	// now we must be in state inFootState, everything else is an error
	if state != inFootState {
		return nil, nil, nil, errors.New("invalid template syntax, must contain %%begin gummibaum repeat and %%end gummibaum repeat")
	}
	return head, body, foot, nil
}

// ExpandConfigJSON parses a config file. The config files must be a dictionary
// mapping "const" to a dictionary of string variable / value pairs and mapping
// "rows" to a dictionary of string variable / value pairs.
func ExpandConfigJSON(r io.Reader) (map[string]string, map[string]string, error) {
	type fileContent struct {
		Const map[string]string
		Rows  map[string]string
	}
	dec := json.NewDecoder(r)
	inst := fileContent{
		make(map[string]string),
		make(map[string]string),
	}
	dec.DisallowUnknownFields()
	err := dec.Decode(&inst)
	if err != nil {
		return nil, nil, err
	}
	return inst.Const, inst.Rows, nil
}

// ExpandConfigFromJSONFile is like ExpandConfigJSON and reads the content from
// a file.
func ExpandConfigFromJSONFile(file string) (map[string]string, map[string]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, nil, err
	}
	var consts, rows map[string]string
	defer func() {
		closeErr := f.Close()
		if err == nil && closeErr != nil {
			consts, rows = nil, nil
			err = closeErr
		}
	}()
	consts, rows, err = ExpandConfigJSON(f)
	return consts, rows, err
}
