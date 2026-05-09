package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
)

/**
 * StaticServer 静态文件服务器
 * 为子应用提供 HTTP 静态文件服务，支持动态端口分配
 * 通过依赖注入方式使用，不再依赖全局变量
 */
type StaticServer struct {
	server *http.Server
	port   int
	dir    string
	mu     sync.Mutex
}

/**
 * 创建静态服务器实例
 */
func NewStaticServer() *StaticServer {
	return &StaticServer{}
}

/**
 * 获取静态服务器状态信息
 * 返回运行状态、监听端口和服务目录
 */
func (s *StaticServer) GetStatus() map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	return map[string]interface{}{
		"running": s.server != nil,
		"port":    s.port,
		"dir":     s.dir,
	}
}

/**
 * 启动静态文件服务器
 * 如果服务器已在运行且目录相同，直接返回当前端口
 * 如果目录不同，先关闭再重新启动
 */
func (s *StaticServer) Start(dir string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil {
		if s.dir == dir {
			return s.port, nil
		}
		s.server.Close()
		s.server = nil
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("分配端口失败: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port

	fs := http.FileServer(http.Dir(dir))
	mux := http.NewServeMux()
	mux.Handle("/", fs)

	s.server = &http.Server{Handler: mux}
	s.port = port
	s.dir = dir

	go func() {
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			s.mu.Lock()
			s.server = nil
			s.mu.Unlock()
		}
	}()

	return port, nil
}

/**
 * 停止静态文件服务器
 * 释放端口并重置状态
 */
func (s *StaticServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server == nil {
		return nil
	}

	err := s.server.Close()
	s.server = nil
	s.port = 0
	return err
}

/**
 * 重启静态文件服务器
 * 先停止再以新目录启动
 */
func (s *StaticServer) Restart(dir string) (int, error) {
	if err := s.Stop(); err != nil {
		return 0, err
	}
	return s.Start(dir)
}

/**
 * 检查服务器是否正在运行
 */
func (s *StaticServer) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.server != nil
}

/**
 * 获取当前监听端口
 * 如果服务器未运行，返回 0
 */
func (s *StaticServer) Port() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.port
}
