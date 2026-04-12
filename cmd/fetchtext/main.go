// 提取待翻译的文本
package main

import (
	"github.com/admpub/i18n/translator"
)

func main() {
	cfg := translator.Config{}
	cfg.ParseCLI()
	translator.Translate(cfg)
}
