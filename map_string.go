package gateway

type MapString struct {
	data map[string]struct{}
	l    int64
}

func NewMapString() *MapString {
	return &MapString{data: map[string]struct{}{}}
}

func (m *MapString) Delete(key string) {
	delete(m.data, key)
}

func (m *MapString) Load(key string) bool {
	_, found := m.data[key]
	return found
}

func (m *MapString) Length() int {
	return len(m.data)
}

func (m *MapString) Range(f func(key string) bool) {
	for key := range m.data {
		if f(key) {
			continue
		}
		break
	}
}
