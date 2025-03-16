package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"publicPost/src/controllers"
	"publicPost/src/models"
)

func main() {

	dsn := "user:password@tcp(ip:port)/database?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库出错: %v", err)
	}
	log.Println("成功连接数据库.")

	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("自动迁移失败: %v", err)
	}

	r := controllers.SetupRouter(db)

	if err := r.Run(":8888"); err != nil {
		log.Fatalf("启动容器服务失败: %v", err)
	}
}
