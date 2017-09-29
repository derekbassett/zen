package zen

type fieldKey struct{}

type fields map[string]interface{}

func (f fields) Merge(m fields) {
	for k, v := range m {
		if _, ok := f[k]; !ok {
			f[k] = v
		}
	}
}
