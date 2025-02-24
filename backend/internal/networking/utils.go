package networking

import (
	"errors"
	"io"
	"net"
	"net/http"
)

var InvalidIPError = errors.New("invalid IP address")

// GetPublicIP retrieves the public IP address of the device, this requires making an external API call
func GetPublicIP() (string, error) {
	req, err := http.Get("https://api.ipify.org/")
	if err != nil {
		return "", err
	}

	bodyData, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyData)

	if parsedIp := net.ParseIP(bodyString); parsedIp == nil {
		return "", InvalidIPError
	}

	return bodyString, nil
}

func GetPrivateIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}
