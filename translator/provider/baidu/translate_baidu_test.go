package translator

import (
	"testing"

	"github.com/admpub/i18n/translator"
	"gopkg.in/stretchr/testify.v1/assert"
)

func TestBaidu(t *testing.T) {
	cfg := translator.Config{Lang: `zh-cn`, TranslatorConfig: `appid=&secret=`}
	cfg.ParseTranslatorConfig()
	text, err := baiduTranslate(cfg, `测试`, `en`)
	assert.Equal(t, nil, err)
	assert.Equal(t, `test`, text)
}
