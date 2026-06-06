package repo

// GithubRepoVersion 表示 GitHub release API 中 CLI 关心的版本字段。
type GithubRepoVersion struct {
	// TagName 是 release tag 名称，用作模板版本号。
	TagName string `json:"tag_name"`
}

// ReadmeEntity 表示渲染新项目 README.md 时使用的数据。
type ReadmeEntity struct {
	// Project 是生成后的项目名。
	Project string
	// Language 是项目开发语言。
	Language string
	// Version 是所使用的模板版本。
	Version string
	// Module 是 Go module 名。
	Module string
}
