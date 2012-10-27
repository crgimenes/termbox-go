// +build !windows
// This file contains a simple and incomplete implementation of the terminfo
// database. Information was taken from the ncurses manpages term(5) and
// terminfo(5). Currently, only the string capabilities for special keys and for
// functions without parameters are actually used. Colors are still done with
// ANSI escape sequences. Other special features that are not (yet?) supported
// are reading from ~/.terminfo, the TERMINFO_DIRS variable, Berkeley database
// format and extended capabilities.

package termbox

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
)

const (
	ti_magic         = 0432
	ti_header_length = 12
)

var fallbackTry bool

func load_terminfo() ([]byte, error) {
	var data []byte
	var err error

	term := os.Getenv("TERM")
	if term == "" {
		return nil, fmt.Errorf("termbox: TERM not set")
	}

	// The following behaviour follows the one described in terminfo(5) as
	// distributed by ncurses.

	terminfo := os.Getenv("TERMINFO")
	if terminfo != "" {
		// if TERMINFO is set, no other directory should be searched
		return ti_try_path(terminfo)
	}

	// next, consider ~/.terminfo
	home := os.Getenv("HOME")
	if home != "" {
		data, err = ti_try_path(home + "/.terminfo")
		if err == nil {
			return data, nil
		}
	}

	// next, TERMINFO_DIRS
	dirs := os.Getenv("TERMINFO_DIRS")
	if dirs != "" {
		for _, dir := range strings.Split(dirs, ":") {
			if dir == "" {
				// "" -> "/usr/share/terminfo"
				dir = "/usr/share/terminfo"
			}
			data, err = ti_try_path(dir)
			if err == nil {
				return data, nil
			}
		}
	}

	// fall back to /usr/share/terminfo
	fallbackTry = true
	return ti_try_path("/usr/share/terminfo")
}

func ti_try_path(path string) (data []byte, err error) {
	// load_terminfo already made sure it is set
	term := os.Getenv("TERM")

	// first try, the typical *nix path
	terminfo := path + "/" + term[0:1] + "/" + term
	data, err = ioutil.ReadFile(terminfo)
	if err == nil {
		return
	}
	if fallbackTry == true && term == "xterm" && runtime.GOOS == "linux" {
		terminfo := "/lib/terminfo/x/" + term
		data, err = ioutil.ReadFile(terminfo)
		if err == nil {
			return
		}

	}
	// fallback to darwin specific dirs structure
	terminfo = path + "/" + hex.EncodeToString([]byte(term[:1])) + "/" + term
	data, err = ioutil.ReadFile(terminfo)

	if fallbackTry == true && err != nil {
		println("")
		panic(" Ubuntu users need to run these commands to install termbox's dependencies.\n" +
			"\t$TERM should also be set to either \"xterm\" or \"xterm-*\"" +
			" for this pkg to function correctly.\n\n" +
			"\tsudo apt-get install ncurses-base\n" +
			"\techo $TERM")
	}
	return
}

func setup_term() (err error) {
	var data []byte
	var header [6]int16
	var str_offset, table_offset int16

	data, err = load_terminfo()
	if err != nil {
		return
	}

	rd := bytes.NewReader(data)
	// 0: magic number, 1: size of names section, 2: size of boolean section, 3:
	// size of numbers section (in integers), 4: size of the strings section (in
	// integers), 5: size of the string table

	err = binary.Read(rd, binary.LittleEndian, header[:])
	if err != nil {
		return
	}

	if (header[1]+header[2])%2 != 0 {
		// old quirk to align everything on word boundaries
		header[2] += 1
	}
	str_offset = ti_header_length + header[1] + header[2] + 2*header[3]
	table_offset = str_offset + 2*header[4]

	keys = make([]string, 0xFFFF-key_min)
	for i, _ := range keys {
		keys[i], err = ti_read_string(rd, str_offset+2*ti_keys[i], table_offset)
		if err != nil {
			return
		}
	}
	funcs = make([]string, t_max_funcs)
	for i, _ := range funcs {
		funcs[i], err = ti_read_string(rd, str_offset+2*ti_funcs[i], table_offset)
		if err != nil {
			return
		}
	}
	err = nil
	return
}

func ti_read_string(rd *bytes.Reader, str_off, table int16) (string, error) {
	var off int16

	_, err := rd.Seek(int64(str_off), 0)
	if err != nil {
		return "", err
	}
	err = binary.Read(rd, binary.LittleEndian, &off)
	if err != nil {
		return "", err
	}
	_, err = rd.Seek(int64(table+off), 0)
	if err != nil {
		return "", err
	}
	var bs []byte
	for {
		b, err := rd.ReadByte()
		if err != nil {
			return "", err
		}
		if b == byte(0x00) {
			break
		}
		bs = append(bs, b)
	}
	return string(bs), nil
}

// "Maps" the function constants from termbox.go to the number of the respective
// string capability in the terminfo file. Taken from (ncurses) term.h.
var ti_funcs = []int16{
	28, 40, 16, 13, 5, 39, 36, 27, 26, 34, 89, 88,
}

// Same as above for the special keys.
var ti_keys = []int16{
	66, 68 /* apparently not a typo; 67 is F10 for whatever reason */, 69, 70,
	71, 72, 73, 74, 75, 67, 216, 217, 77, 59, 76, 164, 82, 81, 87, 61, 79, 83,
}
