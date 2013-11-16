package main

import (
	"code.google.com/p/go-imap/go1/imap"
	"errors"
	"fmt"
)

type xoauth []byte

func GetXOAuth(user string, accessToken string) (xoauth []byte) {
	oauth2string := fmt.Sprint("user=", user, "\001auth=Bearer ", accessToken, "\001\001")
	return []byte(oauth2string)
}

func XoAuth(user string, accessToken string) imap.SASL {
	return xoauth(GetXOAuth(user, accessToken))
}

func (a xoauth) Start(s *imap.ServerInfo) (mech string, ir []byte, err error) {
	return "XOAUTH2", a, nil
}

func (a xoauth) Next(challenge []byte) (response []byte, err error) {
	return nil, errors.New(fmt.Sprint("Authentication failed got ", string(response)))
}
