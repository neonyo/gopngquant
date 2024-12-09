package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type Url struct {
	Url string `json:"url"`
}

func main() {
	r := gin.Default()
	r.Static("/temp", "./temp")
	r.POST("/compress", func(c *gin.Context) {
		file, err := c.FormFile("file")
		ext := filepath.Ext(file.Filename)
		if ext == "" {
			c.JSON(200, gin.H{
				"code": 400,
				"msg":  "no file",
			})
			return
		}
		inputFile := GetLocalFile(ext)
		if err = c.SaveUploadedFile(file, inputFile); err != nil {
			c.String(http.StatusBadRequest, "保存文件错误： %s", err.Error())
			return
		}
		defer RemoveFile(inputFile, 30*time.Second)
		outputFile := GetLocalFile(ext)
		//fmt.Println(fmt.Sprintf("pngquant --quality=65-80 %s --output %s", inputFile, outputFile))
		cmd := exec.Command("pngquant", "--quality=40-80", inputFile, "--output", outputFile, "--speed=10")
		var o bytes.Buffer
		cmd.Stdout = &o
		err = cmd.Run()
		fmt.Println(cmd.String(), "----cmd-----")
		if err != nil {
			fmt.Println("1=====")
			c.JSON(200, gin.H{
				"code": 400,
				"msg":  err.Error(),
			})
			return
		}
		output := o.Bytes()
		fmt.Println(string(output), "=====")
		if len(output) == 0 {
			defer RemoveFile(outputFile, 60*time.Second)
			c.JSON(200, gin.H{
				"code": 200,
				"data": outputFile,
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

func GetLocalFile(ext string) string {
	now := time.Now()
	rand.New(rand.NewSource(time.Now().UnixNano()))
	threeDigitRand := rand.Intn(9000) + 1000
	return fmt.Sprintf("temp/%d%d%d%d%s", now.Year(), now.Month(), now.UnixMilli(), threeDigitRand, ext)
}
