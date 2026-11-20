// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysConfig 对应 sys_config 表。
type SysConfig struct {
	Id            string `json:"id"            orm:"id,primary"      description:"主键"`
	ConfigKey     string `json:"configKey"     orm:"config_key"      description:"配置属性"`
	ConfigValue   string `json:"configValue"   orm:"config_value"    description:"配置值"`
	ConfigEnabled int    `json:"configEnabled" orm:"config_enabled"  description:"是否生效 1生效 0失效"`
	Comment       string `json:"comment"       orm:"comment"         description:"备注"`
}
