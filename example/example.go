package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"syscall"

	"github.com/brk0v/directio"
)

func main() {
	// Open file with O_DIRECT
	flags := os.O_WRONLY | os.O_EXCL | os.O_CREATE | syscall.O_DIRECT
	f, err := os.OpenFile("/tmp/mini.iso", flags, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Use directio writer
	dio, err := directio.New(f)
	if err != nil {
		log.Fatal(err)
	}
	defer dio.Flush()

	// Downloading iso image
	resp, err := http.Get("http://archive.ubuntu.com/ubuntu/dists/bionic/main/installer-amd64/current/images/netboot/mini.iso")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(dio, resp.Body)
}
