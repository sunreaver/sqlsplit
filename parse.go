package sqlsplit

import (
	"fmt"
	"strings"
)

type Mode int

const (
	ModeUnPick          Mode = iota
	ModeRemarkLine           // --
	ModeRemarkMoreLine       // /**/
	ModeDefaultSql           // select/update
	ModeProcedure            // create [or replace] procedure
	ModeMaybeProcedure1      // create
	ModeMaybeProcedure2      // create or
	ModeMaybeProcedure3      // create or replace
	ModeApostrophe           // '
	ModeDoubleQuotes         // "
)

func (m Mode) String() string {
	switch m {
	case ModeUnPick:
		return "unpick"
	case ModeRemarkLine:
		return "--"
	case ModeRemarkMoreLine:
		return "/**/"
	case ModeDefaultSql:
		return "select/update"
	case ModeMaybeProcedure1:
		return "create_"
	case ModeMaybeProcedure2:
		return "create_or"
	case ModeMaybeProcedure3:
		return "create_or_replace"
	case ModeProcedure:
		return "procedure/event"
	case ModeApostrophe:
		return "'"
	case ModeDoubleQuotes:
		return "\""
	}
	return "unknow"
}

type SqlParse struct {
	SQL  string  `json:"sql"`
	Type SQLTYPE `json:"type"`
}

func Split(sqls string) []SqlParse {
	words := NewWords(sqls)
	p := Pick{}
	p.Reset()
	remark := ""
	outs := []SqlParse{}
	words.Range(func(word, space string) (stop bool) {
		over := p.Pick(word, space)
		if over {
			if p.nowmode != ModeUnPick {
				if p.nowmode == ModeRemarkLine || p.nowmode == ModeRemarkMoreLine {
					remark += p.sql
				} else {
					// 针对orcle普通语句不能带分号的规则，这里把非存储过程的sql语句的分号的去掉
					psqlTmp := p.sql
					if p.nowmode != ModeProcedure {
						psqlTmp = RemoveLastSemicolon(psqlTmp)
					}

					// 如果整句未匹配，则证明全是空白字符，直接抛弃
					outs = append(outs, SqlParse{
						SQL:  fmt.Sprintf("%v%v", remark, psqlTmp),
						Type: SQLType(p.sql),
					})
					remark = ""
				}
			}
			p.Reset()
		}
		return false
	})
	if len(p.sql) > 0 {
		for p.modestack.Len() > 0 {
			p.nowmode, _ = p.modestack.Pop().(Mode)
		}
	} else if len(remark) > 0 {
		p.nowmode = ModeRemarkMoreLine
	}
	if p.nowmode != ModeUnPick {
		// 针对orcle普通语句不能带分号的规则，这里把非存储过程的sql语句的分号的去掉
		psqlTmp := p.sql
		if p.nowmode != ModeProcedure {
			psqlTmp = RemoveLastSemicolon(psqlTmp)
		}
		outs = append(outs, SqlParse{
			SQL:  fmt.Sprintf("%v%v", remark, psqlTmp),
			Type: SQLType(p.sql),
		}) // range 存在最后不以结束符结尾，导致最后一条sql丢失
	}
	return outs
}

/**
把字符串中最后一个分号移除
*/
func RemoveLastSemicolon(str string) string {
	str = strings.TrimSpace(str)
	if str == "" {
		return str
	}
	length := len(str)
	if strings.LastIndex(str, ";") == length-1 {
		str = str[0 : length-1]
	}
	return str
}
