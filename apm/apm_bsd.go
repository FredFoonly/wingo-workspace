//+build openbsd freebsd netbsd
//+build cgo

package apm

// #include <stdlib.h>
// #include <fcntl.h>
// #include <machine/apmvar.h>
// int wrapper_getbattstat(struct apm_power_info* pwr)
// {
//   int fd, rc;
//   fd = open("/dev/apm", O_RDONLY);
//   rc = ioctl(fd, APM_IOC_GETPOWER, pwr);
//   close(fd);
//   return rc;
// }
import "C"

//import "fmt"

type APM_Battery_State uint8

const (
	APM_BATT_HIGH      = APM_Battery_State(C.APM_BATT_HIGH)      // Battery has a high state of charge.
	APM_BATT_LOW       = APM_Battery_State(C.APM_BATT_LOW)       // Battery has a low state of charge.
	APM_BATT_CRITICAL  = APM_Battery_State(C.APM_BATT_CRITICAL)  // Battery has a critical state of charge.
	APM_BATT_CHARGING  = APM_Battery_State(C.APM_BATT_CHARGING)  // Battery is not high, low, or critical and is currently charging.
	APM_BATT_UNKNOWN   = APM_Battery_State(C.APM_BATT_UNKNOWN)   // Can not read the current battery state.
	APM_BATTERY_ABSENT = APM_Battery_State(C.APM_BATTERY_ABSENT) // No battery installed.
)

type APM_AC_State uint8

const (
	APM_AC_OFF     = APM_AC_State(C.APM_AC_OFF)     // External power not detected.
	APM_AC_ON      = APM_AC_State(C.APM_AC_ON)      // External power detected.
	APM_AC_BACKUP  = APM_AC_State(C.APM_AC_BACKUP)  // Backup power in use.
	APM_AC_UNKNOWN = APM_AC_State(C.APM_AC_UNKNOWN) // External power state unknown.
)

func GetBattMins() (APM_Power_Source, int, error) {
	apmstat, err := Getapminfo()
	if err != nil {
		return APM_Source_Unknown, -1, err
	}
	if apmstat.Minutes_left < 0 {
		return APM_Source_Wall, -1, nil
	}
	return APM_Source_Battery, apmstat.Minutes_left, nil
}

func Getapminfo() (Apminfo, error) {
	apmpwr := new(C.struct_apm_power_info)
	rc := C.wrapper_getbattstat(apmpwr)
	if rc == 0 {
		apmstat := Apminfo{
			Battery_state: uint8(apmpwr.battery_state),
			Ac_state:      uint8(apmpwr.ac_state),
			Battery_life:  uint8(apmpwr.battery_life),
			Minutes_left:  int(apmpwr.minutes_left),
		}
		if apmstat.Minutes_left > 0x7FFFFFFF {
			apmstat.Minutes_left = -int((^apmpwr.minutes_left) + 1)
		}
		return apmstat, nil
	} else {
		return Apminfo{}, new_apmerror(int(rc))
	}
}
