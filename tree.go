package main

import (
	"flag"
	"fmt"
	tree "github.com/gdey/tree/lib"
)

/*

-foo
 |-bar
 | |-bar0.go
 | \-bar1.go
 \-baz
   \-baz.go


*/
var dir = flag.String("dir", ".", "The directory to get the tree for")

func printFileEntry(fe *tree.FileEntry, depth int, prevprefix string) {
	if fe.Len() == 0 {
		return
	}
	paths := fe.AllPaths()
	for i, p := range paths {
		prefix := prevprefix
		if i == (len(paths) - 1) {
			fmt.Printf("%v \\-%v\n", prefix, p.Filename())
			prefix = prefix + "  "
		} else {
			prefix = prefix + " |"
			fmt.Printf("%v-%v\n", prefix, p.Filename())
		}
		printFileEntry(p, depth+1, prefix)
	}

}
func main() {
	flag.Parse()
	rootEntry := tree.GetFileEntryForDir(*dir)
	fmt.Printf("-%v\n", rootEntry.Filename())
	printFileEntry(rootEntry, 0, "  ")
}
