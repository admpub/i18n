package translator

var translators = map[string]func(tcfg Config, text string, destLang string) (string, error){
	`google`:  googleTranslate,
	`baidu`:   baiduTranslate,
	`tencent`: tencentTranslate,
}

func RegisterTranslator(name string, fn func(tcfg Config, text string, destLang string) (string, error)) {
	translators[name] = fn
}

func GetTranslator(name string) func(tcfg Config, text string, destLang string) (string, error) {
	tr, _ := translators[name]
	return tr
}
