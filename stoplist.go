package services

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
	"strings"
)

type Rule struct {
	l  lexeme
	rr *rootRule
}

type pipeRule struct {
	r []Checker
}

func (s *pipeRule) Check(val string) *CheckResult {
	for i, checker := range s.r {
		cr := checker.Check(val)
		if cr.Found {
			if len(s.r) > 1 {
				cr.Stack = append([]string{fmt.Sprintf("pipe index %v", i)}, cr.Stack...)
			}
			return cr
		}
	}
	return &CheckResult{}
}

type plusRule struct {
	r []Checker
}

func (s *plusRule) Check(val string) *CheckResult {
	res := &CheckResult{
		Found: true,
	}
	for _, checker := range s.r {
		cr := checker.Check(val)
		if !cr.Found {
			return &CheckResult{}
		} else {
			res.Stack = append(res.Stack, cr.Stack...)
		}
	}
	if len(s.r) > 1 {
		res.Stack = append([]string{"plus"}, res.Stack...)
	}
	return res
}

type lineRule struct {
	r []Checker
}

func (s *lineRule) Check(val string) *CheckResult {
	for i, checker := range s.r {
		cr := checker.Check(val)
		if cr.Found {
			if len(s.r) > 1 {
				cr.Stack = append([]string{fmt.Sprintf("line index %v", i)}, cr.Stack...)
			}
			return cr
		}
	}
	return &CheckResult{}
}

type rootRule struct {
	r map[string]Checker
}

func NewRuleFromYaml(data []byte) (Checker, error) {
	y := map[string][]string{}
	err := yaml.Unmarshal(data, y)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal Checker data")
	}
	rr, err := NewRule(y)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to make rule")
	}
	return rr, nil
}

func NewRuleFromYamlFile(path string) (Checker, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read Checker data from file")
	}
	return NewRuleFromYaml(f)
}

func (r *rootRule) Check(val string) *CheckResult {
	return r.r["main"].Check(val)
}

type LexemeType string

const (
	Plus      LexemeType = "plus"
	Pipe      LexemeType = "pipe"
	Regexp    LexemeType = "regexp"
	Reference LexemeType = "reference"
	Text      LexemeType = "text"
)

type lexeme struct {
	Value string
	t     LexemeType
}

func ParseLine(ll string) []lexeme {
	l := []rune(ll)
	var res []lexeme
	value := ""
	reg := false
	ref := false
	next := ""
	for i, c := range l {
		if len(l)-1 == i {
			next = ""
		} else {
			next = string(l[i+1])
		}
		if c == '/' && value == "" {
			reg = true
			continue
		}
		if !reg && c == '{' && value == "" {
			ref = true
			continue
		}
		if !reg && ref && c == '}' && (next == "|" || next == "" || next == "+") {
			ref = false
			res = append(res, lexeme{
				t:     Reference,
				Value: value,
			})
			value = ""
			continue
		}
		if reg && c == '/' && (next == "|" || next == "" || next == "+") {
			reg = false
			res = append(res, lexeme{
				t:     Regexp,
				Value: value,
			})
			value = ""
			continue
		}
		if !reg && !ref && (next == "|" || next == "" || next == "+") {
			value += string(c)
			res = append(res, lexeme{
				t:     Text,
				Value: value,
			})
			value = ""
			continue
		}
		if !reg && c == '|' {
			res = append(res, lexeme{
				t: Pipe,
			})
			continue
		}
		if !reg && c == '+' {
			res = append(res, lexeme{
				t: Plus,
			})
			continue
		}
		value += string(c)
	}
	return res
}

func NewRule(m map[string][]string) (Checker, error) {
	if _, ok := m["main"]; !ok {
		return nil, errors.Errorf("failed to find main rule reference")
	}
	rr := &rootRule{
		r: map[string]Checker{},
	}
	for k, v := range m {
		rule, err := NewLineRule(rr, v)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to make line rule")
		}
		rr.r[k] = rule
	}
	return rr, nil
}

var _ Checker = &rootRule{}

func NewLineRule(rr *rootRule, lines []string) (Checker, error) {
	lr := &lineRule{
		r: []Checker{},
	}
	var rules []Checker
	for _, line := range lines {
		lexemes := ParseLine(line)
		rule, err := NewPlusRule(rr, lexemes)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to make plus rule")
		}
		rules = append(rules, rule)
	}
	lr.r = rules
	return lr, nil
}

var _ Checker = &lineRule{}

func NewPlusRule(rr *rootRule, lms []lexeme) (Checker, error) {
	parts := SplitByLexeme(lms, Plus)
	var rules []Checker
	for _, p := range parts {
		rule, err := NewPipeRule(rr, p)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to make pipe rule")
		}
		rules = append(rules, rule)
	}
	return &plusRule{
		r: rules,
	}, nil
}

var _ Checker = &plusRule{}

type CheckResult struct {
	Found bool
	Stack []string
}

func (s CheckResult) String() string {
	if s.Found {
		return "found: " + strings.Join(s.Stack, ": ")
	}
	return "not found"
}

type Checker interface {
	Check(val string) *CheckResult
}

func NewPipeRule(rr *rootRule, lms []lexeme) (Checker, error) {
	parts := SplitByLexeme(lms, Pipe)
	var rules []Checker
	for _, p := range parts {
		rule, err := newRule(rr, p[0])
		if err != nil {
			return nil, errors.Wrapf(err, "failed to build pipe rule")
		}
		rules = append(rules, rule)
	}
	return &pipeRule{
		r: rules,
	}, nil
}

var _ Checker = &pipeRule{}

type TextRule struct {
	Text string
}

func (s *TextRule) Check(val string) *CheckResult {
	i := strings.Index(val, s.Text)
	if i == -1 {
		return &CheckResult{}
	}
	return &CheckResult{
		Found: true,
		Stack: []string{
			fmt.Sprintf("\"%v\" contains \"%v\" at pos %v", val, s.Text, i),
		},
	}
}

func NewTextRule(text string) (*TextRule, error) {
	return &TextRule{
		Text: text,
	}, nil
}

var _ Checker = &TextRule{}

type RegexpRule struct {
	Regexp *regexp.Regexp
}

func (s *RegexpRule) Check(val string) *CheckResult {
	loc := s.Regexp.FindIndex([]byte(val))
	i := -1
	if loc != nil {
		i = loc[0]
	}
	if i == -1 {
		return &CheckResult{}
	}
	found := val[loc[0]:loc[1]]
	return &CheckResult{
		Found: true,
		Stack: []string{
			fmt.Sprintf("\"%v\" contains \"%v\" by regexp \"%v\" at pos %v", val, found, s.Regexp, i),
		},
	}
}

var _ Checker = &RegexpRule{}

func newRule(rr *rootRule, l lexeme) (Checker, error) {
	switch l.t {
	case Text:
		return NewTextRule(l.Value)
	case Regexp:
		return NewRegexpRule(l.Value)
	case Reference:
		return NewReferenceRule(rr, l.Value)
	default:
		return nil, errors.Errorf("failed to make rule for %v", l.t)
	}
}

type ReferenceRule struct {
	r Checker
	v string
}

func (s *ReferenceRule) Check(val string) *CheckResult {
	cr := s.r.Check(val)
	if cr.Found {
		cr.Stack = append([]string{fmt.Sprintf("reference \"%v\"", s.v)}, cr.Stack...)
	}
	return cr
}

var _ Checker = &ReferenceRule{}

func NewReferenceRule(rr *rootRule, value string) (Checker, error) {
	if _, ok := rr.r[value]; !ok {
		return nil, errors.Errorf("failed to find reference %v", value)
	}
	return &ReferenceRule{
		r: rr.r[value],
		v: value,
	}, nil
}

func NewRegexpRule(value string) (Checker, error) {
	r, err := regexp.Compile(value)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to compile regexp rule for %v", value)
	}

	return &RegexpRule{
		Regexp: r,
	}, nil
}

func SplitByLexeme(lms []lexeme, lt LexemeType) [][]lexeme {
	var parts [][]lexeme
	var cur []lexeme
	for _, l := range lms {
		if l.t != lt {
			cur = append(cur, l)
		} else {
			parts = append(parts, cur)
			cur = nil
		}
	}
	if cur != nil {
		parts = append(parts, cur)
	}
	return parts
}
