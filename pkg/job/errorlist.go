package job

type ErrorList []error

func (el ErrorList) Error() string {
	if len(el) == 0 {
		return ""
	}
	if len(el) == 1 {
		return el[0].Error()
	}
	var s string
	for i, e := range el {
		if i > 0 {
			s += "; "
		}
		s += e.Error()
	}
	return s
}
