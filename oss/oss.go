package oss

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"log"
	"os"
)

var Client *Oss

func init() {
	var err error
	Client, err = NewOss(
		os.Getenv("OssEndPoint"),
		os.Getenv("OssAccessKeyId"),
		os.Getenv("OssAccessKeySecret"),
		os.Getenv("OssBucketName"),
		os.Getenv("OssImgCurl"),
	)
	if err != nil {
		log.Fatalln(err)
	}
}

type Oss struct {
	BucketName string
	ImgCurl    string
	client     *oss.Client
}

func NewOss(Endpoint, AccessKeyId, AccessKeySecret, BucketName, ImgCurl string) (*Oss, error) {
	client, err := oss.New(Endpoint, AccessKeyId, AccessKeySecret)
	if err != nil {
		return nil, err
	}
	return &Oss{
		client:     client,
		ImgCurl:    ImgCurl,
		BucketName: BucketName,
	}, nil
}

func (o *Oss) PutImg(urlPath, fileName, contentType string) (url string, err error) {
	bucket, _ := o.client.Bucket(o.BucketName)
	options := []oss.Option{
		oss.ContentType(contentType),
	}
	err = bucket.PutObjectFromFile(fileName, urlPath, options...)
	url = fmt.Sprintf("https://%s/%s", o.ImgCurl, fileName)
	return
}

func (o *Oss) GetImg(fileName, outFile string) error {
	bucket, _ := o.client.Bucket(o.BucketName)
	return bucket.GetObjectToFile(fileName, outFile)
}
