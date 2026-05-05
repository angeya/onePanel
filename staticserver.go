package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
)

type StaticServer struct {
	server *http.Server
	port   int
	dir    string
	mu     sync.Mutex
}

var staticServer *StaticServer

/**
 * 获取静态服务器状态信息
 */
func (s *StaticServer) GetStatus() map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	status := map[string]interface{}{
		"running": s.server != nil,
		"port":    s.port,
		"dir":     s.dir,
	}
	return status
}

/**
 * 启动静态文件服务器
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
 */
func (s *StaticServer) Restart(dir string) (int, error) {
	if err := s.Stop(); err != nil {
		return 0, err
	}
	return s.Start(dir)
}

func NewStaticServer() *StaticServer {
	staticServer = &StaticServer{}
	return staticServer
}
