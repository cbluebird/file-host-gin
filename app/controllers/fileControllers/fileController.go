package fileControllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"image-host/app/apiException"
	"image-host/app/models"
	"image-host/app/services/nameMapServices"
	"image-host/app/utils"
	"log"
	"net/url"
	"path"
)

func GetFile(c *gin.Context) {
	//uuidName := c.Query("file_name")
	uuidName, _ := url.QueryUnescape(c.Param("file_name"))
	file, err := nameMapServices.QueryByUUID(uuidName)
	if (err != nil || file == models.NameMap{}) {
		_ = c.AbortWithError(404, apiException.NotFound)
		return
	}
	c.FileAttachment("."+file.Path, file.Src)
	if err := nameMapServices.FileDownloadCountIncrement(file.UUID); err != nil {
		log.Println(err)
	}
	return
}

func UploadFile(c *gin.Context) {
	// 存储文件
	file, _ := c.FormFile("file")
	fileName := file.Filename
	uuidName := uuid.NewString() + path.Ext(fileName)
	_ = c.SaveUploadedFile(file, "./public/"+uuidName)
	err := nameMapServices.Insert(models.NameMap{
		Src:  fileName,
		UUID: uuidName,
		Type: file.Header.Get("Content-Type"),
		Path: "./public/" + uuidName,
		Size: file.Size,
	})
	if err != nil {
		_ = c.AbortWithError(200, apiException.ServerError)
	}
	utils.JsonSuccessResponse(c, uuidName)
}
