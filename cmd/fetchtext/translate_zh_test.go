package main

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestConvertChinese(t *testing.T) {
	r, err := convertChinese(``, `zh-CN`, `zh-TW`)
	assert.NoError(t, err)
	t.Log(r)
}
