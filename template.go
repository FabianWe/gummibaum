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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"text/template"
	"unicode/utf8"
)

// LatexEscapeFunc is any function that replaces LaTeX special character within
// a text. Usually it should be clear how to do this, but LaTeX is a very rich
// language and I want to keep it extendable.
type LatexEscapeFunc func(text string) string

// LatexEscapeFromList returns a replacement function given a map of
// substitution pairs. Example: ["&", "\\&"] would replace each occurrence of
// & with \&.
// A list of default replacers can be found in DefaultReplacers,
// EscapeWithDefaults can be used to extend that list.
func LatexEscapeFromList(mapping []string) LatexEscapeFunc {
	replacer := strings.NewReplacer(mapping...)
	return func(s string) string {
		return replacer.Replace(s)
	}
}

// Verb returns a LaTeX verb environment with the given delimiter. For example
// Verb("|", "foo & bar") yields to \verb|foo & bar|. An error is returned if
// the delimiter is contained in the input string or if delimiter has a length
// != 1.
func Verb(del string, args ...interface{}) (string, error) {
	count := utf8.RuneCountInString(del)
	if count != 1 {
		return "", fmt.Errorf(`invalid delimiter length for \verb environment: Expected 1 and got %d`, count)
	}
	// not the way the template packages uses, that does more interesting stuff
	// but I think it should be enough this way
	asStrings := make([]string, len(args))
	for i, arg := range args {
		asStrings[i] = fmt.Sprintf("%v", arg)
	}
	s := strings.Join(asStrings, " ")
	if strings.Contains(s, del) {
		return "", fmt.Errorf(`error executing \verb environment: Input string contains delimiter %s`, del)
	}
	return fmt.Sprintf(`\verb%s%s%s`, del, s, del), nil
}

// Join concatenates the elements of args to create a single string. The separator
// string sep is placed between elements in the resulting string. The function
// parameterized by a LatexEscapeFunc (that can be nil) that is used to prepare
// each arg.
func Join(replace LatexEscapeFunc) func(sep string, args ...interface{}) string {
	return func(sep string, args ...interface{}) string {
		asStrings := make([]string, 0, len(args))
		for _, arg := range args {
			if asSlice, ok := arg.([]string); ok {
				// iterate slice
				for _, a := range asSlice {
					if replace != nil {
						a = replace(a)
					}
					asStrings = append(asStrings, a)
				}
			} else {
				var a = fmt.Sprintf("%v", arg)
				if replace != nil {
					a = replace(a)
				}
				asStrings = append(asStrings, a)
			}
		}
		return strings.Join(asStrings, sep)
	}
}

var (
	// DefaultReplacers describes the default replacer pairs. Note that certain
	// replacements like \textbackslash must have a leading space.
	DefaultReplacers = []string{
		"&", `\&`,
		"%", `\%`,
		"$", `\$`,
		"#", `\#`,
		"_", `\_`,
		"{", `\{`,
		"}", `\}`,
		"~", `\textasciitilde `,
		"^", `\textasciicircum `,
		`\`, `\textbackslash `,
	}
)

// EscapeWithDefaults returns an escape function that uses the content from
// DefaultReplacers and combines it with the replacers from additional.
// Example: ["&", "\\&"] would replace each occurrence of & with \&.
// This replacement is already done by the DefaultReplacers though.
func EscapeWithDefaults(additional []string) LatexEscapeFunc {
	fullReplacers := make([]string, len(DefaultReplacers)+len(additional))
	copy(fullReplacers, DefaultReplacers)
	copy(fullReplacers[len(DefaultReplacers):], additional)
	return LatexEscapeFromList(fullReplacers)
}

// LatexEscaper returns a function that escapes an arbitrary number of
// arguments with the specified escaping function.
func LatexEscaper(replace LatexEscapeFunc) func(args ...interface{}) string {
	return func(args ...interface{}) string {
		// not the way the template packages uses, that does more interesting stuff
		// but I think it should be enough this way
		asStrings := make([]string, len(args))
		for i, arg := range args {
			asStrings[i] = fmt.Sprintf("%v", arg)
		}
		s := strings.Join(asStrings, " ")
		return replace(s)
	}
}

// LatexTemplate adds the functions "latex", "verb", and "join" to the template.
// This function must be called before the template is parsed.
func LatexTemplate(t *template.Template, replace LatexEscapeFunc) *template.Template {
	funcMap := template.FuncMap{
		"latex": LatexEscaper(replace),
		"verb":  Verb,
		"join":  Join(replace),
	}
	return t.Funcs(funcMap)
}

// ParseTemplates parses the templates specified by filenames. See Go
// template documentation for ParseTemplates for details. The functions
// "latex", "verb" and "join" are added. The replace function is used to escape
// special characters, if it is nil no replacement takes place.
// Delims defines which delimiters are used. The default {{ and }} are not nice for latex, so we replace them.
// #( and #) seem to be a good idea. This is what happens when you use the empty string as delims.
func ParseTemplates(replace LatexEscapeFunc, delimLeft, delimRight string, filenames ...string) (*template.Template, error) {
	if delimLeft == "" {
		delimLeft = "#("
	}

	if delimRight == "" {
		delimRight = "#)"
	}

	if len(filenames) > 0 {
		// TODO naming should be fine? I think that's what the comment in ParseFiles
		// in the source code means...
		name := path.Base(filenames[0])
		t, err := LatexTemplate(template.New(name), replace).Delims(delimLeft, delimRight).ParseFiles(filenames...)
		if err != nil {
			return nil, err
		}
		return t, nil
	}
	return nil, errors.New("no template file names given")
}

// TemplateConstJSON parses a constant json file, it must be a dictionary
// mapping strings (replace identifiers) by constant values.
func TemplateConstJSON(r io.Reader) (map[string]string, error) {
	m := make(map[string]string)
	dec := json.NewDecoder(r)
	err := dec.Decode(&m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// TemplateConstFromJSONFile is like TemplateConstJSON and reads the content
// from a file.
func TemplateConstFromJSONFile(file string) (map[string]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	var m map[string]string
	defer func() {
		closeErr := f.Close()
		if err == nil && closeErr != nil {
			m = nil
			err = closeErr
		}
	}()
	m, err = TemplateConstJSON(f)
	return m, err
}
