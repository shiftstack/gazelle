package utils

import "os/user"

func GetUsername() (string) {
	u, err := user.Current()
	if err != nil {
		return "cireport"
	}
	return u.Username
}
