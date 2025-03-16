package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"publicPost/src/models"
	"publicPost/src/response"
	"publicPost/src/services"
	"strconv"
	"strings"
	"time"
)

type SendController struct {
	SendService *services.SendService
	Scheduler   *gocron.Scheduler
	DB          *gorm.DB
}

func NewSendController(sendService *services.SendService, db *gorm.DB) *SendController {
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.StartAsync()

	return &SendController{
		Scheduler:   scheduler,
		SendService: sendService,
		DB:          db,
	}
}

func (c *SendController) GetMediaListByWxid(ctx *gin.Context) {
	var req struct {
		Wxid string `json:"wxid"`
	}
	if err := ctx.ShouldBind(&req); err != nil {
		response.Error(ctx, "请求参数错误")
		return
	}
	list, err := c.SendService.GetAllMediaIdByWxid(req.Wxid)

	if err != nil {
		response.Error(ctx, "请求参数错误")
		return
	}
	response.Success(ctx, "获取MediaId成功", list)

}
func (c *SendController) GetMediaListByWxidP(ctx *gin.Context) {
	var req struct {
		Wxid string `json:"wxid" binding:"required"`

		Q string `json:"q"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "绑定参数失败")
		return
	}

	mediaList, total, err := c.SendService.GetMediaListByWxidP(req.Wxid, req.Q)
	if err != nil {
		response.Error(ctx, "获取MediaId失败")

		return
	}
	response.Success(ctx, "获取MediaId成功", gin.H{
		"code":    200,
		"message": "获取MediaId成功",
		"data": gin.H{
			"data":  mediaList,
			"total": total,
		},
	})

}
func (c *SendController) SetStateStop(ctx *gin.Context) {
	var req struct {
		Wxid string `json:"wxid"`
	}
	if err := ctx.ShouldBind(&req); err != nil {
		response.Error(ctx, "请求参数错误"+req.Wxid)
		return
	}
	state := c.SendService.GetWxNowState(req.Wxid)
	if state == 0 {
		response.Error(ctx, "账号无任何任务"+req.Wxid)
		return
	}
	stop := c.SendService.SetWxStop(req.Wxid)
	if stop != 0 {
		response.Error(ctx, "停止失败"+req.Wxid)
		return
	}
	response.Success(ctx, "已停止", req.Wxid)
}
func (c *SendController) SetStatePost(ctx *gin.Context) {
	var req struct {
		Wxid     string   `json:"wxid"`
		Secret   string   `json:"secret"`
		MediaIds []string `json:"mediaIds" binding:"required,min=1"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		response.Error(ctx, "请求参数错误"+req.Wxid)
		return
	}

	state := c.SendService.GetWxNowState(req.Wxid)
	if state != 0 {
		response.Error(ctx, "账号已经在进行其他任务"+req.Wxid)
		return
	}
	stop := c.SendService.SetWxPost(req.Wxid)
	if stop != 2 {
		response.Error(ctx, "发布失败"+req.Wxid)
		return
	}

	accessToken := c.SendService.GetAccessToken(req.Wxid, req.Secret)
	fmt.Println(req.Wxid, ":", accessToken)
	go func(wxid, accessToken string, mediaIds []string) {
		for _, mediaId := range mediaIds {
			err := c.SendService.PostDart(wxid, accessToken, mediaId)
			if err != nil {

				log.Printf(req.Wxid+"发布 Media ID %s 失败: %v", mediaId, err)
				continue
			}

		}

	}(req.Wxid, accessToken, req.MediaIds)

	response.Success(ctx, "开始发布", req.Wxid)

}

func (c *SendController) SetStateDart(ctx *gin.Context) {
	var req struct {
		Wxid                    string   `json:"wxid"`
		Secret                  string   `json:"secret"`
		MaterialId              string   `json:"materialId"`
		DraftCount              int      `json:"draftCount"`
		Template                string   `json:"template"`
		Titles                  []string `json:"titles"`
		TitleSelectionMethod    string   `json:"titleSelectionMethod" binding:"required"`
		TemplateSelectionMethod string   `json:"templateSelectionMethod" binding:"required"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		response.Error(ctx, "请求参数错误"+req.Wxid)
		return
	}

	state := c.SendService.GetWxNowState(req.Wxid)
	if state != 0 && state != 1 {
		response.Error(ctx, "账号已经在进行其他任务"+req.Wxid)
		return
	}

	stop := c.SendService.SetWxDart(req.Wxid)
	if stop != 1 {
		response.Error(ctx, "草稿失败"+req.Wxid)
		return
	}

	accessToken := c.SendService.GetAccessToken(req.Wxid, req.Secret)
	if accessToken == "" {
		response.Error(ctx, "获取accessToken失败,请检查IP白名单或账号"+req.Wxid)
		return
	}

	var titles []string
	switch req.TitleSelectionMethod {
	case "manual":

		if len(req.Titles) < 1 {
			response.Error(ctx, "无有效任务"+req.Wxid)
			return
		}
		titles = req.Titles

	case "random":

		titles = c.SendService.GetRandomTitles(req.DraftCount)

	case "sequential":

		titles = c.SendService.GetSequentialTitles(req.DraftCount)

	default:
		response.Error(ctx, "无效的标题选择方法")
		return
	}

	if len(titles) < 1 {
		response.Error(ctx, "未找到有效标题"+req.Wxid)
		return
	}

	for i := 0; i < len(titles); i++ {
		title := titles[i]
		err := c.SendService.TitleContain(title, req.Wxid)
		use := c.SendService.GetTitleStateUse(title)
		if use == nil {
			continue
		}
		if err != nil || use.State {

			titles = append(titles[:i], titles[i+1:]...)
			i--
		}
	}

	response.Success(ctx, "开始草稿,共"+strconv.Itoa(len(titles))+"个标题可草稿", titles)

	go func() {
		_, _ = c.SendService.SetTitleStateUse(titles)

		for _, title := range titles {

			err := c.SendService.WrittenDraft(req.Wxid, accessToken, req.MaterialId, req.Template, title, req.TemplateSelectionMethod)

			if err != nil {

				_ = c.SendService.SetTitleStateFalse(title)

				if strings.Contains("access_token expired", err.Error()) {
					accessToken = c.SendService.GetAccessToken(req.Wxid, req.Secret)
					err = c.SendService.WrittenDraft(req.Wxid, accessToken, req.MaterialId, req.Template, title, req.TemplateSelectionMethod)
				}

				delaySeconds := rand.Intn(90) + 50

				delay := time.Duration(delaySeconds) * time.Second

				time.Sleep(delay)
			} else {

			}
		}
	}()
}

func (c *SendController) GetThumbMediaIds(ctx *gin.Context) {
	var req struct {
		Wxid string `json:"wxid"`
	}
	if err := ctx.ShouldBind(&req); err != nil {
		response.Error(ctx, err.Error())
		return
	}
	list := c.SendService.GetThumbMediaIdList(req.Wxid)
	if len(list) == 0 {
		response.Error(ctx, "没有素材"+req.Wxid)
		return
	}
	response.Success(ctx, "", gin.H{"list": list})

}

func (c *SendController) UploadImageGetThumbId(ctx *gin.Context) {

	var req struct {
		Wxid   string `form:"wxid" binding:"required"`
		Secret string `form:"secret" binding:"required"`
		Note   string `form:"note"`
	}

	if err := ctx.ShouldBind(&req); err != nil {

		response.Error(ctx, req.Wxid+"请求参数错误: "+err.Error())
		return
	}

	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, 10<<20)

	file, header, err := ctx.Request.FormFile("image")
	if err != nil {

		response.Error(ctx, req.Wxid+"文件上传失败: "+err.Error())
		return
	}
	defer file.Close()

	log.Printf("开始处理文件: %s\n", header.Filename)

	contentType, err := getFileContentType(file)
	if err != nil {

		response.Error(ctx, req.Wxid+"无法读取文件类型: "+err.Error())
		return
	}

	allowedTypes := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
	}

	ext, ok := allowedTypes[contentType]
	if !ok {

		response.Error(ctx, "不支持的文件类型,请上传jpg、png格式")
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		response.Error(ctx, "指针出错: "+err.Error())
		return
	}

	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		response.Error(ctx, fmt.Sprintf("无法创建上传目录: %v", err))
		return
	}

	tempFile, err := os.CreateTemp(uploadDir, "upload-*"+ext)
	if err != nil {

		response.Error(ctx, fmt.Sprintf("无法创建临时文件: %v", err))
		return
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())

	}()

	if _, err = io.Copy(tempFile, file); err != nil {

		response.Error(ctx, "无法保存文件: "+err.Error())
		return
	}

	accessToken := c.SendService.GetAccessToken(req.Wxid, req.Secret)

	if accessToken == "" {

		response.Error(ctx, req.Wxid+"无法获取微信账号信息")
		return
	}

	wxURL, mediaID, err := c.SendService.UploadToWeChat(tempFile.Name(), accessToken, req.Wxid, req.Note)
	if err != nil {

		response.Error(ctx, fmt.Sprintf(req.Wxid+"上传到微信失败: %v", err))
		return
	}

	response.Success(ctx, "上传成功", gin.H{
		"message":  "上传成功",
		"url":      wxURL,
		"media_id": mediaID,
		"filename": header.Filename,
	})

}
func getFileContentType(file multipart.File) (string, error) {
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}
	return http.DetectContentType(buffer[:n]), nil
}

func (c *SendController) GetTemplateList(ctx *gin.Context) {

	currentDir, err := os.Getwd()

	fmt.Println("Current directory:", currentDir)
	if err != nil {
		response.Error(ctx, "获取当前目录失败")

		return
	}

	templateFiles, err := ioutil.ReadDir(currentDir)
	if err != nil {
		response.Error(ctx, "获取当前目录失败")

		return
	}

	var templates []string
	for _, file := range templateFiles {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") {
			templates = append(templates, file.Name())
		}
	}
	for _, file := range templateFiles {
		fmt.Println("File name:", file.Name())
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") {
			templates = append(templates, file.Name())
		}
	}

	response.Success(ctx, "", gin.H{
		"templates": templates,
	})

}

func (c *SendController) GetTemplate(ctx *gin.Context) {
	templateId := ctx.Param("templateId")

	currentDir, err := os.Getwd()
	fmt.Println("Current directory:", currentDir)

	if err != nil {
		response.Error(ctx, "获取当前目录失败")

		return
	}

	templateFilePath := fmt.Sprintf("%s/%s", currentDir, templateId)

	content, err := ioutil.ReadFile(templateFilePath)
	if err != nil {
		response.Error(ctx, "模板错误")

		return
	}

	response.Success(ctx, "", gin.H{
		"content": string(content),
	})

}

func (c *SendController) SetTemplate(ctx *gin.Context) {
	var request struct {
		TemplateId string `json:"templateId"`
		Content    string `json:"content"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, "请求参数错误")
		return
	}

	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	templateDir := fmt.Sprintf("%s", currentDir)
	filePath := fmt.Sprintf("%s/%s", templateDir, request.TemplateId)

	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		err := os.MkdirAll(templateDir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = ioutil.WriteFile(filePath, []byte(request.Content), 0644)
	if err != nil {
		log.Fatal(err)
	}

	response.Success(ctx, "模板保存成功", nil)

}

func (c *SendController) AddTemplate(ctx *gin.Context) {
	var request struct {
		TemplateId string `json:"templateId"`
		Content    string `json:"content"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.Error(ctx, "请求参数错误")
		return
	}

	if !strings.HasSuffix(request.TemplateId, ".txt") {
		request.TemplateId += ".txt"
	}

	currentDir, err := os.Getwd()
	if err != nil {
		response.Error(ctx, "获取目录失败")
		return
	}

	templateFilePath := fmt.Sprintf("%s/%s", currentDir, request.TemplateId)

	err = ioutil.WriteFile(templateFilePath, []byte(request.Content), 0644)
	if err != nil {
		response.Error(ctx, "模板保存出错")
		return
	}
	response.Success(ctx, "模板保存成功", nil)

}
func (c *SendController) GetTitleList(ctx *gin.Context) {
	var req struct {
		Page     int `form:"page" binding:"gte=1" json:"page"`
		PageSize int `form:"pageSize" binding:"gte=1,lte=100" json:"pageSize"`
	}

	if err := ctx.ShouldBindQuery(&req); err != nil {
		var ve validator.ValidationErrors
		if ok := errors.As(err, &ve); ok {
			out := make([]string, len(ve))
			for i, fe := range ve {
				out[i] = fmt.Sprintf("参数 '%s' 无效", fe.Field())
			}
			response.Error(ctx, strings.Join(out, ", "))
			return
		}

		response.Error(ctx, "参数绑定失败")
		return
	}

	list, total, err := c.SendService.GetTitleList(req.Page, req.PageSize)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}

	response.Success(ctx, "获取成功", gin.H{
		"data":  list,
		"total": total,
	})
}
func (c *SendController) GetTitleListSearch(ctx *gin.Context) {
	var req struct {
		Page     int    `form:"page" binding:"gte=1" json:"page"`
		PageSize int    `form:"pageSize" binding:"gte=1,lte=10" json:"pageSize"`
		Search   string `form:"search" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "请求参数错误")
		return
	}

	list, total, err := c.SendService.GetTitleListSearch(req.Page, req.PageSize, req.Search)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}

	response.Success(ctx, "获取成功", gin.H{
		"data":  list,
		"total": total,
	})
}
func (c *SendController) SetTitle(ctx *gin.Context) {
	var req struct {
		Title string
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "绑定参数错误")
		return
	}

	err := c.SendService.SetTitle(req.Title)

	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, "标题设置完成", req.Title)

}
func (c *SendController) EditTitle(ctx *gin.Context) {
	var req struct {
		Id    int `json:"id"`
		Title string
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "绑定参数错误")
		return
	}

	err := c.SendService.EditTitle(req.Id, req.Title)

	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, "标题设置完成", req.Title)

}
func (c *SendController) DeleteTitle(ctx *gin.Context) {
	var req struct {
		Id    int    `json:"id"`
		Title string `json:"title"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, err.Error())
		return
	}
	err := c.SendService.DeleteTitle(req.Id, req.Title)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, "此标题："+req.Title+" 已删除", nil)

}
func (c *SendController) SetTitleByTxt(ctx *gin.Context) {
	var req struct {
		Titles []string `json:"titles" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "绑定参数错误")
		return
	}

	if err := c.SendService.SetTitlesBatch(req.Titles); err != nil {
		response.Error(ctx, err.Error())
		return
	}

	response.Success(ctx, "批量标题设置完成", nil)
}
func (c *SendController) DeleteTitleBatch(ctx *gin.Context) {
	var req struct {
		Ids []int `json:"ids" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "请求参数错误: "+err.Error())
		return
	}
	if err := c.SendService.DeleteTitleBatch(req.Ids); err != nil {
		response.Error(ctx, "批量删除失败: "+err.Error())
		return
	}
	response.Success(ctx, "批量删除成功", req.Ids)
}
func (c *SendController) GetPreviewLink(ctx *gin.Context) {
	var req struct {
		Wxid    string `json:"wxid" binding:"required"`
		Secret  string `json:"secret" binding:"required"`
		MediaId string `json:"mediaId" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, req.Wxid+"绑定参数错误")
		return
	}

	accessToken := c.SendService.GetAccessToken(req.Wxid, req.Secret)

	if accessToken == "" {
		response.Error(ctx, req.Wxid+"获取到的accessToken为空")
		return
	}
	link, err := c.SendService.GetPreviewLink(accessToken, req.MediaId)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, req.Wxid+"获取预览链接成功", link)
}
func (c *SendController) DeleteMedia(ctx *gin.Context) {
	var req struct {
		Wxid    string `json:"wxid" binding:"required"`
		Secret  string `json:"secret" binding:"required"`
		MediaId string `json:"mediaId" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "绑定参数错误")
		return
	}

	accessToken := c.SendService.GetAccessToken(req.Wxid, req.Secret)
	if accessToken == "" {
		response.Error(ctx, req.Wxid+"获取到的accessToken为空")
		return
	}
	err := c.SendService.DeleteMediaId(accessToken, req.MediaId)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.Success(ctx, req.Wxid+"删除成功", req.MediaId)
}
func (c *SendController) DeleteThumb(ctx *gin.Context) {
	var req struct {
		Wxid         string `json:"wxid" binding:"required"`
		Secret       string `json:"secret" binding:"required"`
		ThumbMediaId string `json:"thumbMediaId" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, err.Error())
	}

	accessToken := c.SendService.GetAccessToken(req.Wxid, req.Secret)
	if accessToken == "" {
		response.Error(ctx, req.Wxid+"获取到的accessToken为空")
		return
	}
	err := c.SendService.DeleteThumbId(accessToken, req.ThumbMediaId)
	if err != nil {
		response.Error(ctx, err.Error())
	}
	response.Success(ctx, req.Wxid+"删除成功", req.ThumbMediaId)
}

type Task struct {
	WxID             string    `json:"wxid" binding:"required"`
	Secret           string    `json:"secret" binding:"required"`
	ScheduledTime    time.Time `json:"scheduledTime" binding:"required"`
	TaskType         string    `json:"taskType" binding:"required,oneof=draft publish"`
	Mode             string    `json:"mode" binding:"required,oneof=random sequential manual"`
	ArticleCount     int       `json:"articleCount" binding:"required,min=1"`
	SelectedArticles []string  `json:"selectedArticles" binding:"omitempty,dive,required"`
	SelectedTitles   []string  `json:"selectedTitles" binding:"omitempty,dive,required"`
	ThumbID          string    `json:"thumbId"`
	TemplateID       string    `json:"templateId"`
}

type BatchScheduleTaskRequest struct {
	Tasks []Task `json:"tasks" binding:"required,dive,required"`
}

type DeleteScheduledTaskRequest struct {
	TaskID uint `json:"taskID" binding:"required"`
}

func (c *SendController) BatchScheduleTask(ctx *gin.Context) {
	var req BatchScheduleTaskRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "请求参数错误: "+err.Error())
		return
	}

	for _, task := range req.Tasks {

		if task.ScheduledTime.Before(time.Now()) {
			response.Error(ctx, "定时发送时间必须在未来")
			return
		}

		if task.TaskType == "draft" {
			if task.ThumbID == "" {
				response.Error(ctx, "草稿任务必须提供 thumbId")
				return
			}
			if task.TemplateID == "" {
				response.Error(ctx, "草稿任务必须提供 templateId")
				return
			}
			if task.Mode == "manual" && len(task.SelectedTitles) == 0 {
				response.Error(ctx, "手动模式下必须选择至少一个标题")
				return
			}
		} else if task.TaskType == "publish" {

			task.ThumbID = ""
			task.TemplateID = ""
		}

		selectedArticlesJSON, err := json.Marshal(task.SelectedArticles)
		if err != nil {

			response.Error(ctx, "处理 SelectedArticles 失败: "+err.Error())
			return
		}

		selectedTitlesJSON, err := json.Marshal(task.SelectedTitles)
		if err != nil {

			response.Error(ctx, "处理 SelectedTitles 失败: "+err.Error())
			return
		}

		scheduledTask := models.ScheduledTask{
			WxID:             task.WxID,
			Secret:           task.Secret,
			ScheduledTime:    task.ScheduledTime,
			TaskType:         task.TaskType,
			Mode:             task.Mode,
			ArticleCount:     task.ArticleCount,
			SelectedArticles: string(selectedArticlesJSON),
			SelectedTitles:   string(selectedTitlesJSON),
			ThumbID:          task.ThumbID,
			TemplateID:       task.TemplateID,
			Status:           "pending",
		}

		if err := c.DB.Create(&scheduledTask).Error; err != nil {

			response.Error(ctx, "保存任务失败: "+err.Error())
			return
		}

		job, err := c.Scheduler.Every(1).Day().StartAt(scheduledTask.ScheduledTime).Do(func(taskID uint) {
			var task models.ScheduledTask
			if err := c.DB.First(&task, taskID).Error; err != nil {

				return
			}
			c.executeTaskFromDB(task)

			_ = c.Scheduler.RemoveByTag(strconv.Itoa(int(task.ID)))
		}, scheduledTask.ID)

		if err != nil {

			response.Error(ctx, "调度任务失败: "+err.Error())
			return
		}

		job.Tag(strconv.Itoa(int(scheduledTask.ID)))

	}

	response.Success(ctx, "批量定时任务已成功创建", nil)
}

func (c *SendController) DeleteScheduledTask(ctx *gin.Context) {
	var req DeleteScheduledTaskRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "请求参数错误: "+err.Error())
		return
	}

	var task models.ScheduledTask
	if err := c.DB.First(&task, req.TaskID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(ctx, "任务未找到")
			return
		}
		response.Error(ctx, "查找任务失败: "+err.Error())
		return
	}

	if err := c.DB.Delete(&task).Error; err != nil {

		response.Error(ctx, "从数据库中删除任务失败: "+err.Error())
		return
	}

	err := c.Scheduler.RemoveByTag(strconv.Itoa(int(task.ID)))
	if err != nil && err.Error() != "gocron: no jobs found with given tag" {

		response.Error(ctx, "从调度器中删除任务失败: "+err.Error())
		return
	}

	response.Success(ctx, "定时任务已成功删除", nil)
}

func (c *SendController) executeTaskFromDB(task models.ScheduledTask) {

	var selectedArticles []string
	if err := json.Unmarshal([]byte(task.SelectedArticles), &selectedArticles); err != nil {

		c.DB.Model(&task).Update("status", "failed")
		return
	}

	taskData := Task{
		WxID:             task.WxID,
		Secret:           task.Secret,
		ScheduledTime:    task.ScheduledTime,
		TaskType:         task.TaskType,
		Mode:             task.Mode,
		ArticleCount:     task.ArticleCount,
		SelectedArticles: selectedArticles,
		ThumbID:          task.ThumbID,
		TemplateID:       task.TemplateID,
	}

	c.executeTask(taskData, task.ID)

	if taskData.TaskType == "publish" || taskData.TaskType == "draft" {
		if task.Status != "failed" {

			if err := c.DB.Model(&task).Update("status", "completed").Error; err != nil {

			}
		}
	}
}

func (c *SendController) executeTask(task Task, scheduledTaskID uint) {

	switch task.TaskType {
	case "draft":
		c.handleDraftTask(task, scheduledTaskID)
	case "publish":
		c.handlePublishTask(task, scheduledTaskID)
	default:

	}
}

func (c *SendController) handleDraftTask(task Task, scheduledTaskID uint) {

	accessToken := c.SendService.GetAccessToken(task.WxID, task.Secret)
	if accessToken == "" {

		c.DB.Model(&models.ScheduledTask{}).Where("id = ?", scheduledTaskID).Update("status", "failed")
		return
	}

	titles, _, err := c.SendService.GetTitleList(0, 0)

	if err != nil {

		c.DB.Model(&models.ScheduledTask{}).Where("id = ?", scheduledTaskID).Update("status", "failed")
		return
	}

	if len(titles) == 0 {

		c.DB.Model(&models.ScheduledTask{}).Where("id = ?", scheduledTaskID).Update("status", "failed")
		return
	}

	for i := 0; i < task.ArticleCount; i++ {

		intn := rand.Intn(len(titles))
		selectedTitle := titles[intn].Title

		err = c.SendService.WrittenDraft(task.WxID, accessToken, task.ThumbID, task.TemplateID, selectedTitle, task.Mode)
		if err != nil {

			c.DB.Model(&models.ScheduledTask{}).Where("id = ?", scheduledTaskID).Update("status", "failed")
			return
		}

	}

	// 更新任务状态为完成
	c.DB.Model(&models.ScheduledTask{}).Where("id = ?", scheduledTaskID).Update("status", "completed")
}

func (c *SendController) handlePublishTask(task Task, scheduledTaskID uint) {

	accessToken := c.SendService.GetAccessToken(task.WxID, task.Secret)
	if accessToken == "" {

		c.DB.Model(&models.ScheduledTask{}).Where("id = ?", scheduledTaskID).Update("status", "failed")
		return
	}

	mediaModels, err := c.SendService.GetAllMediaIdByWxid(task.WxID)
	if err != nil {

		c.DB.Model(&models.ScheduledTask{}).Where("id = ?", scheduledTaskID).Update("status", "failed")
		return
	}

	var mediaIds []string
	if task.Mode == "manual" {
		for _, m := range task.SelectedArticles {

			mediaIds = append(mediaIds, m)
		}
	} else {
		for _, m := range mediaModels {

			mediaIds = append(mediaIds, m.MediaId)
		}
	}

	if len(mediaIds) == 0 {

		c.DB.Model(&models.ScheduledTask{}).Where("id = ?", scheduledTaskID).Update("status", "failed")
		return
	}

	publishCount := task.ArticleCount
	if len(mediaIds) < task.ArticleCount {
		publishCount = len(mediaIds)

	}
	var selectedMediaIds []string
	if task.Mode == "manual" {
		selectedMediaIds = mediaIds
	} else {
		selectedMediaIds = randomSelectArticles(mediaIds, publishCount)
	}

	stop := c.SendService.SetWxPost(task.WxID)
	if stop != 2 {

		c.DB.Model(&models.ScheduledTask{}).Where("id = ?", scheduledTaskID).Update("status", "failed")
		return
	}

	for _, mediaId := range selectedMediaIds {

		stop = c.SendService.SetWxPost(task.WxID)
		err = c.SendService.PostDart(task.WxID, accessToken, mediaId)
		if stop != 2 {

			c.DB.Model(&models.ScheduledTask{}).Where("id = ?", scheduledTaskID).Update("status", "failed")
			continue
		}
		if err != nil {

			c.DB.Model(&models.ScheduledTask{}).Where("id = ?", scheduledTaskID).Update("status", "failed")
			continue
		}
	}

	c.DB.Model(&models.ScheduledTask{}).Where("id = ?", scheduledTaskID).Update("status", "completed")
}
func (c *SendController) CleanInvalidDrafts(ctx *gin.Context) {
	var req struct {
		Wxid     string   `json:"wxid" binding:"required"`
		Secret   string   `json:"secret" binding:"required"`
		MediaIds []string `json:"mediaIds" binding:"required"`
	}
	if err := ctx.ShouldBind(&req); err != nil {
		response.Error(ctx, "请求参数错误: "+err.Error())
		return
	}
	accessToken := c.SendService.GetAccessToken(req.Wxid, req.Secret)

	if accessToken == "" {
		response.Error(ctx, req.Wxid+"获取到的accessToken为空")
		return
	}
	for _, mediaId := range req.MediaIds {
		_, err := c.SendService.GetPreviewLink(accessToken, mediaId)
		if err != nil {

			errs := c.SendService.DeleteMediaIdToDataBase(mediaId)
			fmt.Println(errs)
		}
	}

	response.Success(ctx, req.Wxid+"mediaId已清理", nil)
}
func (c *SendController) GetTasks(ctx *gin.Context) {
	var req struct {
		Wxid string `json:"wxid" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "请求参数错误: "+err.Error())
		return
	}

	var tasks []models.ScheduledTask
	if err := c.DB.Where("wxid = ?", req.Wxid).Find(&tasks).Error; err != nil {

		response.Error(ctx, "获取任务失败: "+err.Error())
		return
	}

	type TaskResponse struct {
		ID            uint    `json:"id"`
		ScheduledTime string  `json:"scheduledTime"`
		TaskType      string  `json:"taskType"`
		Mode          string  `json:"mode"`
		ArticleCount  int     `json:"articleCount"`
		ThumbID       *string `json:"thumbId,omitempty"`
		TemplateID    *string `json:"templateId,omitempty"`
		Status        string  `json:"status"`
		MediaID       string  `json:"mediaId"`
	}

	var taskResponses []TaskResponse
	for _, task := range tasks {
		var mediaID string

		if len(task.SelectedArticles) > 0 {
			var selectedArticles []string
			if err := json.Unmarshal([]byte(task.SelectedArticles), &selectedArticles); err == nil && len(selectedArticles) > 0 {
				mediaID = selectedArticles[0]
			}
		}

		var thumbID *string
		var templateID *string
		if task.TaskType == "draft" {
			thumbID = &task.ThumbID
			templateID = &task.TemplateID
		}

		taskResponses = append(taskResponses, TaskResponse{
			ID:            task.ID,
			ScheduledTime: task.ScheduledTime.Format(time.RFC3339),
			TaskType:      task.TaskType,
			Mode:          task.Mode,
			ArticleCount:  task.ArticleCount,
			ThumbID:       thumbID,
			TemplateID:    templateID,
			Status:        task.Status,
			MediaID:       mediaID,
		})
	}

	response.Success(ctx, "获取任务成功", taskResponses)
}

func randomSelectArticles(mediaIds []string, count int) []string {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(mediaIds), func(i, j int) {
		mediaIds[i], mediaIds[j] = mediaIds[j], mediaIds[i]
	})
	if count > len(mediaIds) {
		count = len(mediaIds)
	}
	return mediaIds[:count]
}
