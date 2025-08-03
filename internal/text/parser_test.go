// SPDX-License-Identifier: Apache-2.0

package text

import (
	"reflect"
	"testing"
)

func TestMatch(t *testing.T) {
	type args struct {
		pattern   string
		value     string
		trimInput bool
	}

	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{"Regex", args{pattern: `/(command|other)/`, value: `command arg`, trimInput: true}, "arg", true},
		{"Regex with casing", args{pattern: `/(COMMAND|OTHER)/`, value: `other arg`, trimInput: true}, "arg", true},
		{"Regular", args{pattern: `command`, value: `command arg`, trimInput: true}, "arg", true},
		{"Casing", args{pattern: `command`, value: `CoMMaND arg`, trimInput: true}, "arg", true},
		{"Casing Keep Msg", args{pattern: `command`, value: `CoMMaND arg`, trimInput: false}, "CoMMaND arg", true},
		{"Space after", args{pattern: `command`, value: `command`, trimInput: true}, "", true},
		{"Nospace", args{pattern: `command`, value: `commandarg`, trimInput: true}, "", false},
		{"Space", args{pattern: `command`, value: `command `, trimInput: true}, "", true},
		{"Fail", args{pattern: `command`, value: `dnammoc`, trimInput: true}, "", false},
		{"Unsupported Regex", args{pattern: `/^(?!.*(hello|goodday|hi)).*issue.*$/`, value: `oh goodday what is the issue`, trimInput: true}, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := Match(tt.args.pattern, tt.args.value, tt.args.trimInput)
			if got != tt.want {
				t.Errorf("Match() got = %v, want %v", got, tt.want)
			}

			if got1 != tt.want1 {
				t.Errorf("Match() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestSubstitute(t *testing.T) {
	type args struct {
		value  string
		tokens map[string]string
	}

	t.Setenv("TEST_ENV_VAR", "1234")

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"Simple", args{value: `${test}`, tokens: map[string]string{"test": "testvalue"}}, "testvalue", false},
		{"Fail", args{value: `${fail}`, tokens: map[string]string{"test": "testvalue"}}, "${fail}", true},
		{"Env var", args{value: `${TEST_ENV_VAR}`, tokens: map[string]string{}}, "1234", false},
		{"Env var and var", args{value: `${TEST_ENV_VAR}`, tokens: map[string]string{"TEST_ENV_VAR": "testvalue"}}, "testvalue", false},
		{"Token exists but value empty", args{value: `${test}`, tokens: map[string]string{"test": ""}}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Substitute(tt.args.value, tt.args.tokens)
			if (err != nil) != tt.wantErr {
				t.Errorf("Substitute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Substitute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleArgTokenizer(t *testing.T) {
	type args struct {
		stripped string
	}

	var empty []string

	tests := []struct {
		name string
		args args
		want []string
	}{
		{"Simple", args{stripped: `simple`}, []string{"simple"}},
		{"Multi", args{stripped: `simple twice`}, []string{"simple", "twice"}},
		{"With quoted text", args{stripped: `simple twice "oh my"`}, []string{"simple", "twice", "oh my"}},
		{"No text", args{stripped: ``}, empty},
		{"Extra spaces", args{stripped: `  simple   twice   wow`}, []string{"simple", "twice", "wow"}},
		{"With slashes", args{stripped: `simple this\that`}, []string{"simple", `this\that`}},
		{"Funny quotes", args{stripped: `simple “quotes for days”`}, []string{"simple", "quotes for days"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RuleArgTokenizer(tt.args.stripped); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecArgTokenizer(t *testing.T) {
	type args struct {
		stripped string
	}

	var empty []string

	tests := []struct {
		name string
		args args
		want []string
	}{
		{"Single", args{stripped: `one`}, []string{"one"}},
		{"Single, Empty", args{stripped: ``}, empty},
		{"Single, Including Slash", args{stripped: `this\that`}, []string{`this\that`}},
		{"Single, Extra Whitespace", args{stripped: ` one `}, []string{"one"}},
		{"Multiple", args{stripped: `one two three`}, []string{"one", "two", "three"}},
		{"Multiple, Extra Whitespace", args{stripped: ` one    two     three  `}, []string{"one", "two", "three"}},
		{"Single, Quoted (Single Quotes)", args{stripped: `one 'quoted arg'`}, []string{"one", "quoted arg"}},
		{"Single, Quoted (Double Quotes)", args{stripped: `one "quoted arg"`}, []string{"one", "quoted arg"}},
		{"Single, Quoted (Smart Quotes)", args{stripped: `one “quoted arg”`}, []string{"one", "quoted arg"}},
		{"Single, Quoted (Empty)", args{stripped: `one “”`}, []string{"one", ""}},
		{"Multiple, Quoted (Single Quotes)", args{stripped: `one two 'quoted arg' three`}, []string{"one", "two", "quoted arg", "three"}},
		{"Multiple, Quoted (Double Quotes)", args{stripped: `one two "quoted arg" three`}, []string{"one", "two", "quoted arg", "three"}},
		{"Multiple, Quoted (Smart Quotes)", args{stripped: `one two “quoted arg” three`}, []string{"one", "two", "quoted arg", "three"}},
		{"Multiple, Quoted (Mixed Quotes)", args{stripped: `one two 'quoted arg1' “quoted arg2” "quoted arg3" three`}, []string{"one", "two", "quoted arg1", "quoted arg2", "quoted arg3", "three"}},
		{"Multiple, Quoted (Embedded Quotes)", args{stripped: `one two 'quoted ”“" arg1' “quoted "' arg2” "quoted ”“' arg3" three`}, []string{"one", "two", `quoted ”“" arg1`, `quoted "' arg2`, `quoted ”“' arg3`, "three"}},
		{"Multiple, Quoted (Some Empty)", args{stripped: `one two '' “” "" three "" four '' five “” six`}, []string{"one", "two", "", "", "", "three", "", "four", "", "five", "", "six"}},
		{"Multiple, Quoted (Mismatched Quotes, Single/Double)", args{stripped: `one two 'quoted arg" three`}, []string{"one", "two", "quoted", "arg", "three"}},
		{"Multiple, Quoted (Mismatched Quotes, Single/Smart Close)", args{stripped: `one two 'quoted arg” three`}, []string{"one", "two", "quoted", "arg", "three"}},
		{"Multiple, Quoted (Mismatched Quotes, Single/Smart Open)", args{stripped: `one two 'quoted arg“ three`}, []string{"one", "two", "quoted", "arg", "three"}},
		{"Multiple, Quoted (Mismatched Quotes, Double/Smart Close)", args{stripped: `one two "quoted arg” three`}, []string{"one", "two", "quoted", "arg", "three"}},
		{"Multiple, Quoted (Mismatched Quotes, Double/Smart Open)", args{stripped: `one two "quoted arg“ three`}, []string{"one", "two", "quoted", "arg", "three"}},
		{"Multiple, Quoted (Mismatched Quotes, Double/Single)", args{stripped: `one two "quoted arg' three`}, []string{"one", "two", "quoted", "arg", "three"}},
		{"Multiple, Quoted (Mismatched Quotes, Smart Close/Single)", args{stripped: `one two ”quoted arg' three`}, []string{"one", "two", "quoted", "arg", "three"}},
		{"Multiple, Quoted (Mismatched Quotes, Smart Open/Single)", args{stripped: `one two “quoted arg' three`}, []string{"one", "two", "quoted", "arg", "three"}},
		{"Multiple, Quoted (Mismatched Quotes, Smart Close/Double)", args{stripped: `one two ”quoted arg" three`}, []string{"one", "two", "quoted", "arg", "three"}},
		{"Multiple, Quoted (Mismatched Quotes, Smart Open/Double)", args{stripped: `one two “quoted arg" three`}, []string{"one", "two", "quoted", "arg", "three"}},
		{"Multiple, Quoted (Mismatched Quotes, Smart Close/Smart Close)", args{stripped: `one two ”quoted arg” three`}, []string{"one", "two", "quoted", "arg", "three"}},
		{"Multiple, Quoted (Mismatched Quotes, Smart Open/Smart Open)", args{stripped: `one two “quoted arg“ three`}, []string{"one", "two", "quoted", "arg", "three"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExecArgTokenizer(tt.args.stripped); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
