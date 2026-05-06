package secret

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/Pimatis/mavetis/src/model"
)

func cacheKey(config model.Config) string {
	parts := []string{"patterns=" + patternKey()}
	values := append([]string{}, config.Allow.Values...)
	regexes := append([]string{}, config.Allow.Regexes...)
	sort.Strings(values)
	sort.Strings(regexes)
	parts = append(parts, "allowValues="+strings.Join(values, "\x00"))
	parts = append(parts, "allowRegexes="+strings.Join(regexes, "\x00"))
	sum := sha256.Sum256([]byte(strings.Join(parts, "\n")))
	return hex.EncodeToString(sum[:])
}

func patternKey() string {
	items := patterns()
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, item.id+":"+item.re.String()+fmt.Sprintf(":%d:%.2f", item.group, item.entropy))
	}
	sort.Strings(parts)
	sum := sha256.Sum256([]byte(strings.Join(parts, "\n")))
	return hex.EncodeToString(sum[:])
}
