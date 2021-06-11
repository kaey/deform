package parse

import "strings"

func contains(v string, s ...string) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}

	return false
}

func concat(s ...string) string {
	if len(s) == 0 {
		return ""
	}
	b := new(strings.Builder)
	b.Grow(10)

	b.WriteString(s[0])
	for i := 0; i < len(s)-1; i++ {
		b.WriteString(", ")
		b.WriteString(s[i])
	}

	b.WriteString(" or ")
	b.WriteString(s[len(s)-1])
	return b.String()
}
