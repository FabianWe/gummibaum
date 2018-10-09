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
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ParseVarValPair parses a string of the form var=val. It returns var and val.
// For syntax errors an error != nil is returned.
func ParseVarValPair(s string) (string, string, error) {
	i := strings.Index(s, "=")
	if i < 0 {
		return "", "", fmt.Errorf("Invalid variable / value pair %s: Must be var=val", s)
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

type ExpandHandler interface {
	HandleLine(line string) string
}

func ApplyExpandHandlers(line string, handlers ...ExpandHandler) string {
	s := line
	for _, handler := range handlers {
		s = handler.HandleLine(s)
	}
	return s
}

func WriteExpandHandlers(w io.Writer, line string, handlers ...ExpandHandler) (int, error) {
	return fmt.Fprintln(w, ApplyExpandHandlers(line, handlers...))
}

type ConstHandler struct {
	replacer *strings.Replacer
}

func NewConstHandler(mapper map[string]string, replaceFunc LatexReplaceFunc) *ConstHandler {
	replaceMap := make([]string, 0, 2*len(mapper))
	for key, value := range mapper {
		valueStr := fmt.Sprint(value)
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

type RowHandler struct {
	replaceVarMap map[string]string
	replaceFunc   LatexReplaceFunc
	currentCol    *Column
}

func NewRowHandler(replaceVarMap map[string]string, replaceFunc LatexReplaceFunc) *RowHandler {
	return &RowHandler{replaceVarMap, replaceFunc, nil}
}

func (h *RowHandler) WithColumn(c *Column) *RowHandler {
	return &RowHandler{h.replaceVarMap, h.replaceFunc, c}
}

func (h *RowHandler) HandleLine(line string) string {
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
		return nil, nil, nil, errors.New("Invalid template syntax, must contain %%begin gummibaum repeat and %%end gummibaum repeat")
	}
	return head, body, foot, nil
}
