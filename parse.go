package sqlsplit

import "fmt"

type Mode int

const (
	ModeUnPick Mode = iota
	ModeRemarkLine
	ModeRemarkMoreLine
	ModeDefaultSql
	ModeProcedure
	ModeMaybeProcedure1 // create
	ModeMaybeProcedure2 // create or
	ModeMaybeProcedure3 // create or replace
	ModeApostrophe      // '
	ModeDoubleQuotes    // "
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
		return "select"
	case ModeMaybeProcedure1:
		return "create_"
	case ModeMaybeProcedure2:
		return "create_or"
	case ModeMaybeProcedure3:
		return "create_or_replace"
	case ModeProcedure:
		return "procedure"
	case ModeApostrophe:
		return "'"
	case ModeDoubleQuotes:
		return "\""
	}
	return "unknow"
}

type SqlParse struct {
	SQL  string `json:"sql"`
	Type string `json:"type"`
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
					// 如果整句未匹配，则证明全是空白字符，直接抛弃
					outs = append(outs, SqlParse{
						SQL:  fmt.Sprintf("%v%v", remark, p.sql),
						Type: p.nowmode.String(),
					})
					remark = ""
				}
			}
			p.Reset()
		}
		return false
	})
	if len(p.sql) > 0 || len(remark) > 0 {
		for p.modestack.Len() > 0 {
			p.nowmode, _ = p.modestack.Pop().(Mode)
		}
		outs = append(outs, SqlParse{
			SQL:  fmt.Sprintf("%v%v", remark, p.sql),
			Type: p.nowmode.String(),
		}) // range 存在最后不以结束符结尾，导致最后一条sql丢失
	}
	return outs
}
