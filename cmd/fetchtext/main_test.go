package main

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestRegexp(t *testing.T) {
	for _, test := range []string{"ctx.T(`text`)", "ctx.T(`%stext`,\"a\")", "ctx.E(`text``)", "ctx.E(`%stext`,\"a\")"} {
		t.Log(test)
		assert.True(t, reFunc.MatchString(test))
	}
	for _, test := range []string{`ctx.T("text")`, `ctx.T("%stext","a")`, `ctx.E("text")`, `ctx.E("%stext","a")`} {
		t.Log(test)
		assert.True(t, reFunc0.MatchString(test))
	}
	for _, test := range []string{".NewError(code.InvalidParameter, `test`"} {
		t.Log(test)
		assert.True(t, reFunc1.MatchString(test))
	}
	for _, test := range []string{`.NewError(code.InvalidParameter, "test")`} {
		t.Log(test)
		assert.True(t, reFunc1_0.MatchString(test))
	}
	for _, test := range []string{`{{$.T "text"}}`, `{{ $.T "text" }}`, `{{- $.T "text" -}}`, `{{$.T "%dtext" 1}}`, `{{printf "other%s" ($.T "%dtext" 1)}}`} {
		t.Log(test)
		assert.True(t, reTplFunc.MatchString(test))
	}
	for _, test := range []string{"{{$.T `text`}}", "{{ $.T `text` }}", "{{- $.T `text` -}}", "{{$.T `%dtext` 1}}", "{{printf \"other%s\" ($.T `%dtext` 1)}}"} {
		t.Log(test)
		assert.True(t, reTplFunc0.MatchString(test))
	}
	for _, test := range []string{`{{"text"|$.T}}`, `{{ "text" | $.T }}`, `{{- "text" | $.T -}}`, `{{"text"|$.T|ToHTML}}`} {
		t.Log(test)
		assert.True(t, reTplFunc1.MatchString(test))
	}
	for _, test := range []string{"{{`text`|$.T}}", "{{`text`|$.T|ToHTML}}"} {
		t.Log(test)
		assert.True(t, reTplFunc1_0.MatchString(test))
	}
	for _, test := range []string{"{{`{{`}}$ads := {{$tag}} `{{`广告位标识`|$.T}}`{{`}}`}}"} {
		t.Log(test)
		assert.Equal(t, `广告位标识`, reTplFunc1_0.FindAllStringSubmatch(test, -1)[0][1])
		assert.True(t, reTplFunc1_0.MatchString(test))
	}
	for _, test := range []string{`{{"te\"xt"|$.T}}`} {
		t.Log(test)
		assert.Equal(t, `te\"xt`, reTplFunc1.FindAllStringSubmatch(test, -1)[0][1])
		t.Log(reTplFunc1.FindAllStringSubmatch(test, -1))
		assert.True(t, reTplFunc1.MatchString(test))
	}
	for _, test := range []string{`{{$.T "te\"xt"}}`} {
		t.Log(test)
		t.Log(reTplFunc.FindAllStringSubmatch(test, -1))
		assert.Equal(t, `te\"xt`, reTplFunc.FindAllStringSubmatch(test, -1)[0][1])
		assert.True(t, reTplFunc.MatchString(test))
	}
}
