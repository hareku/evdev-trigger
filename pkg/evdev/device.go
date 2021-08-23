package evdev

import evdev "github.com/gvalkov/golang-evdev"

type Device interface {
	Read() (*InputEvent, error)
}

type device struct {
	d *evdev.InputDevice
}

func NewDevice(d *evdev.InputDevice) Device {
	return &device{d}
}

func (d *device) Read() (*InputEvent, error) {
	e, err := d.d.ReadOne()
	if err != nil {
		return nil, err
	}
	return &InputEvent{
		Time:  e.Time,
		Type:  e.Type,
		Code:  e.Code,
		Value: e.Value,
	}, nil
}
