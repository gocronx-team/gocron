package crypto

import (
	"sort"
	"strings"
)

// MaskPlaceholder 是脱敏后替换的占位符。
const MaskPlaceholder = "******"

// MinMaskLength 是参与脱敏的机密值最小长度。过短的值(如 "1"、"ok"、"true")
// 若做子串替换会把日志里大量正常内容(数字、布尔、状态码)误打码,破坏可读性,
// 且这类极短值本身不构成有效机密。低于此长度的值不参与脱敏。
const MinMaskLength = 5

// MaskSecrets 把 text 中出现的任意 secret 明文替换为占位符,用于任务输出/命令在
// 落库或展示前的脱敏。空值、过短值(< MinMaskLength)会被忽略;较长的 secret 优先
// 替换,避免某个 secret 是另一个子串时产生残留。
func MaskSecrets(text string, secrets []string) string {
	if text == "" || len(secrets) == 0 {
		return text
	}
	// 去重 + 过滤空值与过短值
	seen := make(map[string]struct{}, len(secrets))
	uniq := make([]string, 0, len(secrets))
	for _, s := range secrets {
		if len(s) < MinMaskLength {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		uniq = append(uniq, s)
	}
	// 按长度降序,先替换长的
	sort.SliceStable(uniq, func(i, j int) bool {
		return len(uniq[i]) > len(uniq[j])
	})
	for _, s := range uniq {
		text = strings.ReplaceAll(text, s, MaskPlaceholder)
	}
	return text
}
