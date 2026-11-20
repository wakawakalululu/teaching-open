// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysConfig is the golang structure of table sys_config for DAO operations like Where/Data.
type SysConfig struct {
	g.Meta        `orm:"table:sys_config, do:true"`
	Id            interface{} // 主键
	ConfigKey     interface{} // 配置属性
	ConfigValue   interface{} // 配置值
	ConfigEnabled interface{} // 是否生效 1生效 0失效
	Comment       interface{} // 备注
}
