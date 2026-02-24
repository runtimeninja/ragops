package ingest

// ChunkText splits text by rune length (MVP).
// size = chunk size, overlap = overlap between chunks.
func ChunkText(s string, size int, overlap int) []string {
	if size <= 0 {
		size = 800
	}
	if overlap < 0 {
		overlap = 0
	}
	r := []rune(s)
	if len(r) == 0 {
		return nil
	}

	out := make([]string, 0, (len(r)/size)+1)
	for start := 0; start < len(r); {
		end := start + size
		if end > len(r) {
			end = len(r)
		}
		out = append(out, string(r[start:end]))
		if end == len(r) {
			break
		}
		start = end - overlap
		if start < 0 {
			start = 0
		}
	}
	return out
}
