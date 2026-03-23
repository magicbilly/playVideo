package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// 创建一个临时的、只有15秒钟的伪造视频文件用于测试
func createTestVideo(path string) error {
	// 将 d=1 改为 d=15，确保视频长度超过 10 秒
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi",
		"-i", "color=c=blue:s=128x72:d=15",
		"-pix_fmt", "yuv420p", path)
	return cmd.Run()
}

func TestVideoProcessor(t *testing.T) {
	// 1. 设置临时测试目录
	testDir, _ := os.MkdirTemp("", "video_test")
	posterDir := filepath.Join(testDir, "poster")
	os.MkdirAll(posterDir, 0755)

	defer os.RemoveAll(testDir) // 测试完自动清理

	testTitle := "test_movie"
	testFilehash := "abc123hash"
	videoFileName := testTitle + ".mkv"
	videoFullPath := filepath.Join(testDir, videoFileName)

	// 2. 生成一个真实的测试视频文件
	err := createTestVideo(videoFullPath)
	if err != nil {
		t.Fatalf("无法创建测试视频文件: %v", err)
	}

	// 3. 初始化处理器
	var vp VideoProcessor
	vp.init(testDir, posterDir, testTitle, testFilehash)

	// 4. 测试封面生成
	t.Run("GeneratePoster", func(t *testing.T) {
		err := vp.generatePoster()
		if err != nil {
			t.Errorf("生成封面失败: %v", err)
		}
		// 检查封面文件是否存在
		posterFile := filepath.Join(posterDir, testFilehash+".png")
		if _, err := os.Stat(posterFile); os.IsNotExist(err) {
			t.Errorf("封面文件未生成: %s", posterFile)
		}
	})

	// 5. 测试视频切片 (HLS)
	t.Run("MkdirVideo", func(t *testing.T) {
		err := vp.mkdirVideo()
		if err != nil {
			t.Errorf("视频切片失败: %v", err)
		}
		// 检查 index.m3u8 是否存在
		m3u8File := filepath.Join(testDir, "m3u8", testFilehash, "index.m3u8")
		if _, err := os.Stat(m3u8File); os.IsNotExist(err) {
			t.Errorf("M3U8文件未生成: %s", m3u8File)
		}
	})
}
