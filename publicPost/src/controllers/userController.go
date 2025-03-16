// src/controllers/user_controller.go

package controllers

import (
	"net/http"
	"publicPost/src/dtos"
	"publicPost/src/response"
	"strings"

	"github.com/gin-gonic/gin"
	"publicPost/src/services"
)

type UserController struct {
	UserService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{
		UserService: userService,
	}
}

func (c *UserController) Register(ctx *gin.Context) {
	req := dtos.UserDTO{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "无效请求")

		return
	}

	user, err := c.UserService.Register(req.Username, req.Password)
	if err != nil {
		response.Error(ctx, "无效请求注册")
		return
	}
	response.Success(ctx, "注册成功", user)

}
func (c *UserController) Delete(ctx *gin.Context) {
	var req struct {
		Id int `json:"id"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "无效请求")
	}
	err := c.UserService.Delete(req.Id)
	if err != nil {
		response.Error(ctx, "删除无效请求")
	}
	response.Success(ctx, "删除成功", nil)

}
func (c *UserController) UpdatePwd(ctx *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Power    int    `json:"power"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {

		response.Error(ctx, "无效请求")
		return
	}

	err := c.UserService.UpdateUserPwd(req.Username, req.Password, req.Power)

	if err != nil {
		response.Error(ctx, "更新无效请求")
		return
	}
	response.Success(ctx, "更新成功", nil)
}
func (c *UserController) GetAllUser(ctx *gin.Context) {
	var req struct {
		Username string `form:"username"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "无效请求")
		return
	}

	users, err := c.UserService.GetAllUsers(req.Username)

	if err != nil {
		response.Error(ctx, "无效请求")
		return
	}
	response.Success(ctx, "注册成功", users)

}

func (c *UserController) Login(ctx *gin.Context) {
	req := dtos.UserDTO{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "无效请求")
		return
	}

	user, accessToken, refreshToken, err := c.UserService.Login(req.Username, req.Password)
	if err != nil {
		response.Error(ctx, "无效请求")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":       "Login successful",
		"user":          user.Username,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (c *UserController) RefreshToken(ctx *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "无效请求")
		return
	}

	newAccessToken, newRefreshToken, err := c.UserService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Error(ctx, "无效请求")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
}

func (c *UserController) Logout(ctx *gin.Context) {

	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		response.Error(ctx, "无效请求")
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		response.Error(ctx, "无效请求")
		return
	}

	accessToken := parts[1]

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, "无效请求")
		return
	}

	err := c.UserService.Logout(accessToken, req.RefreshToken)
	if err != nil {
		response.Error(ctx, "无效请求")
		return
	}

	response.Success(ctx, "ok", nil)
}
