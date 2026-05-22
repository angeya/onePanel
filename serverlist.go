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

const defaultSSHPort = 22

/**
 * ServerListService 服务器列表管理服务。
 * 负责 SSH 密钥对的自动生成、公钥部署、连通性测试以及服务器会话和分类管理。
 */
type ServerListService struct {
	db *Database
}

/**
 * 创建 ServerListService 实例。
 */
func NewServerListService(db *Database) *ServerListService {
	return &ServerListService{db: db}
}

/**
 * sshDir 返回本机用户目录下的 .ssh 目录。
 */
func sshDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败: %w", err)
	}
	return filepath.Join(home, ".ssh"), nil
}

/**
 * GetSSHKeyStatus 获取本地 SSH 密钥状态。
 * 按 ed25519、RSA 的顺序查找第一个可用密钥对。
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
 * ensureSSHKey 确保本地存在可用的 SSH 密钥。
 * 如果不存在，则自动生成默认 ed25519 密钥对。
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
		_ = os.Remove(keyPath)
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
 * GenerateSSHKey 手动生成 SSH 密钥对。
 * 如果密钥已存在则直接返回错误，避免覆盖现有密钥。
 */
func (s *ServerListService) GenerateSSHKey() (*SSHKeyStatus, error) {
	status, err := s.GetSSHKeyStatus()
	if err != nil {
		return nil, err
	}
	if status.KeyExists {
		return nil, fmt.Errorf("SSH 密钥已存在: %s", status.KeyPath)
	}
	return s.ensureSSHKey()
}

/**
 * generateEd25519KeyPair 生成 ed25519 SSH 密钥对。
 * 返回 OpenSSH 格式公钥与 PKCS8 PEM 私钥。
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

	privateKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKeyBytes,
	})
	return pubKeyData, privateKeyPEM, nil
}

/**
 * DeployKey 将公钥部署到远程服务器。
 */
func (s *ServerListService) DeployKey(serverID int64, password string) error {
	status, err := s.ensureSSHKey()
	if err != nil {
		return err
	}

	server, err := s.GetServer(serverID)
	if err != nil {
		return fmt.Errorf("获取服务器信息失败: %w", err)
	}

	client, err := s.dialServerWithPassword(server, password, 15*time.Second)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建 SSH 会话失败: %w", err)
	}
	defer session.Close()

	deployCmd := fmt.Sprintf(
		"mkdir -p ~/.ssh && chmod 700 ~/.ssh && "+
			"echo '%s' >> ~/.ssh/authorized_keys && "+
			"chmod 600 ~/.ssh/authorized_keys && "+
			"sort -u ~/.ssh/authorized_keys -o ~/.ssh/authorized_keys",
		status.PublicKey,
	)
	if err := session.Run(deployCmd); err != nil {
		return fmt.Errorf("部署公钥失败: %w", err)
	}

	return s.markServerKeyDeployed(serverID, true)
}

/**
 * GetLoginCommand 根据服务器配置构建 SSH 登录命令。
 */
func (s *ServerListService) GetLoginCommand(serverID int64) (string, error) {
	server, err := s.GetServer(serverID)
	if err != nil {
		return "", fmt.Errorf("获取服务器信息失败: %w", err)
	}
	return s.buildLoginCommand(server), nil
}

/**
 * TestKeyConnection 测试密钥登录是否可用。
 */
func (s *ServerListService) TestKeyConnection(serverID int64) (string, error) {
	status, err := s.GetSSHKeyStatus()
	if err != nil {
		return "", err
	}
	if !status.KeyExists {
		return "", fmt.Errorf("本地 SSH 密钥不存在")
	}

	server, err := s.GetServer(serverID)
	if err != nil {
		return "", fmt.Errorf("获取服务器信息失败: %w", err)
	}

	keyBytes, err := os.ReadFile(status.KeyPath)
	if err != nil {
		return "", fmt.Errorf("读取私钥失败: %w", err)
	}
	client, err := s.dialServerWithPrivateKey(server, keyBytes, 10*time.Second)
	if err != nil {
		return "", err
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
 * TestConnectivity 检测指定主机和端口是否可达。
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
 * GetSessionCategories 获取所有服务器分类。
 */
func (s *ServerListService) GetSessionCategories() ([]ServerCategory, error) {
	rows, err := s.db.DB().Query(
		"SELECT id, name, sort_order, created_at, updated_at FROM server_category ORDER BY sort_order, id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []ServerCategory
	for rows.Next() {
		var category ServerCategory
		if err := rows.Scan(
			&category.Id,
			&category.Name,
			&category.SortOrder,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	if categories == nil {
		categories = []ServerCategory{}
	}
	return categories, nil
}

/**
 * CreateSessionCategory 创建服务器分类。
 */
func (s *ServerListService) CreateSessionCategory(name string, sortOrder int) (*ServerCategory, error) {
	if name == "" {
		return nil, fmt.Errorf("分类名称不能为空")
	}

	now := NowFormatted()
	result, err := s.db.DB().Exec(
		"INSERT INTO server_category (name, sort_order, created_at, updated_at) VALUES (?, ?, ?, ?)",
		name,
		sortOrder,
		now,
		now,
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
 * DeleteSessionCategory 删除服务器分类。
 */
func (s *ServerListService) DeleteSessionCategory(id int64) error {
	_, err := s.db.DB().Exec("DELETE FROM server_category WHERE id = ?", id)
	return err
}

/**
 * GetServers 获取所有服务器会话。
 */
func (s *ServerListService) GetServers() ([]ServerSession, error) {
	rows, err := s.db.DB().Query(
		"SELECT id, category_id, session_name, host, port, user, use_key_login, key_deployed, created_at, updated_at FROM server_session ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []ServerSession
	for rows.Next() {
		server, err := scanServerSession(rows.Scan)
		if err != nil {
			return nil, err
		}
		servers = append(servers, *server)
	}
	if servers == nil {
		servers = []ServerSession{}
	}
	return servers, nil
}

/**
 * GetServer 根据 ID 获取单个服务器会话。
 */
func (s *ServerListService) GetServer(id int64) (*ServerSession, error) {
	row := s.db.DB().QueryRow(
		"SELECT id, category_id, session_name, host, port, user, use_key_login, key_deployed, created_at, updated_at FROM server_session WHERE id = ?",
		id,
	)
	return scanServerSession(row.Scan)
}

/**
 * AddServer 添加服务器会话。
 * 如果同主机同用户已存在，则直接返回已有记录。
 */
func (s *ServerListService) AddServer(
	categoryID *int64,
	sessionName, host string,
	port int,
	user string,
	useKeyLogin bool,
) (*ServerSession, error) {
	if err := validateServerInput(host, user, &port); err != nil {
		return nil, err
	}

	var existingID int64
	err := s.db.DB().QueryRow(
		"SELECT id FROM server_session WHERE host = ? AND user = ?",
		host,
		user,
	).Scan(&existingID)
	if err == nil {
		return s.GetServer(existingID)
	}

	now := NowFormatted()
	result, err := s.db.DB().Exec(
		"INSERT INTO server_session (category_id, session_name, host, port, user, use_key_login, key_deployed, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?)",
		toNullInt64(categoryID),
		sessionName,
		host,
		port,
		user,
		boolToInt(useKeyLogin),
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("添加服务器失败: %w", err)
	}

	id, _ := result.LastInsertId()
	return &ServerSession{
		Id:          id,
		CategoryId:  categoryID,
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
 * UpdateServer 更新服务器会话信息。
 */
func (s *ServerListService) UpdateServer(
	id int64,
	categoryID *int64,
	sessionName, host string,
	port int,
	user string,
	useKeyLogin bool,
) error {
	if err := validateServerInput(host, user, &port); err != nil {
		return err
	}

	now := NowFormatted()
	_, err := s.db.DB().Exec(
		"UPDATE server_session SET category_id = ?, session_name = ?, host = ?, port = ?, user = ?, use_key_login = ?, updated_at = ? WHERE id = ?",
		toNullInt64(categoryID),
		sessionName,
		host,
		port,
		user,
		boolToInt(useKeyLogin),
		now,
		id,
	)
	return err
}

/**
 * RenameServer 重命名服务器会话。
 */
func (s *ServerListService) RenameServer(id int64, sessionName string) error {
	if sessionName == "" {
		return fmt.Errorf("会话名称不能为空")
	}

	now := NowFormatted()
	_, err := s.db.DB().Exec(
		"UPDATE server_session SET session_name = ?, updated_at = ? WHERE id = ?",
		sessionName,
		now,
		id,
	)
	return err
}

/**
 * DeleteServer 删除服务器会话。
 */
func (s *ServerListService) DeleteServer(id int64) error {
	_, err := s.db.DB().Exec("DELETE FROM server_session WHERE id = ?", id)
	return err
}

/**
 * buildLoginCommand 根据端口构造 SSH 登录命令。
 */
func (s *ServerListService) buildLoginCommand(server *ServerSession) string {
	if server.Port == defaultSSHPort {
		return fmt.Sprintf("ssh %s@%s", server.User, server.Host)
	}
	return fmt.Sprintf("ssh -p %d %s@%s", server.Port, server.User, server.Host)
}

/**
 * markServerKeyDeployed 更新服务器是否已部署密钥的标记。
 */
func (s *ServerListService) markServerKeyDeployed(serverID int64, deployed bool) error {
	now := NowFormatted()
	_, err := s.db.DB().Exec(
		"UPDATE server_session SET key_deployed = ?, updated_at = ? WHERE id = ?",
		boolToInt(deployed),
		now,
		serverID,
	)
	if err != nil {
		return fmt.Errorf("更新服务器状态失败: %w", err)
	}
	return nil
}

/**
 * dialServerWithPassword 使用密码认证连接 SSH。
 */
func (s *ServerListService) dialServerWithPassword(server *ServerSession, password string, timeout time.Duration) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: server.User,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", server.Host, server.Port), config)
	if err != nil {
		return nil, fmt.Errorf("连接服务器失败: %w", err)
	}
	return client, nil
}

/**
 * dialServerWithPrivateKey 使用私钥认证连接 SSH。
 */
func (s *ServerListService) dialServerWithPrivateKey(server *ServerSession, keyBytes []byte, timeout time.Duration) (*ssh.Client, error) {
	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %w", err)
	}

	config := &ssh.ClientConfig{
		User: server.User,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", server.Host, server.Port), config)
	if err != nil {
		return nil, fmt.Errorf("密钥连接失败: %w", err)
	}
	return client, nil
}

/**
 * scanServerSession 统一完成 server_session 记录到模型的扫描与转换。
 */
func scanServerSession(scan func(dest ...interface{}) error) (*ServerSession, error) {
	var server ServerSession
	var categoryID sql.NullInt64
	var useKeyLogin int
	var keyDeployed int

	err := scan(
		&server.Id,
		&categoryID,
		&server.SessionName,
		&server.Host,
		&server.Port,
		&server.User,
		&useKeyLogin,
		&keyDeployed,
		&server.CreatedAt,
		&server.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if categoryID.Valid {
		server.CategoryId = &categoryID.Int64
	}
	server.UseKeyLogin = useKeyLogin == 1
	server.KeyDeployed = keyDeployed == 1
	return &server, nil
}

/**
 * validateServerInput 校验服务器会话的核心字段。
 */
func validateServerInput(host string, user string, port *int) error {
	if host == "" {
		return fmt.Errorf("主机地址不能为空")
	}
	if user == "" {
		return fmt.Errorf("用户名不能为空")
	}
	if *port <= 0 {
		*port = defaultSSHPort
	}
	return nil
}

/**
 * toNullInt64 将可选分类 ID 转为数据库可识别的 NullInt64。
 */
func toNullInt64(value *int64) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *value, Valid: true}
}

/**
 * boolToInt 将布尔值转换为数据库中的整数标记。
 */
func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
