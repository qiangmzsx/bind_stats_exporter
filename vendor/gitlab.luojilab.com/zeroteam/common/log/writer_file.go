// Go support for leveled logs, analogous to https://code.google.com/p/google-glog/
//
// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// File I/O for logs.

package log

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

// MaxSize is the maximum size of a log file in bytes.
var MaxSize uint64 = 1024 * 1024 * 1000

var (
	program  = filepath.Base(os.Args[0])
	host     = "unknownhost"
	userName = "unknownuser"
)

func init() {
	h, err := os.Hostname()
	if err == nil {
		host = shortHostname(h)
	}

	current, err := user.Current()
	if err == nil {
		userName = current.Username
	}

	// Sanitize userName since it may contain filepath separators on Windows.
	userName = strings.Replace(userName, `\`, "_", -1)
}

// shortHostname returns its argument, truncating at the first period.
// For instance, given "www.google.com" it returns "www".
func shortHostname(hostname string) string {
	if i := strings.Index(hostname, "."); i >= 0 {
		return hostname[:i]
	}
	return hostname
}

// logName returns a new log file name containing tag, with start time t, and
// the name for the symlink for tag.
func logName(prefix string, tag string, t time.Time) (name, link string) {
	name = fmt.Sprintf("%s%s.%s%s.%04d%02d%02d-%02d%02d%02d.%d.%d.log",
		prefix,
		host,
		userName,
		tag,
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		t.Nanosecond(),
		pid)
	link = prefix + tag
	if strings.HasSuffix(link, ".") {
		link += "log"
	} else {
		link += ".log"

	}

	return name, link
}

// create creates a new log file and returns the file and its filename, which
// contains tag ("INFO", "FATAL", etc.) and t.  If the file is created
// successfully, create also attempts to update the symlink for that tag, ignoring
// errors.
func createFile(fileName, prefix string, logdir string, symlinks []string, tag string, t time.Time) (f *os.File, filename string, err error) {
	var name string
	var link string
	if fileName != "" {
		name = fileName + tag
		link = fileName + tag
	} else {
		name, link = logName(prefix, tag, t)
	}

	fname := filepath.Join(logdir, name)
	fnameAbs, err := filepath.Abs(fname)
	if err == nil {
		fname = fnameAbs
	} else {
		fmt.Println("common/log/writer_file.createFile err", err)
		return nil, fname, err
	}

	err = ensureParentDirExists(fname)
	if err != nil {
		return nil, fname, err
	}

	f, err = os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err == nil {
		for _, symlinkDir := range symlinks {
			linkFile(fname, symlinkDir, link)
		}
		return f, fname, nil
	}
	return nil, "", fmt.Errorf("log: cannot create log: %v", err)
}

func linkFile(srcFile, symlink string, linkFileName string) {
	if strings.HasSuffix(symlink, "/") {
		symlink = filepath.Join(symlink, linkFileName)
	}

	fnameAbs, err := filepath.Abs(symlink)
	if err == nil {
		symlink = fnameAbs
	} else {
		fmt.Println("common/log/writer_file.linkFile err", err)
		return
	}
	if srcFile == fnameAbs {
		fmt.Printf("common/log/writer_file.linkFile are samefile,ignored! src=%s,dest=%s\n", srcFile, fnameAbs)
		return
	}

	err = ensureParentDirExists(symlink)
	if err != nil {
		fmt.Println(err)
		return
	}
	if isDir(symlink) { // 如果symlink已经存在，且是一个目录，则ignore
		fmt.Printf("%s  exists and is a directory,donot know how to link to it,so create symlink file failed \n", symlink)
		return
	}

	os.Remove(symlink)                 // ignore err
	err = os.Symlink(srcFile, symlink) // ignore err
	if err != nil {
		fmt.Println("symlink failed", err)
	}

}

func isDir(dir string) bool {
	fi, err := os.Stat(dir)
	if err == nil && fi.IsDir() {
		return true
	}
	return false
}
func ensureParentDirExists(fname string) error {
	dir := filepath.Dir(fname)
	fi, err := os.Stat(dir)
	if err == nil && !fi.IsDir() {
		return fmt.Errorf("%s already exists and not a directory", dir)
	}
	if os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create dir %s error: %s", dir, err.Error())
		}
	}
	return nil
}
