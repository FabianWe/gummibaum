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
	"strings"
)

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

type ConstHandler struct {
	replacer *strings.Replacer
}

func NewConstHandler(mapper ConstMapper) *ConstHandler {
	replaceMap := make([]string, 0, 2*len(mapper))
	for key, value := range mapper {
		// add space
		replaceMap = append(replaceMap, key, fmt.Sprint(value))
	}
	return &ConstHandler{strings.NewReplacer(replaceMap...)}
}

func (h *ConstHandler) HandleLine(line string) string {
	return h.replacer.Replace(line)
}
