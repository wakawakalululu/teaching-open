package httpapi

import (
	"io"
	"log/slog"

	"github.com/wakawakalululu/teaching-open/go-server/internal/store"
)

// testLogger 返回丢弃输出的 logger，避免测试噪声。
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
}

// testUser 返回一个用于 DTO 断言的用户对象。
func testUser() *store.User {
	return &store.User{
		ID:       "u-1",
		Username: "alice",
		Realname: "Alice",
		Password: "$2a$12$secret",
		Salt:     "salt",
		Status:   1,
	}
}
