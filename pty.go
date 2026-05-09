package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/UserExistsError/conpty"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ptyInstance struct {
	id      string
	cpty    *conpty.ConPty
	running bool
	cancel  context.CancelFunc
}

type PtyService struct {
	ctx       context.Context
	instances map[string]*ptyInstance
	mu        sync.Mutex
	idCounter atomic.Int64
}

func NewPtyService() *PtyService {
	return &PtyService{
		instances: make(map[string]*ptyInstance),
	}
}

func (p *PtyService) SetContext(ctx context.Context) {
	p.ctx = ctx
}

/**
 * 启动一个新的伪终端实例
 */
func (p *PtyService) Start(req StartRequest) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	shell := req.Shell
	if shell == "" {
		shell = "cmd.exe"
	}
	shell = ResolveShellPath(shell)
	cols := req.Cols
	if cols <= 0 {
		cols = 120
	}
	rows := req.Rows
	if rows <= 0 {
		rows = 30
	}

	cpty, err := conpty.Start(shell, conpty.ConPtyDimensions(cols, rows))
	if err != nil {
		return "", fmt.Errorf("启动伪终端失败: %w", err)
	}

	id := fmt.Sprintf("pty-%d", p.idCounter.Add(1))
	ctx, cancel := context.WithCancel(context.Background())

	inst := &ptyInstance{
		id:      id,
		cpty:    cpty,
		running: true,
		cancel:  cancel,
	}
	p.instances[id] = inst

	go p.readOutput(inst, ctx)

	return id, nil
}

/**
 * 读取伪终端输出并通过事件发送到前端
 * 使用读 goroutine 与 flush goroutine 分离的架构：
 * 读 goroutine 负责持续读取 PTY 输出并写入缓冲区
 * flush goroutine 以约 60fps 的频率将缓冲区内容批量发送到前端
 * 这种设计既避免 EventsEmit 过于频繁导致 Wails WebView2 崩溃，也保证输出延迟可控
 */
func (p *PtyService) readOutput(inst *ptyInstance, ctx context.Context) {
	var outputBuf strings.Builder
	var bufMu sync.Mutex

	go func() {
		ticker := time.NewTicker(16 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				bufMu.Lock()
				if outputBuf.Len() > 0 {
					data := outputBuf.String()
					outputBuf.Reset()
					bufMu.Unlock()
					p.safeEventsEmit("pty-output-"+inst.id, data)
				} else {
					bufMu.Unlock()
				}
			}
		}
	}()

	readBuf := make([]byte, 8192)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		p.mu.Lock()
		cpty := inst.cpty
		p.mu.Unlock()

		if cpty == nil {
			break
		}

		n, err := cpty.Read(readBuf)
		if err != nil {
			if err != io.EOF {
				bufMu.Lock()
				if outputBuf.Len() > 0 {
					p.safeEventsEmit("pty-output-"+inst.id, outputBuf.String())
					outputBuf.Reset()
				}
				bufMu.Unlock()
				p.safeEventsEmit("pty-output-"+inst.id, fmt.Sprintf("\r\n\x1b[31m读取错误: %v\x1b[0m", err))
			}
			p.mu.Lock()
			inst.running = false
			p.mu.Unlock()
			p.safeEventsEmit("pty-exit-"+inst.id, nil)
			break
		}

		if n > 0 {
			bufMu.Lock()
			outputBuf.Write(readBuf[:n])
			if outputBuf.Len() >= 65536 {
				data := outputBuf.String()
				outputBuf.Reset()
				bufMu.Unlock()
				p.safeEventsEmit("pty-output-"+inst.id, data)
			} else {
				bufMu.Unlock()
			}
		}
	}
}

/**
 * 安全的事件发射封装
 * 捕获 EventsEmit 可能产生的 panic，防止 Wails WebView2 控件异常时导致程序崩溃
 */
func (p *PtyService) safeEventsEmit(event string, data ...interface{}) {
	defer func() {
		recover()
	}()
	if p.ctx == nil {
		return
	}
	runtime.EventsEmit(p.ctx, event, data...)
}

/**
 * 向指定伪终端写入数据
 */
func (p *PtyService) Write(id, data string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	inst, ok := p.instances[id]
	if !ok || !inst.running || inst.cpty == nil {
		return nil
	}

	_, err := inst.cpty.Write([]byte(data))
	return err
}

/**
 * 调整指定伪终端的窗口大小
 */
func (p *PtyService) Resize(id string, cols, rows int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	inst, ok := p.instances[id]
	if !ok || inst.cpty == nil {
		return nil
	}

	return inst.cpty.Resize(cols, rows)
}

/**
 * 停止指定伪终端实例
 */
func (p *PtyService) Stop(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	inst, ok := p.instances[id]
	if !ok {
		return nil
	}

	inst.running = false
	inst.cancel()

	if inst.cpty != nil {
		inst.cpty.Close()
		inst.cpty = nil
	}

	delete(p.instances, id)
	return nil
}

/**
 * 检查指定伪终端是否正在运行
 */
func (p *PtyService) IsRunning(id string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	inst, ok := p.instances[id]
	if !ok {
		return false
	}
	return inst.running
}

/**
 * 停止所有伪终端实例
 */
func (p *PtyService) StopAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for id, inst := range p.instances {
		inst.running = false
		inst.cancel()
		if inst.cpty != nil {
			inst.cpty.Close()
			inst.cpty = nil
		}
		delete(p.instances, id)
	}
}
