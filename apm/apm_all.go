package apm

import "fmt"

type APM_Power_Source int

const (
	APM_Source_Unknown = APM_Power_Source(iota)
	APM_Source_Wall
	APM_Source_Battery
)

type Apminfo struct {
	Battery_state, Ac_state, Battery_life uint8
	Minutes_left                          int
}

type Apmerror struct {
	Errcode int
	errmsg  string
}

func new_apmerror(errcode int) error {
	return &Apmerror{Errcode: errcode, errmsg: fmt.Sprintf("APM error 0x%08x", errcode)}
}

func (err *Apmerror) Error() string {
	return err.errmsg
}

func (err *Apmerror) String() string {
	return err.errmsg
}
