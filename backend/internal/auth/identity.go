package auth

import (
	"errors"
	"github.com/go-viper/mapstructure/v2"
)

type Traits struct {
	email    string
	username string
}

type MetadataPublic struct {
	Roles []string `mapstructure:"roles" json:"roles"`
}

var InvalidTypeError = errors.New("invalid input type")

// ParseMetadataPublic parses the public metadata an Ory Kratos Identity
func ParseMetadataPublic(unparsed interface{}) (*MetadataPublic, error) {
	asserted, ok := unparsed.(map[string]interface{})
	if !ok {
		return nil, InvalidTypeError
	}

	var decoded MetadataPublic
	err := mapstructure.Decode(asserted, &decoded)
	if err != nil {
		return nil, err
	}

	return &decoded, nil
}
