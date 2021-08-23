package evdev

import (
	"errors"
	"fmt"

	evdev "github.com/gvalkov/golang-evdev"
)

var ErrDeviceNotFound = errors.New("device not found")

type Finder interface {
	Find(phys string) (Device, error)
}

type finder struct{}

func NewFinder() Finder {
	return &finder{}
}

func (f *finder) Find(phys string) (Device, error) {
	devices, err := evdev.ListInputDevices()
	if err != nil {
		return nil, fmt.Errorf("listing input devices failed: %w", err)
	}

	for _, d := range devices {
		if d.Phys == phys {
			return NewDevice(d), nil
		}
	}
	return nil, ErrDeviceNotFound
}
