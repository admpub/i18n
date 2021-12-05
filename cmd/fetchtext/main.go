// 提取待翻译的文本
package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/admpub/confl"
)

var (
	reFunc       = regexp.MustCompile("\\.(?:SetSucT|SetOkT|SetErrT|T|E)\\(`([^`]+)`") // ctx.T(`text``) ctx.T(`%stext``,"a") ctx.E(`text``) ctx.E(`%stext`,"a")
	reFunc0      = regexp.MustCompile(`\.(?:SetSucT|SetOkT|SetErrT|T|E)\("([^"]+)"`)   // ctx.T("text") ctx.T("%stext","a") ctx.E("text") ctx.E("%stext","a")
	reTplFunc    = regexp.MustCompile(`\{\{(?:[^}]*\()?T[ ]+"(.*?)"`)                  // {{T "text"}} {{T "%dtext" 1}} {{printf "other%s" (T "%dtext" 1)}}
	reTplFunc0   = regexp.MustCompile("\\{\\{(?:[^}]*\\()?T[ ]+`(.*?)`")               // {{T `text``}} {{T `%dtext`` 1}} {{printf "other%s" (T `%dtext`` 1)}}
	reTplFunc1   = regexp.MustCompile(`\{\{"(.*?)"[ ]*\|[ ]*T[ }|]`)                   // {{"text"|T}} {{"text"|T|ToHTML}}
	reTplFunc1_0 = regexp.MustCompile("\\{\\{`(.*?)`[ ]*\\|[ ]*T[ }|]")                // {{`text`|T}} {{`text`|T|ToHTML}}
	reJSFunc     = regexp.MustCompile(`App\.t\('([^']+)'`)                             // App.t('text') App.t('%stext','a')
	reJSFunc0    = regexp.MustCompile(`App\.t\("([^"]+)"`)                             // App.t("text") App.t("%stext",'a')

	//settings
	src                    string
	dist                   string
	exts                   string
	lang                   string
	translate              bool
	translator             string
	translatorConfig       string
	translatorParsedConfig = map[string]string{}
)

func main() {
	flag.StringVar(&src, `src`, `.`, `分析目录`)
	flag.StringVar(&dist, `dist`, `./messages`, `messages文件保存目录`)
	flag.StringVar(&exts, `exts`, `go|html|js`, `正则表达式`)
	flag.StringVar(&lang, `default`, `zh-cn`, `默认语言`)
	flag.StringVar(&translator, `translator`, `google`, `翻译器类型`)
	flag.StringVar(&translatorConfig, `translatorConfig`, `{}`, `翻译器配置`)
	flag.BoolVar(&translate, `translate`, false, `是否自动翻译`)
	flag.Parse()

	data := map[string][]string{} //键保存待翻译的文本，值保存出现待翻译文本的文件名
	goRegexes := []*regexp.Regexp{reFunc, reFunc0}
	htmlRegexes := []*regexp.Regexp{reTplFunc, reTplFunc0, reTplFunc1, reTplFunc1_0}
	jsRegexes := []*regexp.Regexp{reJSFunc, reJSFunc0}
	reExt := regexp.MustCompile(`\.(` + exts + `)$`)
	var err error
	src, err = filepath.Abs(src)
	if err != nil {
		log.Println(err)
		return
	}
	translatorFn := GetTranslator(translator)
	if translatorFn == nil {
		log.Println(`unsupported translator:`, translator)
		return
	}
	if len(translatorConfig) > 0 {
		err = json.Unmarshal([]byte(translatorConfig), &translatorParsedConfig)
		if err != nil {
			log.Println(err)
			return
		}
	}
	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == `vendor` || strings.HasPrefix(info.Name(), `.`) {
				return filepath.SkipDir
			}
			return nil
		}
		if !reExt.MatchString(info.Name()) {
			return nil
		}
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		var regexes []*regexp.Regexp
		switch strings.ToLower(filepath.Ext(info.Name())) {
		case `.go`:
			regexes = goRegexes
		case `.html`, `.htm`:
			regexes = htmlRegexes
		case `.js`:
			regexes = jsRegexes
		default:
			return err
		}
		for _, re := range regexes {
			for _, b := range re.FindAllSubmatch(content, -1) {
				s := string(b[1])
				if len(s) == 0 {
					continue
				}
				if _, y := data[s]; y {
					data[s] = append(data[s], path)
					continue
				}
				data[s] = []string{path}
			}
		}
		log.Println(path)
		return nil
	})
	if err != nil {
		log.Println(err)
		return
	}
	dist, err = filepath.Abs(dist)
	if err != nil {
		log.Println(err)
		return
	}
	if fi, err := os.Stat(dist); err != nil || !fi.IsDir() {
		os.MkdirAll(dist, os.ModePerm)
	}
	var has bool
	err = filepath.Walk(dist, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), `.yaml`) {
			return nil
		}
		has = true
		rows := map[string]string{}
		_, err = confl.DecodeFile(path, &rows)
		if err != nil {
			log.Println(path, `[Error]`)
			return err
		}
		var hasNew bool
		for key := range data {
			if _, y := rows[key]; y {
				continue
			}
			destLang := strings.TrimSuffix(info.Name(), `.yaml`)
			if !translate || lang == destLang {
				continue
			}
			text, err := translatorFn(key, destLang)
			if err != nil {
				return err
			}
			if text == key {
				continue
			}
			rows[key] = text
			hasNew = true
		}
		if hasNew {
			b, err := confl.Marshal(rows)
			if err != nil {
				return err
			}
			return ioutil.WriteFile(path, b, info.Mode())
		}
		return nil
	})
	if err == nil && !has && len(lang) > 0 {
		ppath := filepath.Join(dist, lang+`.yaml`)
		var b []byte
		rows := map[string]string{}
		b, err = confl.Marshal(rows)
		if err == nil {
			err = ioutil.WriteFile(ppath, b, 0766)
		}
	}
	if err != nil {
		log.Println(err)
		return
	}
	//com.Dump(data)
}
