package config

import (
	"fmt"
	"os"
)

type Config struct {
	Server   ServerConfig
	MySQL    MySQLConfig
	Redis    RedisConfig
	Binance  BinanceConfig
	Monitor  MonitorConfig
	Logger   LoggerConfig
}

type ServerConfig struct {
	Host string
	Port string
}

type MySQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	Params   string
}

type RedisConfig struct {
	Enabled bool
	Addr    string
	DB      int
	Pass    string
}

type BinanceConfig struct {
	BaseSpotURL   string
	BaseUSDTMURL  string
	BaseCoinMURL  string
	DepthInterval string
}

type MonitorConfig struct {
	MaxTrackedOrders int
	HeartbeatTimeout int
}

type LoggerConfig struct {
	Level string
	Mode  string
}

func Load() Config {
	return Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		MySQL: MySQLConfig{
			Host:     getEnv("MYSQL_HOST", "47.128.154.233"),
			Port:     getEnv("MYSQL_PORT", "32066"),
			User:     getEnv("MYSQL_USER", "root"),
			Password: getEnv("MYSQL_PASSWORD", "hadamysqlroot@@pass"),
			Database: getEnv("MYSQL_DATABASE", "order_data"),
			Params:   getEnv("MYSQL_PARAMS", "parseTime=true&loc=Asia%2FShanghai"),
		},
		Redis: RedisConfig{
			Enabled: getEnv("REDIS_ENABLED", "false") == "true",
			Addr:    getEnv("REDIS_ADDR", "127.0.0.1:6379"),
			DB:      getEnvInt("REDIS_DB", 0),
			Pass:    getEnv("REDIS_PASSWORD", ""),
		},
		Binance: BinanceConfig{
			BaseSpotURL:   getEnv("BINANCE_SPOT_URL", "wss://stream.binance.com:9443/ws"),
			BaseUSDTMURL:  getEnv("BINANCE_USDTM_URL", "wss://fstream.binance.com/ws"),
			BaseCoinMURL:  getEnv("BINANCE_COINM_URL", "wss://dstream.binance.com/ws"),
			DepthInterval: getEnv("BINANCE_DEPTH_INTERVAL", "100ms"),
		},
		Monitor: MonitorConfig{
			MaxTrackedOrders: getEnvInt("MONITOR_MAX_TRACKED", 5000),
			HeartbeatTimeout: getEnvInt("MONITOR_HEARTBEAT_TIMEOUT", 60),
		},
		Logger: LoggerConfig{
			Level: getEnv("LOG_LEVEL", "info"),
			Mode:  getEnv("LOG_MODE", "console"),
		},
	}
}

func (c ServerConfig) ListenAddr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func (c MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", c.User, c.Password, c.Host, c.Port, c.Database, c.Params)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		var out int
		_, err := fmt.Sscanf(v, "%d", &out)
		if err == nil {
			return out
		}
	}
	return def
}

