package sqlsplit

// Words 用于保存分词后的词语切片与其后紧随的空白字符切片
type Words struct {
	// words 存储切分出的各个独立词语
	words []string
	// wordsSuffix 存储与每个词语对应紧随其后的空白或格式字符
	wordsSuffix []string
}

// Range 遍历所有已切分出的词语及其后缀空白，若回调函数 fn 返回 true 则中止遍历
func (w *Words) Range(fn func(word, space string) (stop bool)) {
	for idx, v := range w.words {
		// 重要判断：检查外部传入的回调是否要求提前终止遍历
		if stop := fn(v, w.wordsSuffix[idx]); stop {
			break
		}
	}
}

// NewWords 根据输入的原始 SQL 字符串，按特殊字符边界与空白切分为 Words 集合实例
func NewWords(input string) *Words {
	w := Words{
		words:       []string{},
		wordsSuffix: []string{},
	}
	// 逐步提取每一个完整的单个词与其后面的空白
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

// oneWholeWords 定义单字符即构成完整词的特殊标点符号集合（如分号、引号、括号、#等）
const oneWholeWords = `;'"()#`

var (
	// oneWholeWord 以 map 形式快速判断 rune 是否为单字符独立词
	oneWholeWord = map[rune]bool{}
	// twoWholeWord 存储双字符独立词集合（如多行注释开头 /*、结尾 */、单行注释 --）
	twoWholeWord = map[string]bool{
		"/*": true, // 块注释开始
		"*/": true, // 块注释结束
		"--": true, // 单行注释开始
	}
)

func init() {
	// 初始化单字符独立词映射表
	for _, v := range oneWholeWords {
		oneWholeWord[v] = true
	}
}

// isOneWholeWord 检查指定字符是否属于单个字符即为独立词的分隔符号
func isOneWholeWord(v rune) bool {
	return oneWholeWord[v]
}

// isTwoWholeWord 检查在字符串 input 的指定偏移 idx 处是否匹配双字符独立词（/*、*/ 或 --）
func isTwoWholeWord(idx int, input string) bool {
	return idx+1 < len(input) && twoWholeWord[input[idx:idx+2]]
}

// isWordBoundary 判定当前字符是否构成单词的边界（空白字符、单字符独立词或双字符独立词）
func isWordBoundary(v rune, idx int, input string) bool {
	return v <= ' ' || isOneWholeWord(v) || isTwoWholeWord(idx, input)
}

// wholeWord 寻找从当前 input 开头算起的一个完整单词的长度。遇到空格、换行、分号或注释标记则结束。
func wholeWord(input string) int {
	nextMustContinue := false
	for idx, v := range input {
		// 重要判断：上一个字符为反斜杠转义符，强制当前字符作为词的一部分，不触发边界中断
		if nextMustContinue {
			nextMustContinue = false
			continue
		}
		// 重要判断：遇到反斜杠时开启转义保护标志
		if v == '\\' {
			nextMustContinue = true
			continue
		}
		// 重要判断：若尚未达到词边界，继续向前扫描
		if !isWordBoundary(v, idx, input) {
			continue
		}
		// 若在索引为 0 处即遇到了独立词边界：
		if idx == 0 {
			// 单字符独立词长度为 1（如 ;、'、"、(、) 等）
			if isOneWholeWord(v) {
				return 1
			}
			// 双字符独立词长度为 2（如 /*、*/、--）
			return 2
		}
		// 遇到了单词的结束边界，返回该单词的长度
		return idx
	}
	// 未遇到边界，整个 input 作为一个完整单词
	return len(input)
}

// spaceEnd 寻找输入文本开头的连续空白字符长度，遇到首个非空白字符（v > ' '）时返回其索引
func spaceEnd(input string) int {
	for idx, v := range input {
		// 重要判断：遇到首个 ASCII 大于空格的可视字符即为非空白
		if v > ' ' {
			return idx
		}
	}
	return len(input)
}
