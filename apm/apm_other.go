//+build !openbsd, !freebsd, !netbsd
//+build !cgo

package apm

import (
	"github.com/BurntSushi/cmd"
	"strings"
	"strconv"
)

func GetBattMins() (APM_Power_Source, int, error) {
	c := cmd.New("apm", "-m")
	if err := c.Run(); err != nil {
		return APM_Source_Unknown, -1, err
	}
	sreslt := strings.TrimSpace(c.BufStdout.String())
	if sreslt == "unknown" {
		return APM_Source_Wall, -1, nil
	} else {
		mins, err := strconv.Atoi(sreslt)
		if err != nil {
			return APM_Source_Unknown, -1, err
		}
		return APM_Source_Battery, mins, nil
	}
}
