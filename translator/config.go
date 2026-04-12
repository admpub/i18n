package translator

import (
	"encoding/json"
	"flag"
	"net/url"

	"github.com/admpub/godotenv"
	"github.com/webx-top/com"
)

type Config struct {
	Src                    string
	Dist                   string
	Exts                   string
	Lang                   string
	Translate              bool
	Translator             string
	TranslatorConfig       string
	TranslatorParsedConfig map[string]string
	OnlyExportParsed       bool
	ForceAll               bool
	Clean                  bool
	VendorDirs             string
	EnvFile                string
	OnlyTranslateIncr      bool
}

func (c *Config) ParseCLI() {
	flag.StringVar(&c.Src, `src`, `.`, `分析目录`)
	flag.StringVar(&c.Dist, `dist`, `./messages`, `messages文件保存目录`)
	flag.StringVar(&c.Exts, `exts`, `go|html|js|form\.json|form\.yaml|form\.yml`, `正则表达式`)
	flag.StringVar(&c.Lang, `default`, `zh-CN`, `默认语言`)
	flag.StringVar(&c.Translator, `translator`, `google`, `翻译器类型`)
	flag.StringVar(&c.TranslatorConfig, `translatorConfig`, ``, `翻译器配置(例如百度翻译配置为: appid=APPID&secret=SECRET)`)
	flag.StringVar(&c.EnvFile, `envFile`, ``, `环境变量配置文件`)
	flag.StringVar(&c.VendorDirs, `vendorDirs`, ``, `依赖子文件夹`)
	flag.BoolVar(&c.Translate, `translate`, false, `是否自动翻译`)
	flag.BoolVar(&c.ForceAll, `forceAll`, false, `是否翻译全部`)
	flag.BoolVar(&c.Clean, `clean`, false, `是否清理不存在的译文`)
	flag.BoolVar(&c.OnlyExportParsed, `onlyExport`, false, `是否仅仅导出解析语言文件后的json数据`)
	flag.BoolVar(&c.OnlyTranslateIncr, `onlyTranslateIncr`, false, `是否仅仅翻译新增的未翻译文本`)
	flag.Parse()
}

func (c *Config) ParseTranslatorConfig() (err error) {
	c.TranslatorParsedConfig = map[string]string{}
	if len(c.TranslatorConfig) > 0 {
		if c.TranslatorConfig[0] == '{' {
			err = json.Unmarshal([]byte(c.TranslatorConfig), &c.TranslatorParsedConfig)
			if err != nil {
				return
			}
		} else {
			var vs url.Values
			vs, err = url.ParseQuery(c.TranslatorConfig)
			if err != nil {
				return
			}
			for k := range vs {
				c.TranslatorParsedConfig[k] = vs.Get(k)
			}
		}
	}
	if len(c.EnvFile) > 0 {
		envMap, _ := godotenv.Read(c.EnvFile)
		for k, v := range envMap {
			c.TranslatorParsedConfig[k] = v
		}
	}
	if com.FileExists(`.translator.env`) {
		envMap, _ := godotenv.Read(".translator.env")
		for k, v := range envMap {
			c.TranslatorParsedConfig[k] = v
		}
	}
	if com.FileExists(`.translator_` + c.Translator + `.env`) {
		envMap, _ := godotenv.Read(`.translator_` + c.Translator + `.env`)
		for k, v := range envMap {
			c.TranslatorParsedConfig[k] = v
		}
	}
	return
}
