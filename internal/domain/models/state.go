package models

type State string

type StateGroup string

func (s State) Group() StateGroup {
	len_state := len(s)

	for i := range len_state {
		if s[i] == ':' {
			return StateGroup(s[:i])
		}
	}

	return ""
}
