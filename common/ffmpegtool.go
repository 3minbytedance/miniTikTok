package common

import (
	"go.uber.org/zap"
	"os/exec"
)

// ffmpeg参数
const (
	inputVideoPathOption = "-i"
	startTimeOption      = "-ss"
	startTime            = "00:00:01" // 截取第1秒的帧
)

// GetVideoFrames ffmpeg 实现，现弃用，改为使用oss的功能
func GetVideoFrames(videoPath string, outputPath string) {
	if videoPath == "" || outputPath == "" {
		zap.L().Error("路径未指定")
		return
	}

	// 设置 ffmpeg 命令行参数，获取第1s的帧
	args := []string{inputVideoPathOption, videoPath, startTimeOption, startTime, "-vframes", "1", outputPath}

	// 创建 *exec.Cmd
	cmd := exec.Command("ffmpeg", args...)

	// 运行 ffmpeg 命令
	if err := cmd.Run(); err != nil {
		zap.L().Error("ffmpeg GetVideoFrames failed", zap.Error(err))
	}
}

// Transcoding 转为h264
func Transcoding(src string, dst string, overwrite bool) {
	args := []string{inputVideoPathOption, src, "-c:v", "libx264", "-strict", "-2", dst}
	if overwrite {
		args = append([]string{"-y"}, args...)
	}
	// 创建 *exec.Cmd
	cmd := exec.Command("ffmpeg", args...)

	// 运行 ffmpeg 命令
	if err := cmd.Run(); err != nil {
		zap.L().Error("ffmpeg出错", zap.Error(err))
		return
	}
}
