package main

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestBaidu(t *testing.T) {
	lang = `zh-cn`
	translatorConfig = `appid=&secret=`
	parseTranslatorConfig()
	text, err := baiduTranslate(`测试`, `en`)
	assert.Equal(t, nil, err)
	assert.Equal(t, `test`, text)
}
