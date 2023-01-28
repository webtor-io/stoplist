package services

import (
	"reflect"
	"testing"
)

func TestParseLine(t *testing.T) {
	type args struct {
		l string
	}
	tests := []struct {
		name string
		args args
		want []lexeme
	}{
		{
			name: "single text",
			args: args{
				l: "text",
			},
			want: []lexeme{
				{
					Value: "text",
					t:     Text,
				},
			},
		},
		{
			name: "single reference",
			args: args{
				l: "{ref}",
			},
			want: []lexeme{
				{
					Value: "ref",
					t:     Reference,
				},
			},
		},
		{
			name: "single regular",
			args: args{
				l: "/reg/",
			},
			want: []lexeme{
				{
					Value: "reg",
					t:     Regexp,
				},
			},
		},
		{
			name: "text with pipes",
			args: args{
				l: "text1|text2|text3",
			},
			want: []lexeme{
				{
					Value: "text1",
					t:     Text,
				},
				{
					t: Pipe,
				},
				{
					Value: "text2",
					t:     Text,
				},
				{
					t: Pipe,
				},
				{
					Value: "text3",
					t:     Text,
				},
			},
		},
		{
			name: "text with pluses",
			args: args{
				l: "text1+text2+text3",
			},
			want: []lexeme{
				{
					Value: "text1",
					t:     Text,
				},
				{
					t: Plus,
				},
				{
					Value: "text2",
					t:     Text,
				},
				{
					t: Plus,
				},
				{
					Value: "text3",
					t:     Text,
				},
			},
		},
		{
			name: "text with plus and pipe",
			args: args{
				l: "text1+text2|text3",
			},
			want: []lexeme{
				{
					Value: "text1",
					t:     Text,
				},
				{
					t: Plus,
				},
				{
					Value: "text2",
					t:     Text,
				},
				{
					t: Pipe,
				},
				{
					Value: "text3",
					t:     Text,
				},
			},
		},
		{
			name: "reg + ref | text",
			args: args{
				l: "/reg/+{ref}|text",
			},
			want: []lexeme{
				{
					Value: "reg",
					t:     Regexp,
				},
				{
					t: Plus,
				},
				{
					Value: "ref",
					t:     Reference,
				},
				{
					t: Pipe,
				},
				{
					Value: "text",
					t:     Text,
				},
			},
		},
		{
			name: "regexp with plus and pipe inside",
			args: args{
				l: "/r+e|g/|text",
			},
			want: []lexeme{
				{
					Value: "r+e|g",
					t:     Regexp,
				},
				{
					t: Pipe,
				},
				{
					Value: "text",
					t:     Text,
				},
			},
		},
		{
			name: "regexp with escaping",
			args: args{
				l: `/re\/g/|text`,
			},
			want: []lexeme{
				{
					Value: `re\/g`,
					t:     Regexp,
				},
				{
					t: Pipe,
				},
				{
					Value: "text",
					t:     Text,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseLine(tt.args.l); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitByLexeme(t *testing.T) {
	type args struct {
		lms []lexeme
		lt  LexemeType
	}
	tests := []struct {
		name string
		args args
		want [][]lexeme
	}{
		{
			name: "empty",
			args: args{
				lms: []lexeme{},
				lt:  Plus,
			},
			want: nil,
		},
		{
			name: "single",
			args: args{
				lms: []lexeme{
					{
						Value: "text1",
						t:     Text,
					},
					{
						t: Plus,
					},
					{
						Value: "text2",
						t:     Text,
					},
				},
				lt: Plus,
			},
			want: [][]lexeme{
				{
					{
						Value: "text1",
						t:     Text,
					},
				},
				{
					{
						Value: "text2",
						t:     Text,
					},
				},
			},
		},
		{
			name: "multiple",
			args: args{
				lms: []lexeme{
					{
						Value: "text1",
						t:     Text,
					},
					{
						t: Plus,
					},
					{
						Value: "text2",
						t:     Text,
					},
					{
						t: Pipe,
					},
					{
						Value: "text3",
						t:     Text,
					},
					{
						t: Plus,
					},
					{
						Value: "text4",
						t:     Text,
					},
				},
				lt: Plus,
			},
			want: [][]lexeme{
				{
					{
						Value: "text1",
						t:     Text,
					},
				},
				{
					{
						Value: "text2",
						t:     Text,
					},
					{
						t: Pipe,
					},
					{
						Value: "text3",
						t:     Text,
					},
				},

				{
					{
						Value: "text4",
						t:     Text,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitByLexeme(tt.args.lms, tt.args.lt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitByLexeme() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRootRule_Check(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		val     string
		message string
		found   bool
	}{
		{
			name: "simple",
			yaml: `
main:
- cadabra
`,
			val:     "abra cadabra",
			message: "found: \"abra cadabra\" contains \"cadabra\" at pos 5",
			found:   true,
		},
		{
			name: "with reference and lines",
			yaml: `
ref:
- burum
- cadabra
main:
- "{ref}"
`,
			val:     "abra cadabra",
			message: "found: reference \"ref\": line index 1: \"abra cadabra\" contains \"cadabra\" at pos 5",
			found:   true,
		},
		{
			name: "with plus",
			yaml: `
main:
- abra+cadabra
`,
			val:     "abra something cadabra",
			message: "found: plus: \"abra something cadabra\" contains \"abra\" at pos 0: \"abra something cadabra\" contains \"cadabra\" at pos 15",
			found:   true,
		},
		{
			name: "with plus not found",
			yaml: `
main:
- abra+burum
`,
			val:     "abra something cadabra",
			message: "not found",
			found:   false,
		},
		{
			name: "with pipe and lines",
			yaml: `
main:
- turum|burum
- abra|cadabra"
`,
			val:     "abra something cadabra",
			message: "found: line index 1: pipe index 0: \"abra something cadabra\" contains \"abra\" at pos 0",
			found:   true,
		},
		{
			name: "with regexp",
			yaml: `
main:
- /c.d.b.a/
`,
			val:     "abra something cadabra",
			message: "found: \"abra something cadabra\" contains \"cadabra\" by regexp \"c.d.b.a\" at pos 15",
			found:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewRuleFromYaml([]byte(tt.yaml))
			if err != nil {
				t.Fatalf("Err = %v, want no error", err)
			}
			got := r.Check(tt.val)
			if !reflect.DeepEqual(got.String(), tt.message) {
				t.Errorf("Check() = %v, want %v", got, tt.message)
			}
			if !reflect.DeepEqual(got.Found, tt.found) {
				t.Errorf("Check() = %v, want %v", got.Found, tt.found)
			}
		})
	}
}
