package config

const CLI_NAME = "firefly"
const CLI_CONFIG_FILE_NAME = "cli"
const CLI_CONFIG_FILE_TYPE = "yaml"

const REPO_OWNER = "lhdhtrc"
const REPO_TOKEN = ""

var Language = []string{
	"Go",
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

var IgnoreDirs = map[string]map[string]bool{
	"go": {
		".git":    true,
		".github": true,
	},
}

var IgnoreFiles = map[string]map[string]bool{
	"go": {
		".gitignore":  true,
		"config.yaml": true,
		"go.sum":      true,
		"LICENSE":     true,
		"run.sh":      true,
		"README.md":   true,
	},
}
