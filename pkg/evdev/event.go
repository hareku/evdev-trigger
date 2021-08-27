package evdev

import "syscall"

const (
	EV_SYN = uint16(0x00)
	EV_KEY = uint16(0x01)
	EV_REL = uint16(0x02)
)

type InputEvent struct {
	Time  syscall.Timeval // time in seconds since epoch at which event occurred
	Type  uint16          // event type - one of ecodes.EV_*
	Code  uint16          // event code related to the event type
	Value int32           // event value related to the event type
}
