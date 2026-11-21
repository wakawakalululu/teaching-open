// Package userreg 单测：不依赖真实 Redis/数据库的纯逻辑路径。
//
// 数据库相关路径（唯一性查询、事务写入）在集成/e2e 环境覆盖；
// 本文件仅覆盖参数校验、限流装配、错误语义与不触达数据库的分支。
package userreg

import (
	"context"
	"errors"
	"strings"
	"testing"

	svcSms "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/sms"
)

func newTestService() *Service {
	return New(svcSms.New(false), nil)
}

func TestPasswordChangeParamMissing(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	cases := [][4]string{
		{"", "Password123", "123456", "13800005555"},
		{"user", "", "123456", "13800005555"},
		{"user", "Password123", "", "13800005555"},
		{"user", "Password123", "123456", ""},
	}
	for _, c := range cases {
		if err := svc.PasswordChange(ctx, c[0], c[1], c[2], c[3]); !errors.Is(err, ErrParamMissing) {
			t.Fatalf("参数缺失用 %v 应返回 ErrParamMissing，得到 %v", c, err)
		}
	}
}

func TestPasswordChangeTooLongPassword(t *testing.T) {
	svc := newTestService()
	long := strings.Repeat("a", maxPasswordBytes+1)
	if err := svc.PasswordChange(context.Background(), "user", long, "123456", "13800005555"); !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("超长密码期望 ErrPasswordTooLong，得到 %v", err)
	}
}

func TestPasswordChangeVerifyBeforeLookup(t *testing.T) {
	// 参数齐全但验证码存储无值：走 sms.VerifyReset（进程内存储）→
	// ErrCodeExpired，不触达数据库（校验先于用户查询，顺序与 Java 一致）
	svc := newTestService()
	err := svc.PasswordChange(context.Background(), "user", "Password123", "123456", "13800006666")
	if !errors.Is(err, svcSms.ErrCodeExpired) {
		t.Fatalf("期望 ErrCodeExpired，得到 %v", err)
	}
}

func TestPhoneVerificationWrongCode(t *testing.T) {
	svc := newTestService()
	if err := svc.PhoneVerification(context.Background(), "13800007777", "000000"); !errors.Is(err, svcSms.ErrCodeWrong) {
		t.Fatalf("期望 ErrCodeWrong，得到 %v", err)
	}
}

func TestRegisterRateGateWiring(t *testing.T) {
	// 限流参数装配自检：同 IP / 同手机号窗口内放行 limit 次，超限拒绝
	svc := newTestService()
	for i := 0; i < regIPLimit; i++ {
		if !svc.ipGate.Allow("ip:1.2.3.4") {
			t.Fatalf("同 IP 第 %d 次应放行", i+1)
		}
	}
	if svc.ipGate.Allow("ip:1.2.3.4") {
		t.Fatal("同 IP 超限后应拒绝")
	}
	for i := 0; i < regPhoneLimit; i++ {
		if !svc.phoneGate.Allow("phone:13800001234") {
			t.Fatalf("同手机号第 %d 次应放行", i+1)
		}
	}
	if svc.phoneGate.Allow("phone:13800001234") {
		t.Fatal("同手机号超限后应拒绝")
	}
}

func TestRandomPassword(t *testing.T) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	pw, err := randomPassword(8)
	if err != nil {
		t.Fatalf("生成随机密码失败: %v", err)
	}
	if len(pw) != 8 {
		t.Fatalf("随机密码长度错误: %s", pw)
	}
	for _, c := range pw {
		if !strings.ContainsRune(alphabet, c) {
			t.Fatalf("随机密码含非法字符: %s", pw)
		}
	}
}

func TestConfigIDListNilCfg(t *testing.T) {
	svc := newTestService()
	out, err := svc.configIDList(context.Background(), "_defaultRole")
	if err != nil || out != nil {
		t.Fatalf("cfg 为 nil 时期望 (nil, nil)，得到 (%v, %v)", out, err)
	}
}

func TestMaskHelpers(t *testing.T) {
	if got := maskPhone("13812345678"); got != "138****5678" {
		t.Fatalf("maskPhone 结果错误: %s", got)
	}
	if got := maskText("administrator"); got != "a***r" {
		t.Fatalf("maskText 结果错误: %s", got)
	}
	if got := maskText("ab"); got != "**" {
		t.Fatalf("maskText 短串结果错误: %s", got)
	}
}
