package arrayops

func NewStringArrayWithValue(size int, value string) []string {
	if size <= 0 {
		return []string{}
	}
	arr := make([]string, size)
	for i := range arr {
		arr[i] = value
	}
	return arr
}
