package reader

// Comment 单行注释
func Comment(r *Reader) string {
	ch, ok := r.ReadByte() // read character after "//"
	for ok && ch != '\n' {
		ch, ok = r.ReadByte()
	}
	r.GoBack()
	return r.ReadText()
}
