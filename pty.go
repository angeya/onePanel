package main

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/UserExistsError/conpty"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type PtyService struct {
	ctx     context.Context
	running bool
	mu      sync.Mutex
	cpty    *conpty.ConPty
}

func NewPtyService() *PtyService {
	return &PtyService{}
}

func (p *PtyService) SetContext(ctx context.Context) {
	p.ctx = ctx
}

func (p *PtyService) Start(shell string, cols, rows int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return nil
	}

	if shell == "" {
		shell = "cmd.exe"
	}

	if cols <= 0 {
		cols = 120
	}
	if rows <= 0 {
		rows = 30
	}

	cpty, err := conpty.Start(shell, conpty.ConPtyDimensions(cols, rows))
	if err != nil {
		return fmt.Errorf("启动伪终端失败: %w", err)
	}

	p.cpty = cpty
	p.running = true

	go p.readOutput()

	return nil
}

func (p *PtyService) readOutput() {
	buf := make([]byte, 8192)

	for {
		p.mu.Lock()
		cpty := p.cpty
		p.mu.Unlock()

		if cpty == nil {
			break
		}

		n, err := cpty.Read(buf)
		if err != nil {
			if err != io.EOF {
				runtime.EventsEmit(p.ctx, "pty-output", fmt.Sprintf("\r\n\x1b[31m读取错误: %v\x1b[0m", err))
			}
			p.mu.Lock()
			p.running = false
			p.mu.Unlock()
			runtime.EventsEmit(p.ctx, "pty-output", "\r\n\x1b[33m进程已退出\x1b[0m")
			break
		}

		if n > 0 {
			output := string(buf[:n])
			runtime.EventsEmit(p.ctx, "pty-output", output)
		}
	}
}

func (p *PtyService) Write(data string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running || p.cpty == nil {
		return nil
	}

	_, err := p.cpty.Write([]byte(data))
	return err
}

func (p *PtyService) Resize(cols, rows int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cpty == nil {
		return nil
	}

	return p.cpty.Resize(cols, rows)
}

func (p *PtyService) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return nil
	}

	p.running = false

	if p.cpty != nil {
		p.cpty.Close()
		p.cpty = nil
	}

	return nil
}

func (p *PtyService) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}
