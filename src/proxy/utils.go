package proxy

import (
	"os"
	"errors"
	"fmt"
	"crypto/md5"
	"io"
	"encoding/hex"
	"strings"
)

func CheckFileIsDirectory(path string) (bool, error)  {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	if fi.IsDir() == false {
		return false, errors.New("target file is not folder")
	}

	return true, nil
}

func GetFileSize(file string) (int64, error) {
	fi, err := os.Stat(file)
	if err != nil {
		return 0, err
	}

	if fi.IsDir() {
		return 0, errors.New(fmt.Sprintf("target file %s is not file", file))
	}

	return fi.Size(), nil
}

func InStringArray(value string, arrays []string) bool {
	for _, v := range arrays {
		if v == value {
			return true
		}
	}

	return false
}

func GetFileMD5sum(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}

	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func HasIntersection(a []string, b []string) bool  {
	if len(a) == 0 || len(b) == 0 {
		return false
	}

	t := strings.Join(b, "%") + "%"
	for _,v := range a {
		if strings.Contains(t, v + "%") {
			return true
		}
	}

	return false
}