package test

import (
	"errors"
	"path/filepath"
	"runtime"
)

var basedir string

func init() {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		panic(errors.New("runtime.Caller error at test init"))
	}
	basedir = filepath.Dir(currentFile)
}

func BaseDir() string {
	return basedir
}

func Path(ref string) string {
	return filepath.Join(basedir, ref)
}
