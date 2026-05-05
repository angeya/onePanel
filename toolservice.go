package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type PortInfo struct {
	Protocol    string `json:"protocol"`
	LocalAddr   string `json:"localAddr"`
	LocalPort   int    `json:"localPort"`
	ForeignAddr string `json:"foreignAddr"`
	ForeignPort int    `json:"foreignPort"`
	State       string `json:"state"`
	Pid         int    `json:"pid"`
	ProcessName string `json:"processName"`
}

type ToolService struct{}

func NewToolService() *ToolService {
	return &ToolService{}
}

/**
 * 获取所有网络端口与进程信息
 */
func (t *ToolService) GetPortList() ([]PortInfo, error) {
	output, err := exec.Command("netstat", "-ano").Output()
	if err != nil {
		return nil, fmt.Errorf("执行 netstat 失败: %w", err)
	}

	pidNameMap, err := getProcessNameMap()
	if err != nil {
		pidNameMap = make(map[int]string)
	}

	var result []PortInfo
	lines := strings.Split(string(output), "\n")

	portPattern := regexp.MustCompile(`^\s*(TCP|UDP)\s+(\S+):(\d+)\s+(\S+):(\d+)\s*(\S*)\s+(\d+)\s*$`)

	for _, line := range lines {
		matches := portPattern.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		protocol := matches[1]
		localPort, _ := strconv.Atoi(matches[3])
		foreignPort, _ := strconv.Atoi(matches[5])
		state := matches[6]
		pid, _ := strconv.Atoi(matches[7])

		if protocol == "UDP" {
			state = ""
		}

		processName := ""
		if name, ok := pidNameMap[pid]; ok {
			processName = name
		}

		result = append(result, PortInfo{
			Protocol:    protocol,
			LocalAddr:   matches[2],
			LocalPort:   localPort,
			ForeignAddr: matches[4],
			ForeignPort: foreignPort,
			State:       state,
			Pid:         pid,
			ProcessName: processName,
		})
	}

	if result == nil {
		result = []PortInfo{}
	}
	return result, nil
}

/**
 * 根据端口号查询进程信息
 */
func (t *ToolService) GetPortInfo(port int) ([]PortInfo, error) {
	all, err := t.GetPortList()
	if err != nil {
		return nil, err
	}

	var result []PortInfo
	for _, info := range all {
		if info.LocalPort == port || info.ForeignPort == port {
			result = append(result, info)
		}
	}

	if result == nil {
		result = []PortInfo{}
	}
	return result, nil
}

/**
 * 终止指定进程
 */
func (t *ToolService) KillProcess(pid int) error {
	err := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/F").Run()
	if err != nil {
		return fmt.Errorf("终止进程 %d 失败: %w", pid, err)
	}
	return nil
}

/**
 * 获取进程名称映射表
 */
func getProcessNameMap() (map[int]string, error) {
	output, err := exec.Command("tasklist", "/FO", "CSV", "/NH").Output()
	if err != nil {
		return nil, fmt.Errorf("执行 tasklist 失败: %w", err)
	}

	pidNameMap := make(map[int]string)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			continue
		}

		name := strings.Trim(parts[0], "\"")
		pidStr := strings.Trim(parts[1], "\"")

		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		pidNameMap[pid] = name
	}

	return pidNameMap, nil
}
