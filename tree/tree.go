package main

import (
	"flag"
	"fmt"

	tree "github.com/gdey/tree"
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
	depthPrefix := "│"
	var sep string

	for i, p := range paths {
		prefix := prevprefix

		if p.Len() == 0 {
			sep = "─"
		} else {
			sep = "┬"
		}

		if i == (len(paths) - 1) {
			fmt.Printf("%v└%v─%v\n", prefix, sep, p.Filename())
			depthPrefix = " "
		} else {
			fmt.Printf("%v├%v─%v\n", prefix, sep, p.Filename())
			depthPrefix = "│"
		}
		printFileEntry(p, depth+1, prefix+depthPrefix)
	}

}
func main() {
	flag.Parse()
	rootEntry := tree.GetFileEntryForDir(*dir)
	fmt.Printf("┌──%v\n", rootEntry.Filename())
	printFileEntry(rootEntry, 0, "")
}
