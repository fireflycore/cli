package config

var CONFIG_PROBLEM = []string{
	"Please select display text language",
}

var CREATE_PROBLEM = map[string][]string{
	"zh": {
		"请输入项目名称.",
		"请选择开发语言.",
	},
	"en": {
		"Please input your project name.",
		"Please choose your development language.",
	},
}

var TIPS_TEXT = map[string][]string{
	"zh": {
		"ctrl+c或q退出cli.",
		"回车确认或下一步.",
	},
	"en": {
		"ctrl+c or q to exit the cli.",
		"enter confirm or next step.",
	},
}
