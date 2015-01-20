package main

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func ExecPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	p, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	return p, nil
}

// WorkDir returns absolute path of work directory.
func WorkDir() (string, error) {
	execPath, err := ExecPath()
	return path.Dir(strings.Replace(execPath, "\\", "/", -1)), err
}

func getWorkDir() string {
	if dir, err := WorkDir(); err != nil {
		return "."
	} else {
		return dir
	}
}

// IsFile returns true if given path is a file,
// or returns false when it's a directory or does not exist.
func IsFile(filePath string) bool {
	f, e := os.Stat(filePath)
	if e != nil {
		return false
	}
	return !f.IsDir()
}

// IsExist checks whether a file or directory exists.
// It returns false when the file or directory does not exist.
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// Create file and parent dirs.
func CreateFile(fpath string) (file *os.File, err error) {
	if err := os.MkdirAll(path.Dir(fpath), os.ModePerm); err != nil {
		return nil, err
	}
	return os.Create(fpath)
}

func staticPath(subdir string) string {
	return path.Join(app.config.static_root, subdir)
}

func siteConfig(cfgname string, devmode bool) string {
	if devmode {
		if dir, err := WorkDir(); err != nil {
			return path.Join("conf", cfgname)
		} else {
			return path.Join(dir, "conf", cfgname)
		}
	} else {
		return path.Join(app.dir.SiteConfig(), cfgname)
	}
}

func userConfig(cfgname string, devmode bool) string {
	if devmode {
		if dir, err := WorkDir(); err != nil {
			return path.Join("custom", cfgname)
		} else {
			return path.Join(dir, "custom", cfgname)
		}
	} else {
		return path.Join(app.dir.UserConfig(), cfgname)
	}
}
