package file

import (
	"fmt"
	"github.com/fireflycore/cli/pkg/config"
	"io"
	"os"
	"path/filepath"
	"strings"
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

func ReplaceInFile(filePath string, oldText, newText string) error {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// 替换文本
	newContent := strings.ReplaceAll(string(content), oldText, newText)

	// 写入新内容到文件
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return err
	}

	return nil
}

func WalkDirAndReplace(language, dirPath, oldText, newText string) error {
	// 遍历目录及其子目录
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // 返回任何遍历时遇到的错误
		}

		// 检查是否需要忽略当前目录
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}
		parts := strings.Split(relPath, string(os.PathSeparator))
		for _, part := range parts {
			if config.IgnoreDirs[language][part] {
				// 如果是要忽略的目录，则返回nil以跳过该目录及其子目录
				if info.IsDir() {
					return filepath.SkipDir
				}
				// 如果当前文件位于要忽略的目录下，则也忽略该文件
				return nil
			}
		}

		// 检查是否需要忽略当前文件
		if info.IsDir() {
			// 目录不需要替换内容，只需检查是否需要忽略
			return nil
		}

		if config.IgnoreFiles[language][filepath.Base(path)] {
			// 如果是要忽略的文件，则直接返回nil
			return nil
		}

		// 替换文件内容
		err = ReplaceInFile(path, oldText, newText)
		if err != nil {
			return err // 返回替换文件内容时遇到的错误
		}
		fmt.Printf("Replaced in %s\n", path)
		return nil
	})
}
