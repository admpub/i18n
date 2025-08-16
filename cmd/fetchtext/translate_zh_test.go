package main

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestConvertChinese(t *testing.T) {
	r, err := convertChinese(`Nging 是一个 Go 语言开发的 Web 服务面板系统，可以配置 Caddy 和 Nginx 站点，并附带了实用的周边工具，例如：计划任务、MySQL 管理、Redis 管理、FTP 管理、SSH 管理、服务器管理等。`, `zh-CN`, `zh-TW`)
	assert.NoError(t, err)
	t.Log(r)
}
