package textstat

import "strings"

var builtinStopWords = map[string]bool{
	"a": true, "an": true, "the": true, "and": true, "or": true, "but": true,
	"is": true, "are": true, "was": true, "were": true, "be": true, "been": true, "being": true,
	"have": true, "has": true, "had": true, "do": true, "does": true, "did": true,
	"will": true, "would": true, "shall": true, "should": true, "can": true, "could": true,
	"may": true, "might": true, "must": true, "need": true, "dare": true, "ought": true, "used": true,
	"i": true, "you": true, "he": true, "she": true, "it": true, "we": true, "they": true,
	"me": true, "him": true, "her": true, "us": true, "them": true,
	"my": true, "your": true, "his": true, "its": true, "our": true, "their": true,
	"mine": true, "yours": true, "hers": true, "ours": true, "theirs": true,
	"this": true, "that": true, "these": true, "those": true,
	"what": true, "which": true, "who": true, "whom": true, "whose": true,
	"where": true, "when": true, "why": true, "how": true,
	"all": true, "each": true, "every": true, "both": true, "few": true, "more": true, "most": true,
	"other": true, "some": true, "such": true, "no": true, "nor": true, "not": true, "only": true,
	"own": true, "same": true, "so": true, "than": true, "too": true, "very": true,
	"just": true, "also": true, "now": true, "here": true, "there": true,
	"new": true, "one": true, "two": true, "first": true, "last": true,
	"if": true, "then": true, "else": true, "for": true, "with": true, "about": true,
	"against": true, "between": true, "into": true, "through": true, "during": true,
	"before": true, "after": true, "above": true, "below": true, "from": true, "up": true,
	"down": true, "in": true, "out": true, "on": true, "off": true, "over": true, "under": true,
	"again": true, "further": true, "once": true, "because": true, "until": true, "while": true,
	"of": true, "at": true, "by": true, "to": true,

	"的": true, "是": true, "在": true, "了": true, "着": true, "过": true, "地": true, "得": true,
	"有": true, "没有": true, "就": true, "都": true, "和": true, "跟": true, "与": true,
	"或": true, "或者": true, "但": true, "但是": true, "然而": true, "而": true, "却": true,
	"如果": true, "假如": true, "要是": true, "只要": true, "只有": true, "除非": true, "否则": true,
	"因为": true, "由于": true, "之所以": true, "所以": true, "因此": true, "于是": true,
	"虽然": true, "尽管": true, "即使": true, "哪怕": true, "既然": true, "既": true,
	"又": true, "也": true, "还": true, "再": true, "更": true, "最": true, "很": true, "非常": true,
	"太": true, "比较": true, "相当": true, "特别": true, "尤其": true, "甚至": true,
	"不": true, "没": true, "别": true, "莫": true, "勿": true, "未": true,
	"这": true, "那": true, "此": true, "彼": true, "该": true, "本": true, "其": true,
	"之": true, "为": true, "以": true, "于": true, "自": true, "从": true, "到": true, "向": true,
	"往": true, "朝": true, "把": true, "被": true, "让": true, "给": true, "叫": true, "令": true,
	"请": true, "您": true, "你": true, "我": true, "他": true, "她": true, "它": true, "们": true,
	"谁": true, "什么": true, "哪": true, "哪里": true, "怎么": true, "怎样": true, "为什么": true,
	"几": true, "多少": true, "多": true, "少": true, "半": true, "一": true, "二": true, "三": true,
	"四": true, "五": true, "六": true, "七": true, "八": true, "九": true, "十": true,
	"百": true, "千": true, "万": true, "亿": true, "第": true, "个": true, "只": true,
	"件": true, "条": true, "根": true, "张": true, "片": true, "块": true, "颗": true,
	"粒": true, "朵": true, "棵": true, "株": true, "座": true, "所": true, "间": true,
	"栋": true, "辆": true, "艘": true, "架": true, "台": true, "部": true,
	"册": true, "页": true, "章": true, "节": true, "段": true, "句": true, "词": true,
	"字": true, "号": true, "次": true, "回": true, "遍": true, "番": true,
	"阵": true, "场": true, "顿": true, "剂": true, "服": true, "帖": true,
	"啊": true, "吧": true, "呢": true, "吗": true, "呀": true, "哇": true,
	"哦": true, "嗯": true, "哈": true, "哎": true, "唉": true, "喂": true, "嗨": true,
	"哟": true, "啦": true, "喽": true, "嘛": true, "呗": true, "而已": true, "罢了": true,
	"及": true, "若": true, "如": true, "似": true,
	"像": true, "好比": true, "如同": true, "犹如": true, "仿佛": true, "似乎": true,
	"大概": true, "大约": true, "可能": true, "也许": true, "或许": true,
	"以及": true, "及其": true, "等等": true, "之类": true, "等": true,
	"能够": true, "会": true, "要": true, "想": true, "愿": true, "肯": true, "敢": true,
	"应": true, "当": true, "必": true, "须": true, "务必": true,
	"来": true, "去": true, "上": true, "进": true,
	"开": true, "关": true, "起": true, "坐": true, "站": true, "躺": true,
	"走": true, "跑": true, "跳": true, "飞": true, "游": true, "爬": true,
	"说": true, "讲": true, "谈": true, "道": true, "问": true, "答": true,
	"看": true, "见": true, "视": true, "望": true, "瞧": true, "盯": true, "瞅": true,
	"听": true, "闻": true, "嗅": true, "尝": true, "觉": true, "知": true, "懂": true,
	"学": true, "习": true, "思": true, "念": true, "忆": true,
	"做": true, "作": true, "搞": true, "办": true, "干": true, "行": true,
}

func GetBuiltinStopWords() map[string]bool {
	sw := make(map[string]bool)
	for k, v := range builtinStopWords {
		sw[k] = v
	}
	return sw
}

type StopWords struct {
	words map[string]bool
}

func NewStopWords() *StopWords {
	return &StopWords{
		words: GetBuiltinStopWords(),
	}
}

func NewStopWordsWithCustom(custom map[string]bool) *StopWords {
	sw := make(map[string]bool)
	for k, v := range custom {
		sw[strings.ToLower(k)] = v
	}
	return &StopWords{words: sw}
}

func (sw *StopWords) IsStopWord(word string) bool {
	return sw.words[strings.ToLower(word)]
}

func (sw *StopWords) Add(word string) {
	sw.words[strings.ToLower(word)] = true
}

func (sw *StopWords) Remove(word string) {
	delete(sw.words, strings.ToLower(word))
}

func (sw *StopWords) Merge(words ...string) {
	for _, w := range words {
		sw.Add(w)
	}
}

func (sw *StopWords) GetWords() map[string]bool {
	copy := make(map[string]bool)
	for k, v := range sw.words {
		copy[k] = v
	}
	return copy
}
