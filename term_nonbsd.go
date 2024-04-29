//go:build !darwin && !freebsd && !netbsd && !openbsd && !windows
// +build !darwin,!freebsd,!netbsd,!openbsd,!windows

package termbox

import (
	"golang.org/x/sys/unix"
)

const (
	getTermios = unix.TCGETS
	setTermios = unix.TCSETS
)
