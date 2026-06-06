package file

import (
	"fmt"
	"github.com/fireflycore/cli/pkg/config"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CopyDir 将源目录完整复制到目标目录。
// src 是源目录路径，dst 是目标目录路径。
func CopyDir(src, dst string) error {
	// 读取源路径信息，用于确认源路径存在以及后续复用权限模式。
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 源路径必须是目录，文件复制交给 CopyFile 处理。
	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	// 先创建目标目录，权限沿用源目录权限。
	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	// 读取源目录下的一级条目。
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// 逐个复制源目录中的文件和子目录。
	for _, entry := range entries {
		// 拼出当前条目的源路径。
		srcPath := filepath.Join(src, entry.Name())
		// 拼出当前条目的目标路径。
		dstPath := filepath.Join(dst, entry.Name())

		// 子目录需要递归复制。
		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// 普通文件直接复制内容。
			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	// 全部条目复制完成。
	return nil
}

// CopyFile 将源文件内容复制到目标文件。
// src 是源文件路径，dst 是目标文件路径。
func CopyFile(src, dst string) error {
	// 打开源文件用于读取。
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	// 复制结束后关闭源文件。
	defer in.Close()

	// 创建或截断目标文件用于写入。
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	// 复制结束后关闭目标文件。
	defer out.Close()

	// 将源文件内容流式复制到目标文件。
	_, err = io.Copy(out, in)
	return err
}

// ReplaceInFile 将文件中的 oldText 全量替换为 newText。
func ReplaceInFile(filePath string, oldText, newText string) error {
	// 读取整个文件内容，模板替换场景中文件通常较小。
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// 执行全量字符串替换。
	newContent := strings.ReplaceAll(string(content), oldText, newText)

	// 将替换后的内容写回原文件。
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return err
	}

	// 替换完成。
	return nil
}

// WalkDirAndReplace 遍历目录并按语言忽略规则替换文件内容。
func WalkDirAndReplace(language, dirPath, oldText, newText string) error {
	// 使用 filepath.Walk 递归遍历目录及其子目录。
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		// 遍历过程中出现的错误直接向上传递。
		if err != nil {
			return err
		}

		// 将当前路径转成相对路径，便于按目录片段匹配忽略规则。
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}
		// 拆分相对路径，逐段检查是否命中忽略目录。
		parts := strings.Split(relPath, string(os.PathSeparator))
		for _, part := range parts {
			// 命中当前语言忽略目录时，目录跳过递归，目录内文件直接忽略。
			if config.IGNORE_DIRS[language][part] {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// 目录本身不需要替换内容。
		if info.IsDir() {
			return nil
		}

		// 命中当前语言忽略文件时直接跳过。
		if config.IGNORE_FILES[language][filepath.Base(path)] {
			return nil
		}

		// 对普通文件执行内容替换。
		err = ReplaceInFile(path, oldText, newText)
		if err != nil {
			return err
		}
		// 输出替换文件路径，方便 create 时观察模板替换过程。
		fmt.Printf("Replaced in %s\n", path)
		return nil
	})
}
