package pkg

var LANGUAGE = []string{
	"Go",
	"Rust",
	"Node.js",
}

var DATABASE = map[string][]string{}

var REGISTER = []string{
	"Etcd",
}

func init() {
	DATABASE["Go"] = []string{
		"Mysql",
		"Mongo",
		"Redis",
	}
}
