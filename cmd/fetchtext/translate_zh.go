package main

import (
	"strings"

	"github.com/admpub/opencc"
)

func convertChinese(content string, srcLang string, destLang string) (string, error) {
	var dictType string
	lowerSrcLang := strings.ToLower(srcLang)
	lowerDestLang := strings.ToLower(destLang)
	switch lowerDestLang {
	case `zh-tw`:
		dictType = opencc.S2TWP
	case `zh-hk`:
		dictType = opencc.S2HK
	default:
		switch lowerSrcLang {
		case `zh`, `zh-cn`:
			dictType = opencc.S2T
		case `zh-tw`:
			dictType = opencc.TW2SP
		case `zh-hk`:
			dictType = opencc.HK2S
		default:
			return content, nil
		}
	}
	cc, err := opencc.NewOpenCC(dictType)
	if err != nil {
		return ``, err
	}
	return cc.ConvertText(content), err
}
