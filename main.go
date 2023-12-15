package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type ctx struct {
	paths   []string
	pattern []string
	regexps []*regexp.Regexp
	found   []string
	verbose bool
	dryRun  bool
	fmutex  sync.Mutex
}

func main() {
	c, err := parseArguments()
	if err != nil {
		usage()
		os.Exit(1)
	}
	b, err := os.ReadFile(".rmcrap")
	if err == nil {
		s := string(b)
		c.pattern = strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
		c.pattern = removeEmptyStrings(c.pattern)
	} else {
		c.pattern = []string{".DS_Store", "Thumbs.db", `^.*.CR2.\$.jpg$`, `^.*.tmp$`}
	}
	c.regexps = make([]*regexp.Regexp, 0)
	for _, p := range c.pattern {
		c.regexps = append(c.regexps, regexp.MustCompile(p))
	}
	dumpCtx(c)
	for _, p := range c.paths {
		c.walkDir(p)
	}

	fmt.Println("\nResult")
	for _, p := range c.found {
		if c.dryRun {
			fmt.Println(p)
		} else {
			err = os.Remove(p)
			if err != nil {
				fmt.Printf("error removing file %s: %v\n", p, err)
			}
		}

	}
}

func parseArguments() (*ctx, error) {
	c := ctx{}
	c.paths = make([]string, 0)
	for _, a := range os.Args[1:] {
		if a == "-v" {
			c.verbose = true
		} else if a == "--dry-run" {
			c.dryRun = true
		} else if strings.HasPrefix(a, "-") {
			return nil, fmt.Errorf("unknown parameter %s", a)
		} else {
			c.paths = append(c.paths, a)
		}
	}
	if len(c.paths) == 0 {
		c.paths = append(c.paths, "//sokrates/Photos")
	}
	return &c, nil
}

func usage() {
	fmt.Printf("usage: %s [-v] [--dry-run] [path] [path]...\n", os.Args[0])
}

func dumpCtx(c *ctx) {
	fmt.Printf("verbose %t\n", c.verbose)
	fmt.Printf("dryrun  %t\n", c.dryRun)
	fmt.Print("Patterns ")
	for _, p := range c.pattern {
		fmt.Print(p + " ")
	}
	fmt.Print("\nPaths ")
	for _, p := range c.paths {
		fmt.Print(p + " ")
	}
}

func (c *ctx) addFound(p string) {
	c.fmutex.Lock()
	c.found = append(c.found, p)
	c.fmutex.Unlock()
}

func (c *ctx) walkDir(p string) {
	fileSystem := os.DirFS(p)
	fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if !d.IsDir() {
			for _, re := range c.regexps {
				if re.MatchString(d.Name()) {
					c.addFound(filepath.Join(p, path))
					break
				}
			}
		}
		return nil
	})
}

func removeEmptyStrings(in []string) []string {
	var r []string
	for _, s := range in {
		if s != "" {
			r = append(r, s)
		}
	}
	return r

}
