package config

import "os"

type Store struct {
	StoreUrl string
}

func NewStore() *Store {
	return &Store{
		StoreUrl: os.Getenv("STORE_URL"),
	}
}
