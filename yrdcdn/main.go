package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Request struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
}

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Result  interface{} `json:"result,omitempty"`
}

type ActionParam struct {
	Key   string `json:"-"`
	Label string `json:"-"`
}

type actionJSON struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Params      orderedParams          `json:"params,omitempty"`
}

type orderedParams []ActionParam

func (op orderedParams) MarshalJSON() ([]byte, error) {
	if len(op) == 0 {
		return []byte("{}"), nil
	}
	var buf []byte
	buf = append(buf, '{')
	for i, p := range op {
		if i > 0 {
			buf = append(buf, ',')
		}
		key, _ := json.Marshal(p.Key)
		val, _ := json.Marshal(p.Label)
		buf = append(buf, key...)
		buf = append(buf, ':')
		buf = append(buf, val...)
	}
	buf = append(buf, '}')
	return buf, nil
}

type pluginMetaJSON struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Version     string         `json:"version"`
	Author      string         `json:"author"`
	Config      orderedParams  `json:"config"`
	Actions     []actionJSON   `json:"actions"`
}

var pluginMeta = pluginMetaJSON{
	Name:        "yrdcdn",
	Description: "融毅盾SSL证书部署插件",
	Version:     "1.0.0",
	Author:      "allinssl",
	Config: orderedParams{
		{Key: "username", Label: "登录邮箱/手机"},
		{Key: "password", Label: "密码"},
		{Key: "id", Label: "域名ID"},
	},
	Actions: []actionJSON{
		{
			Name:        "check",
			Description: "验证账号配置是否正确",
			Params: orderedParams{
				{Key: "username", Label: "登录邮箱/手机"},
				{Key: "password", Label: "密码"},
			},
		},
		{
			Name:        "deploy",
			Description: "部署SSL证书到融毅盾",
			Params: orderedParams{
				{Key: "id", Label: "域名ID"},
			},
		},
	},
}

func main() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		outputError("读取输入失败", err)
		return
	}

	var req Request
	if err := json.Unmarshal(input, &req); err != nil {
		outputError("解析请求失败", err)
		return
	}

	var resp *Response
	switch req.Action {
	case "get_metadata":
		resp = getMetadata()
	case "list_actions":
		resp = listActions()
	case "check":
		resp, err = check(req.Params)
		if err != nil {
			outputError("检查失败", err)
			return
		}
	case "deploy":
		resp, err = deploy(req.Params)
		if err != nil {
			outputError("部署失败", err)
			return
		}
	default:
		outputJSON(&Response{
			Status:  "error",
			Message: "未知 action: " + req.Action,
		})
		return
	}

	outputJSON(resp)
}

func getMetadata() *Response {
	return &Response{
		Status:  "success",
		Message: "插件信息",
		Result:  pluginMeta,
	}
}

func listActions() *Response {
	actionList := make([]actionJSON, 0, len(pluginMeta.Actions))
	for _, a := range pluginMeta.Actions {
		actionList = append(actionList, a)
	}
	return &Response{
		Status:  "success",
		Message: "获取动作列表成功",
		Result: map[string]interface{}{
			"actions": actionList,
		},
	}
}

func outputError(msg string, err error) {
	reason := ""
	if err != nil {
		reason = err.Error()
	}
	outputJSON(&Response{
		Status:  "error",
		Message: fmt.Sprintf("%s: %s", msg, reason),
	})
}

func outputJSON(resp *Response) {
	data, _ := json.Marshal(resp)
	fmt.Println(string(data))
}
