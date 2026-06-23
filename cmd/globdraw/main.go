package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/capnspacehook/glob"
	"github.com/capnspacehook/glob/match"
	"github.com/capnspacehook/glob/match/debug"
)

func main() {
	pattern := flag.String("p", "", "pattern to draw")
	sep := flag.String("s", "", "comma separated list of separators characters")
	flag.Parse()

	if *pattern == "" {
		flag.Usage()
		os.Exit(1)
	}

	var separators []rune
	if len(*sep) > 0 {
		for c := range strings.SplitSeq(*sep, ",") {
			if r, w := utf8.DecodeRuneInString(c); len(c) > w {
				fmt.Println("only single charactered separators are allowed")
				os.Exit(1)
			} else {
				separators = append(separators, r)
			}
		}
	}

	glob, err := glob.Compile(*pattern, separators...)
	if err != nil {
		fmt.Println("could not compile pattern:", err)
		os.Exit(1)
	}

	matcher := glob.(match.Matcher)
	fmt.Fprint(os.Stdout, debug.Graphviz(*pattern, matcher))
}
