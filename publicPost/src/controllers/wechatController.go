// src/controllers/wechatController.go
package controllers

import (
	"net/http"
	"publicPost/src/models"
	"publicPost/src/response"
	"publicPost/src/services"
	"time"

	"github.com/gin-gonic/gin"
)

type WechatController struct {
	WechatService *services.WechatService
}

func NewWechatController(wechatService *services.WechatService) *WechatController {
	return &WechatController{
		WechatService: wechatService,
	}
}

func (wc *WechatController) AddWechatAccount(c *gin.Context) {
	var newAccount models.WechatAccount

	if err := c.ShouldBindJSON(&newAccount); err != nil {
		response.Error(c, "无效的请求参数")
		return
	}

	account, err := wc.WechatService.AddAccount(newAccount.Name, newAccount.Wxid, newAccount.Secret, newAccount.BindWechat)
	if err != nil {
		response.Error(c, "添加微信账号失败")
		return
	}

	response.Success(c, "添加微信账号成功", account)
}

func (wc *WechatController) DeleteWechatAccount(c *gin.Context) {
	var req struct {
		Wxid string `json:"wxid"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {

		response.Error(c, "请求参数错误")
		return
	}

	err := wc.WechatService.DeleteAccount(req.Wxid)
	if err != nil {
		response.Error(c, "删除微信账号失败")
		return
	}

	response.Success(c, "删除微信账号成功", nil)
}

func (wc *WechatController) GetWechat(c *gin.Context) {

	wechatInfo, err := wc.WechatService.GetWechatInfo()

	if err != nil {
		response.Error(c, "获取微信出错")

		return
	}

	c.JSON(http.StatusOK, gin.H{"info": wechatInfo})
}

func (wc *WechatController) UpdateWechatAccount(c *gin.Context) {
	var req models.WechatAccount

	if err := c.ShouldBindJSON(&req); err != nil {

		response.Error(c, "请求参数错误")
		return
	}

	err := wc.WechatService.UpdateAccount(req)
	if err != nil {
		response.Error(c, "更新微信账号失败")
		return
	}

	response.Success(c, "更新微信账号成功", nil)
}

func (wc *WechatController) GetMediaPost(c *gin.Context) {
	var req struct {
		BelongWxid string `json:"belongWxid"`
		StartDate  string `json:"startDate"`
		EndDate    string `json:"endDate"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {

		response.Error(c, "请求参数错误")
		return
	}

	var startDate, endDate time.Time
	var err error

	if req.StartDate == "" {

		startDate = time.Now().Truncate(24 * time.Hour)
	} else {

		startDate, err = time.Parse(time.RFC3339, req.StartDate)
		if err != nil {

			response.Error(c, "时间错误")
			return
		}
	}

	if req.EndDate == "" {

		endDate = time.Now().Truncate(24 * time.Hour).Add(24*time.Hour - time.Nanosecond)
	} else {

		endDate, err = time.Parse(time.RFC3339, req.EndDate)
		if err != nil {

			response.Error(c, "时间错误")
			return
		}
	}

	posts := wc.WechatService.GetWechatMediaByPost(req.BelongWxid, startDate, endDate)

	response.Success(c, "获取到的数据", posts)
}
