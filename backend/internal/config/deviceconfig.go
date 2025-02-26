package config

import "os"

type DeviceConfig struct {
	DeviceId  string `json:"device_id"`
	DeviceKey string `json:"device_key"`
}

func NewDeviceConfig() DeviceConfig {
	return DeviceConfig{
		DeviceId:  os.Getenv("DEVICE_ID"),
		DeviceKey: os.Getenv("DEVICE_KEY"),
	}
}
