package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
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

重要规则：
1. 省份名去掉"省/市/自治区/壮族/回族/维吾尔"等后缀，只保留主体名称（如"浙江"而不是"浙江省"）。
2. 城市名必须是地级市，去掉"市"等后缀（如"杭州"而不是"杭州市"）。
3. 区分县级市和地级市：当地址同时包含地级市和县级市时，输出地级市名称。
   例如"浙江省金华市义乌市" → city应为"金华"而不是"义乌"。
   例如"浙江省温州市瑞安市" → city应为"温州"而不是"瑞安"。
   例如"江苏省苏州市昆山市" → city应为"苏州"而不是"昆山"。
   县级市常见：义乌、瑞安、慈溪、乐清、诸暨、海宁、桐乡、温岭、临海、龙泉、江山、建德、富阳、临安、奉化、余姚、昆山、太仓、常熟、张家港、江阴、宜兴、溧阳、金坛、靖江、泰兴、兴化、如皋、海门、启东、东台、邳州、新沂、丹阳、扬中、句容等。
4. 如果地址中只有一个城市名且它是县级市，则输出该县级市（去掉"市"）。
5. 如果无法确定，province和city都设为空字符串。
6. 只输出JSON，不要任何额外文字或解释。

地址：%s`, address)

	reqBody := dsRequest{
		Model: "deepseek-chat",
		Messages: []dsMessage{
			{Role: "system", Content: "你是一个精确的中国地址解析助手。关键规则：城市必须是地级市，不能是县级市或区。只输出JSON。"},
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

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
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
