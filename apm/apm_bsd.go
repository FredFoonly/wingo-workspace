//+build openbsd freebsd netbsd
//+build cgo

package apm

// #include <stdlib.h>
// #include <fcntl.h>
// #include <machine/apmvar.h>
// int wrapper_getbattstat(unsigned char* battery_state,
//                         unsigned char *ac_state,
//                         unsigned char* battery_life,
//                         int* minutes_left)
// {
//   int fd, rc;
//   struct apm_power_info pwr;
//   fd = open("/dev/apm", O_RDONLY);
//   rc = ioctl(fd, APM_IOC_GETPOWER, &pwr);
//   *battery_state = pwr.battery_state;
//   *ac_state = pwr.ac_state;
//   *battery_life = pwr.battery_life;
//   *minutes_left = pwr.minutes_left;
//   close(fd);
//   return rc;
// }
//
import "C"

import "fmt"

func GetBattMins() string {
	apmstat, err := GetapmStats()
	if err != nil {
		return ""
	}
	if apmstat.Minutes_left < 0 {
		return "unknown"
	}
	return fmt.Sprintf("%d", apmstat.Minutes_left)
}

func GetapmStats() (Apminfo, error) {
	var battery_state, ac_state, battery_life C.uchar
	var minutes_left C.int
	rc := C.wrapper_getbattstat(&battery_state, &ac_state, &battery_life, &minutes_left)
	if rc == 0 {
		apmstat := Apminfo{
			Battery_state: uint8(battery_state),
			Ac_state:      uint8(ac_state),
			Battery_life:  uint8(battery_life),
			Minutes_left:  int(minutes_left),
		}
		return apmstat, nil
	} else {
		return Apminfo{}, new_apmerror(int(rc))
	}
}
