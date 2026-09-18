package sqlsplit

import (
	"fmt"
	"strings"
)

// Pick 为基于词语流的状态拾取器结构体，维护当前拾取的 SQL 文本与状态机上下文
type Pick struct {
	// sql 记录当前正在累积的 SQL 字符串片段
	sql string
	// nowmode 记录当前所处的解析模式（Mode）
	nowmode Mode
	// keypop 标识是否刚刚弹出了关键字
	keypop bool
	// keystack 为关键字嵌套栈，用于配对 begin-end、loop-end、if-end 等结构
	keystack *ItemStack
	// modestack 为模式状态栈，用于在退出单引号、双引号、注释等临时模式后恢复上一层模式
	modestack *ItemStack
}

// Reset 重置拾取器的状态为初始未拾取（ModeUnPick）状态并清空内部栈
func (p *Pick) Reset() {
	p.sql = ""
	p.nowmode = ModeUnPick
	p.keystack = NewStack()
	p.modestack = NewStack()
}

// modeChecker 定义模式检查函数的签名，接收当前小写词与后缀空白，返回计算出的新模式
type modeChecker func(p *Pick, word, space string) Mode

// modeCheckers 维护从 Mode 到对应状态转移检查函数的映射表，有效消除长 switch 语句以控制圈复杂度
var modeCheckers = map[Mode]modeChecker{
	ModeUnPick:          (*Pick).unpickCheck,
	ModeRemarkLine:      (*Pick).remarkLineCheck,
	ModeRemarkMoreLine:  (*Pick).remarkMoreLineCheck,
	ModeDefaultSql:      (*Pick).defaultSqlCheck,
	ModeApostrophe:      (*Pick).apostropheCheck,
	ModeDoubleQuotes:    (*Pick).doubleQuotesCheck,
	ModeMaybeProcedure1: (*Pick).maybeProcedure1Check,
	ModeMaybeProcedure2: (*Pick).maybeProcedure2Check,
	ModeMaybeProcedure3: (*Pick).maybeProcedure3Check,
	ModeProcedure:       (*Pick).procedureCheck,
}

// Pick 逐词输入 word 及其后缀空白 space，驱动状态机流转；当某条 SQL 语句匹配闭合完成时返回 over = true
func (p *Pick) Pick(word, space string) (over bool) {
	newmode := p.nowmode
	lower := strings.ToLower(word)
	// 重要判断：通过映射表分发调用当前模式专属的检查逻辑
	if fn, ok := modeCheckers[p.nowmode]; ok {
		newmode = fn(p, lower, space)
	}
	p.sql = fmt.Sprintf("%v%v%v", p.sql, word, space)

	// 重要判断：如果计算得出的新模式回到了未拾取状态（ModeUnPick）
	if newmode == ModeUnPick {
		// 若模式栈已空，说明最外层的 SQL 语句已完整结束
		if p.modestack.Len() == 0 {
			return true
		}
		// 模式栈非空，弹出恢复上一层被压栈的父模式
		newmode, _ = p.modestack.Pop().(Mode)
		p.nowmode = newmode
	} else if p.nowmode != newmode {
		// 模式发生了状态跃迁，更新当前模式
		p.nowmode = newmode
	}
	return false
}

// unpickCheck 在未拾取状态下检查新词，决定进入何种 SQL 模式
func (p *Pick) unpickCheck(word, _ string) (newMode Mode) {
	// 空词保持未拾取状态
	if len(word) == 0 {
		return ModeUnPick
	}
	// 重要判断：以 create 开头，进入潜在存储过程/函数分析模式
	if word == "create" {
		return ModeMaybeProcedure1
	}
	// 重要判断：检查是否进入单引号或双引号字符串状态
	if newMode, picked := p.quotationCheck(word); picked {
		return newMode
	}
	// 重要判断：检查是否进入注释状态（--、# 或 /*）
	if newMode, picked := p.remarkCheck(word); picked {
		return newMode
	}
	// 默认为普通单分号 SQL 模式
	return ModeDefaultSql
}

// isProducer 检查单词是否属于代表存储过程或复杂代码块的对象类型（procedure/event/package/function）
func isProducer(word string) bool {
	return word == "procedure" || word == "event" || word == "package" || word == "function"
}

// maybeProcedure1Check 处理已识别 "create" 之后的下一个单词
func (p *Pick) maybeProcedure1Check(word, _ string) (newMode Mode) {
	// 重要判断：create 后面直接跟 procedure/function 等，确认进入存储过程模式
	if isProducer(word) {
		return ModeProcedure
	} else if word == "or" {
		// create or -> 进入状态 2
		return ModeMaybeProcedure2
	} else if word == ";" {
		// 提前遇到分号，重置
		return ModeUnPick
	}
	if newMode, picked := p.quotationCheck(word); picked {
		p.modestack.Push(p.nowmode)
		return newMode
	}
	if newMode, picked := p.remarkCheck(word); picked {
		p.modestack.Push(p.nowmode)
		return newMode
	}
	return ModeDefaultSql
}

// maybeProcedure2Check 处理已识别 "create or" 之后的下一个单词
func (p *Pick) maybeProcedure2Check(word, _ string) (newMode Mode) {
	// 重要判断：create or replace -> 进入状态 3
	if word == "replace" {
		return ModeMaybeProcedure3
	} else if word == ";" {
		return ModeUnPick
	}
	if newMode, picked := p.quotationCheck(word); picked {
		p.modestack.Push(p.nowmode)
		return newMode
	}
	if newMode, picked := p.remarkCheck(word); picked {
		p.modestack.Push(p.nowmode)
		return newMode
	}
	return ModeDefaultSql
}

// maybeProcedure3Check 处理已识别 "create or replace" 之后的下一个单词
func (p *Pick) maybeProcedure3Check(word, space string) (newMode Mode) {
	// 重要判断：create or replace 后面跟 procedure/function 等，确认进入存储过程模式
	if isProducer(word) {
		return ModeProcedure
	}
	if newMode, picked := p.quotationCheck(word); picked {
		p.modestack.Push(p.nowmode)
		return newMode
	}
	if newMode, picked := p.remarkCheck(word); picked {
		p.modestack.Push(p.nowmode)
		return newMode
	}
	return ModeDefaultSql
}

// checkProcedureStack 检查存储过程关键字栈顶与当前词的配对情况
func (p *Pick) checkProcedureStack(word string) (Mode, bool) {
	topKey := p.keystack.Look()
	if topKey == nil {
		return p.nowmode, false
	}
	key, _ := topKey.(string)
	// 重要判断：栈顶期待分号 ; 闭合
	if key == ";" {
		if word == ";" {
			p.keystack.Pop()
			// 若关键字栈已完全清空，说明整个存储过程的所有块已闭合，模式结束
			if p.keystack.Len() == 0 {
				return ModeUnPick, true
			}
		}
		return p.nowmode, true
	}
	// 重要判断：栈顶期待 end 关键字闭合
	if key == "end" {
		if word == "end" {
			p.keystack.Pop()
		}
		return p.nowmode, true
	}
	// 可选匹配前缀（如 MAY:case）
	if strings.HasPrefix(key, "MAY:") {
		p.keystack.Pop()
		return p.nowmode, true
	}
	return p.nowmode, false
}

// checkProcedureKeyword 遇到进入代码块的关键字时，将相应的闭合期待压入关键字栈
func (p *Pick) checkProcedureKeyword(word string) {
	switch word {
	case "if", "loop", "begin", "repeat":
		p.keystack.Push(";")   // 期待一个分号结束整个结构
		p.keystack.Push("end") // 期待一个 end 标记该结构块闭合
	case "case":
		p.keystack.Push("MAY:case")
		p.keystack.Push("end")
	}
}

// procedureCheck 处于存储过程模式下处理内部各词，维护嵌套栈平衡
func (p *Pick) procedureCheck(word, space string) (newMode Mode) {
	// 引号与注释优先拦截保护
	if newMode, picked := p.quotationCheck(word); picked {
		p.modestack.Push(p.nowmode)
		return newMode
	}
	if newMode, picked := p.remarkCheck(word); picked {
		p.modestack.Push(p.nowmode)
		return newMode
	}
	// 检查栈顶与当前词的闭合匹配
	if m, handled := p.checkProcedureStack(word); handled {
		return m
	}
	// 检查是否有新的控制结构开启
	p.checkProcedureKeyword(word)
	return ModeProcedure
}

// remarkLineCheck 处理单行注释模式，遇到包含换行符的后缀空白时退出注释模式
func (p *Pick) remarkLineCheck(_, space string) (newMode Mode) {
	// 重要判断：单行注释在遇到换行符时结束
	if strings.Contains(space, "\n") {
		return ModeUnPick
	}
	return ModeRemarkLine
}

// remarkMoreLineCheck 处理多行块注释模式，遇到 */ 标记时退出注释模式
func (p *Pick) remarkMoreLineCheck(word, _ string) (newMode Mode) {
	// 重要判断：多行注释在遇到 */ 闭合符时结束
	if word == "*/" {
		return ModeUnPick
	}
	return ModeRemarkMoreLine
}

// defaultSqlCheck 处理普通 SQL 模式，遇到未在引号或注释内的独立分号时标志语句结束
func (p *Pick) defaultSqlCheck(word, _ string) (newMode Mode) {
	// 重要判断：分号结束当前普通 SQL 语句
	if word == ";" {
		return ModeUnPick
	}
	if newMode, picked := p.quotationCheck(word); picked {
		p.modestack.Push(p.nowmode)
		return newMode
	}
	if newMode, picked := p.remarkCheck(word); picked {
		p.modestack.Push(p.nowmode)
		return newMode
	}
	return ModeDefaultSql
}

// remarkCheck 检查单词是否代表注释开始（--、#、/*）
func (p *Pick) remarkCheck(word string) (newMode Mode, picked bool) {
	if word == "--" || word == "#" {
		return ModeRemarkLine, true
	} else if word == "/*" {
		return ModeRemarkMoreLine, true
	}
	return ModeUnPick, false
}

// quotationCheck 检查单词是否代表字符串引号开始（' 或 "）
func (p *Pick) quotationCheck(word string) (newMode Mode, picked bool) {
	if word == "'" {
		return ModeApostrophe, true
	} else if word == "\"" {
		return ModeDoubleQuotes, true
	}
	return ModeUnPick, false
}

// apostropheCheck 处理单引号模式，遇到闭合单引号时退出
func (p *Pick) apostropheCheck(word, _ string) Mode {
	if word == "'" {
		return ModeUnPick
	}
	return ModeApostrophe
}

// doubleQuotesCheck 处理双引号模式，遇到闭合双引号时退出
func (p *Pick) doubleQuotesCheck(word, _ string) Mode {
	if word == "\"" {
		return ModeUnPick
	}
	return ModeDoubleQuotes
}
