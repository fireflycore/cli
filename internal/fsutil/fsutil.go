package fsutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ModuleLineRegexp 用于从 go.mod 中识别 module 声明。
var ModuleLineRegexp = regexp.MustCompile(`^\s*module\s+(\S+)\s*$`)

// CopyDir 将一个目录完整复制到另一个目录。
func CopyDir(src, dst string) error {
	// 读取源目录信息，确认源路径存在并获取权限模式。
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	// 复制目录时源路径必须是目录。
	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}
	// 目标目录已存在时拒绝覆盖，避免误删或混入旧文件。
	if _, err = os.Stat(dst); err == nil {
		return fmt.Errorf("target already exists: %s", dst)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	// 创建目标目录，权限沿用源目录。
	if err = os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}
	// 复制目录内容。
	return copyDirContents(src, dst)
}

// copyDirContents 递归复制目录内容，调用方负责创建目标根目录。
func copyDirContents(src, dst string) error {
	// 读取当前源目录的子项。
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	// 逐个复制文件或子目录。
	for _, entry := range entries {
		// 组织当前子项的源路径。
		srcPath := filepath.Join(src, entry.Name())
		// 组织当前子项的目标路径。
		dstPath := filepath.Join(dst, entry.Name())
		// 读取子项详细信息以保留权限。
		info, err := entry.Info()
		if err != nil {
			return err
		}
		// 目录需要递归创建和复制。
		if entry.IsDir() {
			if err = os.MkdirAll(dstPath, info.Mode()); err != nil {
				return err
			}
			if err = copyDirContents(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}
		// 普通文件直接复制内容。
		if err = CopyFile(srcPath, dstPath, info.Mode()); err != nil {
			return err
		}
	}
	return nil
}

// CopyFile 将源文件复制为目标文件。
func CopyFile(src, dst string, mode os.FileMode) error {
	// 打开源文件用于读取。
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// 创建目标文件用于写入。
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	// 流式复制文件内容。
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

// ReplaceInTextFiles 遍历目录，对文本文件执行字符串替换。
func ReplaceInTextFiles(root string, replacements map[string]string, skipNames map[string]bool) error {
	// 空替换表无需遍历文件系统。
	if len(replacements) == 0 {
		return nil
	}
	// 递归遍历目标目录。
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// 跳过常见元数据目录。
		if entry.IsDir() {
			if skipNames[entry.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		// 跳过调用方指定的文件名。
		if skipNames[entry.Name()] {
			return nil
		}
		// 读取文件内容。
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		// 二进制文件不参与文本替换。
		if isBinary(data) {
			return nil
		}
		// 执行所有替换规则。
		next := string(data)
		for oldText, newText := range replacements {
			next = strings.ReplaceAll(next, oldText, newText)
		}
		// 内容没有变化时不写回，减少无意义的文件时间戳变动。
		if next == string(data) {
			return nil
		}
		// 保留原文件权限。
		info, err := entry.Info()
		if err != nil {
			return err
		}
		return os.WriteFile(path, []byte(next), info.Mode())
	})
}

// isBinary 用简单的 NUL 字节规则判断文件是否为二进制。
func isBinary(data []byte) bool {
	// 空文件按文本处理。
	if len(data) == 0 {
		return false
	}
	// 出现 NUL 字节通常意味着二进制内容。
	return strings.ContainsRune(string(data), '\x00')
}

// ReadModule 从 go.mod 中读取 module 名。
func ReadModule(root string) (string, error) {
	// 读取 go.mod 文件。
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", err
	}
	// 逐行查找 module 声明。
	for _, line := range strings.Split(string(data), "\n") {
		matches := ModuleLineRegexp.FindStringSubmatch(line)
		if len(matches) == 2 {
			return matches[1], nil
		}
	}
	return "", fmt.Errorf("module line not found in go.mod")
}

// WriteModule 将 go.mod 中的 module 声明改为指定值。
func WriteModule(root, module string) error {
	// 读取 go.mod 内容。
	path := filepath.Join(root, "go.mod")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	// 替换 module 声明。
	lines := strings.Split(string(data), "\n")
	for index, line := range lines {
		if ModuleLineRegexp.MatchString(line) {
			lines[index] = "module " + module
			return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
		}
	}
	return fmt.Errorf("module line not found in go.mod")
}

// UpdateJSONFile 读取 JSON 文件并允许调用方修改通用 map。
func UpdateJSONFile(path string, update func(map[string]any)) error {
	// 读取原始 JSON 文件。
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	// 用 map 保留调用方不关心的字段。
	var raw map[string]any
	if err = json.Unmarshal(data, &raw); err != nil {
		return err
	}
	// 调用方执行具体修改。
	update(raw)
	// 格式化输出，保持配置可读。
	next, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	// 文本文件末尾补换行。
	next = append(next, '\n')
	return os.WriteFile(path, next, 0644)
}

// SetNestedString 按路径写入嵌套字符串字段。
func SetNestedString(raw map[string]any, path []string, value string) {
	// 空路径或空值不写入。
	if len(path) == 0 || strings.TrimSpace(value) == "" {
		return
	}
	// 只剩最后一段时直接赋值。
	if len(path) == 1 {
		raw[path[0]] = value
		return
	}
	// 确保下一层是 map。
	next, ok := raw[path[0]].(map[string]any)
	if !ok {
		next = make(map[string]any)
		raw[path[0]] = next
	}
	// 递归写入剩余路径。
	SetNestedString(next, path[1:], value)
}
