package types

import (
	"encoding/json"
	"fmt"
	"strings"
)

const setMark = "set"

// Set is a type for Sets
// in OVSDBs JSON representation Sets are represented as a two element array with
// the first element being the string "set" and the second element being an array of the elements
// one element sets are represented as a single element JSON value
// e.g. ["set", [["uuid", "8cc2eb5c-8e66-4554-af1d-8fa5b9321f99"], ["uuid", "8cc2eb5c-8e66-4554-af1d-8fa5b9321f99"]]]
// or ["set", []] for an empty set
// but ["uuid", "8cc2eb5c-8e66-4554-af1d-8fa5b9321f99"] or just "some_string"  for a set with one element
type Set[T AtomicType] []T

func (s *Set[T]) UnmarshalJSON(data []byte) error {
	// one element case
	var one T
	if err := json.Unmarshal(data, &one); err == nil {
		*s = []T{one}
		return nil
	}

	// set case
	parts := [2]any{new(string), new([]T)}
	if err := json.Unmarshal(data, &parts); err != nil || *parts[0].(*string) != setMark {
		return fmt.Errorf("invalid set: %w", err)
	}
	*s = *parts[1].(*[]T)
	return nil
}

func (s Set[T]) MarshalJSON() ([]byte, error) {
	if len(s) == 1 {
		return json.Marshal(s[0])
	} else if s == nil {
		s = []T{}
	}
	return json.Marshal([]any{setMark, []T(s)})
}

func (s *Set[T]) String() string {
	var sb strings.Builder
	sb.WriteString("[")
	d := ""
	for _, v := range *s {
		_, _ = fmt.Fprintf(&sb, "%s%v", d, v)
		d = ", "
	}
	sb.WriteString("]")
	return sb.String()
}

func (s Set[T]) Update2(_s2 any) (any, error) {
	s2, ok := _s2.(Set[T])
	s3 := s
	if !ok {
		return nil, fmt.Errorf("unsuitable Set update2")
	}
	if s2 == nil {
		return s, nil
	}
	for _, v2 := range s2 {
		found := false
		for i, v := range s3 {
			if v == v2 {
				found = true
				s3 = append((s3)[:i], (s3)[i+1:]...)
				break
			}
		}
		if !found {
			s3 = append(s3, v2)
		}
	}
	return s3, nil
}

func (s Set[T]) Has(_val T) bool {
	for _, val := range s {
		if val == _val {
			return true
		}
	}
	return false
}
