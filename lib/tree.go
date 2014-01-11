// Copyright 2014 Guatam Dey. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Pacakge tree implements the methods that are used by the
// implementation of the tree command. This command display
// a text based graphical representation of a directory.
package tree

import (
	"container/list"
	"fmt"
	"io/ioutil"
	"path"
)

// This is a FileEntry that will contain all the information. If the parent is
// Null then it is root directory entry
type FileEntry struct {
	name    string     // The name of the directory or file
	parent  *FileEntry // This could be nil if there is no parent.
	entries *list.List // Children this is nil if this is a file, other contains
	// the directory entries.
	Error error // The error retrieving this file entry.
}

type reqFEChan struct {
	fileEntry *FileEntry
	err       error
}

type FileEntryElementTypeError string

func (err FileEntryElementTypeError) Error() string {
	return "Unknown type" + string(err)
}

// Create a new file Entry, it will also the initialize the parent dir pointer
// to nil
func New(name string) *FileEntry { return new(FileEntry).Init(name, nil) }

// Return the filename
func (fe *FileEntry) Filename() string { return fe.name }

func (fe *FileEntry) AllPaths() []*FileEntry {
	if fe.Len() == 0 {
		return []*FileEntry{}
	}
	elements := make([]*FileEntry, fe.Len())

	for i, e := 0, fe.entries.Front(); e != nil; i, e = i+1, e.Next() {
		ele, err := FileEntryValueFrom(e)
		if err != nil {
			continue
		}
		elements[i] = ele
	}
	return elements
}

// Return the full filename including the parent directories.
func (fe *FileEntry) FullFilename() string {
	fepath := ""
	parent := fe.parent
	for parent != nil {
		fepath = path.Join(parent.Filename(), fepath)
		parent = parent.parent
	}
	return path.Join(fepath, fe.Filename())
}

func (fe *FileEntry) ParentDir() *FileEntry { return fe.parent }

func (fe *FileEntry) hasParentDir() bool { return fe.parent != nil }

// Set the parent to the given parent file entry; returns the originating
// object; not the given parent
func (fe *FileEntry) setParentDir(parent *FileEntry) *FileEntry {
	fe.parent = parent
	return fe
}

func (fe *FileEntry) Init(name string, parent *FileEntry) *FileEntry {
	if parent != nil {
		fe.parent = parent
	}
	fe.name = name
	fe.entries = list.New()
	return fe
}

func (fe *FileEntry) AddEntry(newEntry *FileEntry) *FileEntry {
	if newEntry != nil {
		newEntry.parent = fe
		fe.entries.PushFront(newEntry)
	}
	return fe
}

func (fe *FileEntry) Len() int {
	if fe.entries == nil {
		return 0
	}
	return fe.entries.Len()
}

func (fe *FileEntry) Front() *list.Element { return fe.entries.Front() }
func (fe *FileEntry) Back() *list.Element  { return fe.entries.Back() }
func FileEntryValueFrom(i *list.Element) (*FileEntry, error) {
	if i == nil { // no opt on nil
		return nil, nil
	}
	switch t := i.Value.(type) {
	default:
		return nil, FileEntryElementTypeError(fmt.Sprintf("Unexpected type %t\n", t))
	case *FileEntry:
		return t, nil
	}
}

func (fe *FileEntry) RemoveElement(e *list.Element) *FileEntry {
	efe, err := FileEntryValueFrom(e)
	if err != nil {
		panic(fmt.Sprintf("Don't know what to do with %v", err))
	}
	efe.parent = nil
	fe.entries.Remove(e)
	return efe
}
func getFileEntryForFile(parent *FileEntry, name string) *FileEntry {
	file := New(name)
	file.setParentDir(parent)
	return file
}

/***
 * HELPER FUNCTIONS
 ***/

func getFileEntryForDir(parent *FileEntry, name string, ch chan *FileEntry, sem chan int) {
	head := New(name)
	head.setParentDir(parent)
	files, err := ioutil.ReadDir(head.FullFilename())
	if err != nil {
		fmt.Println("Got error when looking at ", name, "error:", err)
		head.Error = err
		ch <- head
		return
	}
	sch := make(chan *FileEntry, cap(sem))
	count := 0
	for _, file := range files {
		select {
		case a := <-sch:
			// Let's take care of the directory entry that is done.
			head.AddEntry(a)
			count -= 1
		default:
			if file.IsDir() {
				select {
				case sem <- 1:
					go getFileEntryForDir(head, file.Name(), sch, sem)
					defer func() { <-sem }()
				default:
					getFileEntryForDir(head, file.Name(), sch, sem)
				}
				count += 1
			} else {
				fe := getFileEntryForFile(head, file.Name())
				head.AddEntry(fe)
			}
		}
	}
	for count != 0 {
		head.AddEntry(<-sch)
		count -= 1
	}
	ch <- head
	return

}

// GetFileEntryForDirWithThreadSize function returns a FileEntry structure that
// contains the entire directory tree.
func GetFileEntryForDirWithThreadSize(name string, threadSize int) *FileEntry {
	ch := make(chan *FileEntry, 1)
	sem := make(chan int, threadSize)
	go getFileEntryForDir(nil, name, ch, sem)
	return <-ch
}

// GetFileEntryForDir This is a convience function that just calls GetFileEntryForDirWithThreadSize,
// with the threadSize set to 200.
func GetFileEntryForDir(name string) *FileEntry { return GetFileEntryForDirWithThreadSize(name, 200) }
