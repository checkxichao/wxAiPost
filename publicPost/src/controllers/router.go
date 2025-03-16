package controllers

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"publicPost/src/config"
	"publicPost/src/middleware"
	"publicPost/src/models"
	"publicPost/src/repositories"
	"publicPost/src/services"
	"strconv"
	"time"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:6608"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
	}))
	cfg := config.LoadConfig()

	userRepo := repositories.NewUserRepository(db)
	tokenRepo := repositories.NewTokenRepository(db)
	gptRepo := repositories.NewGptRepository(db)
	wechatRepo := repositories.NewWechatRepository(db)

	userService := services.NewUserService(userRepo, tokenRepo, cfg)
	wechatService := services.NewWechatService(wechatRepo)
	gptService := services.NewGptService(gptRepo)
	sendService := services.NewSendService(userRepo, wechatRepo, gptService)

	userController := NewUserController(userService)
	wechatController := NewWechatController(wechatService)
	gptController := NewGptController(gptService)
	sendController := NewSendController(sendService, db)

	authGroup := r.Group("/auth")
	authGroup.POST("/register", userController.Register)
	authGroup.POST("/login", userController.Login)
	authGroup.POST("/refresh", userController.RefreshToken)
	authGroup.POST("/logout", userController.Logout)

	protectedGroup := r.Group("/protected")
	protectedGroup.Use(middleware.JWTAuthMiddleware(db, cfg))

	userGroup := r.Group("/user")
	userGroup.POST("/getUser", middleware.JWTAuthMiddleware(db, cfg), userController.GetAllUser)
	userGroup.POST("/editUser", middleware.JWTAuthMiddleware(db, cfg), userController.UpdatePwd)
	userGroup.POST("/deleteUser", middleware.JWTAuthMiddleware(db, cfg), userController.Delete)

	wechatGroup := r.Group("/wechat")
	wechatGroup.GET("/getWechat", middleware.JWTAuthMiddleware(db, cfg), wechatController.GetWechat)
	wechatGroup.POST("/addWechat", middleware.JWTAuthMiddleware(db, cfg), wechatController.AddWechatAccount)
	wechatGroup.POST("/editWechat", middleware.JWTAuthMiddleware(db, cfg), wechatController.UpdateWechatAccount)
	wechatGroup.POST("/deleteWechat", middleware.JWTAuthMiddleware(db, cfg), wechatController.DeleteWechatAccount)
	wechatGroup.POST("/getMediaPost", middleware.JWTAuthMiddleware(db, cfg), wechatController.GetMediaPost)

	gptGroup := r.Group("/GPT")
	gptGroup.GET("/getGptInfo", middleware.JWTAuthMiddleware(db, cfg), gptController.GetGptInfo)
	gptGroup.POST("/SetGptInfo", middleware.JWTAuthMiddleware(db, cfg), gptController.SetGptInfo)

	sendGroup := r.Group("/send")
	sendGroup.POST("/getMedias", middleware.JWTAuthMiddleware(db, cfg), sendController.GetMediaListByWxid)
	sendGroup.POST("/getMediaListByWxidP", middleware.JWTAuthMiddleware(db, cfg), sendController.GetMediaListByWxidP)
	sendGroup.POST("/postDart", middleware.JWTAuthMiddleware(db, cfg), sendController.SetStatePost)
	sendGroup.POST("/writeDart", middleware.JWTAuthMiddleware(db, cfg), sendController.SetStateDart)
	sendGroup.POST("/getThumbList", middleware.JWTAuthMiddleware(db, cfg), sendController.GetThumbMediaIds)
	sendGroup.POST("/uploadThumb", middleware.JWTAuthMiddleware(db, cfg), sendController.UploadImageGetThumbId)
	sendGroup.GET("/getTemplate/:templateId", middleware.JWTAuthMiddleware(db, cfg), sendController.GetTemplate)
	sendGroup.GET("/getTemplateList", middleware.JWTAuthMiddleware(db, cfg), sendController.GetTemplateList)
	sendGroup.POST("/setTemplate", middleware.JWTAuthMiddleware(db, cfg), sendController.SetTemplate)
	sendGroup.POST("/addTemplate", middleware.JWTAuthMiddleware(db, cfg), sendController.AddTemplate)
	sendGroup.POST("/setStop", middleware.JWTAuthMiddleware(db, cfg), sendController.SetStateStop)
	sendGroup.GET("/getTitleList", middleware.JWTAuthMiddleware(db, cfg), sendController.GetTitleList)
	sendGroup.POST("/getTitleListSearch", middleware.JWTAuthMiddleware(db, cfg), sendController.GetTitleListSearch)
	sendGroup.POST("/setTitle", middleware.JWTAuthMiddleware(db, cfg), sendController.SetTitle)
	sendGroup.POST("/editTitle", middleware.JWTAuthMiddleware(db, cfg), sendController.EditTitle)
	sendGroup.POST("/deleteTitle", middleware.JWTAuthMiddleware(db, cfg), sendController.DeleteTitle)
	sendGroup.POST("/deleteTitles", middleware.JWTAuthMiddleware(db, cfg), sendController.DeleteTitleBatch)
	sendGroup.POST("/setTitleBatch", middleware.JWTAuthMiddleware(db, cfg), sendController.SetTitleByTxt)
	sendGroup.POST("/getPreviewLink", middleware.JWTAuthMiddleware(db, cfg), sendController.GetPreviewLink)
	sendGroup.POST("/deleteMedia", middleware.JWTAuthMiddleware(db, cfg), sendController.DeleteMedia)
	sendGroup.POST("/batchScheduleTask", middleware.JWTAuthMiddleware(db, cfg), sendController.BatchScheduleTask)
	sendGroup.POST("/deleteScheduledTask", middleware.JWTAuthMiddleware(db, cfg), sendController.DeleteScheduledTask)
	sendGroup.POST("/getTasks", middleware.JWTAuthMiddleware(db, cfg), sendController.GetTasks)
	sendGroup.POST("/cleanInvalidDrafts", middleware.JWTAuthMiddleware(db, cfg), sendController.CleanInvalidDrafts)
	loadAndScheduleTasks(sendController, db)
	return r
}

func loadAndScheduleTasks(sendController *SendController, db *gorm.DB) {
	var tasks []models.ScheduledTask

	if err := db.Where("scheduled_time > ?", time.Now()).Find(&tasks).Error; err != nil {

		return
	}

	for _, task := range tasks {

		job, err := sendController.Scheduler.Every(1).Day().StartAt(task.ScheduledTime).Do(func(taskID uint) {
			var task models.ScheduledTask
			if err := db.First(&task, taskID).Error; err != nil {

				return
			}
			sendController.executeTaskFromDB(task)
		}, task.ID)

		if err != nil {

			continue
		}

		job.Tag(strconv.Itoa(int(task.ID)))

	}
}
