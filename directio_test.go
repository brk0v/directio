// +build linux

package directio

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var (
	bufsizes = []int{
		0, 16, 23, 32, 46, 64, 93, 128, 1024, 4096, 4197, 12384, 16384, 32754,
	}

	writesizes = []int{
		8192, 16384, 16, 23, 32, 0, 46, 8192, 64, 93, 128, 1024, 4096, 4197, 12384, 16384, 32754,
	}
)

func tmpDir(t testing.TB) (string, func()) {
	dir, err := ioutil.TempDir("", "directio-test-")
	if err != nil {
		t.Fatal(err)
	}

	clean := func() {
		os.RemoveAll(dir)
	}

	return dir, clean
}

func tmpFile(t testing.TB, dir string, prefix string) *os.File {
	flags := os.O_WRONLY | os.O_EXCL | os.O_CREATE | O_DIRECT
	f, err := os.OpenFile(filepath.Join(dir, fmt.Sprintf("foo-%s", prefix)), flags, 0666)
	if err != nil {
		t.Fatal(err)
	}

	return f
}

func TestWriter(t *testing.T) {
	data := make([]byte, 2<<16) // 128KB test data
	for i := 0; i < len(data); i++ {
		data[i] = byte(' ' + i%('~'-' '))
	}

	dir, clean := tmpDir(t)
	defer clean()

	// Write nwrite bytes using buffer size bs.
	// Check that the right amount makes it out
	// and that the data is correct.
	for i := 0; i < len(bufsizes); i++ {
		var context string
		twrite := 0
		bs := bufsizes[i]

		f := tmpFile(t, dir, fmt.Sprintf("%d", bs))
		dio, err := NewSize(f, bs)
		if err != nil {
			t.Fatal(err)
		}

		point := 0
		for j := 0; j < len(writesizes); j++ {
			nwrite := writesizes[j]
			twrite += nwrite
			context = fmt.Sprintf("nwrite=%d bufsize=%d", nwrite, bs)

			n, e1 := dio.Write(data[point : point+nwrite])
			if e1 != nil || n != nwrite {
				t.Errorf("%s: buf.Write %d = %d, %v", context, nwrite, n, e1)
				continue
			}
			point += nwrite
		}

		if e := dio.Flush(); e != nil {
			t.Errorf("%s: buf.Flush = %v", context, e)
		}

		fname := f.Name()
		f.Close()

		written, err := ioutil.ReadFile(fname)
		if err != nil {
			t.Fatal(err)
		}

		if len(written) != twrite {
			t.Errorf("%s: %d bytes written", context, len(written))
		}

		// Check content
		if !bytes.Equal(written, data[:twrite]) {
			t.Fatal("wrong bytes were written")
		}
	}
}
