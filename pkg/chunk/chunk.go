package chunk

func chunk[T any](slice []T, size int) [][]T {
	if len(slice) == 0 || size <= 0 {
		return nil
	}
	var chunks [][]T
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}
