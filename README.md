# Linux Direct IO Writer

Direct IO writer using O_DIRECT

Example:

```go
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

```

Check that dio bypass linux pagecache using `vmtouch`:

```bash
$ vmtouch /tmp/mini.iso
           Files: 1
     Directories: 0
  Resident Pages: 1/16384  4K/64M  0.0061%
         Elapsed: 0.000356 seconds
```

or using my `https://github.com/brk0v/cpager` to check per cgroup pagecache usage:

```bash
$ sudo ~/go/bin/cpager /tmp/mini.iso
         Files: 1
   Directories: 0
Resident Pages: 1/16385 4K/64M 0.0%

 cgmem inode    percent       pages        path
           -     100.0%       16384        not charged
        2187       0.0%           1        /sys/fs/cgroup/memory/user.slice/user-1000.slice/session-3.scope
```
