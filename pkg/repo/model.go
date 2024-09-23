package repo

type GithubRepoVersion struct {
	TagName string `json:"tag_name"`
}

type ReadmeEntity struct {
	Project  string
	Language string
	Version  string
}
