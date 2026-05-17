package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// DeepSeek API key from env
var deepseekKey = os.Getenv("DEEPSEEK_API_KEY")

const deepseekURL = "https://api.deepseek.com/chat/completions"

type dsMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type dsRequest struct {
	Model    string      `json:"model"`
	Messages []dsMessage `json:"messages"`
}

type dsChoice struct {
	Message dsMessage `json:"message"`
}

type dsResponse struct {
	Choices []dsChoice `json:"choices"`
}

type addrResult struct {
	Province string `json:"province"`
	City     string `json:"city"`
	Error    string `json:"error,omitempty"`
}

// AnalyzeAddress 调用 DeepSeek 解析地址，提取省份和城市
func AnalyzeAddress(address string) (province, city string, err error) {
	if deepseekKey == "" {
		return "", "", fmt.Errorf("未设置 DEEPSEEK_API_KEY 环境变量")
	}

	prompt := fmt.Sprintf(
		`从以下地址中提取省份和城市名称。以JSON格式返回：{"province":"省名","city":"城市名"}。
省份名去掉"省/市/自治区/壮族/回族/维吾尔"等后缀，只保留主体名称（如"浙江"而不是"浙江省"）。
城市名去掉"市/区/县/州/地区"等后缀（如"杭州"而不是"杭州市"）。
如果无法确定，province和city都设为空字符串。
地址：%s`, address)

	reqBody := dsRequest{
		Model: "deepseek-chat",
		Messages: []dsMessage{
			{Role: "system", Content: "你是一个精确的地址解析助手。只输出JSON，不要任何额外文字。"},
			{Role: "user", Content: prompt},
		},
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", deepseekURL, strings.NewReader(string(body)))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+deepseekKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("API请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("API返回 %d: %s", resp.StatusCode, string(respBody))
	}

	var dsResp dsResponse
	if err := json.Unmarshal(respBody, &dsResp); err != nil {
		return "", "", fmt.Errorf("解析响应失败: %w", err)
	}
	if len(dsResp.Choices) == 0 {
		return "", "", fmt.Errorf("API无返回内容")
	}

	content := dsResp.Choices[0].Message.Content
	// 去除可能的 markdown 代码块标记
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result addrResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return "", "", fmt.Errorf("解析地址结果失败: %w\n原始内容: %s", err, content)
	}

	return result.Province, result.City, nil
}
