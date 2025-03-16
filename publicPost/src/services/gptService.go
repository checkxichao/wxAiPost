package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"publicPost/src/models"
	"publicPost/src/repositories"
)

type GptService struct {
	gptRepo repositories.GptRepository // 添加 gptRepo 字段
}

func NewGptService(gptRepo repositories.GptRepository) *GptService {
	return &GptService{
		gptRepo: gptRepo,
	}
}

func (g *GptService) GetInfo() *models.GptModel {
	gptInfo, err := g.gptRepo.GetGPTInfo()
	if err != nil {
		return &models.GptModel{}
	}
	return gptInfo
}

func (g *GptService) SetInfo(info models.GptModel) error {
	err := g.gptRepo.SetGPTInfo(info)
	if err != nil {
		return err
	}
	return nil
}

func (g *GptService) CallProxyAPI(messages []map[string]interface{}, key string, model string) (string, error) {
	url := "https://xiaoai.plus/v1/chat/completions" // 确保 URL 正确

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", key),
	}

	data := ChatRequest{
		Model:               model,
		Messages:            make([]interface{}, len(messages)),
		MaxCompletionTokens: 1024, // 降低 max_tokens 以防 400 错误
		Temperature:         0.7,
		TopP:                1.0,
		Stream:              false,
	}

	// 转换 messages 格式
	for i, msg := range messages {
		data.Messages[i] = msg
	}

	reqBody, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("JSON 编码失败: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 处理 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API 请求失败: 状态码 %d, 响应: %s", resp.StatusCode, string(body))
	}

	var responseJson map[string]interface{}
	if err := json.Unmarshal(body, &responseJson); err != nil {
		return "", fmt.Errorf("JSON 解析失败: %v", err)
	}

	choices, ok := responseJson["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", errors.New("无效返回")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", errors.New("无效返回1")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", errors.New("无效返回2")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", errors.New("无效返回3")
	}

	return content, nil
}

type ChatRequest struct {
	Model               string        `json:"model"`
	Messages            []interface{} `json:"messages"`
	MaxCompletionTokens int           `json:"max_completion_tokens"`
	Temperature         float64       `json:"temperature,omitempty"`
	TopP                float64       `json:"top_p,omitempty"`
	Stream              bool          `json:"stream"`
}
