package sqlsplit

import (
	"regexp"
	"strings"
)

// SQLTYPE 表示 SQL 语句的语法分类大类
type SQLTYPE string

const (
	// DCL 数据控制语言（Data Control Language）：负责权限分配、用户/角色管理及数据库访问安全配置（如 GRANT、REVOKE、CREATE USER、ALTER USER 等）
	DCL SQLTYPE = "DCL"
	// DDL 数据定义语言（Data Definition Language）：负责数据库结构、对象定义的创建与变更（如 CREATE TABLE/VIEW/PROCEDURE、ALTER、DROP、TRUNCATE 等）
	DDL SQLTYPE = "DDL"
	// DML 数据操纵语言（Data Manipulation Language）：负责表中数据的插入、更新、删除及过程调用（如 INSERT、UPDATE、DELETE、CALL、MERGE 等）
	DML SQLTYPE = "DML"
	// DQL 数据查询语言（Data Query Language）：负责数据检索与只读查询（如 SELECT、SHOW、DESCRIBE、EXPLAIN 等）
	DQL SQLTYPE = "DQL"
	// TTL 事务处理语言（Transaction Control Language / TCL）：负责事务的提交、回滚与保存点标记（如 COMMIT、ROLLBACK、SAVEPOINT 等）
	TTL SQLTYPE = "TTL"
)

// dclExactVerbs 包含直接通过首动词即可无条件判定为 DCL 类别的关键字集合。
// 这些指令均用于权限控制、会话管理或数据库级系统/游标资源配置。
var dclExactVerbs = map[string]bool{
	"AUDIT":      true, // Oracle 审计控制
	"CONNECT":    true, // 数据库连接控制
	"DISCONNECT": true, // 断开连接控制
	"EXIT":       true, // 退出客户端控制
	"GRANT":      true, // 授予用户或角色权限
	"NOAUDIT":    true, // 取消审计
	"QUIT":       true, // 退出控制
	"SHUTDOWN":   true, // 关闭数据库系统实例
	"REVOKE":     true, // 回收用户或角色权限
	"SET":        true, // 设置会话/全局变量（如 MySQL SET FOREIGN_KEY_CHECKS=0，归为 DCL 可避免参与 explain 计划分析）
	"USE":        true, // 切换当前活动数据库
	"LOCK":       true, // 表级锁控制（如 MySQL LOCK TABLES）
	"UNLOCK":     true, // 解锁控制（如 MySQL UNLOCK TABLES）
	"OPEN":       true, // 游标/资源开启控制
	"CLOSE":      true, // 游标/资源关闭控制
}

// ddlVerbs 包含首动词通常代表数据定义语言（DDL）的关键字集合。
// 注意：其中的 CREATE、DROP、ALTER、RENAME、COMMENT 可能是管理用户/角色，后续需进一步结合 istrDCL 甄别。
var ddlVerbs = map[string]bool{
	"DROP":     true, // 删除数据库对象（表、视图、索引等）
	"ALTER":    true, // 修改数据库对象结构
	"COMMENT":  true, // 对象注释说明（COMMENT ON TABLE/COLUMN）
	"TRUNCATE": true, // 截断表数据（DDL 快速清空）
	"CREATE":   true, // 创建数据库对象（表、视图、索引、过程等）
	"REINDEX":  true, // 重建索引（PostgreSQL/SQLite）
	"MOVE":     true, // 移动对象表空间或游标
	"RENAME":   true, // 重命名对象（如 RENAME TABLE；RENAME USER 会由 istrDCL 归为 DCL）
}

// dmlVerbs 包含首动词直接代表数据操纵语言（DML）的关键字集合。
// 用于对表内部的数据行进行变更或过程执行。
var dmlVerbs = map[string]bool{
	"INSERT":  true, // 插入数据记录（含 INSERT INTO ... SELECT）
	"UPDATE":  true, // 更新现有数据记录
	"DELETE":  true, // 删除数据记录
	"REPLACE": true, // 替换/写入数据记录（MySQL/SQLite REPLACE INTO）
	"CALL":    true, // 调用存储过程或函数
	"DECLARE": true, // 声明游标、变量或匿名 PL/SQL 块
	"MERGE":   true, // 合并写入操作（UPSERT 语义，如 Oracle/PostgreSQL MERGE INTO）
}

// ttlVerbs 包含直接代表事务处理语言（TTL/TCL）的关键字集合。
var ttlVerbs = map[string]bool{
	"COMMIT":    true, // 提交当前活动事务
	"ROLLBACK":  true, // 回滚当前活动事务
	"SAVEPOINT": true, // 设定事务保存点
	"BEGIN":     true, // 开启事务（如 BEGIN;、BEGIN TRANSACTION 等）
	"START":     true, // 开启事务（如 MySQL START TRANSACTION）
}

// dclSQL 定义用于检测属于安全/权限控制（DCL）的正则模式。
// 背景：像 CREATE USER、DROP ROLE、ALTER ROLE、RENAME USER、COMMENT ON POLICY 等
// 虽然以 CREATE/DROP/ALTER/RENAME/COMMENT 开头，但其实质管理的是用户、角色或安全策略，应准确判定为 DCL 而非普通 DDL。
const dclSQL = `^(?i)(CREATE|DROP|ALTER|RENAME|COMMENT ON)\s+(USER|ROLE|POLICY)`

// dclReg 为编译后的 DCL 识别正则表达式，全局单例复用以提升性能
var dclReg = regexp.MustCompile(dclSQL)

// istrDCL 使用正则表达式检测语句是否属于涉及用户、角色、安全策略的 DCL 语句
func istrDCL(raw string) bool {
	return dclReg.MatchString(raw)
}

// skipWhitespace 跳过当前字符串中的连续空白字符（空格、制表符、回车换行）
func skipWhitespace(raw string, pos int) int {
	for pos < len(raw) && raw[pos] <= ' ' {
		pos++
	}
	return pos
}

// skipLineComment 扫描单行注释（-- 或 #）直到行末换行符并返回换行后的下一个字符位置
func skipLineComment(raw string, pos int) int {
	for pos < len(raw) && raw[pos] != '\n' {
		pos++
	}
	// 重要判断：如果由于遇到换行符结束，则跨过该换行符以定位到下一行行首
	if pos < len(raw) && raw[pos] == '\n' {
		pos++
	}
	return pos
}

// skipBlockComment 扫描块注释（/* ... */）并返回关闭标记 */ 之后的位置；若未闭合则跳至末尾
func skipBlockComment(raw string, pos int) int {
	idx := strings.Index(raw[pos+2:], "*/")
	// 重要判断：若块注释未闭合，直接跳至字符串末尾，避免死循环
	if idx == -1 {
		return len(raw)
	}
	return pos + 2 + idx + 2
}

// stripLeadingComments 剥离 SQL 语句开头的全部前置空白字符与注释内容（单行 --、# 与块注释 /* */）。
//
// 设计背景：
// 很多生产环境 SQL 语句开头包含作者信息或版权注释，如：
//
//	/* 作者: 张三 */ CREATE TABLE t (...);
//
// 若不剥离前导注释，直接检查首字符会因为首动词被注释阻隔而误判为默认的 DQL 类型。
func stripLeadingComments(raw string) string {
	pos := 0
	for pos < len(raw) {
		pos = skipWhitespace(raw, pos)
		// 重要判断：若已扫描至末尾，则跳出循环
		if pos >= len(raw) {
			break
		}
		// 重要判断：检测标准 SQL 单行注释 --
		if raw[pos] == '-' && pos+1 < len(raw) && raw[pos+1] == '-' {
			pos = skipLineComment(raw, pos+2)
			continue
		}
		// 重要判断：检测 MySQL 风格单行注释 #
		if raw[pos] == '#' {
			pos = skipLineComment(raw, pos+1)
			continue
		}
		// 重要判断：检测通用块注释 /* ... */
		if raw[pos] == '/' && pos+1 < len(raw) && raw[pos+1] == '*' {
			pos = skipBlockComment(raw, pos)
			continue
		}
		// 遇到非注释且非空白的真正 SQL 代码内容，退出剥离循环
		break
	}
	return raw[pos:]
}

// getLeadingWord 提取输入字符串中的首个完整词（连续字母、数字或下划线），并转换为大写返回
func getLeadingWord(s string) string {
	s = strings.TrimSpace(s)
	end := 0
	for end < len(s) && (isWordChar(s[end])) {
		end++
	}
	return strings.ToUpper(s[:end])
}

// isWordChar 检查指定字节是否属于 SQL 词字符（英文字母、数字或下划线）
func isWordChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

// checkCTEVerb 检查通用表表达式（WITH 子句）之后的主语句动词属于 DML 还是 DQL
func checkCTEVerb(word string) (SQLTYPE, bool) {
	switch word {
	case "INSERT", "UPDATE", "DELETE":
		// 重要判断：WITH 后跟 INSERT/UPDATE/DELETE，属于通过 CTE 进行的数据操纵（DML）
		return DML, true
	case "SELECT":
		// 重要判断：WITH 后跟 SELECT，属于常规 CTE 查询（DQL）
		return DQL, true
	default:
		return DQL, false
	}
}

// determineCTEType 分析 Common Table Expression（CTE，即以 WITH 开头的语句）的主操作类型。
//
// 设计背景：
// WITH 语句可以是只读查询（WITH cte AS (...) SELECT ... -> DQL），
// 也可以是数据写入（WITH cte AS (...) INSERT INTO ... SELECT ... -> DML）。
// 本方法通过追踪括号匹配深度，跳过 CTE 视图定义区，寻找括号外层的主操作动词。
func determineCTEType(raw string) SQLTYPE {
	depth := 0
	for i := 0; i < len(raw); i++ {
		ch := raw[i]
		// 跳过引号字符串和注释，避免其中的括号干扰深度计数
		if nextPos, skipped := skipQuoteOrComment(raw, i); skipped {
			i = nextPos - 1
			continue
		}
		// 重要判断：进入 CTE 括号定义，增加深度
		if ch == '(' {
			depth++
			continue
		}
		// 重要判断：退出 CTE 括号定义，减少深度
		if ch == ')' {
			if depth > 0 {
				depth--
			}
			continue
		}
		// 重要判断：在深度为 0 的外层找到首个标识符动词，即为主语句动词
		if depth == 0 && isWordChar(ch) {
			w := getLeadingWord(raw[i:])
			if tp, ok := checkCTEVerb(w); ok {
				return tp
			}
		}
	}
	// 默认兜底为 DQL
	return DQL
}

// skipQuoteOrComment 检查指定位置是否处于引号字符串或注释的起始位置，若是则跳过并返回结束位置
func skipQuoteOrComment(raw string, pos int) (int, bool) {
	if pos >= len(raw) {
		return pos, false
	}
	// 尝试跳过引号字面量
	if nextPos, ok := skipQuoteAt(raw, pos); ok {
		return nextPos, true
	}
	// 尝试跳过注释
	return skipCommentAt(raw, pos)
}

// skipQuoteAt 检查并跳过单引号或双引号字面量
func skipQuoteAt(raw string, pos int) (int, bool) {
	ch := raw[pos]
	if ch == '\'' {
		return scanSingleQuote(raw, pos), true
	}
	if ch == '"' {
		return scanDoubleQuote(raw, pos), true
	}
	return pos, false
}

// skipCommentAt 检查并跳过块注释（/* */）或单行注释（-- 和 #）
func skipCommentAt(raw string, pos int) (int, bool) {
	ch := raw[pos]
	if ch == '/' && pos+1 < len(raw) && raw[pos+1] == '*' {
		return skipBlockComment(raw, pos), true
	}
	if (ch == '-' && pos+1 < len(raw) && raw[pos+1] == '-') || ch == '#' {
		return findEndOfLine(raw, pos), true
	}
	return pos, false
}

// SQLType 根据输入的原始 SQL 字符串，准确分析并推导出对应的 SQL 分类类型（DDL、DML、DQL、TTL、DCL）。
//
// 处理流程：
// 1. 自动剥离前导多行/单行注释及空白字符，提取出首个有效关键字。
// 2. 优先匹配 DCL 专属关键字（GRANT、REVOKE、SET 等）。
// 3. 匹配 DDL 关键字（DROP、ALTER、CREATE 等），并进一步通过正则甄别 CREATE USER/ROLE 等特殊 DCL。
// 4. 匹配 DML 关键字（INSERT、UPDATE、DELETE、CALL、MERGE 等）。
// 5. 匹配 TTL 事务控制关键字（COMMIT、ROLLBACK 等）。
// 6. 对 WITH 子句深入解析主操作动词；默认兜底返回 DQL（覆盖 SELECT、SHOW、EXPLAIN 等）。
func SQLType(raw string) SQLTYPE {
	// 步骤 1：剥离前导注释与空白，获得纯净的起始代码
	clean := stripLeadingComments(raw)
	if clean == "" {
		return DQL
	}

	// 步骤 2：提取首个关键字（大写）
	verb := getLeadingWord(clean)

	// 步骤 3：精确匹配 DCL
	if dclExactVerbs[verb] {
		return DCL
	}

	// 步骤 4：匹配 DDL 与特殊 DCL（如 CREATE USER / ALTER ROLE）
	if ddlVerbs[verb] {
		// 重要判断：针对 CREATE USER/ROLE/DATABASE 等语句，重定向归类为 DCL
		if istrDCL(clean) {
			return DCL
		}
		return DDL
	}

	// 步骤 5：匹配 DML（含 INSERT INTO ... SELECT 复核语句）
	if dmlVerbs[verb] {
		return DML
	}

	// 步骤 6：匹配 TTL（事务控制）
	if ttlVerbs[verb] {
		return TTL
	}

	// 步骤 7：特殊处理 WITH 通用表表达式
	if verb == "WITH" {
		return determineCTEType(clean)
	}

	// 默认返回 DQL（包含常规 SELECT、SELECT ... FOR UPDATE 复核语句等）
	return DQL
}
