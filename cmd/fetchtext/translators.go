package main

var translators = map[string]func(text string, destLang string) (string, error){
	`google`:  googleTranslate,
	`baidu`:   baiduTranslate,
	`tencent`: tencentTranslate,
}

func RegisterTranslator(name string, fn func(text string, destLang string) (string, error)) {
	translators[name] = fn
}

func GetTranslator(name string) func(text string, destLang string) (string, error) {
	tr, _ := translators[name]
	return tr
}
