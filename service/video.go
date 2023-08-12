package service

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	cos "github.com/tencentyun/cos-go-sdk-v5"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"project/dao/mysql"
	"project/models"
	"project/utils"
	"strings"
	"sync"
)

const oss = "https://tiktok-1319971229.cos.ap-nanjing.myqcloud.com/"

// todo io优化，待测
func UploadIOVideo(file *multipart.FileHeader) (string, error) {
	// 生成 UUID
	fileId := strings.Replace(uuid.New().String(), "-", "", -1)

	// 修改文件名
	fileName := fileId + ".mp4"

	// 创建临时文件
	tempFile, err := createTempFile(fileName)
	if err != nil {
		return "", err
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// 创建缓冲写入器
	dest := bufio.NewWriter(tempFile)

	// 将上传的文件内容写入临时文件
	_, err = io.Copy(dest, src)
	if err != nil {
		return "", err
	}

	// 清空缓冲区并确保文件已写入磁盘
	if err = dest.Flush(); err != nil {
		return "", err
	}
	return fileName, nil
}

func createTempFile(fileName string) (*os.File, error) {
	tempDir := "./public/" // 临时文件夹路径

	// 创建临时文件夹（如果不存在）
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		err := os.Mkdir(tempDir, 0755)
		if err != nil {
			return nil, err
		}
	}

	// 在临时文件夹中创建临时文件
	tmpfile, err := os.Create(tempDir + fileName)
	if err != nil {
		return nil, err
	}

	return tmpfile, nil
}

// UploadToOSS  上传至腾讯OSS
func UploadToOSS(localPath string, remotePath string) error {
	reqUrl := viper.GetString("oss.tencent")
	u, _ := url.Parse(reqUrl)
	b := &cos.BaseURL{BucketURL: u}
	c := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:     viper.GetString("oss.SecretID"),
			SecretKey:    viper.GetString("oss.SecretKey"),
			SessionToken: "SECRETTOKEN",
		},
	})

	// 通过文件流上传对象
	fd, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer fd.Close()
	_, err = c.Object.Put(context.Background(), remotePath, fd, nil)
	if err != nil {
		return err
	}
	return nil
}

func GetVideoCover(fileName string) string {
	// 生成图片 UUID
	imgId := uuid.New().String()
	// 修改文件名
	imgName := strings.Replace(imgId, "-", "", -1) + ".jpg"
	//调用ffmpeg 获取封面图
	utils.GetVideoFrame("./public/"+fileName, "./public/"+imgName)
	return imgName
}

// todo
func StoreVideoAndImg(videoName string, imgName string, authorId uint, title string) {
	//视频存储到oss
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		if err := UploadToOSS("./public/"+videoName, videoName); err != nil {
			log.Fatal(err)
			return
		}
	}()

	// 图片存储到oss
	go func() {
		defer wg.Done()
		if err := UploadToOSS("./public/"+imgName, imgName); err != nil {
			log.Fatal(err)
			return
		}
	}()

	go func() {
		defer wg.Done()
		mysql.InsertVideo(videoName, imgName, authorId, title)
		//TODO redis用户上传的视频数+1
	}()
	wg.Wait()
}

func GetPublishList(userID uint) ([]models.VideoResponse, bool) {
	videos, found := mysql.FindVideosByAuthorId(userID)
	if !found {
		return []models.VideoResponse{}, false
	}
	// 将查询结果转换为VideoResponse类型
	var videoResponses []models.VideoResponse
	for _, video := range videos {
		user, _ := GetUserInfoByUserId(userID)
		commentCount, _ := GetCommentCount(video.VideoId)
		videoResponse := models.VideoResponse{
			Id:            video.VideoId,
			Author:        user,
			PlayUrl:       oss + video.VideoUrl,
			CoverUrl:      oss + video.CoverUrl,
			FavoriteCount: 0, // TODO
			CommentCount:  commentCount,
			IsFavorite:    isUserFavorite(111, video.VideoId), // TODO  userId,videoID
			Title:         video.Title,
		}
		videoResponses = append(videoResponses, videoResponse)
	}

	return videoResponses, true
}

func GetFeedList(latestTime string) ([]models.VideoResponse, int64, error) {
	videos := mysql.GetLatestVideos(latestTime)
	if len(videos) == 0 {
		return []models.VideoResponse{}, 0, errors.New("no videos")
	}
	// 将查询结果转换为VideoResponse类型
	var videoResponses []models.VideoResponse
	for _, video := range videos {
		user, _ := GetUserInfoByUserId(video.AuthorId)
		commentCount, _ := GetCommentCount(video.VideoId)
		videoResponse := models.VideoResponse{
			Id:            video.VideoId,
			Author:        user,
			PlayUrl:       oss + video.VideoUrl,
			CoverUrl:      oss + video.CoverUrl,
			FavoriteCount: 0, // TODO
			CommentCount:  commentCount,
			IsFavorite:    isUserFavorite(111, video.VideoId), // TODO  userId,videoID
			Title:         video.Title,
		}

		videoResponses = append(videoResponses, videoResponse)
	}
	//本次返回的视频中，发布最早的时间
	nextTime := videos[len(videos)-1].CreatedAt
	fmt.Println(nextTime)
	return videoResponses, nextTime, nil
}

func getFavoriteCount(uint) uint { return 1 }

func isUserFavorite(uint, uint) bool { return false }
