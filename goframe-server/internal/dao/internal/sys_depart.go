// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysDepartDao is the data access object for table sys_depart.
type SysDepartDao struct {
	table   string           // table is the underlying table name of the DAO.
	group   string           // group is the database configuration group name of current DAO.
	columns SysDepartColumns // columns contains all the column names of Table for convenient usage.
}

// SysDepartColumns defines and stores the column names for table sys_depart.
// 列名以 api/db/teachingopen2.8.sql L570 为准（共 20 列，逐列核对无遗漏）。
type SysDepartColumns struct {
	Id             string // ID
	ParentId       string // 父机构ID
	DepartName     string // 机构/部门名称
	DepartNameEn   string // 英文名
	DepartNameAbbr string // 缩写
	DepartOrder    string // 排序
	Description    string // 描述
	OrgCategory    string // 机构类别 1组织机构，2岗位
	OrgType        string // 机构类型 1一级部门 2子部门
	OrgCode        string // 机构编码
	Mobile         string // 手机号
	Fax            string // 传真
	Address        string // 地址
	Memo           string // 备注
	Status         string // 状态（1启用，0不启用）
	DelFlag        string // 删除状态（0，正常，1已删除）
	CreateBy       string // 创建人
	CreateTime     string // 创建日期
	UpdateBy       string // 更新人
	UpdateTime     string // 更新日期
}

// sysDepartColumns holds the columns for table sys_depart.
var sysDepartColumns = SysDepartColumns{
	Id:             "id",
	ParentId:       "parent_id",
	DepartName:     "depart_name",
	DepartNameEn:   "depart_name_en",
	DepartNameAbbr: "depart_name_abbr",
	DepartOrder:    "depart_order",
	Description:    "description",
	OrgCategory:    "org_category",
	OrgType:        "org_type",
	OrgCode:        "org_code",
	Mobile:         "mobile",
	Fax:            "fax",
	Address:        "address",
	Memo:           "memo",
	Status:         "status",
	DelFlag:        "del_flag",
	CreateBy:       "create_by",
	CreateTime:     "create_time",
	UpdateBy:       "update_by",
	UpdateTime:     "update_time",
}

// NewSysDepartDao creates and returns a new DAO object for table data access.
func NewSysDepartDao() *SysDepartDao {
	return &SysDepartDao{
		group:   "default",
		table:   "sys_depart",
		columns: sysDepartColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *SysDepartDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *SysDepartDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *SysDepartDao) Columns() SysDepartColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *SysDepartDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *SysDepartDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *SysDepartDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
