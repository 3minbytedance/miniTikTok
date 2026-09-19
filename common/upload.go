package common

import (
	"context"
	"douyin/config"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/tencentyun/cos-go-sdk-v5"
	"go.uber.org/zap"
)

// LocalFallbackDir 返回本地降级存储目录（OSS 未配置时使用）。
func LocalFallbackDir() string {
	dir := bizFileLocalPath
	if config.Conf.OSSConfig != nil && config.Conf.LocalFallbackPath != "" {
		dir = config.Conf.LocalFallbackPath
	}
	return dir
}

const bizFileLocalPath = "./tmp/"

// CreateDirectoryIfNotExist 确保本地存储目录存在。
func CreateDirectoryIfNotExist() error {
	dir := LocalFallbackDir()
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if mkErr := os.MkdirAll(dir, 0700); mkErr != nil {
				return mkErr
			}
			return nil
		}
		return err
	}
	return nil
}

// OSSConfigured 判断是否启用了 OSS（enabled 且 bucket/密钥配置完整）。
func OSSConfigured() bool {
	c := config.Conf.OSSConfig
	if c == nil || !c.Enabled {
		return false
	}
	return c.BucketURL != "" && c.SecretID != "" && c.SecretKey != ""
}

func getClient() *cos.Client {
	c := config.Conf.OSSConfig
	if c == nil {
		return nil
	}
	u, err := url.Parse(c.BucketURL)
	if err != nil {
		zap.L().Error("parse OSS bucket url error", zap.Error(err))
		return nil
	}
	base := &cos.BaseURL{BucketURL: u}
	if c.CIBucketURL != "" {
		cu, err := url.Parse(c.CIBucketURL)
		if err != nil {
			zap.L().Error("parse OSS ci url error", zap.Error(err))
			return nil
		}
		base.CIURL = cu
	}
	return cos.NewClient(base, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:     c.SecretID,
			SecretKey:    c.SecretKey,
			SessionToken: c.SessionToken,
		},
	})
}

// PersistFile 将本地文件持久化：OSS 已配置则上传到 COS，否则保留/拷贝到本地降级目录。
// object 为对象名（相对路径）。
func PersistFile(localPath, object string) error {
	if OSSConfigured() {
		c := getClient()
		if c == nil {
			return fmt.Errorf("oss enabled but client is nil, check oss config")
		}
		if _, err := c.Object.PutFromFile(context.Background(), object, localPath, nil); err != nil {
			return err
		}
		return nil
	}

	// 本地降级：拷贝到 fallback 目录（若源/目标一致则跳过）
	if err := CreateDirectoryIfNotExist(); err != nil {
		return err
	}
	dst := filepath.Join(LocalFallbackDir(), object)
	if absEqual(localPath, dst) {
		return nil
	}
	return copyFile(localPath, dst)
}

// ResolveURL 将对象名解析为可访问 URL：OSS 返回 COS URL，本地返回 /static 前缀路径。
func ResolveURL(object string) string {
	if object == "" {
		return ""
	}
	c := config.Conf.OSSConfig
	if OSSConfigured() {
		return strings.TrimRight(c.BucketURL, "/") + "/" + strings.TrimLeft(object, "/")
	}
	prefix := "/static"
	if c != nil && c.StaticURLPrefix != "" {
		prefix = c.StaticURLPrefix
	}
	return strings.TrimRight(prefix, "/") + "/" + strings.TrimLeft(object, "/")
}

// GenerateCover 生成视频封面。
// OSS 模式调用 COS 数据万象截帧（视频对象已上传）；本地模式用 ffmpeg 从本地视频截帧。
func GenerateCover(localVideoPath, videoObject string) (string, error) {
	imgName := strings.ReplaceAll(uuid.New().String(), "-", "") + ".jpg"
	if OSSConfigured() {
		if err := postSnapShot(videoObject, imgName); err != nil {
			return "", err
		}
		return imgName, nil
	}

	// 本地降级：ffmpeg 截帧到本地目录
	if err := CreateDirectoryIfNotExist(); err != nil {
		return "", err
	}
	outPath := filepath.Join(LocalFallbackDir(), imgName)
	args := []string{"-i", localVideoPath, "-ss", "00:00:01", "-vframes", "1", outPath}
	if err := exec.Command("ffmpeg", args...).Run(); err != nil {
		zap.L().Error("ffmpeg extract cover failed", zap.Error(err))
		return "", err
	}
	return imgName, nil
}

func postSnapShot(videoObject, imgName string) error {
	c := getClient()
	if c == nil {
		return fmt.Errorf("oss client is nil, check oss config")
	}
	cfg := config.Conf.OSSConfig
	opt := &cos.PostSnapshotOptions{
		Input:  &cos.JobInput{Object: videoObject},
		Time:   "1",
		Width:  720,
		Height: 1280,
		Format: "jpg",
		Output: &cos.JobOutput{
			Region: cfg.Region,
			Bucket: cfg.Bucket,
			Object: imgName,
		},
	}
	_, _, err := c.CI.PostSnapshot(context.Background(), opt)
	return err
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func absEqual(a, b string) bool {
	aa, err1 := filepath.Abs(a)
	bb, err2 := filepath.Abs(b)
	if err1 != nil || err2 != nil {
		return a == b
	}
	return aa == bb
}
