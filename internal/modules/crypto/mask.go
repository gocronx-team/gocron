package crypto

import (
	"sort"
	"strings"
)

// MaskPlaceholder 是脱敏后替换的占位符。
const MaskPlaceholder = "******"

// MaskSecrets 把 text 中出现的任意 secret 明文替换为占位符,用于任务输出/命令在
// 落库或展示前的脱敏。空值会被忽略;较长的 secret 优先替换,避免某个 secret 是
// 另一个子串时产生残留。
func MaskSecrets(text string, secrets []string) string {
	if text == "" || len(secrets) == 0 {
		return text
	}
	// 去重 + 过滤空值
	seen := make(map[string]struct{}, len(secrets))
	uniq := make([]string, 0, len(secrets))
	for _, s := range secrets {
		if s == "" {
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
