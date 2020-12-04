package gateway

type MapString struct {
	data map[string]string
	l    int64
}

func NewMapString() *MapString {
	return &MapString{data: map[string]string{}}
}

func (m *MapString) Delete(key string) {
	delete(m.data, key)
}

func (m *MapString) Load(key string) (string, bool) {
	val, found := m.data[key]
	return val, found
}

func (m *MapString) Length() int {
	return len(m.data)
}
