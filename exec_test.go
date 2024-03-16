package main

import (
	"io/ioutil"
	"k8s.io/klog/v2"
	"os"
	"syscall"
	"testing"
)

func TestGetPdf(t *testing.T) {
	if err := syscall.Exec("/usr/bin/sleep", []string{"sleep", "10"}, os.Environ()); err != nil {
		ioutil.WriteFile("log3", []byte("exec err : "+err.Error()), 0644)
		klog.Fatalf(err.Error())
	}
}
