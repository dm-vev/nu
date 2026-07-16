package components

import "strings"

func markdownRenderInline(source string, opts MarkdownOptions) string {
	var out strings.Builder
	for i := 0; i < len(source); {
		if strings.HasPrefix(source[i:], "**") {
			if end := strings.Index(source[i+2:], "**"); end >= 0 {
				out.WriteString(opts.StrongStyle(source[i+2 : i+2+end]))
				i += end + 4
				continue
			}
		}
		if source[i] == '`' {
			if end := strings.IndexByte(source[i+1:], '`'); end >= 0 {
				out.WriteString(opts.CodeStyle(source[i+1 : i+1+end]))
				i += end + 2
				continue
			}
		}
		if source[i] == '*' || source[i] == '_' {
			marker := source[i]
			if end := strings.IndexByte(source[i+1:], marker); end >= 0 {
				out.WriteString(opts.EmphasisStyle(source[i+1 : i+1+end]))
				i += end + 2
				continue
			}
		}
		next := markdownNextMarker(source, i)
		out.WriteString(opts.TextStyle(source[i:next]))
		i = next
	}
	return out.String()
}

func markdownNextMarker(source string, start int) int {
	for i := start + 1; i < len(source); i++ {
		switch source[i] {
		case '*', '_', '`':
			return i
		}
	}
	return len(source)
}
