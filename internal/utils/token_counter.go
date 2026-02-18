package utils

import (
	"github.com/tiktoken-go/tokenizer"
)

// CountTokensTiktoken counts the total tokens for the combined messages
// Each message has .role and .content
func CountTokensTiktoken(messages []map[string]string, encodingName string) (int, error) {
	enc, err := tokenizer.Get(tokenizer.Encoding(encodingName))
	if err != nil {
		return 0, err
	}
	total := 0
	for _, msg := range messages {
		txt := msg["content"]
		ids, _, e := enc.Encode(txt)
		if e != nil {
			return 0, e
		}
		total += len(ids)
		// Also consider role => typically small, but let's do it:
		roleIds, _, _ := enc.Encode(msg["role"])
		total += len(roleIds)
	}
	return total, nil
}
