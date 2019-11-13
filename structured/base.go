package structured

type Encoder struct{}

// AppendKey appends a new key to the output JSON.
func (e Encoder) AppendKey(dst []byte, key string) []byte {
	dst = append(dst, key...)
	return append(dst, '=')
}
