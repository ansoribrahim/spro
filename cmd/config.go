package main

import (
	"context"
	"time"

	"github.com/jpillora/backoff"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	Env       *EnvConfig
	postgesDB *gorm.DB
	// StopTickerCh signal for closing ticker channel
	StopTickerCh chan bool
)

type EnvConfig struct {
	Postgres Postgres `mapstructure:"postgres"`
	Redis    Redis    `mapstructure:"redis"`
}

type Postgres struct {
	DSN             string        `mapstructure:"dsn"`
	LogLevel        string        `mapstructure:"log_level"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	PingInterval    time.Duration `mapstructure:"ping_interval"`
	RetryAttempts   float64       `mapstructure:"retry_attempts"`
}

type Redis struct {
	CacheHost       string        `mapstructure:"cache_host"`
	WorkerCacheHost string        `mapstructure:"worker_cache_host"`
	DialTimeout     time.Duration `mapstructure:"dial_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
}

func LoadConfig() {
	setupEnv()
	initializePostgresConn()
	initializeRedisConn()
}

func setupEnv() {
	viper.SetConfigFile("config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatal("Failed to read config file: ", err)
	}

	err = viper.Unmarshal(&Env)
	if err != nil {
		logrus.Fatal("Failed to unmarshal config file: ", err)
	}
}

func initializePostgresConn() *gorm.DB {
	conn, err := openPostgresConn(Env.Postgres.DSN)
	if err != nil {
		logrus.WithField("databaseDSN", Env.Postgres.DSN).Fatal("failed to connect postgresql database: ", err)
	}

	StopTickerCh = make(chan bool)

	go checkConnection(time.NewTicker(Env.Postgres.PingInterval), Env.Postgres.DSN)

	switch Env.Postgres.LogLevel {
	case "error":
		conn.Logger = conn.Logger.LogMode(logger.Error)
	case "warn":
		conn.Logger = conn.Logger.LogMode(logger.Warn)
	case "silent":
		conn.Logger = conn.Logger.LogMode(logger.Silent)
	default:
		conn.Logger = conn.Logger.LogMode(logger.Info)
	}

	postgesDB = conn

	return postgesDB
}

func openPostgresConn(dsn string) (*gorm.DB, error) {
	psqlDialector := postgres.Open(dsn)
	db, err := gorm.Open(psqlDialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	conn, err := db.DB()
	if err != nil {
		logrus.Fatal(err)
	}
	conn.SetMaxIdleConns(Env.Postgres.MaxIdleConns)
	conn.SetMaxOpenConns(Env.Postgres.MaxOpenConns)
	conn.SetConnMaxLifetime(Env.Postgres.ConnMaxLifetime)

	return db, nil
}

func checkConnection(ticker *time.Ticker, dsn string) {
	for {
		select {
		case <-StopTickerCh:
			ticker.Stop()
			return
		case <-ticker.C:
			if _, err := postgesDB.DB(); err != nil {
				reconnectPostgresConn(dsn)
			}
		}
	}
}

func reconnectPostgresConn(dsn string) {
	b := backoff.Backoff{
		Factor: 2,
		Jitter: true,
		Min:    100 * time.Millisecond,
		Max:    1 * time.Second,
	}

	postgresRetryAttempts := Env.Postgres.RetryAttempts

	for b.Attempt() < postgresRetryAttempts {
		conn, err := openPostgresConn(dsn)
		if err != nil {
			logrus.WithField("databaseDSN", dsn).Error("failed to connect postgresql database: ", err)
		}

		if conn != nil {
			postgesDB = conn
			break
		}
		time.Sleep(b.Duration())
	}

	if b.Attempt() >= postgresRetryAttempts {
		logrus.Fatal("maximum retry to connect database")
	}
	b.Reset()
}

func initializeRedisConn() *redis.Client {
	opts, err := redis.ParseURL(Env.Redis.CacheHost)
	if err != nil {
		logrus.Fatal(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         opts.Addr,
		Username:     opts.Username,
		Password:     opts.Password,
		DB:           opts.DB,
		DialTimeout:  Env.Redis.DialTimeout,
		WriteTimeout: Env.Redis.WriteTimeout,
		ReadTimeout:  Env.Redis.ReadTimeout,
	})

	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		logrus.Fatal("err connect to Redis: ", err)
	}

	return rdb
}
