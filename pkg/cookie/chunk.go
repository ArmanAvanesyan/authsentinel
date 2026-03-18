package cookie

// maxCookieValueSize is a conservative limit for a single cookie value.
// It is not exposed; callers interact with Codec interfaces instead.
const maxCookieValueSize = 3800

// chunkValue splits v into chunks if it exceeds maxCookieValueSize.
// The first chunk is v itself when no splitting is required.
func chunkValue(v string) []string {
	if len(v) <= maxCookieValueSize {
		return []string{v}
	}
	out := make([]string, 0, (len(v)/maxCookieValueSize)+1)
	for start := 0; start < len(v); start += maxCookieValueSize {
		end := start + maxCookieValueSize
		if end > len(v) {
			end = len(v)
		}
		out = append(out, v[start:end])
	}
	return out
}

// unchunkValue joins previously chunked parts back into a single value.
func unchunkValue(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	total := 0
	for _, p := range parts {
		total += len(p)
	}
	buf := make([]byte, 0, total)
	for _, p := range parts {
		buf = append(buf, p...)
	}
	return string(buf)
}

