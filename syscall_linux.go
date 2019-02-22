package directio

import (
	"errors"
	"syscall"
)

const (
	O_DIRECT = syscall.O_DIRECT
)

var ErrNotSetDirectIO = errors.New("O_DIRECT flag is absent")

func fcntl(fd uintptr, cmd uintptr, arg uintptr) (uintptr, error) {
	r0, _, e1 := syscall.Syscall(syscall.SYS_FCNTL, fd, uintptr(cmd), uintptr(arg))
	if e1 != 0 {
		return 0, e1
	}

	return r0, nil
}

func checkDirectIO(fd uintptr) error {
	flags, err := fcntl(fd, syscall.F_GETFL, 0)
	if err != nil {
		return err
	}

	if (flags & O_DIRECT) == O_DIRECT {
		return nil
	}

	return ErrNotSetDirectIO
}

func setDirectIO(fd uintptr, dio bool) error {
	flag, err := fcntl(fd, syscall.F_GETFL, 0)
	if err != nil {
		return err
	}

	if dio {
		flag |= O_DIRECT
	} else {
		flag &^= O_DIRECT
	}

	_, err = fcntl(fd, syscall.F_SETFL, flag)
	return err
}
