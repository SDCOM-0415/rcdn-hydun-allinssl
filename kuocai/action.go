package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type checkResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func getBaseURL(params map[string]interface{}) (string, error) {
	baseUrl, _ := params["baseUrl"].(string)
	baseUrl = strings.TrimSpace(baseUrl)
	if baseUrl == "" {
		return "", errors.New("平台地址不能为空")
	}
	baseUrl = strings.TrimRight(baseUrl, "/")
	return baseUrl, nil
}

func Upload(params map[string]interface{}) (*Response, error) {
	certStr, _ := params["cert"].(string)
	keyStr, _ := params["key"].(string)
	username, _ := params["username"].(string)
	password, _ := params["password"].(string)
	domainId, _ := params["domainId"].(string)

	if username == "" || password == "" {
		return nil, errors.New("登录邮箱/手机或密码不能为空")
	}

	if domainId == "" {
		return nil, errors.New("域名ID不能为空")
	}

	if certStr == "" || keyStr == "" {
		return nil, errors.New("证书或私钥不能为空")
	}

	token, err := doLoginRequest(params, username, password)
	if err != nil {
		return nil, err
	}

	tokenStr, _ := token.(string)
	if tokenStr == "" {
		return nil, errors.New("获取token失败")
	}

	cookies := fmt.Sprintf("kuocai_cdn_token=%s", tokenStr)

	res, err := doRequest(params, "/CdnDomainHttps/httpsConfiguration", map[string]interface{}{
		"doMainId": domainId,
		"https": map[string]interface{}{
			"certificate_name":   generateUniqID(),
			"certificate_source": "0",
			"certificate_value":  certStr,
			"https_status":       "on",
			"private_key":        keyStr,
		},
	}, &cookies)

	if err != nil {
		return nil, err
	}

	return &Response{
		Status:  "success",
		Message: fmt.Sprintf("域名ID:%s 更新成功", domainId),
		Result:  res.(map[string]interface{}),
	}, nil
}

func doLoginRequest(params map[string]interface{}, username, password string) (interface{}, error) {
	baseURL, err := getBaseURL(params)
	if err != nil {
		return nil, err
	}
	formData := url.Values{}
	formData.Set("userAccount", username)
	formData.Set("userPwd", password)
	formData.Set("remember", "true")

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Post(baseURL+"/login/loginUser", "application/x-www-form-urlencoded", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var result checkResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if result.Code == "SUCCESS" {
		return result.Data, nil
	} else if result.Message != "" {
		return nil, errors.New(result.Message)
	}
	return nil, fmt.Errorf("请求失败(httpCode=%d)", resp.StatusCode)
}

func doRequest(params map[string]interface{}, path string, bodyParams map[string]interface{}, cookies *string) (interface{}, error) {
	baseURL, err := getBaseURL(params)
	if err != nil {
		return nil, err
	}
	requestURL := baseURL + path

	var body []byte

	body, err = json.Marshal(bodyParams)
	if err != nil {
		return nil, fmt.Errorf("编码参数失败: %v", err)
	}

	req, err := http.NewRequest("POST", requestURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if cookies != nil && *cookies != "" {
		req.Header.Set("Cookie", *cookies)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var result checkResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if result.Code == "SUCCESS" {
		return result.Data, nil
	} else if result.Message != "" {
		return nil, errors.New(result.Message)
	} else {
		return nil, fmt.Errorf("请求失败(httpCode=%d)", resp.StatusCode)
	}
}

func generateUniqID() string {
	return fmt.Sprintf("cert_%d", time.Now().UnixNano())
}
