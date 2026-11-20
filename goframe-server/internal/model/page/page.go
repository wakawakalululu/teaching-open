// Package page 定义业务域共用的分页参数与响应结构。
//
// 形状对齐 Java MyBatis-Plus IPage 的前端消费字段
// （records/total/size/current/pages）。
package page

import "math"

// Param 分页参数。
type Param struct {
	PageNo   int
	PageSize int
}

// Result 分页响应。
type Result struct {
	Records interface{} `json:"records"`
	Total   int         `json:"total"`
	Size    int         `json:"size"`
	Current int         `json:"current"`
	Pages   int         `json:"pages"`
}

// New 以 records 组装分页结构。
func New(records interface{}, total, pageNo, pageSize int) *Result {
	return &Result{
		Records: records,
		Total:   total,
		Size:    pageSize,
		Current: pageNo,
		Pages:   int(math.Ceil(float64(total) / float64(pageSize))),
	}
}
