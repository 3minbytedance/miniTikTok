package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"douyin/common"
	"douyin/config"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"
)

// TestMain 初始化配置后即可注册路由，测试不依赖下游 RPC。
func TestMain(m *testing.M) {
	if err := os.Setenv("DOUYIN_CONFIG", "../../config/app.yaml"); err != nil {
		panic(err)
	}
	if err := config.Init(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func newTestEngine() *server.Hertz {
	h := server.New(server.WithHostPorts("127.0.0.1:0"))
	register(h)
	return h
}

// /static/<file> 应直接映射到本地降级目录（回归：此前 h.Static 会重复拼接前缀）。
func TestStaticFileServed(t *testing.T) {
	dir := common.LocalFallbackDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("创建静态目录失败: %v", err)
	}
	content := "hello-douyin-static"
	if err := os.WriteFile(dir+"/ut-test.txt", []byte(content), 0o644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}
	defer os.Remove(dir + "/ut-test.txt")

	h := newTestEngine()
	resp := ut.PerformRequest(h.Engine, "GET", "/static/ut-test.txt", nil)
	if resp.Code != 200 {
		t.Fatalf("GET /static/ut-test.txt = %d, want 200; body: %s", resp.Code, resp.Body.String())
	}
	if got := resp.Body.String(); got != content {
		t.Errorf("静态内容 = %q, want %q", got, content)
	}
}

// 不存在的静态文件返回 404。
func TestStaticFileNotFound(t *testing.T) {
	h := newTestEngine()
	resp := ut.PerformRequest(h.Engine, "GET", "/static/not-exist-file.mp4", nil)
	if resp.Code != 404 {
		t.Errorf("GET 不存在的静态文件 = %d, want 404", resp.Code)
	}
}

// 路径穿越必须被拦截，不能读到仓库根目录文件（go.mod / 配置）。
func TestStaticFileTraversalBlocked(t *testing.T) {
	h := newTestEngine()
	for _, target := range []string{
		"/static/../go.mod",
		"/static/../../config/app.yaml",
		"/static/..%2F..%2Fgo.mod",
	} {
		resp := ut.PerformRequest(h.Engine, "GET", target, nil)
		if resp.Code == 200 && strings.Contains(resp.Body.String(), "module douyin") {
			t.Errorf("路径穿越未拦截 %q，泄漏内容", target)
		}
	}
}

// /metrics 应暴露 Prometheus 指标。
// promhttp 流式写入与 ut 假 writer 不兼容，改用真实监听端口验证。
func TestMetricsEndpoint(t *testing.T) {
	h := server.New(server.WithHostPorts("127.0.0.1:18099"))
	register(h)
	go h.Spin()

	addr := "127.0.0.1:18099"
	deadline := time.Now().Add(3 * time.Second)
	for {
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("测试服务器未启动: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	resp, err := http.Get("http://" + addr + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics 失败: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("GET /metrics = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("/metrics 响应为空")
	}

	_ = h.Shutdown(context.Background())
}
