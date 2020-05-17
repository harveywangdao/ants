package util

import (
	"github.com/satori/go.uuid"
	"os"
	"strings"
)

func GetUUID() string {
	u := uuid.NewV4()
	return strings.Replace(u.String(), "-", "", -1)
}

// 文件或目录是否存在
func IsPathExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	/*  if os.IsNotExist(err) {
	    return false
	  }*/

	return false
}

// 目录是否存在
func IsDir(dirPath string) bool {
	dir, err := os.Stat(dirPath)
	if err != nil {
		return false
	}

	return dir.IsDir()
}
