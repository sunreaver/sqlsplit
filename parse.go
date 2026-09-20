package sqlsplit

import (
	"strings"
)

// Mode 表示 SQL 词法/状态机解析过程中的当前工作模式（状态）
type Mode int

const (
	// ModeUnPick 未拾取/初始空闲状态（等待识别下一条语句或注释的起始）
	ModeUnPick Mode = iota
	// ModeRemarkLine 单行注释状态（以 -- 或 # 开头，直到换行符结束）
	ModeRemarkLine
	// ModeRemarkMoreLine 多行/块注释状态（以 /* 开头，直到 */ 结束）
	ModeRemarkMoreLine
	// ModeDefaultSql 普通 SQL 语句状态（常规 DDL/DML/DQL，以单一分号结束）
	ModeDefaultSql
	// ModeProcedure 存储过程/函数/触发器等复合代码块状态（内部包含多个分号与嵌套块）
	ModeProcedure
	// ModeMaybeProcedure1 潜在存储过程判定状态1（已识别到 CREATE 关键字）
	ModeMaybeProcedure1
	// ModeMaybeProcedure2 潜在存储过程判定状态2（已识别到 CREATE OR 关键字）
	ModeMaybeProcedure2
	// ModeMaybeProcedure3 潜在存储过程判定状态3（已识别到 CREATE OR REPLACE 关键字）
	ModeMaybeProcedure3
	// ModeApostrophe 单引号字符串字面量状态（以 ' 开头，处理转义与闭合）
	ModeApostrophe
	// ModeDoubleQuotes 双引号标识符或字符串状态（以 " 开头，处理转义与闭合）
	ModeDoubleQuotes
)

// modeNames 维护 Mode 枚举到其可读文本标识的映射关系表，用于调试与打印输出
var modeNames = map[Mode]string{
	ModeUnPick:          "unpick",
	ModeRemarkLine:      "--",
	ModeRemarkMoreLine:  "/**/",
	ModeDefaultSql:      "select/update",
	ModeMaybeProcedure1: "create_",
	ModeMaybeProcedure2: "create_or",
	ModeMaybeProcedure3: "create_or_replace",
	ModeProcedure:       "procedure/event/function",
	ModeApostrophe:      "'",
	ModeDoubleQuotes:    "\"",
}

// String 将 Mode 状态转换为对应的直观字符串描述；未知状态返回 "unknow"
func (m Mode) String() string {
	// 判断：通过 map 快速查找状态名称，保持圈复杂度严格小于 10
	if name, ok := modeNames[m]; ok {
		return name
	}
	return "unknow"
}

// SqlParse 表示单条解析后的完整 SQL 语句结构体
type SqlParse struct {
	// SQL 为提取出的完整单条 SQL 文本（已附加前置注释，普通语句去除了尾部分号）
	SQL string `json:"sql"`
	// Type 为自动推导出的 SQL 语句类别（DDL、DML、DQL、TTL、DCL）
	Type SQLTYPE `json:"type"`
}

// MaxInputSize 为 Split 函数接受的最大输入字节数，默认 100MB。
// 超过此限制的输入将直接返回 nil，以避免因超大 SQL 文件引发内存溢出。
const MaxInputSize = 100 * 1024 * 1024

// Split 是本库对外导出的核心切分入口方法。
// 该方法接收多行或多段原始 SQL 脚本文本，将其准确拆分为逻辑上独立、语法完整的单条 SQL 语句列表。
//
// 使用约束：
//   - 输入必须为 UTF-8 编码文本，不支持 GBK/GB18030 等其他编码。
//   - 输入大小不应超过 MaxInputSize（100MB），超出将返回 nil。
//
// 核心兼容与处理特性：
// 1. 多方言兼容：全面支持 Oracle（PL/SQL 过程、声明区分号保护、/ 独立行结束符）、MySQL（反引号、# 注释、DELIMITER 切换）、PostgreSQL（$$ 引用块）。
// 2. 空语句过滤：自动忽略连续空分号（如 ;;;）、前置空分号，确保返回的每一项均为有效 SQL。
// 3. 注释规范化：前置注释紧密附着于下一条 SQL；文件末尾无归属的孤立注释直接丢弃。
// 4. 分号处理：普通 SQL 移除末尾分号（适配 Oracle 驱动执行规范），存储过程完整保留内部及末尾分号。
func Split(sqls string) []SqlParse {
	// 重要判断：输入超过最大允许大小时，直接返回 nil 以防止 OOM
	if len(sqls) > MaxInputSize {
		return nil
	}
	// 调用词法扫描引擎统一执行状态驱动切分
	return parseSQLScript(sqls)
}

// RemoveLastSemicolon 将输入字符串末尾的最后一个分号移除。
//
// 设计背景：
// Oracle 数据库驱动（如 godror/go-ora）在执行普通 SQL（SELECT/INSERT/UPDATE 等）时，
// 若末尾携带分号会抛出 "ORA-00911: invalid character" 错误；
// 因此普通语句在输出前必须剥离尾部分号，而存储过程/匿名块结尾的分号则必须保留。
func RemoveLastSemicolon(str string) string {
	// 重要判断：先去除首尾空白字符，若为空字符串则直接返回，避免越界访问
	str = strings.TrimSpace(str)
	if str == "" {
		return str
	}

	// 重要判断：检查字符串末尾是否包含分号，若有则精准截取掉最后一个分号并再次去除空白
	if idx := strings.LastIndex(str, ";"); idx >= 0 && idx == len(str)-1 {
		return strings.TrimSpace(str[:idx])
	}
	return str
}
