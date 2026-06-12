package form

type Errors map[string][]string

func (e Errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

func (e Errors) HasAny() bool {
	return len(e) > 0
}
