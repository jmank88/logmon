package logmon

import "testing"

func TestAdd(t *testing.T) {
	for _, testCase := range []struct {
		input    []string
		topSize  int
		expected []entry
	}{
		{
			[]string{"a"},
			5,
			[]entry{{"a", 1}},
		},
		{
			[]string{"a", "a", "b"},
			1,
			[]entry{{"a", 2}},
		},
		{
			[]string{"a", "a", "b", "b", "b", "c", "c", "c", "c"},
			2,
			[]entry{{"c", 4}, {"b", 3}},
		},
	} {
		s := newStats()
		for _, key := range testCase.input {
			s.add(key)
		}
		got := s.top(testCase.topSize)
		if !equals(got, testCase.expected) {
			t.Errorf("case: %v; expected %v but got %v", testCase.input, testCase.expected, got)
		}
	}
}

func equals(a, b []entry) bool {
	if len(a) != len(b) {
		return false
	}
	for i, _ := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
