package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/evilCYH/NodeHub/internal/models/config"
	"github.com/evilCYH/NodeHub/internal/utils"
)

var (
	baseConfig = config.DefaultBase()
	configMu   sync.RWMutex
	loaded     bool
)

func defaultConfigPath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取可执行文件路径失败: %v", err)
	}
	execDir := filepath.Dir(execPath)
	return filepath.Join(execDir, "config.json"), nil
}

func resolveConfigPath(configPath string) (string, error) {
	if configPath == "" {
		defaultPath, err := defaultConfigPath()
		if err != nil {
			return "", err
		}
		configPath = defaultPath
	}
	if !filepath.IsAbs(configPath) {
		absPath, err := filepath.Abs(configPath)
		if err != nil {
			return "", fmt.Errorf("无法转换为绝对路径: %v", err)
		}
		configPath = absPath
	}
	return configPath, nil
}

func Load(configPath string) error {
	resolvedPath, err := resolveConfigPath(configPath)
	if err != nil {
		return err
	}

	cfg := config.DefaultBase()
	if err := loadFromFile(&cfg, resolvedPath); err != nil {
		if os.IsNotExist(err) {
			if err := createDefaultConfig(resolvedPath, &cfg); err != nil {
				return fmt.Errorf("创建默认配置文件失败: %v", err)
			}
			if err := loadFromFile(&cfg, resolvedPath); err != nil {
				return fmt.Errorf("加载默认配置文件失败: %v", err)
			}
		} else {
			return fmt.Errorf("加载配置文件失败: %v", err)
		}
	}

	setupPaths(&cfg, resolvedPath)

	loadFromEnv(&cfg)

	if err := validateConfig(&cfg); err != nil {
		return fmt.Errorf("配置验证失败: %v", err)
	}

	configMu.Lock()
	baseConfig = cfg
	loaded = true
	configMu.Unlock()

	return nil
}

func MustLoad(configPath string) {
	if err := Load(configPath); err != nil {
		panic(err)
	}
}

func Base() config.Base {
	configMu.RLock()
	defer configMu.RUnlock()
	if !loaded {
		panic("config not loaded: call config.Load or config.MustLoad first")
	}
	return baseConfig
}

func setupPaths(config *config.Base, configPath string) {
	configDir := filepath.Dir(configPath)

	if config.Server.UIPath == "" {
		config.Server.UIPath = filepath.Join(configDir, "ui")
	}

	if config.Database.Path == "" {
		config.Database.Path = filepath.Join(configDir, "data", "nodehub.db")
	}

	if config.Log.Path == "" {
		config.Log.Path = filepath.Join(configDir, "log")
	}

	if config.Session.NodePath == "" {
		config.Session.NodePath = filepath.Join(configDir, "session", "node.session")
	}

}

func loadFromFile(config *config.Base, filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return err
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}
	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	return nil
}

func loadFromEnv(config *config.Base) {
	if port := os.Getenv("NODEHUB_SERVER_PORT"); port != "" {
		if p, err := parsePort(port); err == nil {
			config.Server.Port = p
		}
	}
	if host := os.Getenv("NODEHUB_SERVER_HOST"); host != "" {
		config.Server.Host = host
	}
	if dbPath := os.Getenv("NODEHUB_DATABASE_PATH"); dbPath != "" {
		config.Database.Path = dbPath
	}
	if dbType := os.Getenv("NODEHUB_DATABASE_TYPE"); dbType != "" {
		config.Database.Type = dbType
	}
	if logLevel := os.Getenv("NODEHUB_LOG_LEVEL"); logLevel != "" {
		config.Log.Level = logLevel
	}
	if logOutput := os.Getenv("NODEHUB_LOG_OUTPUT"); logOutput != "" {
		config.Log.Output = logOutput
	}
	if logDir := os.Getenv("NODEHUB_LOG_DIR"); logDir != "" {
		config.Log.Path = logDir
	}
	if jwtSecret := os.Getenv("NODEHUB_JWT_SECRET"); jwtSecret != "" {
		config.JWT.Secret = jwtSecret
	}
}

func parsePort(portStr string) (int, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("无效的端口号: %s", portStr)
	}
	if port <= 0 || port > 65535 {
		return 0, fmt.Errorf("端口号超出范围: %d", port)
	}
	return port, nil
}

func createDefaultConfig(filePath string, cfg *config.Base) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Errorf("生成JWT密钥失败: %v", err)
	}
	cfg.JWT.Secret = hex.EncodeToString(bytes)

	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return fmt.Errorf("序列化默认配置失败: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

func validateConfig(config *config.Base) error {
	if err := validateServerConfig(&config.Server); err != nil {
		return fmt.Errorf("服务器配置验证失败: %v", err)
	}

	if err := validateDatabaseConfig(&config.Database); err != nil {
		return fmt.Errorf("数据库配置验证失败: %v", err)
	}

	if err := validateLogConfig(&config.Log); err != nil {
		return fmt.Errorf("日志配置验证失败: %v", err)
	}

	if err := validateJWTConfig(&config.JWT); err != nil {
		return fmt.Errorf("JWT配置验证失败: %v", err)
	}

	if err := validateSessionConfig(&config.Session); err != nil {
		return fmt.Errorf("会话配置验证失败: %v", err)
	}

	return nil
}

func validateServerConfig(config *config.ServerConfig) error {
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("端口号必须在1-65535范围内，当前值: %d", config.Port)
	}

	if config.Host == "" {
		return fmt.Errorf("主机地址不能为空")
	}

	if ip := net.ParseIP(config.Host); ip == nil {
		return fmt.Errorf("无效的主机地址格式: %s", config.Host)
	}

	return nil
}

func validateDatabaseConfig(config *config.DatabaseConfig) error {
	if config.Type == "" {
		return fmt.Errorf("数据库类型不能为空")
	}
	if config.Path == "" {
		return fmt.Errorf("数据库路径不能为空")
	}
	dir := filepath.Dir(config.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("无法创建数据库目录 %s: %v", dir, err)
	}
	if !utils.IsWritableDir(dir) {
		return fmt.Errorf("数据库目录 %s 不可写", dir)
	}
	return nil
}

func validateLogConfig(config *config.LogConfig) error {
	validLevels := []string{"debug", "info", "warn", "error"}
	if !utils.Contains(validLevels, strings.ToLower(config.Level)) {
		return fmt.Errorf("无效的日志等级: %s，支持的等级: %v", config.Level, validLevels)
	}

	validOutputs := []string{"console", "file", "both"}
	if !utils.Contains(validOutputs, strings.ToLower(config.Output)) {
		return fmt.Errorf("无效的日志输出方式: %s，支持的方式: %v", config.Output, validOutputs)
	}

	if config.Output == "file" || config.Output == "both" {
		if config.Path == "" {
			return fmt.Errorf("日志输出到文件时，文件路径不能为空")
		}

		dir := config.Path
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("无法创建日志目录 %s: %v", dir, err)
		}

		if !utils.IsWritableDir(dir) {
			return fmt.Errorf("日志目录 %s 不可写", dir)
		}
	}

	return nil
}

func validateJWTConfig(config *config.JWTConfig) error {
	if config.Secret == "" {
		return fmt.Errorf("JWT密钥不能为空")
	}

	if len(config.Secret) < 16 {
		return fmt.Errorf("JWT密钥长度不能少于16个字符，当前长度: %d", len(config.Secret))
	}

	if strings.Contains(config.Secret, "change-me") || config.Secret == "nodehub-jwt-secret" {
		return fmt.Errorf("请修改默认的JWT密钥以确保安全性")
	}

	return nil
}

func validateSessionConfig(config *config.SessionConfig) error {
	if config.NodePath == "" {
		return fmt.Errorf("节点会话文件路径不能为空")
	}

	dir := filepath.Dir(config.NodePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("无法创建会话目录 %s: %v", dir, err)
	}

	if !utils.IsWritableDir(dir) {
		return fmt.Errorf("会话目录 %s 不可写", dir)
	}

	return nil
}
