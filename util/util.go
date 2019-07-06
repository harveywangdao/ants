package util

import (
	"github.com/satori/go.uuid"
	"os"
	"strings"
)

func GetUUID() string {
	u, err := uuid.NewV4()
	if err != nil {
		return ""
	}

	return strings.Replace(u.String(), "-", "", -1)
}

func IsFileExist(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}

	/*  if os.IsNotExist(err) {
	    return false
	  }*/

	return false
}
