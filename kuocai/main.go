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
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Params      orderedParams `json:"params,omitempty"`
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
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Version     string        `json:"version"`
	Author      string        `json:"author"`
	Config      orderedParams `json:"config"`
	Actions     []actionJSON  `json:"actions"`
}

var pluginMeta = pluginMetaJSON{
	Name:        "kuocai",
	Description: "括彩CDN SSL证书部署插件，支持所有基于括彩CDN系统的平台",
	Version:     "1.0.0",
	Author:      "allinssl",
	Config: orderedParams{
		{Key: "username", Label: "登录邮箱/手机"},
		{Key: "password", Label: "密码"},
		{Key: "baseUrl", Label: "平台地址"},
	},
	Actions: []actionJSON{
		{
			Name:        "check",
			Description: "验证账号配置是否正确",
			Params: orderedParams{
				{Key: "username", Label: "登录邮箱/手机"},
				{Key: "password", Label: "密码"},
				{Key: "baseUrl", Label: "平台地址"},
			},
		},
		{
			Name:        "upload",
			Description: "部署SSL证书到括彩CDN平台",
			Params: orderedParams{
				{Key: "domainId", Label: "域名ID"},
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

	switch req.Action {
	case "get_metadata":
		outputJSON(&Response{Status: "success", Message: "插件信息", Result: pluginMeta})
	case "list_actions":
		outputJSON(&Response{Status: "success", Message: "支持的动作", Result: map[string]interface{}{"actions": pluginMeta.Actions}})
	case "check":
		resp, err := check(req.Params)
		if err != nil {
			outputError("检查失败", err)
			return
		}
		outputJSON(resp)
	case "upload":
		resp, err := Upload(req.Params)
		if err != nil {
			outputError("部署失败", err)
			return
		}
		outputJSON(resp)
	default:
		outputJSON(&Response{
			Status:  "error",
			Message: "未知 action: " + req.Action,
		})
		return
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
