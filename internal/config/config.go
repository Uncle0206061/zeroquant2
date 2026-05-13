// Package config 提供配置加载、数据库和 Redis 连接管理
// 包含：连接池优化、索引自动创建、结构化日志
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/Uncle0206061/zeroquant2/backend/pkg/logger"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// Config 定义应用配置结构
type Config struct {
	// 服务配置
	ServerPort string `yaml:"server_port"` // 后端端口：8080
	Mode       string `yaml:"mode"`        // 运行模式：debug, release

	// 日志配置
	LogLevel string `yaml:"log_level"` // 日志级别：debug, info, warn, error
	LogPath  string `yaml:"log_path"`  // 日志文件路径

	// 数据库配置
	DBHost     string `yaml:"db_host"`     // PostgreSQL 主机
	DBPort     string `yaml:"db_port"`     // PostgreSQL 端口：5432
	DBUser     string `yaml:"db_user"`     // 用户名
	DBPassword string `yaml:"db_password"` // 密码
	DBName     string `yaml:"db_name"`     // 数据库名：biz
	DBSSLMode  string `yaml:"db_sslmode"`  // SSL 模式：disable

	// 数据库连接池配置
	DBMaxIdleConns int `yaml:"db_max_idle_conns"` // 最大空闲连接数
	DBMaxOpenConns int `yaml:"db_max_open_conns"` // 最大打开连接数
	DBConnMaxLife  int `yaml:"db_conn_max_life"`  // 连接最大存活时间（秒）

	// Redis 配置
	RedisHost     string `yaml:"redis_host"`     // Redis 主机
	RedisPort     string `yaml:"redis_port"`     // Redis 端口：6379
	RedisPassword string `yaml:"redis_password"` // 密码
	RedisDB       int    `yaml:"redis_db"`       // 数据库编号

	// JWT 配置
	JWTSecret string `yaml:"jwt_secret"` // JWT 密钥
	JWTExpire int    `yaml:"jwt_expire"` // 过期时间（小时）

	// 数据服务配置（Python 服务）
	DataServiceURL string `yaml:"data_service_url"` // Python 数据服务地址
}

var (
	DB     *gorm.DB
	DBPool *redis.Client
	cfg    *Config
)

// Load 加载配置文件
func Load() (*Config, error) {
	// 默认配置（含连接池优化参数）
	cfg = &Config{
		ServerPort:     "8080",
		Mode:           "debug",
		LogLevel:       "info",
		LogPath:        "./logs/app.log",
		DBHost:         "localhost",
		DBPort:         "5432",
		DBUser:         "postgres",
		DBPassword:     "postgres",
		DBName:         "biz",
		DBSSLMode:      "disable",
		DBMaxIdleConns: 10,
		DBMaxOpenConns: 30,
		DBConnMaxLife:  3600,
		RedisHost:      "localhost",
		RedisPort:      "6379",
		RedisPassword:  "",
		RedisDB:        0,
		JWTSecret:      "zeroquant2026",
		JWTExpire:      24 * 7,
	}

	// 尝试从配置文件加载
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "./config.yaml"
	}

	if data, err := os.ReadFile(cfgPath); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			logger.Warn("Failed to parse config file, using defaults: %v", err)
		}
	}

	// 从环境变量覆盖
	if v := os.Getenv("SERVER_PORT"); v != "" {
		cfg.ServerPort = v
	}
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.DBHost = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.DBPassword = v
	}
	if v := os.Getenv("REDIS_HOST"); v != "" {
		cfg.RedisHost = v
	}

	logger.Info("Config loaded successfully")
	return cfg, nil
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return DB
}

// GetRedis 获取 Redis 连接
func GetRedis() *redis.Client {
	return DBPool
}

// GetConfig 获取配置实例
func GetConfig() *Config {
	return cfg
}

// InitDB 初始化数据库连接（含连接池优化）
func InitDB(c *Config) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(c.DBMaxIdleConns)                                        // 最大空闲连接
	sqlDB.SetMaxOpenConns(c.DBMaxOpenConns)                                        // 最大打开连接
	sqlDB.SetConnMaxLifetime(time.Duration(c.DBConnMaxLife) * time.Second)         // 连接最大存活
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)                                     // 空闲连接最大存活

	logger.Info("Database connected, pool: max_open=%d, max_idle=%d, max_life=%ds",
		c.DBMaxOpenConns, c.DBMaxIdleConns, c.DBConnMaxLife)
	return nil
}

// InitRedis 初始化 Redis 连接（含连接池优化）
func InitRedis(c *Config) error {
	DBPool = redis.NewClient(&redis.Options{
		Addr:         c.RedisHost + ":" + c.RedisPort,
		Password:     c.RedisPassword,
		DB:           c.RedisDB,
		PoolSize:     10,           // 连接池大小
		MinIdleConns: 5,            // 最小空闲连接
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	logger.Info("Redis connected, pool_size=10, min_idle=5")
	return nil
}

// CloseDB 关闭数据库连接
func CloseDB() {
	if DB != nil {
		sqlDB, _ := DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

// CloseRedis 关闭 Redis 连接
func CloseRedis() {
	if DBPool != nil {
		DBPool.Close()
	}
}
