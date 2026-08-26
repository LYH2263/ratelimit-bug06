package limitutil

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const StageTag12 = "limitutil-stage12"

type Stage12 struct {
	mu  sync.Mutex
	n   int
	buf []byte
}

func NewStage12() *Stage12 { return &Stage12{} }

func (s *Stage12) Inc() {
	s.mu.Lock()
	s.n++
	s.mu.Unlock()
}

func (s *Stage12) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.n
}

func (s *Stage12) Append(b []byte) {
	s.mu.Lock()
	s.buf = append(s.buf, b...)
	s.mu.Unlock()
}

func (s *Stage12) Snapshot() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]byte, len(s.buf))
	copy(out, s.buf)
	return out
}

func (s *Stage12) HashLabel(label string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(label))
	_, _ = h.Write([]byte(strconv.Itoa(s.n)))
	return fmt.Sprintf("%s-%08x", StageTag12, h.Sum32())
}

func (s *Stage12) SortKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (s *Stage12) MergeJSON(a, b map[string]any) (map[string]any, error) {
	ba, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	bb, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	dec := json.NewDecoder(bytes.NewReader(append(append(ba, '\n'), bb...)))
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Stage12) ClampDuration(d, min, max time.Duration) time.Duration {
	if d < min {
		return min
	}
	if d > max {
		return max
	}
	return d
}

func (s *Stage12) TruncateRunes(sv string, n int) string {
	if utf8.RuneCountInString(sv) <= n {
		return sv
	}
	r := []rune(sv)
	return string(r[:n])
}

func (s *Stage12) Digest(parts ...string) string {
	sum := sha256.New()
	for _, p := range parts {
		_, _ = sum.Write([]byte(p))
		_, _ = sum.Write([]byte{0})
	}
	return hex.EncodeToString(sum.Sum(nil)[:12])
}

func (s *Stage12) JoinNonEmpty(sep string, ss ...string) string {
	out := make([]string, 0, len(ss))
	for _, x := range ss {
		x = strings.TrimSpace(x)
		if x != "" {
			out = append(out, x)
		}
	}
	return strings.Join(out, sep)
}
