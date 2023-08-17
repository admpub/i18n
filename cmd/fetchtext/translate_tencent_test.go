package main

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestTencent(t *testing.T) {
	lang = `zh-cn`
	translatorConfig = `appid=&secret=`
	parseTranslatorConfig()
	text, err := tencentTranslate(`测试`, `en`)
	assert.Equal(t, nil, err)
	assert.Equal(t, `test`, text)
}
