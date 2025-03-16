package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"publicPost/src/models"
	"publicPost/src/repositories"
	"strconv"
	"strings"
	"time"
)

type SendService struct {
	userRepo   repositories.UserRepository   // 用户仓库
	wechatRepo repositories.WechatRepository // 微信仓库
	gptService *GptService                   // GPT 服务实例
}

func NewSendService(userRepo repositories.UserRepository, wechatRepo repositories.WechatRepository, gptService *GptService) *SendService {
	return &SendService{
		userRepo:   userRepo,
		wechatRepo: wechatRepo,
		gptService: gptService,
	}
}

func (se *SendService) GetAccessToken(wxid string, secret string) string {
	wx, err := se.wechatRepo.GetWxInfoByWxid(wxid)
	if err != nil {
		fmt.Printf("微信信息获取失败: %v\n", err)
		return err.Error()
	}

	accessToken := wx.AccessToken
	fmt.Printf("当前accToken: %s, 有效期至: %v, 现在是: %v\n", accessToken, wx.EndTime, time.Now())

	timePK := time.Now().After(wx.EndTime)

	if accessToken == "" || timePK {
		fmt.Println("此时accToken过期,重新申请一个")

		url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", wxid, secret)

		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("HTTP 请求出错: %v\n", err)
			return ""
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("读取响应出错: %v\n", err)
			return ""
		}

		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			fmt.Printf("反序列化json失败: %v\n", err)
			return ""
		}
		fmt.Printf("微信API响应: %v\n", result)

		newAccessToken, ok := result["access_token"].(string)

		if ok {
			err := se.wechatRepo.SetAccessToken(wxid, newAccessToken)
			if err != nil {
				fmt.Printf("更新token失败: %v\n", err)
				return ""
			}
			fmt.Println("成功更新token")
			return newAccessToken
		} else {
			fmt.Printf("更新token时出错: %v\n", result)
			return ""
		}
	}

	return accessToken
}

func (se *SendService) GetWxNowState(wxid string) int {
	state, err := se.wechatRepo.GetWxNowState(wxid)
	if err != nil {
		return -1
	}
	return state
}
func (se *SendService) SetWxStop(wxid string) int {
	err := se.wechatRepo.SetWxNowState(wxid, 0)
	if err != nil {
		return -1
	}
	return 0
}
func (se *SendService) SetWxPost(wxid string) int {
	err := se.wechatRepo.SetWxNowState(wxid, 2)
	if err != nil {
		return -1
	}
	return 2
}
func (se *SendService) SetWxDart(wxid string) int {
	err := se.wechatRepo.SetWxNowState(wxid, 1)
	if err != nil {
		return -1
	}
	return 1
}

func (se *SendService) GetAllMediaIdByWxid(wxid string) ([]*models.WechatMediaModel, error) {

	info, err := se.wechatRepo.GetAllInfoByWxid(wxid)
	if err != nil {
		return nil, err
	}
	return info, nil
}
func (se *SendService) GetMediaListByWxidP(wxid string, query string) ([]*models.WechatMediaModel, int, error) {
	return se.wechatRepo.GetMediaListByWxid(wxid, query)
}
func (se *SendService) PostDart(wxid string, accessToken string, mediaId string) error {

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/freepublish/submit?access_token=%s", accessToken)
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	data := map[string]string{
		"media_id": mediaId,
	}

	state, err := se.wechatRepo.GetWxNowState(wxid)

	if err != nil {
		return err
	}
	if state != 2 {
		return errors.New("已停止")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		_ = se.wechatRepo.SetWxNowState(wxid, 0)
		return fmt.Errorf("序列化json失败: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		_ = se.wechatRepo.SetWxNowState(wxid, 0)
		return fmt.Errorf("创建请求失败: %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		_ = se.wechatRepo.SetWxNowState(wxid, 0)
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		_ = se.wechatRepo.SetWxNowState(wxid, 0)
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if resp.StatusCode == 200 {
		if errCode, ok := result["errcode"].(float64); ok && errCode == 0 {
			publishID, _ := result["publish_id"].(string)
			log.Printf("草稿 %s 发布成功，发布任务ID: %s", mediaId, publishID)

			err := se.wechatRepo.SetMediaState(mediaId, 1)
			if err != nil {
				log.Printf("更新媒体 ID 状态失败: %v", err)
			}
		} else {
			_ = se.wechatRepo.SetWxNowState(wxid, 0)
			_ = se.wechatRepo.SetMediaState(mediaId, 1)
			log.Printf("%s上传失败！错误信息: %v", mediaId, result)
		}
	} else {
		_ = se.wechatRepo.SetWxNowState(wxid, 0)
		log.Printf("上传失败！错误码: %d", resp.StatusCode)
	}
	delay := rand.Intn(121) + 60
	log.Printf("等待 %d 秒后继续...", delay)
	time.Sleep(time.Duration(delay) * time.Second)

	log.Printf("%s任务已完成", mediaId)
	err = se.wechatRepo.SetWxNowState(wxid, 0)
	if err != nil {
		return err
	}

	return nil
}

func (se *SendService) UploadToWeChat(filePath string, acctoken string, wxid string, note string) (string, string, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/material/add_material?access_token=%s&type=image", acctoken)

	file, err := os.Open(filePath)
	if err != nil {
		return "", "", fmt.Errorf("无法打开文件: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("media", filepath.Base(filePath))
	if err != nil {
		return "", "", fmt.Errorf("无法创建表单文件: %v", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", "", fmt.Errorf("无法写入文件数据: %v", err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", "", fmt.Errorf("无法创建请求: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("请求微信接口失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("无法读取响应体: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", fmt.Errorf("反序列化json失败: %v", err)
	}

	if errCode, exists := result["errcode"]; exists && errCode != float64(0) {
		errmsg, _ := result["errmsg"].(string)
		return "", "", fmt.Errorf("微信API错误: %v", errmsg)
	}

	mediaID, ok := result["media_id"].(string)
	if !ok {
		return "", "", fmt.Errorf("缺少 media_id 或类型错误: %v", result)
	}

	wxURL, ok := result["url"].(string)
	if !ok {
		wxURL = ""
	}
	if wxURL == "" || mediaID == "" {
		return "", "", errors.New("图片url或media_id缺失")
	}
	err = se.wechatRepo.SetThumbMediaId(wxid, mediaID, note, wxURL)
	if err != nil {
		return "", "", err
	}
	return wxURL, mediaID, nil
}
func (se *SendService) TitleContain(title, wxid string) error {
	containTitles, err := se.wechatRepo.GetTitlesByWxid(wxid)
	if err != nil {
		_ = se.wechatRepo.SetWxNowState(wxid, 0)
		return err
	}
	for _, containTitle := range containTitles {
		if containTitle.Title == title {
			return fmt.Errorf("标题重复: %s", title)
		}
	}
	return nil
}
func (se *SendService) WrittenDraft(wxid string, accessToken string, thumbMediaId string, template string, selfTtile string, templateMethod string) error {
	url := "https://api.weixin.qq.com/cgi-bin/draft/add?access_token=" + accessToken
	wx, err := se.wechatRepo.GetWxInfoByWxid(wxid)
	if err != nil {
		_ = se.wechatRepo.SetWxNowState(wxid, 0)
		return err
	}

	gptInfo := se.gptService.GetInfo()
	if gptInfo == nil {
		_ = se.wechatRepo.SetWxNowState(wxid, 0)
		return fmt.Errorf("未能获取有效的GPT信息")
	}

	containTitles, err := se.wechatRepo.GetTitlesByWxid(wxid)

	if err != nil {
		_ = se.wechatRepo.SetWxNowState(wxid, 0)
		return err
	}
	for _, containTitle := range containTitles {
		if containTitle.Title == selfTtile {
			return fmt.Errorf("标题重复: %s", selfTtile)
		}
	}

	var messages []map[string]interface{}

	if templateMethod == "intro" { //介绍

		messages = append(messages, map[string]interface{}{
			"role":    "user",
			"content": "请以这个游戏：" + selfTtile + "为主题虚构一个游戏的介绍、玩法、特色，介绍一下游戏的玩法和内容,最后进行总结",
		})
	} else { //排行榜
		rands := strconv.Itoa(rand.Intn(9) + 3)
		messages = append(messages, map[string]interface{}{
			"role":    "user",
			"content": "请虚构一个游戏排行榜,介绍一下传奇游戏的玩法和内容,然后随机介绍" + rands + "款游戏,最后进行总结",
		})
	}

	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	filePath := fmt.Sprintf("%s/%s", currentDir, template)

	content, err := ioutil.ReadFile(filePath)

	contentMessage, err := se.gptService.CallProxyAPI(messages, gptInfo.Key, gptInfo.Model)
	contentMessage = addPTagsToText(contentMessage)

	finalContent := string(content) + contentMessage
	if err != nil {
		_ = se.wechatRepo.SetWxNowState(wxid, 0)
		return err
	}
	if len(wx.Name) > 8 {
		wx.Name = ""
	}
	article := models.Article{
		Title:        selfTtile,
		Author:       wx.Name,
		Content:      finalContent,
		ThumbMediaId: thumbMediaId,
	}

	draftRequest := models.DraftRequest{
		Articles: []models.Article{article},
	}
	jsonData, err := json.Marshal(draftRequest)
	if err != nil {
		_ = se.wechatRepo.SetWxNowState(wxid, 0)
		return fmt.Errorf("序列化json失败: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		_ = se.wechatRepo.SetWxNowState(wxid, 0)
		return fmt.Errorf("HTTP request error: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		_ = se.wechatRepo.SetWxNowState(wxid, 0)
		return fmt.Errorf("响应解析失败: %v", err)
	}

	if mediaID, exists := result["media_id"]; exists {

		if mediaIDStr, ok := mediaID.(string); ok {

			err = se.wechatRepo.SetMediaId(wxid, mediaIDStr, selfTtile)
			if err != nil {
				_ = se.wechatRepo.SetWxNowState(wxid, 0)
				return fmt.Errorf("写入数据库失败: %v", err)
			}
			err = se.wechatRepo.DeleteTitleByTitle(selfTtile)
			if err != nil {
				_ = se.wechatRepo.SetWxNowState(wxid, 0)
				return fmt.Errorf("删除标题失败: %v", err)
			}

		} else {
			_ = se.wechatRepo.SetWxNowState(wxid, 0)
			return fmt.Errorf("media_id 不是字符串类型: %v", mediaID)
		}
	} else {
		_ = se.wechatRepo.SetWxNowState(wxid, 0)
		return fmt.Errorf("上传失败！错误信息: %v", result)
	}

	delaySeconds := rand.Intn(90) + 50

	time.Sleep(time.Duration(delaySeconds) * time.Second)

	_ = se.wechatRepo.SetWxNowState(wxid, 0)

	return nil
}
func (se *SendService) GetThumbMediaIdList(wxid string) []*models.ThumbMediaModel {

	result, err := se.wechatRepo.GetThumbMediaIdList(wxid)
	if err != nil {
		return make([]*models.ThumbMediaModel, 0)
	}
	return result
}
func addPTagsToText(text string) string {

	paragraphs := strings.Split(text, "\n")

	var wrappedParagraphs []string
	for _, para := range paragraphs {

		if len(strings.TrimSpace(para)) > 0 {
			wrappedParagraphs = append(wrappedParagraphs, fmt.Sprintf("<p>%s</p></br>", strings.TrimSpace(para)))
		}
	}

	return strings.Join(wrappedParagraphs, "\n")
}
func (se *SendService) SetTitle(title string) error {
	return se.wechatRepo.SetTitle(title)
}
func (se *SendService) EditTitle(id int, title string) error {
	return se.wechatRepo.EditTitle(id, title)
}
func (se *SendService) DeleteTitle(id int, title string) error {
	return se.wechatRepo.DeleteTitle(id, title)
}

func (se *SendService) SetTitlesBatch(titles []string) error {
	return se.wechatRepo.SetTitlesBatch(titles)
}
func (se *SendService) DeleteTitleBatch(ids []int) error {
	return se.wechatRepo.DeleteTitleBatch(ids)
}
func (se *SendService) GetTitleList(page int, pageSize int) ([]*models.TitleListModel, int64, error) {
	return se.wechatRepo.GetTitleList(page, pageSize)
}
func (se *SendService) GetTitleListSearch(page int, pageSize int, search string) ([]*models.TitleListModel, int64, error) {
	return se.wechatRepo.GetTitleListSearch(page, pageSize, search)
}
func (se *SendService) GetRandomTitles(count int) []string {
	titles, err := se.wechatRepo.GetRandomTitles(count)
	if err != nil {
		fmt.Printf("服务层获取随机标题失败: %v\n", err)
		return nil
	}
	return titles
}

func (se *SendService) GetSequentialTitles(count int) []string {
	titles, err := se.wechatRepo.GetSequentialTitles(count)
	if err != nil {
		fmt.Printf("服务层获取顺序标题失败: %v\n", err)
		return nil
	}
	return titles
}
func (se *SendService) GetPreviewLink(accessToken, mediaId string) (string, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/draft/get?access_token=%s", accessToken)
	payload := struct {
		MediaID string `json:"media_id"`
	}{
		MediaID: mediaId,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("序列化json失败: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送 HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应体失败: %w", err)
	}

	var result struct {
		ErrCode   int    `json:"errcode"`
		ErrMsg    string `json:"errmsg"`
		NewsItems []struct {
			Title        string `json:"title"`
			Author       string `json:"author,omitempty"`
			Digest       string `json:"digest,omitempty"`
			Content      string `json:"content"`
			ThumbMediaID string `json:"thumb_media_id"`
			URL          string `json:"url"`
		} `json:"news_item"`
	}

	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		return "", fmt.Errorf("反序列化JSON 失败: %w", err)
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("API 错误: %s, 错误码: %d", result.ErrMsg, result.ErrCode)
	}

	if len(result.NewsItems) == 0 {
		return "", fmt.Errorf("响应中未找到news_item")
	}

	var previewURLs []string
	for _, item := range result.NewsItems {
		if item.URL != "" {
			previewURLs = append(previewURLs, item.URL)
		}
	}

	previewLink := strings.Join(previewURLs, "\n")

	return previewLink, nil
}
func (se *SendService) DeleteMediaId(accessToken, mediaId string) error {
	url := "https://api.weixin.qq.com/cgi-bin/draft/delete?access_token=" + accessToken
	body := map[string]string{
		"media_id": mediaId,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化json失败: %v", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if errcode, ok := response["errcode"].(float64); ok && errcode != 0 {
		fmt.Println("Error at", time.Now())
		return fmt.Errorf("API errcode: %v", errcode)
	}

	err = se.wechatRepo.DeleteMediaId(mediaId)

	return err
}
func (se *SendService) DeleteMediaIdToDataBase(mediaId string) error {
	return se.wechatRepo.DeleteMediaId(mediaId)
}
func (se *SendService) DeleteThumbId(accessToken, thumbId string) error {

	url := "https://api.weixin.qq.com/cgi-bin/material/del_material?access_token=" + accessToken

	body := map[string]string{
		"media_id": thumbId,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化json失败 %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	var response struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if response.ErrCode != 0 {
		return fmt.Errorf("APIerrcode: %v, errmsg: %v", response.ErrCode, response.ErrMsg)
	}

	err = se.wechatRepo.DeleteThumbId(thumbId)
	if err != nil {
		return fmt.Errorf("删除失败: %v", err)
	}

	return nil
}
func (se *SendService) SetTitleStateUse(titles []string) (error, int) {
	if len(titles) == 0 {
		return fmt.Errorf("没传入任何标题"), -1
	}
	count := 0
	for _, title := range titles {
		err := se.wechatRepo.SetTitleState(title, true)
		if err != nil {
			count++
		}
	}

	if count > 0 {
		return fmt.Errorf("共有%d个标题出现了异常", count), count
	}
	return nil, 0
}
func (se *SendService) SetTitleStateFalse(titles string) error {
	return se.wechatRepo.SetTitleState(titles, false)

}
func (se *SendService) GetTitleStateUse(title string) *models.TitleListModel {
	return se.wechatRepo.GetTitleState(title)
}
