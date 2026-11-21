// teachingwork_test.go 作业域零 DB 依赖单测：排行榜排序判定、KV 存储行为、
// 快照格式、数值辅助。
package teachingwork

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
)

func TestLeaderboardOrder(t *testing.T) {
	cases := []struct {
		orderBy string
		want    string
	}{
		{"view", "teaching_work.view_num DESC"},
		{"time", "teaching_work.create_time DESC"},
		{"star", "teaching_work.star_num DESC"},
		// 前端已传 create_time 但后端 switch 不认 → 无排序（Java 现状保留）
		{"create_time", ""},
		{"", ""},
	}
	for _, tc := range cases {
		if got := leaderboardOrder(tc.orderBy); got != tc.want {
			t.Errorf("leaderboardOrder(%q) = %q, 期望 %q", tc.orderBy, got, tc.want)
		}
	}
}

func TestPlaceholders(t *testing.T) {
	if got := placeholders(3); got != "?,?,?" {
		t.Errorf("placeholders(3) = %q", got)
	}
	if got := placeholders(1); got != "?" {
		t.Errorf("placeholders(1) = %q", got)
	}
}

func TestMemStore(t *testing.T) {
	s := &memStore{sets: map[string]map[string]bool{}, vals: map[string]memVal{}}
	ctx := context.Background()
	s.Set(ctx, "k1", "v1")
	if v, ok := s.Get(ctx, "k1"); !ok || v != "v1" {
		t.Fatalf("Get/Set 异常: %v %v", v, ok)
	}
	s.SetEx(ctx, "k2", "v2", 1*time.Hour)
	if _, ok := s.Get(ctx, "k2"); !ok {
		t.Fatal("SetEx 后应可读")
	}
	s.SAdd(ctx, "s1", "a")
	s.SAdd(ctx, "s1", "b")
	if members := s.SGet(ctx, "s1"); len(members) != 2 {
		t.Fatalf("SGet = %v, 期望 2 个成员", members)
	}
	if !s.SHasKey(ctx, "s1", "a") || s.SHasKey(ctx, "s1", "c") {
		t.Fatal("SHasKey 异常")
	}
	s.SRemove(ctx, "s1", "a")
	if members := s.SGet(ctx, "s1"); len(members) != 1 {
		t.Fatalf("SRemove 后 = %v", members)
	}
	s.Del(ctx, "k1")
	if _, ok := s.Get(ctx, "k1"); ok {
		t.Fatal("Del 后应不可读")
	}
}

func TestWorkSnapshot(t *testing.T) {
	data := do.TeachingWork{
		Id:       "w1",
		UserId:   "u1",
		WorkName: "测试作业",
	}
	got := workSnapshot(data)
	if !strings.Contains(got, `"workName":"测试作业"`) || !strings.Contains(got, `"id":"w1"`) {
		t.Fatalf("快照格式异常: %s", got)
	}
}

func TestFormatHelpers(t *testing.T) {
	if intKey(4) != "4" || boolToInt(true) != 1 || boolToInt(false) != 0 {
		t.Fatal("数值辅助函数异常")
	}
	if derefStr(nil) != "" || derefInt(nil) != 0 {
		t.Fatal("指针解引用辅助函数异常")
	}
}
