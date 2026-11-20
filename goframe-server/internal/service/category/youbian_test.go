package category

import "testing"

// 用例覆盖 Java 版注释中的全部示例与边界（Z99 回绕、追加段、子级首码）。
func TestNextYouBianCode(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", "A01"},          // 首码
		{"A01", "A02"},       // 同级递增
		{"A99", "B01"},       // 数字满 99 字母进位
		{"D01A04", "D01A05"}, // 多级尾段递增
		{"Z99", "Z99A01"},    // Z99 追加新段
		{"A01Z99", "A01Z99A01"},
	}
	for _, c := range cases {
		if got := nextYouBianCode(c.in); got != c.want {
			t.Errorf("nextYouBianCode(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNextYouBianCodeInvalid(t *testing.T) {
	for _, in := range []string{"A", "A1", "A###", "A0X"} {
		if got := nextYouBianCode(in); got != "" {
			t.Errorf("nextYouBianCode(%q) = %q, want empty for invalid input", in, got)
		}
	}
}

func TestSubYouBianCode(t *testing.T) {
	cases := []struct {
		parent string
		local  string
		want   string
	}{
		{"A01", "", "A01A01"},       // 父节点下首个子码
		{"A01", "A01B03", "A01B04"}, // 有同级时按同级递增
	}
	for _, c := range cases {
		if got := subYouBianCode(c.parent, c.local); got != c.want {
			t.Errorf("subYouBianCode(%q,%q) = %q, want %q", c.parent, c.local, got, c.want)
		}
	}
}
