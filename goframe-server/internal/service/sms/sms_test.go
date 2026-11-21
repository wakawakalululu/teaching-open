// Package sms 单测：验证码存储/校验语义与行为限流。
// 全部基于进程内存储（New(false)），不依赖真实 Redis/数据库。
package sms

import (
	"context"
	"errors"
	"testing"
	"time"
)

func testCtx() context.Context {
	return context.Background()
}

func TestGenCode(t *testing.T) {
	for i := 0; i < 50; i++ {
		code, err := genCode(codeLength)
		if err != nil {
			t.Fatalf("生成验证码失败: %v", err)
		}
		if len(code) != codeLength {
			t.Fatalf("验证码长度错误: %s", code)
		}
		for _, c := range code {
			if c < '0' || c > '9' {
				t.Fatalf("验证码含非数字字符: %s", code)
			}
		}
	}
}

func TestSendRejectsEmptyMobile(t *testing.T) {
	svc := New(false)
	if err := svc.Send(testCtx(), "", ModeRegister, "1.2.3.4"); !errors.Is(err, ErrMobileRequired) {
		t.Fatalf("期望 ErrMobileRequired，得到 %v", err)
	}
}

func TestSendResendBlockPrecedesModeDispatch(t *testing.T) {
	// Java 语义：Redis 中已有存活验证码时，无论 smsmode 是什么都在
	// 模式校验（涉及查库）之前拒绝重发。内存存储下该路径不触达数据库。
	svc := New(false)
	ctx := testCtx()
	const phone = "13800001111"
	if err := svc.store.Save(ctx, phone, "123456", CodeTTL); err != nil {
		t.Fatalf("预置验证码失败: %v", err)
	}
	for _, mode := range []string{ModeLogin, ModeRegister, ModeForgetPassword, "9"} {
		if err := svc.Send(ctx, phone, mode, "1.2.3.4"); !errors.Is(err, ErrCodeStillValid) {
			t.Fatalf("mode=%s 期望 ErrCodeStillValid，得到 %v", mode, err)
		}
	}
}

func TestSendUnknownMode(t *testing.T) {
	// 未知 smsmode：Java 走 b=false → 发送失败话术；不触达数据库
	svc := New(false)
	if err := svc.Send(testCtx(), "13800001111", "9", "1.2.3.4"); !errors.Is(err, ErrSendFailed) {
		t.Fatalf("期望 ErrSendFailed，得到 %v", err)
	}
}

func TestVerifyFailClosed(t *testing.T) {
	svc := New(false)
	ctx := testCtx()
	const (
		phone = "13800002222"
		code  = "654321"
	)
	// 无码 → 错误（Java：redis null → equals false）
	if err := svc.Verify(ctx, phone, code); !errors.Is(err, ErrCodeWrong) {
		t.Fatalf("无码时期望 ErrCodeWrong，得到 %v", err)
	}
	if err := svc.store.Save(ctx, phone, "111111", CodeTTL); err != nil {
		t.Fatalf("预置验证码失败: %v", err)
	}
	// 不匹配
	if err := svc.Verify(ctx, phone, code); !errors.Is(err, ErrCodeWrong) {
		t.Fatalf("不匹配时期望 ErrCodeWrong，得到 %v", err)
	}
	// 命中
	if err := svc.Verify(ctx, phone, "111111"); err != nil {
		t.Fatalf("命中时期望 nil，得到 %v", err)
	}
	// 空码恒不通过（fail-closed）
	if err := svc.Verify(ctx, phone, ""); !errors.Is(err, ErrCodeWrong) {
		t.Fatalf("空码时期望 ErrCodeWrong，得到 %v", err)
	}
}

func TestVerifyResetSemantics(t *testing.T) {
	// passwordChange 契约区分两种话术：无值=失效、不匹配=不匹配
	svc := New(false)
	ctx := testCtx()
	const phone = "13800003333"
	if err := svc.VerifyReset(ctx, phone, "123456"); !errors.Is(err, ErrCodeExpired) {
		t.Fatalf("无值时期望 ErrCodeExpired，得到 %v", err)
	}
	if err := svc.store.Save(ctx, phone, "999999", CodeTTL); err != nil {
		t.Fatalf("预置验证码失败: %v", err)
	}
	if err := svc.VerifyReset(ctx, phone, "123456"); !errors.Is(err, ErrCodeMismatch) {
		t.Fatalf("不匹配时期望 ErrCodeMismatch，得到 %v", err)
	}
	if err := svc.VerifyReset(ctx, phone, "999999"); err != nil {
		t.Fatalf("命中时期望 nil，得到 %v", err)
	}
}

func TestReArmAndConsume(t *testing.T) {
	svc := New(false)
	ctx := testCtx()
	store, ok := svc.store.(*memCodeStore)
	if !ok {
		t.Fatal("New(false) 应装配进程内存储")
	}
	var clock = time.Unix(1700000000, 0)
	store.now = func() time.Time { return clock }
	const phone = "13800004444"

	// ReArm 重存验证码（phoneVerification 通过后的续期）
	if err := svc.ReArm(ctx, phone, "333333"); err != nil {
		t.Fatalf("ReArm 失败: %v", err)
	}
	if err := svc.Verify(ctx, phone, "333333"); err != nil {
		t.Fatalf("ReArm 后校验应通过，得到 %v", err)
	}
	// ReArm 写入带 TTL：滑出窗口后失效（Java 原实现无 TTL，Go 加固）
	clock = clock.Add(CodeTTL + time.Second)
	if err := svc.Verify(ctx, phone, "333333"); !errors.Is(err, ErrCodeWrong) {
		t.Fatalf("TTL 过后期望 ErrCodeWrong，得到 %v", err)
	}
	// Consume 一次性作废
	if err := svc.ReArm(ctx, phone, "444444"); err != nil {
		t.Fatalf("ReArm 失败: %v", err)
	}
	if err := svc.Consume(ctx, phone); err != nil {
		t.Fatalf("Consume 失败: %v", err)
	}
	if err := svc.Verify(ctx, phone, "444444"); !errors.Is(err, ErrCodeWrong) {
		t.Fatalf("Consume 后期望 ErrCodeWrong，得到 %v", err)
	}
}

func TestMemStoreExpiry(t *testing.T) {
	m := newMemCodeStore()
	ctx := testCtx()
	var clock = time.Unix(1700000000, 0)
	m.now = func() time.Time { return clock }
	if err := m.Save(ctx, "k", "123456", 10*time.Second); err != nil {
		t.Fatalf("Save 失败: %v", err)
	}
	if _, found, err := m.Load(ctx, "k"); err != nil || !found {
		t.Fatalf("有效期内应可读取 found=%v err=%v", found, err)
	}
	clock = clock.Add(11 * time.Second)
	if _, found, err := m.Load(ctx, "k"); err != nil || found {
		t.Fatalf("过期键应不可见 found=%v err=%v", found, err)
	}
}

func TestRateWindow(t *testing.T) {
	w := NewRateWindow(2, time.Minute)
	var clock = time.Unix(1700000000, 0)
	w.now = func() time.Time { return clock }
	if !w.Allow("a") || !w.Allow("a") {
		t.Fatal("窗口内前 limit 次应放行")
	}
	if w.Allow("a") {
		t.Fatal("超限后应拒绝")
	}
	clock = clock.Add(61 * time.Second)
	if !w.Allow("a") {
		t.Fatal("窗口滑动后应放行")
	}
}

func TestMaskPhone(t *testing.T) {
	if got := maskPhone("13812345678"); got != "138****5678" {
		t.Fatalf("脱敏结果错误: %s", got)
	}
	if got := maskPhone("123456"); got != "123456" {
		t.Fatalf("短号码应原样返回: %s", got)
	}
}
