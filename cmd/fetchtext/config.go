package main

import (
	"encoding/json"
	"net/url"

	"github.com/admpub/godotenv"
	"github.com/webx-top/com"
)

func parseTranslatorConfig() (err error) {
	if len(translatorConfig) > 0 {
		if translatorConfig[0] == '{' {
			err = json.Unmarshal([]byte(translatorConfig), &translatorParsedConfig)
			if err != nil {
				return
			}
		} else {
			var vs url.Values
			vs, err = url.ParseQuery(translatorConfig)
			if err != nil {
				return
			}
			for k := range vs {
				translatorParsedConfig[k] = vs.Get(k)
			}
		}
	}
	if com.FileExists(`.translator.env`) {
		envMap, _ := godotenv.Read(".translator.env")
		if envMap != nil {
			for k, v := range envMap {
				translatorParsedConfig[k] = v
			}
		}
	}
	if com.FileExists(`.translator_` + translator + `.env`) {
		envMap, _ := godotenv.Read(`.translator_` + translator + `.env`)
		if envMap != nil {
			for k, v := range envMap {
				translatorParsedConfig[k] = v
			}
		}
	}
	return
}
