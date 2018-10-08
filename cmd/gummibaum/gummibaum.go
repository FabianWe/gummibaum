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

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/FabianWe/gummibaum"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	var constFlag arrayFlags
	expansion := flag.NewFlagSet("expand", flag.ExitOnError)
	fileFlag := expansion.String("file", "", "Input template file")
	expansion.Var(&constFlag, "const", "variable / value pair: var=value")
	noEscape := expansion.Bool("no-escape", false, "Set to true to globally suppress LaTeX escaping of input")
	if len(os.Args) == 1 {
		fmt.Println("NO")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "expand":
		expansion.Parse(os.Args[2:])
		constMap, constMapErr := gummibaum.ParseConstPairs(constFlag)
		if constMapErr != nil {
			panic(constMapErr)
		}
		replacer := gummibaum.LatexReplacer(gummibaum.LatexReplaceFromList(gummibaum.DefaultReplacers))
		if !*noEscape {
			for key, val := range constMap {
				constMap[key] = replacer(val)
			}
		}
		constHandler := gummibaum.NewConstHandler(constMap)
		if *fileFlag == "" {
			panic("no file")
		}
		f, openErr := os.Open(*fileFlag)
		if openErr != nil {
			panic(openErr)
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			line = gummibaum.ApplyExpandHandlers(line, constHandler)
			fmt.Println(line)
		}
		if scannErr := scanner.Err(); scannErr != nil {
			panic(scannErr)
		}
	default:
		fmt.Println("NO 2")
		os.Exit(1)
	}
}
