// hub_impl.go Hub 依赖的 fmt/time 助手（独立实现避免导入环）。
package ws

import (
	"fmt"
	"time"
)

// fmtSprintf fmt.Sprintf 别名。
func fmtSprintf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

// timeNowNano 当前纳秒时间戳。
func timeNowNano() int64 { return time.Now().UnixNano() }
