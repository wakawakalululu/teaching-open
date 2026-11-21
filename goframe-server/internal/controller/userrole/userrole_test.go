package userrole

import (
	"reflect"
	"testing"
)

// splitIDs 批量删除 ids 参数解析健壮性：空串、纯空白、连续逗号等非法输入
// 均不得产生空段（Java 原版空段会进入 IN (”) 恒不命中，丢弃后语义等价）。
func TestSplitIDs(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{",", nil},
		{" , , ", nil},
		{"u1", []string{"u1"}},
		{"u1,u2", []string{"u1", "u2"}},
		{"u1, u2 ,", []string{"u1", "u2"}},
		{",u1,,u2,", []string{"u1", "u2"}},
	}
	for _, c := range cases {
		if got := splitIDs(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("splitIDs(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
