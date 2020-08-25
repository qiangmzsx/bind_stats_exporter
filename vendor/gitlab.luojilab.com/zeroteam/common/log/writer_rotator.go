package log

import (
	"io"
	"os"
	"sync/atomic"
	"time"

	"gitlab.luojilab.com/zeroteam/common/env"
)

type syncWriter interface {
	Sync() error
	io.Writer
}

type filerotator struct {
	logdir          string
	symlinks        []string
	filenamePrefix  string
	filename        string
	file            *os.File
	lv              Level
	writeToSameFile bool
	nbytes          uint64 // The number of bytes written to this file
	logStats        bool
}

func (fr *filerotator) Sync() error {
	if fr.file != nil {
		return fr.file.Sync()
	}
	return nil
}
func (fr *filerotator) Close() (err error) {
	if fr.file != nil {
		err = fr.file.Close()
		fr.file = nil
	}
	return err
}

func (fr *filerotator) Write(p []byte) (n int, err error) {
	var start time.Time
	start = time.Now()
	if fr.file != nil {
		n, err = fr.file.Write(p)
	}
	dur := time.Since(start) / time.Millisecond
	atomic.AddInt64(&stats.WaitingWriteMS, int64(dur))
	atomic.AddUint64(&fr.nbytes, uint64(n))

	if atomic.LoadUint64(&fr.nbytes) >= MaxSize {
		if atomic.LoadUint64(&fr.nbytes) >= MaxSize { // double check after get lock
			if err := fr.rotateFile(false, time.Now()); err != nil {
				return 0, err
			}
		}
	}

	return
}

// rotateFile closes the filerotator's file and starts a new one.
func (fr *filerotator) rotateFile(init bool, now time.Time) error {
	if fr.filename != "" && !init { // 对于filename非空的情况，使用固定文件名，不支持滚动
		return nil
	}

	var err error
	var tag string
	if fr.writeToSameFile {
		tag = ""
	} else {
		tag = "." + fr.lv.String()
	}

	if fr.filenamePrefix == "" {
		fr.filenamePrefix = env.GetEnv().GetAppName() + "."
	}

	var file *os.File
	file, _, err = createFile(fr.filename, fr.filenamePrefix, fr.logdir, fr.symlinks, tag, now)
	if file != nil && err == nil {
		if fr.file != nil {
			if fr.logStats {
				fr.file.WriteString(GetStats().String() + "\n")
			}

			fr.file.Close()
		}
		fr.file = file
		atomic.SwapUint64(&fr.nbytes, 0)
	}
	return err
}
