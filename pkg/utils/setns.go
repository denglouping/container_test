package utils

import (
	"fmt"
	"syscall"
)

func Setns(filePath string) {
	fd, err := syscall.Open(filePath, syscall.O_RDONLY, 0)
	if err != nil {
		fmt.Println("无法打开命名空间文件:", err)
		return
	}
	defer syscall.Close(fd)

	// 调用setns函数切换到指定的命名空间
	_, _, errno := syscall.Syscall(syscall., uintptr(fd), syscall.CLONE_NEWNS, 0)
	if errno != 0 {
		fmt.Println("setns调用失败:", errno)
		return
	}

	fmt.Println("成功切换到命名空间")
}