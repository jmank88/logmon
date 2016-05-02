package logmon

import "sort"

type stats map[string]int

func newStats() stats {
	return stats(make(map[string]int))
}

func (s stats) add(key string) {
	if cnt, ok := s[key]; ok {
		s[key] = cnt + 1
	} else {
		s[key] = 1
	}
}

// The top method returns the top cnt entries from s, from greatest to least.
func (s stats) top(cnt int) []entry {
	var es entries
	for k, v := range s {
		es = append(es, entry{k, v})
	}
	sort.Sort(es)
	if len(es) < cnt {
		return es
	}
	return es[:cnt]
}

type entry struct {
	string
	int
}

type entries []entry

func (es entries) Len() int {
	return len(es)
}

//Note: This actually does 'More', since we want a reverse sort.
func (es entries) Less(i, j int) bool {
	return es[i].int > es[j].int
}

func (es entries) Swap(i, j int) {
	es[i], es[j] = es[j], es[i]
}
