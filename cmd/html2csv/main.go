/* SPDX-License-Identifier: BSD-2-Clause */

package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/ricardobranco777/html2csv/htmltable"
)

import flag "github.com/spf13/pflag"

const Version = "0.2.0"

func main() {
	var delim, tablesel string
	var version bool

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] FILE\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVarP(&delim, "delimiter", "d", ",", "delimiter")
	flag.StringVarP(&tablesel, "table", "t", "", "select tables by index or name")
	flag.BoolVarP(&version, "version", "", false, "print version and exit")
	flag.Parse()

	if version {
		fmt.Printf("html2csv v%s %v %s/%s\n", Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	log.SetFlags(0)
	log.SetPrefix("ERROR: ")

	f := os.Stdin
	var err error

	if flag.NArg() > 1 {
		flag.Usage()
		os.Exit(1)
	}
	if flag.NArg() == 1 {
		f, err = os.Open(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
	}

	r := []rune(delim)
	if len(r) != 1 {
		fmt.Fprintf(os.Stderr, "delimiter must be a single character\n")
		os.Exit(1)
	}
	delimiter := r[0]

	tables, err := htmltable.Parse(f)
	if err != nil {
		log.Fatal(err)
	}

	sel, err := htmltable.ParseSelector(tablesel)
	if err != nil {
		log.Fatal(err)
	}
	tables = sel.Apply(tables)

	enc := htmltable.NewCSVEncoder()
	enc.Comma = delimiter

	if err := enc.Encode(os.Stdout, tables); err != nil {
		log.Fatal(err)
	}
}
