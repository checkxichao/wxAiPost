package controllers

import (
	"github.com/gin-gonic/gin"

	"publicPost/src/models"
	"publicPost/src/response"
	"publicPost/src/services"
)

type GptController struct {
	GptService *services.GptService
}

func NewGptController(gptService *services.GptService) *GptController {
	return &GptController{
		GptService: gptService,
	}
}

func (g *GptController) GetGptInfo(c *gin.Context) {
	info := g.GptService.GetInfo()
	if info != nil {
		response.Success(c, "获取成功", info)
	} else {
		response.Success(c, "没有 GPT 设置", nil)
	}
}

func (g *GptController) SetGptInfo(c *gin.Context) {
	var newAccount models.GptModel

	if err := c.ShouldBindJSON(&newAccount); err != nil {
		response.Error(c, "无效的请求参数")
		return
	}

	err := g.GptService.SetInfo(newAccount)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, "设置成功", nil)
}
