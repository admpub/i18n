// 提取待翻译的文本
package translator

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/admpub/confl"
	"github.com/admpub/i18n"
	"github.com/coscms/forms"
	formsconfig "github.com/coscms/forms/config"
	"github.com/webx-top/com"
	"gopkg.in/yaml.v3"
)

const tplTagS = `(?:-?[ ]*)?`
const tplTxtE = `[ )}|]`

var (
	reFunc  = regexp.MustCompile("\\.(?:T|E)\\(`([^`]+)`")  // ctx.T(`text``) ctx.T(`%stext``,"a") ctx.E(`text``) ctx.E(`%stext`,"a")
	reFunc0 = regexp.MustCompile(`\.(?:T|E)\("(.*?)"[ ,)]`) // ctx.T("text") ctx.T("%stext","a") ctx.E("text") ctx.E("%stext","a")

	reFunc1   = regexp.MustCompile("\\.NewError\\((?:code|stdCode|codes|xcode)\\.[\\w]+,[ ]?`([^`]+)`") //.NewError(code.InvalidParameter, `aaa`
	reFunc1_0 = regexp.MustCompile(`\.NewError\((?:code|stdCode|codes|xcode)\.[\w]+,[ ]?"(.*?)"[ ,)]`)  //.NewError(code.InvalidParameter, "")

	reTplFunc    = regexp.MustCompile(`\{\{` + tplTagS + `(?:[^}]*\()?\$\.(?:T|RawT)[ ]+"(.*?)"` + tplTxtE)  // {{$.T "text"}} {{$.T "%dtext" 1}} {{printf "other%s" ($.T "%dtext" 1)}}
	reTplFunc0   = regexp.MustCompile("\\{\\{" + tplTagS + "(?:[^}]*\\()?\\$\\.(?:T|RawT)[ ]+`([^`]+)`")     // {{$.T `text``}} {{$.T `%dtext`` 1}} {{printf "other%s" ($.T `%dtext`` 1)}}
	reTplFunc1   = regexp.MustCompile(`\{\{` + tplTagS + `"(.*?)"[ ]*\|[ ]*\$\.(?:T|RawT)` + tplTxtE)        // {{"text"|$.T}} {{"text"|$.T|ToHTML}}
	reTplFunc1_0 = regexp.MustCompile("\\{\\{" + tplTagS + "`([^`]+)`[ ]*\\|[ ]*\\$\\.(?:T|RawT)" + tplTxtE) // {{`text`|$.T}} {{`text`|$.T|ToHTML}}
	reJSFunc     = regexp.MustCompile(`App\.t\('(.*?)'[ ,)]`)                                                // App.t('text') App.t('%stext','a')
	reJSFunc0    = regexp.MustCompile(`App\.t\("(.*?)"[ ,)]`)                                                // App.t("text") App.t("%stext",'a')
	reChinese    = regexp.MustCompile(`[\p{Han}]+`)
	nltrReplacer = strings.NewReplacer(`\n`, "\n", `\t`, "\t", `\r`, "\r")
)

func Translate(tcfg Config) {
	data := map[string][]string{} //键保存待翻译的文本，值保存出现待翻译文本的文件名
	goRegexes := []*regexp.Regexp{reFunc, reFunc0, reFunc1, reFunc1_0}
	htmlRegexes := []*regexp.Regexp{reTplFunc, reTplFunc0, reTplFunc1, reTplFunc1_0}
	jsRegexes := []*regexp.Regexp{reJSFunc, reJSFunc0}
	reExt := regexp.MustCompile(`\.(` + tcfg.Exts + `)$`)
	stripSlashes := []*regexp.Regexp{reFunc0, reFunc1_0, reTplFunc, reTplFunc1, reJSFunc0}
	var readResourceFile = func(path string, info os.FileInfo) error {
		if !reExt.MatchString(info.Name()) {
			return nil
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
			var texts map[string]struct{}
			if strings.HasSuffix(info.Name(), `.form.json`) {
				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				var cfg *formsconfig.Config
				cfg, err = forms.Unmarshal(content, path)
				if err != nil {
					return err
				}
				texts = cfg.GetMultilingualText()
			} else if strings.HasSuffix(info.Name(), `.form.yaml`) || strings.HasSuffix(info.Name(), `.form.yml`) {
				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				cfg := &formsconfig.Config{}
				err = yaml.Unmarshal(content, cfg)
				if err != nil {
					return err
				}
				texts = cfg.GetMultilingualText()
			}
			for k := range texts {
				if _, y := data[k]; y {
					data[k] = append(data[k], path)
					continue
				}
				data[k] = []string{path}
			}
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, re := range regexes {
			needStripSlashes := slices.Contains(stripSlashes, re)
			for _, b := range re.FindAllSubmatch(content, -1) {
				s := string(b[1])
				if len(s) == 0 {
					continue
				}
				if needStripSlashes {
					s = com.StripSlashesOnlyQuote(s)
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
	tcfg.Src, err = filepath.Abs(tcfg.Src)
	if err != nil {
		log.Println(err)
		return
	}
	translatorFn := GetTranslator(tcfg.Translator)
	if translatorFn == nil {
		log.Println(`unsupported translator:`, tcfg.Translator)
		return
	}
	if err = tcfg.ParseTranslatorConfig(); err != nil {
		log.Println(err)
	}
	var vendorDirsSlice []string = []string{
		`github.com/nging-plugins`,
		`github.com/coscms/webcore`,
		`github.com/coscms/webfront`,
	}
	if len(tcfg.VendorDirs) > 0 {
		for _, v := range strings.Split(tcfg.VendorDirs, `|`) {
			v = strings.TrimSpace(v)
			if len(v) > 0 {
				vendorDirsSlice = append(vendorDirsSlice, v)
			}
		}
	}
	err = filepath.Walk(tcfg.Src, func(path string, info os.FileInfo, err error) error {
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
	tcfg.Dist, err = filepath.Abs(tcfg.Dist)
	if err != nil {
		log.Println(err)
		return
	}
	if fi, err := os.Stat(tcfg.Dist); err != nil || !fi.IsDir() {
		os.MkdirAll(tcfg.Dist, os.ModePerm)
	}
	var has bool
	err = filepath.Walk(tcfg.Dist, func(path string, info os.FileInfo, err error) error {
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
		if tcfg.OnlyExportParsed {
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
		if tcfg.Lang == destLang {
			return err
		}
		var hasNew bool
		for key := range data {
			oldText, existsText := rows[key]
			noPrefixKey := i18n.TrimGroupPrefix(key)
			var text string
			if tcfg.Translate {
				if !existsText {
					oldText = noPrefixKey
				} else if len(oldText) > 0 {
					if tcfg.OnlyTranslateIncr {
						continue
					}
				}
				needTr := tcfg.ForceAll || needTranslation(oldText, destLang)
				if !needTr {
					continue
				}
				srcText := nltrReplacer.Replace(noPrefixKey)
				if destLang == `zh-TW` || destLang == `zh-HK` {
					text, err = convertChinese(srcText, tcfg.Lang, destLang)
				} else {
					text, err = translatorFn(tcfg, srcText, destLang)
					if err != nil {
						wait := time.Second * 2
						for i := 0; i < 5; i++ {
							log.Println(err.Error(), `Will retry after `+wait.String(), fmt.Sprintf(`(%d/%d)`, i+1, 5))
							time.Sleep(wait)
							text, err = translatorFn(tcfg, srcText, destLang)
							if err == nil {
								break
							}
						}
					}
				}
				if err != nil {
					log.Println(err)
					break
				}
				if text == noPrefixKey {
					continue
				}
			} else {
				text = noPrefixKey
			}
			rows[key] = text
			hasNew = true
		}
		if tcfg.Clean {
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
	if err == nil && !has && len(tcfg.Lang) > 0 {
		ppath := filepath.Join(tcfg.Dist, tcfg.Lang+`.yaml`)
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
	if destLang == `zh-CN` || destLang == `zh-cn` || destLang == `zh` {
		return !reChinese.MatchString(text)
	}
	return reChinese.MatchString(text)
}
