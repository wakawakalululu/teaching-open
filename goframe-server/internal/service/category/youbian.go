// 流水号编码生成规则（移植自 Java org.jeecg.common.util.YouBianCodeUtil，
// 语义逐字节对齐，供分类字典 code 自动生成使用）。
//
// 规则：尾段 = 字母 + 2 位数字，从 A01 起递增；
//   - 数字满 99 → 字母进位、数字归 01（A99 → B01）；
//   - Z99 → 追加新段 A01（A01Z99 → A01Z99A01）。
package category

import (
	"fmt"
	"strconv"
)

// numLength 尾段数字位数（与 Java 常量一致）。
const numLength = 2

// nextYouBianCode 根据前一个 code 获取同级下一个 code；code 为空时返回首码 A01。
func nextYouBianCode(code string) string {
	if code == "" {
		return "A" + fmt.Sprintf("%0*d", numLength, 1)
	}
	segLen := 1 + numLength
	if len(code) < segLen {
		// 不足以构成完整尾段，按 Java NumberFormatException 路径视为非法输入
		return ""
	}
	before := code[:len(code)-segLen]
	after := code[len(code)-segLen:]
	afterLetter := after[0]
	afterNum, err := strconv.Atoi(after[1:])
	if err != nil {
		return ""
	}

	const maxNum = 99
	var nextNum string
	nextLetter := afterLetter
	if afterNum == maxNum {
		nextNum = fmt.Sprintf("%0*d", numLength, 1)
		nextLetter = nextLetterOf(afterLetter)
	} else {
		nextNum = fmt.Sprintf("%0*d", numLength, afterNum+1)
	}

	// Z99：追加新段（A01Z99 → A01Z99A01）
	if afterLetter == 'Z' && afterNum == maxNum {
		return code + string(nextLetter) + nextNum
	}
	return before + string(nextLetter) + nextNum
}

// subYouBianCode 根据父 code 获取首个子级 code；localCode 非空时为同级递增。
func subYouBianCode(parentCode, localCode string) string {
	if localCode != "" {
		return nextYouBianCode(localCode)
	}
	return parentCode + "A" + fmt.Sprintf("%0*d", numLength, 1)
}

// nextLetterOf 字母递增（Z 回绕到 A）。
func nextLetterOf(zimu byte) byte {
	if zimu == 'Z' {
		return 'A'
	}
	return zimu + 1
}
