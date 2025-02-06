package config

import (
	"fmt"
	"net/url"
	"os"
)

type OryService struct {
	PrivateAddress url.URL
	PublicAddress  url.URL
}

type Ory struct {
	Kratos      OryService
	KratosAdmin OryService
	Hydra       OryService
}

func newOryService(urlString string) (*OryService, error) {
	privateUrl, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	publicUrl := *privateUrl
	publicUrl.Host = fmt.Sprintf("%s.%s:%s", privateUrl.Hostname(), os.Getenv("HOMECLOUD_HOST"), os.Getenv("HOMECLOUD_PORT"))
	fmt.Println(publicUrl.String())

	return &OryService{
		PrivateAddress: *privateUrl,
		PublicAddress:  publicUrl,
	}, nil
}

func OryFromEnv() (*Ory, error) {
	kratosService, err := newOryService(os.Getenv("KRATOS_URL"))
	if err != nil {
		return nil, err
	}

	kratosAdminService, err := newOryService(os.Getenv("KRATOS_ADMIN_URL"))
	if err != nil {
		return nil, err
	}

	hydraService, err := newOryService(os.Getenv("HYDRA_URL"))
	if err != nil {
		return nil, err
	}

	return &Ory{
		Kratos:      *kratosService,
		KratosAdmin: *kratosAdminService,
		Hydra:       *hydraService,
	}, nil
}
