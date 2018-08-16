package gost

import (
	"io"
	"sync"
)

const (
	BUFSZ = 64 * 1024
)

var (
	bufpool = sync.Pool{
		New: func() interface{} {
			return make([]byte, BUFSZ)
		},
	}
)

func IOCopy(dst io.Writer, src io.Reader) (written int64, err error) {
	buf := bufpool.Get().([]byte)
	written, err = io.CopyBuffer(dst, src, buf)
	bufpool.Put(buf)
	return written, err
}

type ReaderWrapper struct {
	reader io.Reader
	ended *bool
	readfromclient bool
}

func newReaderWrapper(r io.Reader, ended *bool, readfromclient bool) *ReaderWrapper {
	return &ReaderWrapper{
		reader: r,
		ended: ended,
		readfromclient: readfromclient,
	}
}

func (wp *ReaderWrapper) Read(b []byte) (n int, err error) {
	if (*wp.ended) {
		return 0, io.EOF
	}
	n, err = wp.reader.Read(b)
	// if wp.readfromclient {
	// 	glog.Infof("wrapper: read %v bytes, err = %v", n, err)
	// }
	return
}

type WriterWrapper struct {
	writer io.Writer
	ended *bool
	readfromclient bool
}

func newWriterWrapper(w io.Writer, ended *bool, readfromclient bool) *WriterWrapper {
	return &WriterWrapper{
		writer: w,
		ended: ended,
		readfromclient: readfromclient,
	}
}

func (wp *WriterWrapper) Write(b []byte) (n int, err error) {
	if (*wp.ended) {
		return 0, io.EOF
	}
	// if !wp.readfromclient {
	// 	glog.Infof("wrapper: trying to write %v bytes", len(b))
	// }
	n, err = wp.writer.Write(b)
	// if !wp.readfromclient {
	// 	glog.Infof("wrapper: written %v bytes, err = %v", n, err)
	// }
	return
}

func IOCopyDebug(dst io.ReadWriter, src io.ReadWriter, readfromclient bool, ended *bool, die chan int) (written int64, err error) {
	// c := readfromclient
	// glog.Warningf("IOCopy(read from client = %v) started", c)

	written, err = IOCopy(newWriterWrapper(dst, ended, readfromclient), newReaderWrapper(src, ended, readfromclient))
	*ended = true

	// if err != nil {
	// 	glog.Warningf("error in IOCopy (read from client = %v): written = %v, err = %v", c, written, err)
	// } else {
	// 	glog.Warningf("IOCopy(read from client = %v) ended with no error, written = %v", c, written)
	// }
	// glog.Warningf("IOCopy(read from client = %v) close called!", c)
	if !readfromclient {
		die <- 0
	}
	return
}

func CopyTwoStreams(a, b io.ReadWriter) {
	ended := false
	die := make(chan int)
	go IOCopyDebug(a, b, false, &ended, die)
	IOCopyDebug(b, a, true, &ended, die)
	<-die
}
