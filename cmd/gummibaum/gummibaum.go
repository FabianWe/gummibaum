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
	"io"
	"log"
	"os"
	"path/filepath"

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

func getWriter(path string) (io.Writer, func(), error) {
	if len(path) == 0 {
		return os.Stdout, func() {}, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, func() {}, err
	}
	done := func() {
		f.Close()
	}
	return f, done, nil
}

func openCSVExpand(path string) (*gummibaum.CSVReader, error) {
	if len(path) == 0 {
		return nil, nil
	}
	return gummibaum.NewCSVFileReader(path, ',', true)
}

// sorry, really ugly code
func expand(args []string) {
	expansion := flag.NewFlagSet("expand", flag.ExitOnError)
	var constFlag arrayFlags
	expansion.Var(&constFlag, "const", "replace variable / value pair: var=value")
	var rowFlag arrayFlags
	expansion.Var(&rowFlag, "row", "replace variable / row name pair: var=row-name")
	fileFlag := expansion.String("file", "", "Input template file")
	noEscape := expansion.Bool("no-escape", false, "Set to true to globally suppress LaTeX escaping of input")
	outFilePath := expansion.String("out", "", "If given write to a file instead of std out. Must be a directory if single-file is false")
	singleFile := expansion.Bool("single-file", true, "If a collection is inserted output to a single file")
	dataSource := expansion.String("csv", "", "Path to the csv file containing the data")
	config := expansion.String("config", "", "Path to a json file containing the config")
	expansion.Parse(args)
	// first parse config from json if given
	var jsonConst, jsonRows map[string]string
	if len(*config) > 0 {
		var jsonErr error
		jsonConst, jsonRows, jsonErr = gummibaum.ExpandConfigFromJSONFile(*config)
		if jsonErr != nil {
			panic(jsonErr)
		}
	}
	constMap, constMapErr := gummibaum.ParseVarValList(constFlag)
	if constMapErr != nil {
		panic(constMapErr)
	}
	rowMap, rowMapErr := gummibaum.ParseVarValList(rowFlag)
	if rowMapErr != nil {
		panic(rowMapErr)
	}
	// now update both maps, values from the command line take precedence
	constMap = gummibaum.MergeStringMaps(jsonConst, constMap)
	rowMap = gummibaum.MergeStringMaps(jsonRows, rowMap)
	var replacer gummibaum.LatexEscapeFunc
	if !*noEscape {
		replacer = gummibaum.LatexEscapeFromList(gummibaum.DefaultReplacers)
	}
	constHandler := gummibaum.NewConstHandler(constMap, replacer)
	var rowHandler *gummibaum.RowHandler
	if len(rowMap) > 0 {
		rowHandler = gummibaum.NewRowHandler(rowMap, replacer)
	}
	if *fileFlag == "" {
		panic("no file")
	}
	f, openErr := os.Open(*fileFlag)
	if openErr != nil {
		panic(openErr)
	}
	defer f.Close()
	if rowHandler == nil {
		// check where to write output to
		out, done, outErr := getWriter(*outFilePath)
		if outErr != nil {
			panic(outErr)
		}
		defer done()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			_, writeErr := gummibaum.WriteExpandHandlers(out, line, constHandler)
			if writeErr != nil {
				panic(writeErr)
			}
		}
		if scannErr := scanner.Err(); scannErr != nil {
			panic(scannErr)
		}
	} else {
		// parse whole file content and compute parts
		head, body, foot, splitErr := gummibaum.ExpandParseTex(f)
		if splitErr != nil {
			panic(splitErr)
		}
		if *singleFile {
			// just apply each one after the other
			out, done, outErr := getWriter(*outFilePath)
			if outErr != nil {
				panic(outErr)
			}
			defer done()
			// apply each row in head
			csv, csvErr := openCSVExpand(*dataSource)
			if csvErr != nil {
				panic(csvErr)
			}
			for _, line := range head {
				_, writeErr := gummibaum.WriteExpandHandlers(out, line, constHandler)
				if writeErr != nil {
					panic(writeErr)
				}
			}
			// iterate body
			if csv != nil {
				// iterate each row and apply handlers
				collection, collectionErr := gummibaum.NewCollection(csv)
				if collectionErr != nil {
					panic(collectionErr)
				}
				for _, col := range collection.Columns {
					// create new row handler with col, that's how we should use it
					newRowHandler := rowHandler.WithColumn(col)
					// now apply handlers for each line in body
					for _, line := range body {
						_, writeErr := gummibaum.WriteExpandHandlers(out, line, constHandler, newRowHandler)
						if writeErr != nil {
							panic(writeErr)
						}
					}
				}
			}
			// iterate foot
			for _, line := range foot {
				_, writeErr := gummibaum.WriteExpandHandlers(out, line, constHandler)
				if writeErr != nil {
					panic(writeErr)
				}
			}
		} else {
			// now outfile must be a directory
			csv, csvErr := openCSVExpand(*dataSource)
			if csvErr != nil {
				panic(csvErr)
			}
			if csv == nil {
				return
			}
			// now iterate each column
			collection, collectionErr := gummibaum.NewCollection(csv)
			if collectionErr != nil {
				panic(collectionErr)
			}
			for i, col := range collection.Columns {
				// open a file
				fPath := filepath.Join(*outFilePath, fmt.Sprintf("out%d.tex", i+1))
				outFile, outFileErr := os.Create(fPath)
				if outFileErr != nil {
					log.Printf("Unable to create file %s\n", fPath)
					continue
				}
				// we don't defer the call to Close, this would mean that we could
				// end up with thousands of deferred calls
				// write head
				for _, line := range head {
					_, writeErr := gummibaum.WriteExpandHandlers(outFile, line, constHandler)
					if writeErr != nil {
						// close file
						outFile.Close()
						panic(writeErr)
					}
				}
				// apply body with current column
				// create new row handler with col, that's how we should use it
				newRowHandler := rowHandler.WithColumn(col)
				for _, line := range body {
					_, writeErr := gummibaum.WriteExpandHandlers(outFile, line, constHandler, newRowHandler)
					if writeErr != nil {
						outFile.Close()
						panic(writeErr)
					}
				}
				// write foot
				for _, line := range foot {
					_, writeErr := gummibaum.WriteExpandHandlers(outFile, line, constHandler)
					if writeErr != nil {
						outFile.Close()
						panic(writeErr)
					}
				}
				outFile.Close()
			}
		}
	}
}

func template(args []string) {
	templateFlags := flag.NewFlagSet("template", flag.ExitOnError)
	templateFlags.Parse(args)
	replacer := gummibaum.LatexEscapeFromList(gummibaum.DefaultReplacers)
	filenames := templateFlags.Args()
	template, templateErr := gummibaum.ParseTemplates(replacer, filenames...)
	if templateErr != nil {
		panic(templateErr)
	}
	data := map[string]interface{}{
		"REPLNAME": "John",
	}
	err := template.Execute(os.Stdout, data)
	if err != nil {
		panic(err)
	}
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("NO")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "expand":
		expand(os.Args[2:])
	case "template":
		template(os.Args[2:])
	default:
		fmt.Println("NO 2")
		os.Exit(1)
	}
}
