//+build !openbsd, !freebsd, !netbsd
//+build !cgo

package apm

import (
	"github.com/BurntSushi/cmd"
	"strings"
)

func GetBattMins() string {
	c := cmd.New("apm", "-m")
	if err := c.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(c.BufStdout.String())
}
