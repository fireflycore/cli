package view

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fireflycore/cli/pkg/config"
	"regexp"
	"strings"
)

// inputReg 用于过滤项目名中不适合目录名的字符。
var inputReg = regexp.MustCompile("[^a-zA-Z0-9_]+")

// CreateFormEntity 表示 create 交互式表单的全部状态。
type CreateFormEntity struct {
	// problemIndex 表示当前正在回答第几个问题。
	problemIndex int
	// languageIndex 表示语言列表中当前高亮的选项。
	languageIndex int

	// input 是项目名输入框模型。
	input textinput.Model

	// Project 是用户最终确认的项目名。
	Project string
	// Language 是用户最终确认的开发语言。
	Language string
}

// Init 初始化 Bubble Tea 表单，目前没有额外启动命令。
func (model *CreateFormEntity) Init() tea.Cmd {
	return nil
}

// Update 根据键盘消息更新表单状态。
func (model *CreateFormEntity) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// cmd 用于保存子组件返回的后续命令。
	var cmd tea.Cmd

	// create 表单当前只处理键盘消息。
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// 根据按键执行退出、移动或确认逻辑。
		switch msg.String() {
		case "q", "ctrl+c":
			// q 和 ctrl+c 都直接退出表单。
			return model, tea.Quit
		case "up":
			// 只有语言选择问题支持向上移动。
			if model.problemIndex == 1 && model.languageIndex > 0 {
				model.languageIndex--
			}
		case "down":
			// 只有语言选择问题支持向下移动，且不能越过列表末尾。
			if model.problemIndex == 1 && (model.languageIndex < len(config.LANGUAGE)-1) {
				model.languageIndex++
			}
		case "enter":
			// enter 根据当前问题确认输入或选择。
			switch model.problemIndex {
			case 0:
				// 读取项目名输入并过滤非法字符。
				input := model.input.Value()
				model.Project = inputReg.ReplaceAllString(input, "")
				// 项目名确认后进入语言选择。
				model.problemIndex++
			case 1:
				// 语言统一转小写，保持和模板语言 key 一致。
				model.Language = strings.ToLower(config.LANGUAGE[model.languageIndex])
				// 语言确认后表单完成。
				model.problemIndex++
			}
			// 所有问题完成后退出 Bubble Tea 程序。
			if model.problemIndex+1 > len(CREATE_PROJECT_PROBLEM) {
				return model, tea.Quit
			}
		}
	}

	// 只有项目名问题需要把消息继续传给 textinput。
	if model.problemIndex == 0 {
		model.input, cmd = model.input.Update(msg)
	}

	// 返回更新后的模型和可能存在的子命令。
	return model, cmd
}

// View 根据当前问题状态渲染交互式表单。
func (model *CreateFormEntity) View() string {
	// strings.Builder 用于高效拼接终端输出。
	var str strings.Builder

	// 默认提示前缀使用主色。
	prefix := PrimaryColor.Render("<-")
	// 项目名尚未回答时渲染输入框。
	if model.problemIndex == 0 && len(model.Project) == 0 {
		str.WriteString(fmt.Sprintf("%s %s\n", prefix, InfoColor.Render(CREATE_PROJECT_PROBLEM[0])))
		str.WriteString(fmt.Sprintf("-> %s\n", model.input.View()))
	}

	// 语言尚未回答时渲染语言列表。
	if model.problemIndex == 1 && len(model.Language) == 0 {
		// 先回显已经确认的项目名。
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, InfoColor.Render(fmt.Sprintf("%s -", CREATE_PROJECT_PROBLEM[0])), PrimaryColor.Render(model.Project)))
		// 再渲染当前问题标题。
		str.WriteString(fmt.Sprintf("%s %s\n", prefix, InfoColor.Render(CREATE_PROJECT_PROBLEM[1])))

		// 遍历支持语言并渲染当前焦点。
		for ii, item := range config.LANGUAGE {
			// 非焦点项使用空白占位，保持列表对齐。
			selected := "  "
			// 焦点项使用箭头和高亮颜色。
			if model.languageIndex == ii {
				selected = FocusColor.Render("->")
				item = FocusColor.Render(item)
			}
			// 写入当前语言选项。
			str.WriteString(fmt.Sprintf("%s %s\n", selected, item))
		}
	}

	// 所有问题完成后回显最终答案。
	if model.problemIndex == len(CREATE_PROJECT_PROBLEM) {
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, InfoColor.Render(fmt.Sprintf("%s -", CREATE_PROJECT_PROBLEM[0])), PrimaryColor.Render(model.Project)))
		str.WriteString(fmt.Sprintf("%s %s %s\n", prefix, InfoColor.Render(fmt.Sprintf("%s -", CREATE_PROJECT_PROBLEM[1])), PrimaryColor.Render(model.Language)))
	} else {
		// 表单未完成时在底部显示操作提示。
		prefix = WarningColor.Render("->")
		// TIPS_TEXT 保存多行快捷键提示。
		tips := TIPS_TEXT
		// 逐行渲染提示文本。
		for index, item := range tips {
			str.WriteString(fmt.Sprintf("\n%s %s", prefix, item))
			// 最后一条提示后补换行，避免终端提示符贴在文本后。
			if index == len(tips)-1 {
				str.WriteString("\n")
			}
		}
	}

	// 返回完整终端视图。
	return str.String()
}

// NewCreate 启动 create 交互式表单并返回用户输入结果。
func NewCreate() (*CreateFormEntity, error) {
	// 创建项目名输入框。
	input := textinput.New()
	// 不展示 textinput 默认 prompt，外层已经渲染问题标题。
	input.Prompt = ""
	// 进入表单时默认聚焦项目名输入框。
	input.Focus()

	// 初始化表单状态。
	form := &CreateFormEntity{
		input: input,
	}

	// 启动 Bubble Tea 程序并等待用户完成表单。
	p := tea.NewProgram(form)
	if _, err := p.Run(); err != nil {
		return nil, err
	}

	// 用户中途退出或未完成必要字段时返回错误。
	if form.Project == "" || form.Language == "" {
		return nil, errors.New(DangerColor.Render("缺少必要参数"))
	}

	// 返回完整的创建表单结果。
	return form, nil
}
