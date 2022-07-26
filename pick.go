package sqlsplit

import (
	"fmt"
	"strings"
)

type Pick struct {
	sql       string
	nowmode   Mode
	keypop    bool
	keystack  *ItemStack // 关键字栈
	modestack *ItemStack // 模式栈
}

func (p *Pick) Reset() {
	p.sql = ""
	p.nowmode = ModeUnPick
	p.keystack = NewStack()
	p.modestack = NewStack()
}

func (p *Pick) Pick(word, space string) (over bool) {
	newmode := p.nowmode
	switch p.nowmode {
	case ModeUnPick:
		newmode = p.unpickCheck(strings.ToLower(word), space)
	case ModeRemarkLine:
		newmode = p.remarkLineCheck(strings.ToLower(word), space)
	case ModeRemarkMoreLine:
		newmode = p.remarkMoreLineCheck(strings.ToLower(word), space)
	case ModeDefaultSql:
		newmode = p.defaultSqlCheck(strings.ToLower(word), space)
	case ModeApostrophe:
		newmode = p.apostropheCheck(strings.ToLower(word), space)
	case ModeDoubleQuotes:
		newmode = p.doubleQuotesCheck(strings.ToLower(word), space)
	case ModeMaybeProcedure1:
		newmode = p.maybeProcedure1Check(strings.ToLower(word), space)
	case ModeMaybeProcedure2:
		newmode = p.maybeProcedure2Check(strings.ToLower(word), space)
	case ModeMaybeProcedure3:
		newmode = p.maybeProcedure3Check(strings.ToLower(word), space)
	case ModeProcedure:
		newmode = p.procedureCheck(strings.ToLower(word), space)
	}
	p.sql = fmt.Sprintf("%v%v%v", p.sql, word, space)
	if newmode == ModeUnPick {
		// 上一个模式结束，如果栈空，则本次模式匹配完毕
		if p.modestack.Len() == 0 {
			// 上上一个模式为未匹配，则意味着本次匹配结束
			return true
		}
		newmode, _ = p.modestack.Pop().(Mode)
		p.nowmode = newmode
	} else if p.nowmode != newmode {
		p.nowmode = newmode
	}
	return false
}

func (p *Pick) unpickCheck(word, _ string) (newMode Mode) {
	if len(word) == 0 {
		return ModeUnPick
	}
	if word == "create" {
		return ModeMaybeProcedure1
	}
	if newMode, picked := p.quotationCheck(word); picked {
		return newMode
	}
	if newMode, picked := p.remarkCheck(word); picked {
		return newMode
	}
	return ModeDefaultSql
}

func (p *Pick) maybeProcedure1Check(word, _ string) (newMode Mode) {
	// create procedure
	// create or replace procedure
	if word == "procedure" || word == "event" || word == "package" {
		return ModeProcedure
	} else if word == "or" {
		return ModeMaybeProcedure2
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

func (p *Pick) maybeProcedure2Check(word, _ string) (newMode Mode) {
	// create procedure
	// create or replace procedure
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

func (p *Pick) maybeProcedure3Check(word, space string) (newMode Mode) {
	// create procedure
	// create or replace procedure
	if word == "procedure" || word == "event" || word == "package" {
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

func (p *Pick) procedureCheck(word, space string) (newMode Mode) {
	// create procedure xxx begin end
	if newMode, picked := p.quotationCheck(word); picked {
		p.modestack.Push(p.nowmode)
		return newMode
	}
	if newMode, picked := p.remarkCheck(word); picked {
		p.modestack.Push(p.nowmode)
		return newMode
	}
	if topKey := p.keystack.Look(); topKey != nil {
		key, _ := topKey.(string)
		if key == ";" {
			if word == ";" {
				p.keystack.Pop()
				if p.keystack.Len() == 0 {
					// 关键字栈顶空，则模式匹配完毕
					return ModeUnPick
				}
			}
			return p.nowmode
		} else if key == "end" {
			if word == "end" {
				p.keystack.Pop()
				return p.nowmode
			}
		} else if strings.HasPrefix(key, "MAY:") {
			// 可是，可不是
			p.keystack.Pop()
			if word == key[len("MAY:"):] {
				return p.nowmode
			}
		}
	}
	switch word {
	case "if", "loop", "begin":
		p.keystack.Push(";") // 等待一个 ; 结束end
		p.keystack.Push("end")
	case "case":
		p.keystack.Push("MAY:case")
		p.keystack.Push("end")
	}
	return ModeProcedure
}

func (p *Pick) remarkLineCheck(_, space string) (newMode Mode) {
	// --  xxx \n
	// # xxx \n
	if strings.Contains(space, "\n") {
		return ModeUnPick
	}
	return ModeRemarkLine
}

func (p *Pick) remarkMoreLineCheck(word, _ string) (newMode Mode) {
	// /*  xxx */
	if word == "*/" {
		return ModeUnPick
	}
	return ModeRemarkMoreLine
}

func (p *Pick) defaultSqlCheck(word, _ string) (newMode Mode) {
	// select * from xxx;
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

func (p *Pick) remarkCheck(word string) (newMode Mode, picked bool) {
	// select * from xxx;
	if word == "--" || word == "#" {
		return ModeRemarkLine, true
	} else if word == "/*" {
		return ModeRemarkMoreLine, true
	}
	return ModeUnPick, false
}

func (p *Pick) quotationCheck(word string) (newMode Mode, picked bool) {
	// ' "
	if word == "'" {
		return ModeApostrophe, true
	} else if word == "\"" {
		return ModeDoubleQuotes, true
	}
	return ModeUnPick, false
}

func (p *Pick) apostropheCheck(word, _ string) Mode {
	// '
	if word == "'" {
		return ModeUnPick
	}
	return ModeApostrophe
}

func (p *Pick) doubleQuotesCheck(word, _ string) Mode {
	// "
	if word == "\"" {
		return ModeUnPick
	}
	return ModeDoubleQuotes
}
