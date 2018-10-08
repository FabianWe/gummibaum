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
	"unicode/utf8"
)

// LatexReplaceFunc is any function that replaces LaTeX special character within
// a text. Usually it should be clear how to do this, but LaTeX is a very rich
// language and I want to keep it extendable.
type LatexReplaceFunc func(text string) string

// LatexReplaceFromList returns a replacement function given a map of
// substitution pairs. Example: ["&", "\\&"] would replace each occurrence of
// & with \&.
// A list of default replacers can be found in DefaultReplacers,
// ReplacerWithDefaults can be used to extend that list.
func LatexReplaceFromList(mapping []string) LatexReplaceFunc {
	replacer := strings.NewReplacer(mapping...)
	return func(s string) string {
		return replacer.Replace(s)
	}
}

// Verb returns a LaTeX verb environment with the given delimiter. For example
// Verb("|", "foo & bar") yields to \verb|foo & bar|. An error is returned if
// the delimiter is contained in the input string or if delimiter has a length
// != 1.
func Verb(del, s string) (string, error) {
	count := utf8.RuneCountInString(s)
	if count != 1 {
		return "", fmt.Errorf(`Invalid delimiter length for \verb environment: Expected 1 and got %d`, count)
	}
	if strings.Contains(s, del) {
		return "", fmt.Errorf(`Error executing \verb environment: Input string contains delimiter %s`, del)
	}
	return fmt.Sprintf(`\verb%s%s%s`, del, s, del), nil
}

var (
	// DefaultReplacers describes the default replacer pairs.
	DefaultReplacers = []string{
		"&", `\&`,
		"%", `\%`,
		"$", `\$`,
		"#", `\#`,
		"_", `\_`,
		"{", `\{`,
		"}", `\}`,
		"~", `\textasciitilde`,
		"^", `\textasciicircum`,
		`\`, `\textbackslash`,
	}
)

// ReplacerWithDefaults returns a replacer function that uses the content from
// DefaultReplacers and combines it with the replacers from additional.
// Example: ["&", "\\&"] would replace each occurrence of & with \&.
// This replacement is already done by the DefaultReplacers though.
func ReplacerWithDefaults(additional []string) LatexReplaceFunc {
	fullReplacers := make([]string, len(DefaultReplacers)+len(additional))
	copy(fullReplacers, DefaultReplacers)
	copy(fullReplacers[len(DefaultReplacers):], additional)
	return LatexReplaceFromList(fullReplacers)
}

// LatexReplacer returns a function that replaces an arbitrary number of
// arguments with the specified replace function.
func LatexReplacer(replace LatexReplaceFunc) func(args ...interface{}) string {
	return func(args ...interface{}) string {
		// not the way the template packages uses, that does more interesting stuff
		// but I think it should be enough this way
		s := fmt.Sprint(args...)
		return replace(s)
	}
}
