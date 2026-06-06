package config

// CLI_NAME 是 CLI 在缓存目录中的固定名称。
const CLI_NAME = "firefly"

// CLI_CONFIG_FILE_NAME 是全局配置文件名，不包含扩展名。
const CLI_CONFIG_FILE_NAME = "cli"

// CLI_CONFIG_FILE_TYPE 是全局配置文件格式。
const CLI_CONFIG_FILE_TYPE = "yaml"

// RELEASE 是 CLI 当前发布版本。
const RELEASE = "v0.0.6"

// REPO_OWNER 是模板仓库所在的 GitHub 组织。
const REPO_OWNER = "fireflycore"

// REPO_TOKEN 是访问 GitHub API 的可选 token，当前默认不内置。
const REPO_TOKEN = ""

// LANGUAGE 是 create 命令当前支持的语言列表。
var LANGUAGE = []string{
	// Go 是当前唯一支持的主线模板语言。
	"Go",
	// 下面的语言是未来扩展预留项，当前不参与交互选择。
	//"Rust",
	//"Dart",
	//"Swift",
	//"Kotlin",
	//"Python",
	//"Node.js",
	//"Java",
	//"PHP",
	//"C++",
	//"C#",
	//"Ruby",
}

// IGNORE_DIRS 定义模板替换时需要跳过的目录。
var IGNORE_DIRS = map[string]map[string]bool{
	// go 模板不替换 git 元数据和 GitHub workflow。
	"go": {
		".git":    true,
		".github": true,
	},
}

// IGNORE_FILES 定义模板替换时需要跳过的文件。
var IGNORE_FILES = map[string]map[string]bool{
	// go 模板中的这些文件不参与全局字符串替换。
	"go": {
		".gitignore":  true,
		"config.yaml": true,
		"go.sum":      true,
		"LICENSE":     true,
		"run.sh":      true,
		"README.md":   true,
	},
}
