package imageController

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"image-host/app/apiException"
	"image-host/app/config"
	"image-host/app/models"
	"image-host/app/services/nameMapServices"
	"image-host/app/utils"
	"image-host/config/database"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type ImageReqData struct {
	Date string `json:"date"`
	Src  string `json:"src"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DeleteImgForm struct {
	ID string `json:"id"`
}

func GetImage(c *gin.Context) {
	PageNum := c.Query("page_num")
	PageSize := c.Query("page_size")
	PN, _ := strconv.Atoi(PageNum)
	PS, _ := strconv.Atoi(PageSize)
	if PN <= 0 {
		PN = 1
	}
	var imageList []models.NameMap
	err := database.DB.Where(&models.NameMap{Type: "image/webp"}).Limit(PS).Offset((PN - 1) * PS).Order("Date desc").Find(&imageList).Error
	if err != nil {
		_ = c.AbortWithError(200, apiException.ServerError)
		return
	}
	var total int64
	if err := database.DB.Model(&models.NameMap{}).Count(&total).Error; err != nil {
		_ = c.AbortWithError(200, apiException.ServerError)
		return
	}

	reqList := make([]ImageReqData, len(imageList))
	for i, image := range imageList {
		reqList[i].ID = image.UUID
		reqList[i].Name = image.Src
		reqList[i].Src = config.GetWebpUrlKey() + image.Path
		reqList[i].Date = image.Date.String()
	}
	utils.JsonSuccessResponse(c, gin.H{
		"list":  reqList,
		"total": total,
	})
}

func UploadImg(c *gin.Context) {
	// 存储文件
	form, _ := c.MultipartForm()
	images := form.File["img"]
	for _, img := range images {
		imgName := img.Filename
		if err := c.SaveUploadedFile(img, "./tmp/"+imgName); err != nil {
			//log.Println(err)
			_ = c.AbortWithError(200, apiException.ImgTypeError)
			return
		}
		// 打开并判断文件类型
		file, _ := os.Open("./tmp/" + imgName)

		buffer := make([]byte, 512)
		_, _ = file.Read(buffer)
		contentType := http.DetectContentType(buffer)
		file.Close()

		// 重启文件并转换类型
		file, _ = os.Open("./tmp/" + imgName)
		imgPrefix := strings.TrimSuffix(img.Filename, path.Ext(imgName))
		if contentType == "image/png" {
			// 为了处理一些仅修改了后缀而并未重新编码的图片，所有 png 文件都改为正确后缀
			newTypeName := "./tmp/" + imgPrefix + ".png"
			_ = os.Rename("./tmp/"+imgName, newTypeName)

			// png2jpg
			imgNew, err := png.Decode(file)
			if err != nil {
				fmt.Println(err)
				_ = c.AbortWithError(200, apiException.ImgTypeError)
				return
			}
			out, err := os.Create("./tmp/" + imgPrefix + ".jpg")
			if err != nil {
				fmt.Println(err)
				_ = c.AbortWithError(200, apiException.ImgTypeError)
				return
			}

			err = jpeg.Encode(out, imgNew, &jpeg.Options{Quality: 95})
			out.Close()
			if err != nil {
				fmt.Println(err)
				_ = c.AbortWithError(200, apiException.ImgTypeError)
				return
			}
			_ = os.Remove(newTypeName)
			imgName = imgPrefix + ".jpg"
			file.Close()
			file, _ = os.Open("./tmp/" + imgName)
		}

		// jpg2webp
		imgNew, err := jpeg.Decode(file)
		file.Close()
		if err != nil {
			_ = c.AbortWithError(300, apiException.ImgTypeError)
			return
		}
		fileName := uuid.NewString() + ".webp"
		output, _ := os.Create("./img/" + fileName)
		options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 75)
		if err != nil {
			_ = c.AbortWithError(200, apiException.ImgTypeError)
			return
		}
		err = webp.Encode(output, imgNew, options)
		output.Close()
		if err != nil {
			_ = c.AbortWithError(200, apiException.ImgTypeError)
			return
		}
		_ = os.Remove("./tmp/" + imgName)
		fileInfo, err := os.Stat("./img/" + fileName)
		err = nameMapServices.Insert(models.NameMap{
			Src:  imgPrefix + ".webp",
			UUID: fileName,
			Type: "image/webp",
			Path: "/img/" + fileName,
			Size: fileInfo.Size(),
			Date: time.Now(),
		})
		if err != nil {
			_ = c.AbortWithError(200, apiException.ServerError)
			return
		}
	}
	utils.JsonSuccessResponse(c, nil)
}

func DeleteImg(c *gin.Context) {
	var postForm DeleteImgForm
	errBind := c.ShouldBindJSON(&postForm)
	if errBind != nil {
		_ = c.AbortWithError(200, apiException.ParamError)
		return
	}
	image, err := nameMapServices.QueryByUUID(postForm.ID)
	if err != nil {
		_ = c.AbortWithError(200, apiException.ServerError)
		return
	}
	err = nameMapServices.DeleteImage(&image)
	if err != nil {
		log.Println(err)
		_ = c.AbortWithError(400, apiException.ServerError)
		return
	}
	err = os.Remove("." + image.Path)
	if err != nil {
		_ = c.AbortWithError(300, apiException.ServerError)
		return
	}
	utils.JsonSuccessResponse(c, nil)
}
