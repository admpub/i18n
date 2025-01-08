// 提取待翻译的文本
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/admpub/confl"
	"github.com/webx-top/com"
)

var (
	reFunc  = regexp.MustCompile("\\.(?:T|E)\\(`([^`]+)`") // ctx.T(`text``) ctx.T(`%stext``,"a") ctx.E(`text``) ctx.E(`%stext`,"a") .NewError(code.InvalidParameter, `
	reFunc0 = regexp.MustCompile(`\.(?:T|E)\("([^"]+)"`)   // ctx.T("text") ctx.T("%stext","a") ctx.E("text") ctx.E("%stext","a")

	reFunc1   = regexp.MustCompile("\\.NewError\\((?:code|stdCode|codes)\\.[\\w]+,[ ]?`([^`]+)`")
	reFunc1_0 = regexp.MustCompile(`\.NewError\((?:code|stdCode|codes)\.[\w]+,[ ]?"([^"]+)"`)

	reTplFunc    = regexp.MustCompile(`\{\{(?:[^}]*\()?\$\.T[ ]+"(.*?)"`)      // {{$.T "text"}} {{$.T "%dtext" 1}} {{printf "other%s" ($.T "%dtext" 1)}}
	reTplFunc0   = regexp.MustCompile("\\{\\{(?:[^}]*\\()?\\$\\.T[ ]+`(.*?)`") // {{$.T `text``}} {{$.T `%dtext`` 1}} {{printf "other%s" ($.T `%dtext`` 1)}}
	reTplFunc1   = regexp.MustCompile(`\{\{"(.*?)"[ ]*\|[ ]*\$\.T[ }|]`)       // {{"text"|$.T}} {{"text"|$.T|ToHTML}}
	reTplFunc1_0 = regexp.MustCompile("\\{\\{`(.*?)`[ ]*\\|[ ]*\\$\\.T[ }|]")  // {{`text`|$.T}} {{`text`|$.T|ToHTML}}
	reJSFunc     = regexp.MustCompile(`App\.t\('([^']+)'`)                     // App.t('text') App.t('%stext','a')
	reJSFunc0    = regexp.MustCompile(`App\.t\("([^"]+)"`)                     // App.t("text") App.t("%stext",'a')
	reChinese    = regexp.MustCompile(`[\p{Han}]+`)

	//settings
	src                    string
	dist                   string
	exts                   string
	lang                   string
	translate              bool
	translator             string
	translatorConfig       string
	translatorParsedConfig = map[string]string{}
	onlyExportParsed       bool
	forceAll               bool
	clean                  bool
	vendorDirs             string
)

func main() {
	flag.StringVar(&src, `src`, `.`, `分析目录`)
	flag.StringVar(&dist, `dist`, `./messages`, `messages文件保存目录`)
	flag.StringVar(&exts, `exts`, `go|html|js`, `正则表达式`)
	flag.StringVar(&lang, `default`, `zh-cn`, `默认语言`)
	flag.StringVar(&translator, `translator`, `google`, `翻译器类型`)
	flag.StringVar(&translatorConfig, `translatorConfig`, ``, `翻译器配置(例如百度翻译配置为: appid=APPID&secret=SECRET)`)
	flag.StringVar(&vendorDirs, `vendorDirs`, ``, `依赖子文件夹`)
	flag.BoolVar(&translate, `translate`, false, `是否自动翻译`)
	flag.BoolVar(&forceAll, `forceAll`, false, `是否翻译全部`)
	flag.BoolVar(&clean, `clean`, false, `是否清理不存在的译文`)
	flag.BoolVar(&onlyExportParsed, `onlyExport`, false, `是否仅仅导出解析语言文件后的json数据`)
	flag.Parse()

	data := map[string][]string{} //键保存待翻译的文本，值保存出现待翻译文本的文件名
	goRegexes := []*regexp.Regexp{reFunc, reFunc0, reFunc1, reFunc1_0}
	htmlRegexes := []*regexp.Regexp{reTplFunc, reTplFunc0, reTplFunc1, reTplFunc1_0}
	jsRegexes := []*regexp.Regexp{reJSFunc, reJSFunc0}
	reExt := regexp.MustCompile(`\.(` + exts + `)$`)
	var readResourceFile = func(path string, info os.FileInfo) error {
		if !reExt.MatchString(info.Name()) {
			return nil
		}
		content, err := os.ReadFile(path)
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
	}
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
	if err = parseTranslatorConfig(); err != nil {
		log.Println(err)
	}
	var vendorDirsSlice []string = []string{
		`github.com/nging-plugins`,
		`github.com/coscms/webcore`,
		`github.com/coscms/webfront`,
	}
	if len(vendorDirs) > 0 {
		for _, v := range strings.Split(vendorDirs, `|`) {
			v = strings.TrimSpace(v)
			if len(v) > 0 {
				vendorDirsSlice = append(vendorDirsSlice, v)
			}
		}
	}
	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), `.`) {
				return filepath.SkipDir
			}
			if info.Name() == `vendor` {
				for _, vendorSubdir := range vendorDirsSlice {
					pluginsDir := filepath.Join(path, vendorSubdir)
					if com.FileExists(pluginsDir) {
						filepath.Walk(pluginsDir, func(path string, info os.FileInfo, err error) error {
							if err != nil {
								return err
							}
							if info.IsDir() {
								return nil
							}
							return readResourceFile(path, info)
						})
					}
				}
				return filepath.SkipDir
			}
			return nil
		}

		return readResourceFile(path, info)
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
		if onlyExportParsed {
			file, err := os.Create(info.Name() + `.json`)
			if err != nil {
				return err
			}
			defer file.Close()
			enc := json.NewEncoder(file)
			enc.SetEscapeHTML(false)
			enc.SetIndent(``, `  `)
			err = enc.Encode(rows)
			return err
		}
		destLang := strings.TrimSuffix(info.Name(), `.yaml`)
		if lang == destLang {
			return err
		}
		var hasNew bool
		for key := range data {
			oldText, existsText := rows[key]
			var text string
			if translate {
				if !existsText {
					oldText = key
				}
				needTr := forceAll || needTranslation(oldText, destLang)
				if !needTr {
					continue
				}
				text, err = translatorFn(key, destLang)
				if err != nil {
					wait := time.Second * 2
					for i := 0; i < 5; i++ {
						log.Println(err.Error(), `Will retry after `+wait.String(), fmt.Sprintf(`(%d/%d)`, i+1, 5))
						time.Sleep(wait)
						text, err = translatorFn(key, destLang)
						if err == nil {
							break
						}
					}
				}
				if err != nil {
					log.Println(err)
					break
				}
				if text == key {
					continue
				}
			} else {
				text = key
			}
			rows[key] = text
			hasNew = true
		}
		if clean {
			for key := range rows {
				if _, ok := data[key]; !ok {
					delete(rows, key)
					if !hasNew {
						hasNew = true
					}
				}
			}
		}
		if hasNew {
			b, err := confl.Marshal(rows)
			if err != nil {
				return err
			}
			return os.WriteFile(path, b, info.Mode())
		}
		return nil
	})
	if err == nil && !has && len(lang) > 0 {
		ppath := filepath.Join(dist, lang+`.yaml`)
		var b []byte
		rows := map[string]string{}
		b, err = confl.Marshal(rows)
		if err == nil {
			err = os.WriteFile(ppath, b, 0766)
		}
	}
	if err != nil {
		log.Println(err)
		return
	}
	//com.Dump(data)
}

func needTranslation(text string, destLang string) bool {
	if destLang == `zh-cn` {
		return !reChinese.MatchString(text)
	}
	return reChinese.MatchString(text)
}
