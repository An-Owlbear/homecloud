package deviceinfo

import (
	"github.com/alexedwards/argon2id"
)

// DeviceInfo stores information about a device
type DeviceInfo struct {
	DeviceId  string `json:"device_id" dynamodbav:"device_id"`
	DeviceKey string `json:"device_key" dynamodbav:"device_key"`
	Subdomain string `json:"subdomain" dynamodbav:"subdomain"`
}

func HashKey(key string) (string, error) {
	hash, err := argon2id.CreateHash(key, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func CheckKey(key string, hashedKey string) (bool, error) {
	matches, err := argon2id.ComparePasswordAndHash(key, hashedKey)
	if err != nil {
		return false, err
	}

	return matches, nil
}
