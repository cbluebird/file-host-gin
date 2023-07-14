package router

import (
	"github.com/gin-gonic/gin"
	"image-host/app/controllers/fileControllers"
	"image-host/app/controllers/imageController"
	"image-host/app/midwares"
)

func imageRouterInit(r *gin.RouterGroup) {
	fun := r.Group("/image", midwares.CheckLogin)
	{
		fun.GET("/get", imageController.GetImage)
		fun.POST("/upload_img", imageController.UploadImg)
		fun.POST("/delete", imageController.DeleteImg)
		fun.GET("/download/:file_name", fileControllers.GetFile)
	}
}
