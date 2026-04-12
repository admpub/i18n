// 提取待翻译的文本
package main

import (
	"github.com/admpub/i18n/translator"

	_ "github.com/admpub/i18n/translator/provider/baidu"
	_ "github.com/admpub/i18n/translator/provider/google"
	_ "github.com/admpub/i18n/translator/provider/tencent"
)

func main() {
	cfg := translator.Config{}
	cfg.ParseCLI()
	translator.Translate(cfg)
}
