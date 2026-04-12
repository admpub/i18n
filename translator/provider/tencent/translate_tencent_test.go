package translator

import (
	"testing"

	"github.com/admpub/i18n/translator"
	"gopkg.in/stretchr/testify.v1/assert"
)

func TestTencent(t *testing.T) {
	cfg := translator.Config{Lang: `zh-cn`, Translator: `tencent`, TranslatorConfig: `appid=&secret=`}
	err := cfg.ParseTranslatorConfig()
	assert.NoError(t, err)
	t.Logf(`config: %+v`, cfg.TranslatorParsedConfig)
	text, err := tencentTranslate(cfg, `测试`, `en`)
	assert.Equal(t, nil, err)
	assert.Equal(t, `test`, text)
}
