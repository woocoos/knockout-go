package test

import (
	"errors"
	"os"
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

func Tmp(ref string) string {
	if filepath.IsAbs(ref) {
		return ref
	}

	tmpPath := filepath.Join(basedir, "tmp")
	if err := os.MkdirAll(tmpPath, 0755); err != nil {
		panic(err)
	}
	return filepath.Join(tmpPath, ref)
}
