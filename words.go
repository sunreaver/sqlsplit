package sqlsplit

type Words struct {
	words       []string
	wordsSuffix []string
}

func (w *Words) Range(fn func(word, space string) (stop bool)) {
	for idx, v := range w.words {
		if stop := fn(v, w.wordsSuffix[idx]); stop {
			break
		}
	}
}

func NewWords(input string) *Words {
	w := Words{
		words:       []string{},
		wordsSuffix: []string{},
	}
	for len(input) > 0 {
		whole := input[:wholeWord(input)]
		input = input[len(whole):]
		space := input[:spaceEnd(input)]
		input = input[len(space):]
		w.words = append(w.words, whole)
		w.wordsSuffix = append(w.wordsSuffix, space)
	}
	return &w
}

const oneWholeWords = `;'"()` // )begin 和 if(abc)

var (
	oneWholeWord = map[rune]bool{}
	twoWholeWord = map[string]bool{
		"/*": true,
		"*/": true,
		"--": true,
	}
)

func init() {
	for _, v := range oneWholeWords {
		oneWholeWord[v] = true
	}
}

func isOneWholeWord(v rune) bool {
	return oneWholeWord[v]
}

func isTwoWholeWord(idx int, input string) bool {
	return idx+1 < len(input) && twoWholeWord[input[idx:idx+2]]
}

// 查找完整单次，遇见空格，换行，分号则结束
func wholeWord(input string) int {
	for idx, v := range input {
		if v > ' ' && !isOneWholeWord(v) && !isTwoWholeWord(idx, input) {
			continue
		}
		if idx == 0 {
			if isOneWholeWord(v) {
				return 1
			} else if isTwoWholeWord(idx, input) {
				return 2
			}
		}
		return idx
	}
	return len(input)
}

// 查找完整单次，遇见空格，换行，分号则结束
func spaceEnd(input string) int {
	for idx, v := range input {
		if v > ' ' {
			return idx
		}
	}
	return len(input)
}
