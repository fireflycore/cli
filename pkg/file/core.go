package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyDir 复制一个目录到另一个位置
// src 是源目录的路径，dst 是目标目录的路径
func CopyDir(src, dst string) error {
	// 获取源目录的信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 如果源不是一个目录，返回错误
	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	// 创建目标目录
	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	// 打开源目录
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// 遍历源目录中的每个条目
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// 如果是文件，则复制文件
		if entry.IsDir() {
			// 如果是目录，则递归复制
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// 复制文件
			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyFile 复制一个文件到另一个位置
// src 是源文件的路径，dst 是目标文件的路径
func CopyFile(src, dst string) error {
	// 打开源文件
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// 创建目标文件
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	// 复制文件内容
	_, err = io.Copy(out, in)
	return err
}
