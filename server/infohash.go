package main

import (
	"fmt"
	"net/url"
	"regexp"
)

func GetInfoHash(magnet string) (string, error) {
	u, err := url.Parse(magnet)
	if err != nil {
		return "", err
	}
	if u.Scheme != "magnet" {
		return "", fmt.Errorf("invalid magnet: url of is of scheme %s", u.Scheme)
	}

	q := u.Query()

	xt, ok := q["xt"]
	if !ok {
		return "", fmt.Errorf("invalid magnet: missing \"xt\" parameter")
	}
	if len(xt) != 1 {
		return "", fmt.Errorf("invalid magnet: invalid \"xt\" parameter")
	}

	urn := xt[0]
	if urn[0:9] != "urn:btih:" {
		return "", fmt.Errorf("invalid magnet: invalid urn")
	}

	hash := urn[9:]

	if m, _ := regexp.Match(`^[0-9A-Fa-f]{40}$`, []byte(hash)); !m {
		return "", fmt.Errorf("invalid magnet: invalid hash")
	}

	return hash, nil
}
