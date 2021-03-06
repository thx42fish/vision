package vision

import (
	"os"
	"runtime"
	"strings"
)

func HomeDir() string {
	if runtime.GOOS == "windows" {
		if home := os.Getenv("HOME"); len(home) > 0 {
			if _, err := os.Stat(home); err == nil {
				return home
			}
		}
		if homeDrive, homePath := os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"); len(homeDrive) > 0 && len(homePath) > 0 {
			homeDir := homeDrive + homePath
			if _, err := os.Stat(homeDir); err == nil {
				return homeDir
			}
		}
		if userProfile := os.Getenv("USERPROFILE"); len(userProfile) > 0 {
			if _, err := os.Stat(userProfile); err == nil {
				return userProfile
			}
		}
	}
	return os.Getenv("HOME")
}

func HomeAbs(path string) string {
	if strings.Index(path, "~") == 0 {
		path = strings.Replace(path, "~", HomeDir(), -1)
	}
	return path
}
