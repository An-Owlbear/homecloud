package config

import "os"

type Storage struct {
	DataPath string
}

func NewStorage() *Storage {
	return &Storage{
		DataPath: os.Getenv("DATA_PATH"),
	}
}
