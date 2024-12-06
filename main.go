package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"pngquant/oss"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Url struct {
	Url string `json:"url"`
}

func main() {
	r := gin.Default()
	r.POST("/compress", func(c *gin.Context) {
		var req Url
		err := c.ShouldBindJSON(&req)
		if err != nil {
			c.JSON(200, gin.H{
				"code": 400,
				"msg":  err.Error(),
			})
			return
		}
		ext := filepath.Ext(req.Url)
		if ext == "" {
			c.JSON(200, gin.H{
				"code": 400,
				"msg":  "no file",
			})
			return
		}
		parsedURL, _ := url.Parse(req.Url)
		path := strings.TrimLeft(parsedURL.Path, "/")
		inputFile := GetLocalFile(ext)
		fmt.Println(path, inputFile, "---")
		err = oss.Client.GetImg(path, inputFile)
		if err != nil {
			c.JSON(200, gin.H{
				"code": 400,
				"msg":  err.Error(),
			})
			return
		}
		defer RemoveFile(inputFile, 30*time.Second)
		outFile := GetLocalFile(ext)
		cmd := exec.Command("pngquant", "--quality=65-80", inputFile, "--output", outFile)
		var o bytes.Buffer
		cmd.Stdout = &o
		err = cmd.Run()
		if err != nil {
			c.JSON(200, gin.H{
				"code": 400,
				"msg":  err.Error(),
			})
			return
		}
		output := o.Bytes()

		if len(output) == 0 {
			defer RemoveFile(outFile, 30*time.Second)
			ossFilePath := GetOssFile(ext)
			contentType := GetContentType(ext)
			var ossUrl string
			ossUrl, err = oss.Client.PutImg(outFile, ossFilePath, contentType)
			c.JSON(200, gin.H{
				"code": 200,
				"data": ossUrl,
				"msg":  "ok",
			})
			return
		}
		c.JSON(200, gin.H{
			"code": 400,
			"msg":  string(output),
		})
	})
	err := r.Run(":8082") // 监听并在 0.0.0.0:8080 上启动服务
	if err != nil {
		log.Fatalln(err)
	}
}

func RemoveFile(filePath string, delay time.Duration) {
	go func() {
		// 使用定时器来等待指定的延迟时间
		timer := time.NewTimer(delay)
		defer timer.Stop() // 确保在 goroutine 结束时停止定时器

		// 阻塞等待定时器触发
		<-timer.C

		// 尝试删除文件
		err := os.Remove(filePath)
		if err != nil {
			fmt.Printf("无法删除文件 %s: %v\n", filePath, err)
		} else {
			fmt.Printf("文件 %s 已成功删除\n", filePath)
		}
	}()
}

func GetOssFile(ext string) string {
	now := time.Now()
	rand.New(rand.NewSource(time.Now().UnixNano()))
	threeDigitRand := rand.Intn(9000) + 1000
	return fmt.Sprintf("newCrm/%d/%d/%d%d%s", now.Year(), now.Month(), now.UnixMilli(), threeDigitRand, ext)
}

func GetLocalFile(ext string) string {
	now := time.Now()
	rand.New(rand.NewSource(time.Now().UnixNano()))
	threeDigitRand := rand.Intn(9000) + 1000
	return fmt.Sprintf("temp/%d%d%d%d%s", now.Year(), now.Month(), now.UnixMilli(), threeDigitRand, ext)
}

func GetContentType(ext string) string {
	if ext == "" {
		return ""
	}
	switch strings.ToUpper(ext[1:]) {
	case "PNG":
		return "image/png"
	case "JPG", "JPEG":
		return "image/jpeg"
	default:
		return ""
	}
}
