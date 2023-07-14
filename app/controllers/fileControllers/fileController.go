package fileControllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"image-host/app/apiException"
	"image-host/app/models"
	"image-host/app/services/nameMapServices"
	"image-host/app/utils"
	"image-host/config/database"
	"log"
	"net/url"
	"path"
	"strconv"
	"time"
)

type FileReqData struct {
	Type     string `json:"type"`
	FileName string `json:"name"`
	UUid     string `json:"uuid"`
	Size     int64  `json:"size"`
	Date     string `json:"date"`
}

func GetFile(c *gin.Context) {
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
	//file, _ := c.FormFile("file")
	//fileName := file.Filename
	form, _ := c.MultipartForm()
	files := form.File["file"]
	log.Println(files)
	for _, file := range files {
		fileName := file.Filename
		uuidName := uuid.NewString() + path.Ext(fileName)
		_ = c.SaveUploadedFile(file, "./public/"+uuidName)
		err := nameMapServices.Insert(models.NameMap{
			Src:  fileName,
			UUID: uuidName,
			Type: file.Header.Get("Content-Type"),
			Path: "/public/" + uuidName,
			Size: file.Size,
			Date: time.Now(),
		})
		if err != nil {
			_ = c.AbortWithError(200, apiException.ServerError)
		}
	}
	utils.JsonSuccessResponse(c, nil)
}

func GetFileList(c *gin.Context) {
	PageNum := c.Query("page_num")
	PageSize := c.Query("page_size")
	PN, _ := strconv.Atoi(PageNum)
	PS, _ := strconv.Atoi(PageSize)
	if PN <= 0 {
		PN = 1
	}
	var fileList []models.NameMap
	err := database.DB.Where(&models.NameMap{}).Limit(PS).Offset((PN - 1) * PS).Order("Date desc").Find(&fileList).Error
	if err != nil {
		_ = c.AbortWithError(200, apiException.ServerError)
		return
	}
	var total int64
	if err := database.DB.Model(&models.NameMap{}).Count(&total).Error; err != nil {
		_ = c.AbortWithError(200, apiException.ServerError)
		return
	}

	reqList := make([]FileReqData, len(fileList))
	for i, file := range fileList {
		reqList[i].UUid = file.UUID
		reqList[i].FileName = file.Src
		reqList[i].Size = file.Size
		reqList[i].Date = file.Date.String()
		reqList[i].Type = file.Type
	}
	utils.JsonSuccessResponse(c, gin.H{
		"list":  reqList,
		"total": total,
	})
}
