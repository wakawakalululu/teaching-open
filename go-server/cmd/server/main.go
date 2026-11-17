// Command server 是 teaching-open 的 Go 服务端入口。
//
// 启动顺序刻意保持为「配置校验 -> 依赖连接 -> 迁移自检 -> 启动服务」，
// 任一环节失败立即退出，不带病启动（Java 版存在大量配置缺省可用的情况，
// 例如 JWT 密钥直接取用户密码 MD5，缺乏统一校验）。
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"github.com/wakawakalululu/teaching-open/go-server/internal/auth"
	"github.com/wakawakalululu/teaching-open/go-server/internal/config"
	"github.com/wakawakalululu/teaching-open/go-server/internal/httpapi"
	"github.com/wakawakalululu/teaching-open/go-server/internal/service"
	"github.com/wakawakalululu/teaching-open/go-server/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("服务启动失败", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger.Info("配置加载完成",
		slog.String("env", cfg.Logistics.Env),
		slog.String("version", cfg.Logistics.Version),
	)

	// 数据库
	db, err := sqlx.Open("mysql", cfg.MySQL.DSN)
	if err != nil {
		return errors.New("打开数据库失败: " + err.Error())
	}
	defer db.Close()
	db.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MySQL.ConnMaxLifetime)

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelStartup()
	if err := db.PingContext(startupCtx); err != nil {
		return errors.New("数据库连接不可用: " + err.Error())
	}
	// 强制使用 UTC 会话时区，避免跨时区部署时 create_time 出现偏移
	if _, err := db.ExecContext(startupCtx, "SET time_zone = '+00:00'"); err != nil {
		logger.Warn("设置数据库会话时区失败", slog.String("error", err.Error()))
	}

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
	defer rdb.Close()
	if err := rdb.Ping(startupCtx).Err(); err != nil {
		return errors.New("Redis 连接不可用: " + err.Error())
	}

	// 令牌与吊销
	revoker := service.NewRedisRevoker(rdb)
	tokens, err := auth.NewManager(cfg.JWT, revoker)
	if err != nil {
		return err
	}
	limiter := auth.NewLoginLimiter(rdb, cfg.Security.LoginFailLimit, cfg.Security.LoginFailWindow)
	userRepo := store.NewUserRepo(db)
	loginSvc := service.NewLoginService(userRepo, tokens, limiter, logger)

	// 迁移自检：提示仍处于弱算法保护下的账号规模
	if err := reportLegacyAccounts(startupCtx, db, logger); err != nil {
		logger.Warn("存量密码统计失败", slog.String("error", err.Error()))
	}

	api := httpapi.New(cfg, logger, loginSvc, tokens)
	srv := &http.Server{
		Addr:              cfg.Server.Addr,
		Handler:           api.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("HTTP 服务已启动", slog.String("addr", cfg.Server.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	// 优雅停机
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-stop:
		logger.Info("收到停机信号", slog.String("signal", sig.String()))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return errors.New("优雅停机失败: " + err.Error())
	}
	logger.Info("服务已停止")
	return nil
}

// reportLegacyAccounts 统计仍使用 Java 版 PBE 密文的账号数量，
// 用于评估迁移进度；该查询只读，且表名/列名为常量。
func reportLegacyAccounts(ctx context.Context, db *sqlx.DB, logger *slog.Logger) error {
	const query = `
		SELECT
			COUNT(*) AS total,
			SUM(CASE WHEN password LIKE '$2a$%' OR password LIKE '$2b$%' OR password LIKE '$2y$%' THEN 1 ELSE 0 END) AS bcrypt_count
		FROM sys_user
		WHERE del_flag = 0`
	var row struct {
		Total       int `db:"total"`
		BcryptCount int `db:"bcrypt_count"`
	}
	if err := db.GetContext(ctx, &row, query); err != nil {
		return err
	}
	logger.Info("密码散列迁移进度",
		slog.Int("total_users", row.Total),
		slog.Int("bcrypt_users", row.BcryptCount),
		slog.Int("legacy_users", row.Total-row.BcryptCount),
	)
	return nil
}
