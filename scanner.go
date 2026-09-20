package sqlsplit

import (
	"strings"
)

// sqlScanner 是整个 SQL 切分引擎的核心扫描器结构体，负责流式解析输入脚本并维护语法状态机。
type sqlScanner struct {
	// input 为待切分的完整 SQL 脚本文本内容
	input string
	// pos 为当前扫描指针在 input 中的字节索引位置
	pos int
	// delim 为当前活动的语句切分分隔符，默认初始为分号 ";"，可被 MySQL 的 DELIMITER 指令动态修改
	delim string
	// pendingComments 用于暂存语句尚未开始前遇到的注释内容（单行/多行注释），将在遇到首个有效 SQL 词时附着到该 SQL 前面
	pendingComments string
	// currSQL 用于累积当前正在提取的 SQL 语句文本
	currSQL strings.Builder
	// isBlock 标记当前语句是否为复合过程块（如 CREATE PROCEDURE/FUNCTION/TRIGGER/PACKAGE 或 DECLARE/BEGIN 块）
	isBlock bool
	// inDecl 标记当前过程块是否处于变量/游标声明区（如 Oracle 中 AS/IS 与 BEGIN 之间），处于该状态时内部的分号属于声明语句，不能切分
	inDecl bool
	// hasBegun 标记过程块是否已经进入了主体代码段（即已遇到首个外层 BEGIN）
	hasBegun bool
	// blockDepth 记录当前过程块内 BEGIN...END、IF...END IF、LOOP...END LOOP、CASE...END CASE 的嵌套深度
	blockDepth int
	// justSawEnd 标记上一个非空白词是否为 END，用于防止紧跟其后的 IF/LOOP/CASE/REPEAT（如 END IF、END LOOP）被误识别为开启新块
	justSawEnd bool
	// lastWord 记录最近读取的一个大写关键字
	lastWord string
	// isPkgBody 标记当前复合块是否为 Oracle PACKAGE BODY
	isPkgBody bool
	// leadingTokens 收集当前语句开头的若干个关键字（最多8个），用于判断该语句是否属于存储过程/触发器/函数等复合块
	leadingTokens []string
	// results 存储切分产出的有效完整 SQL 语句列表
	results []SqlParse
}

// newScanner 创建并初始化一个具有默认分号分隔符的 SQL 扫描器实例
func newScanner(input string) *sqlScanner {
	return &sqlScanner{
		input: input,
		delim: ";",
	}
}

// parseSQLScript 为扫描引擎的主驱动入口函数。
// 循环调用 step() 进行状态推演与字符消费，直到文本结束并完成收尾。
func parseSQLScript(sqls string) []SqlParse {
	s := newScanner(sqls)
	// 重要判断：逐步步进扫描，直到输入文本消费完毕
	for s.step() {
	}
	// 扫描结束时处理末尾未闭合的语句（如最后一条语句没有以分号结尾）
	s.finish()
	if s.results == nil {
		return []SqlParse{}
	}
	return s.results
}

// finish 在扫描到达文件末尾时执行收尾检查。
// 若最后一条语句没有以分号结尾但存在有效 SQL 内容，则予以提取产出；若仅为孤立注释，则安全丢弃。
func (s *sqlScanner) finish() {
	// 重要判断：仅当累积的 SQL 包含非空白实际字符时才输出，避免将文件尾部的无归属孤立注释当成假 SQL 输出
	if strings.TrimSpace(s.currSQL.String()) != "" {
		s.emitStatement()
	}
}

// isAtLineStart 检查指定位置 pos 是否位于所在行的行首（即前面仅有空白字符，或直接处于换行符之后）。
// 用于精确识别 DELIMITER 客户端指令与 Oracle 单独行斜杠 /。
func isAtLineStart(input string, pos int) bool {
	// 重要判断：整个文本的起始位置即为第一行的行首
	if pos == 0 {
		return true
	}
	// 向前反向扫描前一个字符，若遇到换行符则说明正处于新行行首
	for i := pos - 1; i >= 0; i-- {
		if input[i] == '\n' {
			return true
		}
		// 重要判断：若在遇到换行前先遇到了非空白字符，说明当前位置并非行首
		if input[i] != ' ' && input[i] != '\t' && input[i] != '\r' {
			return false
		}
	}
	return true
}

// skipSpacesOnLine 跳过当前行内的水平空白字符（空格与水平制表符），遇到换行符或非空白字符时停止
func skipSpacesOnLine(input string, pos int) int {
	for pos < len(input) && (input[pos] == ' ' || input[pos] == '\t' || input[pos] == '\r') {
		pos++
	}
	return pos
}

// findEndOfLine 从当前位置向后扫描直到遇到换行符（\n），返回换行符之后的位置（若跨越换行）或文本末尾
func findEndOfLine(input string, pos int) int {
	for pos < len(input) && input[pos] != '\n' {
		pos++
	}
	// 重要判断：若因遇到换行符结束，则跨过该换行符定位到下一行开始
	if pos < len(input) && input[pos] == '\n' {
		pos++
	}
	return pos
}

// matchesKeywordAt 在指定偏移 pos 处以不区分大小写的方式匹配目标关键字 kw，并确保其后具有词边界（如空格、分号、换行或末尾）
func matchesKeywordAt(input string, pos int, kw string) bool {
	n := len(kw)
	// 重要判断：剩余长度不足或字符串不匹配时返回 false
	if pos+n > len(input) || !strings.EqualFold(input[pos:pos+n], kw) {
		return false
	}
	// 重要判断：若正好匹配到文本末尾，属于有效词边界
	if pos+n == len(input) {
		return true
	}
	next := input[pos+n]
	// 重要判断：验证下一个字符必须为空白字符、分号或换行，防止前缀误匹配（如 DELIMITER_EXTRA）
	return next <= ' ' || next == ';' || next == '\n'
}

// checkDelimiterDirective 检查当前行首是否存在 MySQL 的 DELIMITER 指令（如 DELIMITER // 或 DELIMITER ;）。
// 若存在，提取出新的分隔符字符串以及指令结束后的下一字符位置。
func (s *sqlScanner) checkDelimiterDirective() (string, int, bool) {
	// 重要判断：DELIMITER 指令必须位于行首（允许前置水平空白）
	if !isAtLineStart(s.input, s.pos) {
		return "", s.pos, false
	}
	p := skipSpacesOnLine(s.input, s.pos)
	// 重要判断：检查是否匹配 DELIMITER 关键字
	if !matchesKeywordAt(s.input, p, "DELIMITER") {
		return "", s.pos, false
	}
	// 定位该指令所在的行末位置并截取新指定的分隔符
	lineEnd := findEndOfLine(s.input, p+9)
	newDelim := strings.TrimSpace(s.input[p+9 : lineEnd])
	return newDelim, lineEnd, true
}

// handleDelimiterDirective 消费并处理当前遇到的 DELIMITER 切换指令。
//
// 设计背景：
// DELIMITER 是 MySQL 客户端工具指令，并非服务端合法 SQL。
// 若将其发往数据库执行，会直接报错 "1064: You have an error in your SQL syntax"。
// 因此扫描引擎需在行首静默消费该指令以更新活动切分符，同时不生成任何 SqlParse 条目。
func (s *sqlScanner) handleDelimiterDirective() bool {
	newDelim, nextPos, ok := s.checkDelimiterDirective()
	if !ok {
		return false
	}
	// 重要判断：若成功提取出非空的新分隔符，则将其应用为当前切分符
	if newDelim != "" {
		s.delim = newDelim
	}
	s.pos = nextPos
	return true
}

// isStandaloneSlash 检查当前位置是否为 Oracle 独立成行的斜杠 / 结束标记。
//
// 设计背景：
// 在 Oracle PL/SQL 脚本（SQL*Plus 规范）中，存储过程、触发器或匿名块经常使用独立占一行的 / 作为执行并提交当前块的结束标记。
func isStandaloneSlash(input string, pos int) (int, bool) {
	// 重要判断：字符必须是 '/' 且必须位于行首
	if pos >= len(input) || input[pos] != '/' || !isAtLineStart(input, pos) {
		return pos, false
	}
	// 检查 '/' 之后直到行末是否全为空白字符
	p := skipSpacesOnLine(input, pos+1)
	if p == len(input) {
		return p, true
	}
	// 重要判断：'/' 后面紧跟换行符，确认其为独立行斜杠
	if input[p] == '\n' {
		return p + 1, true
	}
	return pos, false
}

// tryHandleSlash 在合适时机触发独立行斜杠切分（非过程块且已有累积语句时不切分，避免误伤除号）
func (s *sqlScanner) tryHandleSlash() bool {
	if !s.isBlock && strings.TrimSpace(s.currSQL.String()) != "" {
		return false
	}
	if nextPos, ok := isStandaloneSlash(s.input, s.pos); ok {
		s.handleSlash(nextPos)
		return true
	}
	return false
}

// handleSlash 处理独立行斜杠 / 结束标记，若当前有累积的未结束语句，则触发其闭合并输出
func (s *sqlScanner) handleSlash(nextPos int) {
	s.pos = nextPos
	// 重要判断：若当前已累积有实际 SQL 语句，独立行斜杠触发该语句的切分闭合
	if strings.TrimSpace(s.currSQL.String()) != "" {
		s.emitStatement()
	}
}

// scanCommentAt 检测并扫描位于当前位置的注释（支持单行 --、# 与块注释 /* ... */），返回注释结束后的偏移位置
func scanCommentAt(input string, pos int) (int, bool) {
	if pos >= len(input) {
		return pos, false
	}
	// 重要判断：检测 MySQL 专有单行注释符 #
	if input[pos] == '#' {
		return findEndOfLine(input, pos+1), true
	}
	if pos+1 < len(input) {
		// 重要判断：检测标准 SQL 单行注释符 --
		if input[pos] == '-' && input[pos+1] == '-' {
			return findEndOfLine(input, pos+2), true
		}
		// 重要判断：检测多行/块注释符 /*
		if input[pos] == '/' && input[pos+1] == '*' {
			return skipBlockComment(input, pos), true
		}
	}
	return pos, false
}

// tryScanComment 尝试识别并提取当前位置的注释内容。
// 若当前语句尚未开始，则将注释追加到 pendingComments 以便作为前缀绑定到下一条语句；
// 若当前语句已在读取中，则直接作为语句内部注释写入当前 SQL 缓冲区。
func (s *sqlScanner) tryScanComment() bool {
	nextPos, ok := scanCommentAt(s.input, s.pos)
	if !ok {
		return false
	}
	comment := s.input[s.pos:nextPos]
	// 重要判断：若当前语句已有内容，注释写入语句体内部；否则作为前置注释留存
	if s.currSQL.Len() > 0 {
		s.currSQL.WriteString(comment)
	} else {
		s.pendingComments += comment
	}
	s.pos = nextPos
	return true
}

// isDollarTagChar 检查指定字节是否属于 PostgreSQL Dollar-quote 标签字符（字母、数字或下划线）
func isDollarTagChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

// matchDollarTag 匹配形如 $$ 或 $tag$ 的 PostgreSQL 块引用标签名，若匹配成功返回完整标签（含两侧的 $）
func matchDollarTag(input string, pos int) string {
	if pos >= len(input) || input[pos] != '$' {
		return ""
	}
	i := pos + 1
	for i < len(input) && isDollarTagChar(input[i]) {
		i++
	}
	// 重要判断：标签必须以另一个 $ 闭合，且标签长度一般在合理范围内（<= 64），避免把非标签变量误判
	if i < len(input) && input[i] == '$' && i-pos <= 64 {
		return input[pos : i+1]
	}
	return ""
}

// scanDollarQuoteAt 扫描 PostgreSQL 的 Dollar-quote 块（$$ ... $$ 或 $tag$ ... $tag$）。
//
// 设计背景：
// PostgreSQL 中几乎所有的 PL/pgSQL 存储过程与函数定义都将函数体包裹在 $$ 中。
// 处于该代码块内部的所有分号、引号均属于函数体字面量，绝不能触发切分，直到遇到成对的闭合标签。
func scanDollarQuoteAt(input string, pos int) (int, bool) {
	tag := matchDollarTag(input, pos)
	if tag == "" {
		return pos, false
	}
	// 寻找匹配的闭合 tag
	closeIdx := strings.Index(input[pos+len(tag):], tag)
	// 重要判断：若未找到匹配闭合 tag，扫描至末尾防止越界
	if closeIdx == -1 {
		return len(input), true
	}
	return pos + len(tag) + closeIdx + len(tag), true
}

// isQuoteEndContext 判定字符是否处于闭合单引号的合法外部上下文
func isQuoteEndContext(b byte) bool {
	return b <= ' ' || b == ';' || b == ',' || b == ')'
}

// checkEscapedQuote 检查反斜杠后紧随的单引号是否为路径末尾的闭合引号
func checkEscapedQuote(input string, i int) (int, bool) {
	if input[i+1] == '\'' && (i+2 >= len(input) || isQuoteEndContext(input[i+2])) {
		return i + 2, true
	}
	return i + 2, false
}

// scanSingleQuote 扫描单引号字符串字面量（'...'），妥善处理转义单引号（\' 与 ''），防止将字符串内部的分号当成切分符
func scanSingleQuote(input string, pos int) int {
	i := pos + 1
	for i < len(input) {
		// 重要判断：跳过反斜杠转义字符 \'
		if input[i] == '\\' && i+1 < len(input) {
			nextI, isEnd := checkEscapedQuote(input, i)
			if isEnd {
				return nextI
			}
			i = nextI
			continue
		}
		if input[i] == '\'' {
			// 重要判断：跳过 SQL 标准双单引号转义 ''
			if i+1 < len(input) && input[i+1] == '\'' {
				i += 2
				continue
			}
			return i + 1 // 找到合法闭合单引号
		}
		i++
	}
	return len(input)
}

// scanDoubleQuote 扫描双引号标识符或字符串（"..."），妥善处理 \" 与 "" 转义
func scanDoubleQuote(input string, pos int) int {
	i := pos + 1
	for i < len(input) {
		// 重要判断：跳过反斜杠转义 \"
		if input[i] == '\\' && i+1 < len(input) {
			i += 2
			continue
		}
		if input[i] == '"' {
			// 重要判断：跳过双双引号转义 ""
			if i+1 < len(input) && input[i+1] == '"' {
				i += 2
				continue
			}
			return i + 1 // 找到合法闭合双引号
		}
		i++
	}
	return len(input)
}

// scanBacktick 扫描 MySQL 反引号标识符（`...`），妥善处理包含分号的表名/字段名（如 `db;1`.`col;name`）
func scanBacktick(input string, pos int) int {
	i := pos + 1
	for i < len(input) {
		// 重要判断：跳过反斜杠转义 \`
		if input[i] == '\\' && i+1 < len(input) {
			i += 2
			continue
		}
		if input[i] == '`' {
			// 重要判断：跳过双反引号转义 ``
			if i+1 < len(input) && input[i+1] == '`' {
				i += 2
				continue
			}
			return i + 1 // 找到合法闭合反引号
		}
		i++
	}
	return len(input)
}

// scanQuoteAt 根据当前引号字符（'、"、`）分发调用对应的引号扫描函数
func scanQuoteAt(input string, pos int) (int, bool) {
	if pos >= len(input) {
		return pos, false
	}
	q := input[pos]
	if q == '\'' {
		return scanSingleQuote(input, pos), true
	}
	if q == '"' {
		return scanDoubleQuote(input, pos), true
	}
	if q == '`' {
		return scanBacktick(input, pos), true
	}
	return pos, false
}

// handleLiteralAt 尝试匹配并处理字符串字面量、反引号标识符或 PostgreSQL Dollar-quote 块。
//
// 设计背景：
// PostgreSQL 函数语法形如：CREATE FUNCTION ... AS $$ <body> $$ LANGUAGE plpgsql;
// 函数体在 AS 之后完全封闭在 $$ 块内部。因此一旦 Dollar 块扫描完毕，声明区已彻底结束，
// 必须将 inDecl 明确重置为 false，确保函数末尾的分号能够正常切分。
func (s *sqlScanner) handleLiteralAt() bool {
	// 重要判断：优先尝试匹配 PostgreSQL Dollar-quote（$$ 或 $tag$）
	if nextPos, ok := scanDollarQuoteAt(s.input, s.pos); ok {
		s.ensureStatementPrefix()
		s.currSQL.WriteString(s.input[s.pos:nextPos])
		s.pos = nextPos
		// 重要设计：Dollar 块扫描完成即意味着过程体结束，不再处于声明区状态
		if s.isBlock {
			s.inDecl = false
		}
		return true
	}
	// 重要判断：匹配普通单引号、双引号与 MySQL 反引号
	if nextPos, ok := scanQuoteAt(s.input, s.pos); ok {
		s.ensureStatementPrefix()
		s.currSQL.WriteString(s.input[s.pos:nextPos])
		s.pos = nextPos
		// 若处于存储过程且单引号紧跟 AS（如 PG 函数体 AS 'SELECT ...'），清除声明区状态
		if s.isBlock && s.lastWord == "AS" {
			s.inDecl = false
		}
		return true
	}
	return false
}

// isAtDelimiter 检查当前扫描指针处是否与当前活动切分符（s.delim，默认 ;）完全匹配
func (s *sqlScanner) isAtDelimiter() bool {
	if len(s.delim) == 0 || s.pos+len(s.delim) > len(s.input) {
		return false
	}
	return strings.EqualFold(s.input[s.pos:s.pos+len(s.delim)], s.delim)
}

// canTerminateBlock 判定当前复合代码块（存储过程/函数）是否具备闭合条件。
//
// 设计背景：
// 1. 若处于声明区（!hasBegun && inDecl，即 AS/IS 与 BEGIN 之间），内部的变量声明分号（如 v_cnt NUMBER;）绝不能结束过程。
// 2. 若嵌套深度 blockDepth > 0（内部尚未完全配对 END），也不能结束过程。
// 仅当深度归零且不在声明区时，分号才标志整个存储过程的结束。
func (s *sqlScanner) canTerminateBlock() bool {
	// 重要判断：处于 Oracle 变量声明区时，分号为变量声明语句分隔符，不能终止过程
	if !s.hasBegun && s.inDecl {
		return false
	}
	// 重要判断：嵌套块未平衡（如内部的 BEGIN、IF、LOOP 尚未遇到对应的 END），不能终止过程
	if s.blockDepth > 0 {
		return false
	}
	return true
}

// handleDelimiter 处理匹配到活动切分符时的切分逻辑
func (s *sqlScanner) handleDelimiter() {
	// 分支 1：处于存储过程/复合块中
	if s.isBlock {
		if !s.canTerminateBlock() {
			// 重要判断：若过程尚未达到终止条件，该分号属于过程内部代码，保留并继续累积
			s.currSQL.WriteString(s.delim)
			s.pos += len(s.delim)
			return
		}
		// 重要判断：若过程达到闭合条件，对于标准分号完整保留其尾部分号；若为自定义切分符（如 //）则不追加
		if s.delim == ";" {
			s.currSQL.WriteString(s.delim)
		}
		s.pos += len(s.delim)
		s.emitStatement()
		return
	}

	// 分支 2：普通 SQL 语句，若有累积实际内容则触发闭合
	if strings.TrimSpace(s.currSQL.String()) != "" {
		s.pos += len(s.delim)
		s.emitStatement()
		return
	}

	// 分支 3：连续分号（如 ;;;）或前导空分号，直接跨过，不产出空对象
	s.pos += len(s.delim)
}

// scanWord 从当前位置向后扫描连续的单词字符（英文字母、数字、下划线）
func (s *sqlScanner) scanWord() (string, int) {
	p := s.pos
	for p < len(s.input) && isWordChar(s.input[p]) {
		p++
	}
	return s.input[s.pos:p], p
}

// blockKeywords 维护能开启复合过程块的关键字集合
var blockKeywords = map[string]bool{
	"PROCEDURE": true,
	"FUNCTION":  true,
	"TRIGGER":   true,
	"PACKAGE":   true,
	"EVENT":     true,
	"BODY":      true,
}

// nonBlockKeywords 维护虽然以 CREATE 开头但属于普通单分号 DDL 的对象关键字集合（如视图、表、索引）
var nonBlockKeywords = map[string]bool{
	"VIEW":     true,
	"TABLE":    true,
	"INDEX":    true,
	"DATABASE": true,
	"SCHEMA":   true,
}

// checkBlockStart 分析语句开头的若干 Token，判定当前语句是否为存储过程等复合块。
//
// 判断规则：
// 1. 以 DECLARE 或 BEGIN 开头：直接认定为匿名过程块。
// 2. 以 CREATE 开头：检查后续关键字中是否存在 PROCEDURE/FUNCTION/TRIGGER/PACKAGE 等；
//    若先遇到了 VIEW/TABLE/INDEX 等，则明确排除，认定为普通 DDL。
func checkBlockStart(tokens []string) bool {
	if len(tokens) == 0 {
		return false
	}
	first := tokens[0]
	// 重要判断：直接以 DECLARE 或 BEGIN 开头的属于匿名块
	if first == "DECLARE" || first == "BEGIN" {
		return true
	}
	if first != "CREATE" {
		return false
	}
	for i := 1; i < len(tokens); i++ {
		t := tokens[i]
		// 重要判断：遇到存储过程核心动词，判定为复合代码块
		if blockKeywords[t] {
			return true
		}
		// 重要判断：遇到视图、表等普通 DDL 关键字，判定为非复合块
		if nonBlockKeywords[t] {
			return false
		}
	}
	return false
}

// handleControlWord 处理过程块内部具有匹配 END 的控制流关键字（IF、LOOP、CASE、REPEAT）。
//
// 设计背景：
// 在 PL/SQL 中，语法为：END IF; 或 END LOOP;。
// 当刚遇到 END 时（s.justSawEnd 为 true），紧随其后的 IF/LOOP 仅为配对标记，绝不能当作新块开启并增加深度。
func (s *sqlScanner) handleControlWord(word string) {
	// 重要判断：紧跟在 END 后的控制关键字属于标签修饰符，消费该标记并不递增深度
	if s.justSawEnd {
		s.justSawEnd = false
		return
	}
	s.blockDepth++
}

// handleEndWord 处理 END 关键字引起的块嵌套深度递减与声明区状态流转。
//
// 设计背景：
// 1. 若深度递减至 0 且处于 Oracle 包体（isPkgBody），说明内部子过程（PROCEDURE/FUNCTION）结束，
//    应恢复包体顶层声明区状态（inDecl = true, hasBegun = false），防止子过程的分号将包体截断。
// 2. 若在深度已为 0 的声明区再次遇到 END（包体自身的 END），则解除声明区（inDecl = false），允许包体末尾分号正常闭合。
func (s *sqlScanner) handleEndWord() {
	if s.blockDepth > 0 {
		s.blockDepth--
		if s.isPkgBody && s.blockDepth == 0 {
			s.inDecl = true
			s.hasBegun = false
		}
	} else if s.inDecl {
		s.inDecl = false
	}
	s.justSawEnd = true
}

// updateBlockState 根据当前读取到的关键字推进并更新复合过程块的语法状态机
func (s *sqlScanner) updateBlockState(word string) {
	upper := strings.ToUpper(word)
	switch upper {
	case "AS", "IS":
		// 重要判断：在主体 BEGIN 之前遇到 AS 或 IS，标志进入 Oracle 变量/游标声明区
		if !s.hasBegun {
			s.inDecl = true
		}
	case "BEGIN":
		// 重要判断：遇到 BEGIN 标志着声明区结束，进入主体执行段，嵌套深度自增
		s.inDecl = false
		s.hasBegun = true
		s.blockDepth++
		s.justSawEnd = false
	case "END":
		s.handleEndWord()
	case "IF", "LOOP", "CASE", "REPEAT":
		s.handleControlWord(upper)
	default:
		s.justSawEnd = false
	}
}

// ensureStatementPrefix 当语句首次写入实际有效内容时，将之前累积的前置注释（pendingComments）作为前缀写入当前 SQL
func (s *sqlScanner) ensureStatementPrefix() {
	if s.currSQL.Len() == 0 && len(s.pendingComments) > 0 {
		s.currSQL.WriteString(s.pendingComments)
		s.pendingComments = ""
	}
}

// hasPkgBody 检查前导关键字中是否同时包含 PACKAGE 和 BODY
func hasPkgBody(tokens []string) bool {
	hasPkg, hasBody := false, false
	for _, t := range tokens {
		if t == "PACKAGE" {
			hasPkg = true
		} else if t == "BODY" {
			hasBody = true
		}
	}
	return hasPkg && hasBody
}

// onWordRead 在扫描到一个常规单词时触发状态流转：
// 1. 将前置注释附着并累积当前单词。
// 2. 收集语句前缀 Token 并触发是否为过程块（checkBlockStart）判定。
// 3. 若为过程块，驱动其内部状态机演进（updateBlockState）。
func (s *sqlScanner) onWordRead(word string) {
	s.ensureStatementPrefix()
	s.currSQL.WriteString(word)
	upper := strings.ToUpper(word)
	s.lastWord = upper

	// 收集前导关键字用于复合块首部判定
	if len(s.leadingTokens) < 8 {
		s.leadingTokens = append(s.leadingTokens, upper)
		if !s.isBlock && checkBlockStart(s.leadingTokens) {
			s.isBlock = true
			if s.leadingTokens[0] == "BEGIN" {
				s.hasBegun = true
			} else if s.leadingTokens[0] == "DECLARE" {
				s.inDecl = true
			}
		}
		if s.isBlock && !s.isPkgBody && upper == "BODY" {
			s.isPkgBody = hasPkgBody(s.leadingTokens)
		}
	}

	// 若处于过程块中，依据当前词推进状态
	if s.isBlock {
		s.updateBlockState(word)
	}
}

// onOtherChar 处理非单词字符（如空白、括号、逗号、操作符等）。
//
// 设计背景：
// 若上一词为 END，且后续遇到任何非空白标点（如括号 )），必须立即重置 justSawEnd 为 false。
// 例如在 FOR X IN (SELECT CASE ... END) LOOP 场景中，END 后面紧跟 )，
// 紧跟的右括号说明该 END 并非 END LOOP，其后的 LOOP 必须作为正规循环正常递增 blockDepth。
func (s *sqlScanner) onOtherChar(b byte) {
	// 重要判断：空白字符不破坏 justSawEnd 状态（因为 END IF 之间可能有空格/换行）
	if b <= ' ' {
		if s.currSQL.Len() > 0 {
			s.currSQL.WriteByte(b)
		}
		s.pos++
		return
	}
	// 重要判断：遇到非空白标点符号（如 )、,、;），清除 justSawEnd 标记
	s.justSawEnd = false
	s.ensureStatementPrefix()
	s.currSQL.WriteByte(b)
	s.pos++
}

// emitStatement 组装并将当前提取出的单条完整 SQL 输出到结果集合中。
//
// 规范化处理：
// 1. 普通 SQL：调用 RemoveLastSemicolon 剥离尾部分号（兼容 Oracle 普通语句执行限制）。
// 2. 存储过程：完整保留结尾分号。
// 3. MySQL 自定义分隔符（如 //）：剥离末尾的临时分隔符。
// 4. 调用 SQLType 精确推导语句类型。
func (s *sqlScanner) emitStatement() {
	raw := s.currSQL.String()
	s.currSQL.Reset()
	s.leadingTokens = nil

	sqlStr := strings.TrimSpace(raw)
	if sqlStr == "" {
		return
	}
	// 重要判断：对于非复合块的普通 SQL 语句，剥离最后一个尾部分号
	if !s.isBlock {
		sqlStr = RemoveLastSemicolon(sqlStr)
	}
	// 重要判断：若使用了非分号的自定义切分符（如 MySQL //），剔除末尾多余的分隔符标记
	if s.delim != ";" {
		sqlStr = strings.TrimSuffix(sqlStr, s.delim)
		sqlStr = strings.TrimSpace(sqlStr)
	}
	if sqlStr == "" {
		return
	}

	// 自动推断 SQL 类型并加入结果切片
	s.results = append(s.results, SqlParse{
		SQL:  sqlStr,
		Type: SQLType(sqlStr),
	})

	// 重置过程块状态
	s.isBlock = false
	s.isPkgBody = false
	s.lastWord = ""
	s.inDecl = false
	s.hasBegun = false
	s.blockDepth = 0
	s.justSawEnd = false
}

// step 为扫描器的单步步进方法，按照语法优先级依次探测并消费输入流。
//
// 探测优先级：
// 1. 行首 MySQL DELIMITER 指令。
// 2. 行首 Oracle 独立行斜杠 /。
// 3. 单行与多行注释。
// 4. PostgreSQL Dollar 块或普通字符串/反引号。
// 5. 活动切分符。
// 6. 单词。
// 7. 空白及其他字符。
func (s *sqlScanner) step() bool {
	if s.pos >= len(s.input) {
		return false
	}
	// 1. 探测 DELIMITER 客户端指令
	if s.handleDelimiterDirective() {
		return true
	}
	// 2. 探测独立行斜杠
	if s.tryHandleSlash() {
		return true
	}
	// 3. 探测单行与块注释
	if s.tryScanComment() {
		return true
	}
	// 4. 探测 Dollar-quote 与字符串字面量
	if s.handleLiteralAt() {
		return true
	}
	// 5. 探测活动语句切分符
	if s.isAtDelimiter() {
		s.handleDelimiter()
		return true
	}
	// 6. 探测连续词字符
	if isWordChar(s.input[s.pos]) {
		word, nextPos := s.scanWord()
		s.onWordRead(word)
		s.pos = nextPos
		return true
	}
	// 7. 处理空白与其他字符
	s.onOtherChar(s.input[s.pos])
	return true
}
