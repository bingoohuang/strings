// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Strings is a more capable, UTF-8 aware version of the standard strings utility.
//
// Flags(=default) are:
//
//	-ascii(=false)    restrict strings to ASCII
//	-search=abc       search string abc
//	-min(=6)          minimum length of UTF-8 strings printed, in runes
//	-max(=256)        maximum length of UTF-8 strings printed, in runes
//	-offset(=true)    show file name and offset of start of each string
package main // import "robpike.io/cmd/strings"

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	min    = flag.Int("min", 6, "minimum length of UTF-8 strings printed, in runes")
	max    = flag.Int("max", 256, "maximum length of UTF-8 strings printed, in runes")
	ascii  = flag.Bool("ascii", false, "restrict strings to ASCII")
	search = flag.String("search", "", "search ASCII string")
	n      = flag.Int("n", 1, "search at most n places")
	offset = flag.Bool("offset", true, "show file name and offset of start of each string")
)

var stdout *bufio.Writer

func main() {
	log.SetFlags(0)
	log.SetPrefix("strings: ")
	stdout = bufio.NewWriter(os.Stdout)
	defer stdout.Flush()

	flag.Parse()

	if *search != "" {
		*min = len(*search)
	}

	if *max < *min {
		*max = *min
	}

	if flag.NArg() == 0 {
		do("<stdin>", os.Stdin)
	} else {
		for _, arg := range flag.Args() {
			fd, err := os.Open(arg)
			if err != nil {
				log.Print(err)
				continue
			}
			do(arg, fd)
			stdout.Flush()
			fd.Close()
		}
	}
}

func do(name string, file *os.File) {
	in := bufio.NewReader(file)
	str := make([]rune, 0, *max)
	filePos := int64(0)
	searchTimes := 0
	print := func() {
		if len(str) >= *min {
			s := string(str)
			if *search == "" || strings.Contains(s, *search) {
				if searchTimes > 0 || *search != "" && strings.Contains(s, *search) {
					searchTimes++
				}

				if *offset {
					fmt.Printf("%s:#%d:\t%s\n", name, filePos-int64(len(s)), s)
				} else {
					fmt.Println(s)
				}

				if *n > 0 && searchTimes >= *n {
					os.Exit(0)
				}
			}
		}
		str = str[:0]
	}
	for {
		var (
			r   rune
			wid int
			err error
		)
		// One string per loop.
		for ; ; filePos += int64(wid) {
			if r, wid, err = in.ReadRune(); err != nil {
				if err != io.EOF {
					log.Print(err)
				}
				return
			}
			if !strconv.IsPrint(r) || *ascii && r >= 0xFF {
				print()
				continue
			}
			// It's printable. Keep it.
			if len(str) >= *max {
				print()
			}
			str = append(str, r)
		}
	}
}
