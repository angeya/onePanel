package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

/**
 * ServerListService 服务器列表管理服务
 * 负责 SSH 密钥对的自动生成、公钥部署、服务器会话和分类管理
 * 通过依赖注入持有 Database 引用
 */
type ServerListService struct {
	db *Database
}

/**
 * 创建 ServerListService 实例
 * 注入 Database 依赖
 */
func NewServerListService(db *Database) *ServerListService {
	return &ServerListService{db: db}
}

/**
 * 获取用户主目录下的 .ssh 目录路径
 */
func sshDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败: %w", err)
	}
	return filepath.Join(home, ".ssh"), nil
}

/**
 * 获取本地 SSH 密钥状态
 * 检查 ed25519 和 RSA 密钥是否存在，返回第一个找到的密钥信息
 */
func (s *ServerListService) GetSSHKeyStatus() (*SSHKeyStatus, error) {
	dir, err := sshDir()
	if err != nil {
		return nil, err
	}

	keyPairs := []struct {
		privateKey string
		publicKey  string
	}{
		{filepath.Join(dir, "id_ed25519"), filepath.Join(dir, "id_ed25519.pub")},
		{filepath.Join(dir, "id_rsa"), filepath.Join(dir, "id_rsa.pub")},
	}

	for _, pair := range keyPairs {
		_, err := os.Stat(pair.privateKey)
		if err == nil {
			pubKeyData, err := os.ReadFile(pair.publicKey)
			if err != nil {
				pubKeyData = []byte("")
			}
			return &SSHKeyStatus{
				KeyExists:  true,
				PublicKey:  string(pubKeyData),
				KeyPath:    pair.privateKey,
				PubKeyPath: pair.publicKey,
			}, nil
		}
	}

	return &SSHKeyStatus{
		KeyExists:  false,
		PublicKey:  "",
		KeyPath:    filepath.Join(dir, "id_ed25519"),
		PubKeyPath: filepath.Join(dir, "id_ed25519.pub"),
	}, nil
}

/**
 * 确保 SSH 密钥存在
 * 如果密钥不存在则自动生成，返回密钥状态
 * 用于部署公钥前的自动准备，用户无需手动操作
 */
func (s *ServerListService) ensureSSHKey() (*SSHKeyStatus, error) {
	status, err := s.GetSSHKeyStatus()
	if err != nil {
		return nil, err
	}
	if status.KeyExists {
		return status, nil
	}

	dir, err := sshDir()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("创建 .ssh 目录失败: %w", err)
	}

	pubKeyData, privKeyData, err := generateEd25519KeyPair()
	if err != nil {
		return nil, fmt.Errorf("生成密钥对失败: %w", err)
	}

	keyPath := filepath.Join(dir, "id_ed25519")
	pubPath := filepath.Join(dir, "id_ed25519.pub")

	if err := os.WriteFile(keyPath, privKeyData, 0600); err != nil {
		return nil, fmt.Errorf("写入私钥失败: %w", err)
	}

	if err := os.WriteFile(pubPath, pubKeyData, 0644); err != nil {
		os.Remove(keyPath)
		return nil, fmt.Errorf("写入公钥失败: %w", err)
	}

	return &SSHKeyStatus{
		KeyExists:  true,
		PublicKey:  string(pubKeyData),
		KeyPath:    keyPath,
		PubKeyPath: pubPath,
	}, nil
}

/**
 * 生成 ed25519 SSH 密钥对
 * 如果密钥已存在则返回错误，避免覆盖
 * 生成的私钥权限设置为 0600，公钥权限设置为 0644
 */
func (s *ServerListService) GenerateSSHKey() (*SSHKeyStatus, error) {
	status, err := s.GetSSHKeyStatus()
	if err != nil {
		return nil, err
	}
	if status.KeyExists {
		return nil, fmt.Errorf("SSH 密钥已存在: %s", status.KeyPath)
	}

	result, err := s.ensureSSHKey()
	if err != nil {
		return nil, err
	}
	return result, nil
}

/**
 * 使用 crypto/ed25519 生成 ed25519 SSH 密钥对
 * 返回 OpenSSH 格式的公钥和 PKCS8 格式的私钥
 */
func generateEd25519KeyPair() (publicKeyOpenSSH []byte, privateKeyPEM []byte, err error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("生成 ed25519 密钥失败: %w", err)
	}

	sshPubKey, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		return nil, nil, fmt.Errorf("转换公钥格式失败: %w", err)
	}
	pubKeyData := ssh.MarshalAuthorizedKey(sshPubKey)

	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return nil, nil, fmt.Errorf("序列化私钥失败: %w", err)
	}

	privKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKeyBytes,
	})

	return pubKeyData, privKeyPEM, nil
}

/**
 * 将公钥部署到远程服务器
 * 使用密码认证连接远程服务器，将公钥追加到 authorized_keys 文件
 * 部署成功后更新数据库中的 key_deployed 状态
 * 如果本地密钥不存在，会自动先生成
 */
func (s *ServerListService) DeployKey(serverId int64, password string) error {
	status, err := s.ensureSSHKey()
	if err != nil {
		return err
	}

	server, err := s.GetServer(serverId)
	if err != nil {
		return fmt.Errorf("获取服务器信息失败: %w", err)
	}

	config := &ssh.ClientConfig{
		User: server.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", server.Host, server.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("连接服务器失败: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建 SSH 会话失败: %w", err)
	}
	defer session.Close()

	pubKey := status.PublicKey
	deployCmd := fmt.Sprintf(
		"mkdir -p ~/.ssh && chmod 700 ~/.ssh && "+
			"echo '%s' >> ~/.ssh/authorized_keys && "+
			"chmod 600 ~/.ssh/authorized_keys && "+
			"sort -u ~/.ssh/authorized_keys -o ~/.ssh/authorized_keys",
		pubKey,
	)

	if err := session.Run(deployCmd); err != nil {
		return fmt.Errorf("部署公钥失败: %w", err)
	}

	now := NowFormatted()
	_, err = s.db.DB().Exec(
		"UPDATE server_session SET key_deployed = 1, updated_at = ? WHERE id = ?",
		now, serverId,
	)
	if err != nil {
		return fmt.Errorf("更新服务器状态失败: %w", err)
	}

	return nil
}

/**
 * 获取 SSH 登录命令
 * 根据服务器配置返回对应的 ssh 命令字符串
 * 如果密钥已部署，使用 ssh 命令即可自动密钥认证
 */
func (s *ServerListService) GetLoginCommand(serverId int64) (string, error) {
	server, err := s.GetServer(serverId)
	if err != nil {
		return "", fmt.Errorf("获取服务器信息失败: %w", err)
	}

	var cmd string
	if server.Port == 22 {
		cmd = fmt.Sprintf("ssh %s@%s", server.User, server.Host)
	} else {
		cmd = fmt.Sprintf("ssh -p %d %s@%s", server.Port, server.User, server.Host)
	}
	return cmd, nil
}

/**
 * 测试 SSH 密钥连接是否可用
 * 使用本地私钥尝试连接远程服务器，验证密钥登录是否正常
 */
func (s *ServerListService) TestKeyConnection(serverId int64) (string, error) {
	status, err := s.GetSSHKeyStatus()
	if err != nil {
		return "", err
	}
	if !status.KeyExists {
		return "", fmt.Errorf("本地 SSH 密钥不存在")
	}

	server, err := s.GetServer(serverId)
	if err != nil {
		return "", fmt.Errorf("获取服务器信息失败: %w", err)
	}

	keyBytes, err := os.ReadFile(status.KeyPath)
	if err != nil {
		return "", fmt.Errorf("读取私钥失败: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return "", fmt.Errorf("解析私钥失败: %w", err)
	}

	config := &ssh.ClientConfig{
		User: server.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", server.Host, server.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return "", fmt.Errorf("密钥连接失败: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput("echo 'SSH密钥登录成功' && hostname && whoami")
	if err != nil {
		return "", fmt.Errorf("执行命令失败: %w", err)
	}

	return string(output), nil
}

/**
 * 检测指定主机和端口是否可达
 * 用于在添加服务器前验证网络连通性
 */
func (s *ServerListService) TestConnectivity(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("无法连接到 %s: %w", addr, err)
	}
	conn.Close()
	return nil
}

/**
 * 获取所有服务器分类
 * 按排序字段和 ID 升序排列
 */
func (s *ServerListService) GetSessionCategories() ([]ServerCategory, error) {
	rows, err := s.db.DB().Query(
		"SELECT id, name, sort_order, created_at, updated_at " +
			"FROM server_category ORDER BY sort_order, id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []ServerCategory
	for rows.Next() {
		var c ServerCategory
		if err := rows.Scan(&c.Id, &c.Name, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	if categories == nil {
		categories = []ServerCategory{}
	}
	return categories, nil
}

/**
 * 创建服务器分类
 */
func (s *ServerListService) CreateSessionCategory(name string, sortOrder int) (*ServerCategory, error) {
	if name == "" {
		return nil, fmt.Errorf("分类名称不能为空")
	}

	now := NowFormatted()
	result, err := s.db.DB().Exec(
		"INSERT INTO server_category (name, sort_order, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?)",
		name, sortOrder, now, now,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &ServerCategory{
		Id:        id,
		Name:      name,
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

/**
 * 删除服务器分类
 */
func (s *ServerListService) DeleteSessionCategory(id int64) error {
	_, err := s.db.DB().Exec("DELETE FROM server_category WHERE id = ?", id)
	return err
}

/**
 * 获取所有服务器会话列表
 */
func (s *ServerListService) GetServers() ([]ServerSession, error) {
	rows, err := s.db.DB().Query(
		"SELECT id, category_id, session_name, host, port, user, use_key_login, "+
			"key_deployed, created_at, updated_at FROM server_session ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []ServerSession
	for rows.Next() {
		var svr ServerSession
		var categoryId sql.NullInt64
		var useKeyLogin int
		var keyDeployed int
		if err := rows.Scan(
			&svr.Id, &categoryId, &svr.SessionName, &svr.Host, &svr.Port,
			&svr.User, &useKeyLogin, &keyDeployed, &svr.CreatedAt, &svr.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if categoryId.Valid {
			svr.CategoryId = &categoryId.Int64
		}
		svr.UseKeyLogin = useKeyLogin == 1
		svr.KeyDeployed = keyDeployed == 1
		servers = append(servers, svr)
	}
	if servers == nil {
		servers = []ServerSession{}
	}
	return servers, nil
}

/**
 * 根据 ID 获取单个服务器会话
 */
func (s *ServerListService) GetServer(id int64) (*ServerSession, error) {
	var svr ServerSession
	var categoryId sql.NullInt64
	var useKeyLogin int
	var keyDeployed int
	err := s.db.DB().QueryRow(
		"SELECT id, category_id, session_name, host, port, user, use_key_login, "+
			"key_deployed, created_at, updated_at FROM server_session WHERE id = ?",
		id,
	).Scan(
		&svr.Id, &categoryId, &svr.SessionName, &svr.Host, &svr.Port,
		&svr.User, &useKeyLogin, &keyDeployed, &svr.CreatedAt, &svr.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if categoryId.Valid {
		svr.CategoryId = &categoryId.Int64
	}
	svr.UseKeyLogin = useKeyLogin == 1
	svr.KeyDeployed = keyDeployed == 1
	return &svr, nil
}

/**
 * 添加服务器会话记录
 * sessionName 为空时使用 host@user 格式
 * 如果同主机同用户已存在，则返回已有记录
 */
func (s *ServerListService) AddServer(
	categoryId *int64, sessionName, host string, port int, user string, useKeyLogin bool,
) (*ServerSession, error) {
	if host == "" {
		return nil, fmt.Errorf("主机地址不能为空")
	}
	if user == "" {
		return nil, fmt.Errorf("用户名不能为空")
	}
	if port <= 0 {
		port = 22
	}
	if sessionName == "" {
		sessionName = fmt.Sprintf("%s@%s", user, host)
	}

	var existingId int64
	err := s.db.DB().QueryRow(
		"SELECT id FROM server_session WHERE host = ? AND user = ?",
		host, user,
	).Scan(&existingId)
	if err == nil {
		return s.GetServer(existingId)
	}

	now := NowFormatted()
	var catId sql.NullInt64
	if categoryId != nil {
		catId = sql.NullInt64{Int64: *categoryId, Valid: true}
	}

	useKey := 0
	if useKeyLogin {
		useKey = 1
	}

	result, err := s.db.DB().Exec(
		"INSERT INTO server_session (category_id, session_name, host, port, user, "+
			"use_key_login, key_deployed, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?)",
		catId, sessionName, host, port, user, useKey, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("添加服务器失败: %w", err)
	}

	id, _ := result.LastInsertId()
	return &ServerSession{
		Id:          id,
		CategoryId:  categoryId,
		SessionName: sessionName,
		Host:        host,
		Port:        port,
		User:        user,
		UseKeyLogin: useKeyLogin,
		KeyDeployed: false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

/**
 * 更新服务器会话信息
 */
func (s *ServerListService) UpdateServer(
	id int64, categoryId *int64, sessionName, host string, port int, user string, useKeyLogin bool,
) error {
	if host == "" {
		return fmt.Errorf("主机地址不能为空")
	}
	if user == "" {
		return fmt.Errorf("用户名不能为空")
	}
	if port <= 0 {
		port = 22
	}
	if sessionName == "" {
		sessionName = fmt.Sprintf("%s@%s", user, host)
	}

	now := NowFormatted()
	var catId sql.NullInt64
	if categoryId != nil {
		catId = sql.NullInt64{Int64: *categoryId, Valid: true}
	}

	useKey := 0
	if useKeyLogin {
		useKey = 1
	}

	_, err := s.db.DB().Exec(
		"UPDATE server_session SET category_id = ?, session_name = ?, host = ?, "+
			"port = ?, user = ?, use_key_login = ?, updated_at = ? WHERE id = ?",
		catId, sessionName, host, port, user, useKey, now, id,
	)
	return err
}

/**
 * 重命名服务器会话
 */
func (s *ServerListService) RenameServer(id int64, sessionName string) error {
	if sessionName == "" {
		return fmt.Errorf("会话名称不能为空")
	}

	now := NowFormatted()
	_, err := s.db.DB().Exec(
		"UPDATE server_session SET session_name = ?, updated_at = ? WHERE id = ?",
		sessionName, now, id,
	)
	return err
}

/**
 * 删除服务器会话记录
 */
func (s *ServerListService) DeleteServer(id int64) error {
	_, err := s.db.DB().Exec("DELETE FROM server_session WHERE id = ?", id)
	return err
}
