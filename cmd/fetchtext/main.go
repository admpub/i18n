// 提取待翻译的文本
package main

import (
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
	reFunc  = regexp.MustCompile("\\.(?:SetSucT|SetOkT|SetErrT|T|E)\\([`\"](.*)[\"`]\\)")
	reFunc1 = regexp.MustCompile(`\{\{(?:[^}]*\()?T[ ]+"(.*?)"`)
	reFunc2 = regexp.MustCompile(`\{\{"(.*?)"[ ]*\|[  ]*T[ }|]`)
	reTKK   = regexp.MustCompile(`(?i)TKK\=eval\('\(\(function\(\)\{var\s+a\\x3d(-?\d+);var\s+b\\x3d(-?\d+);return\s+(\d+)\+`)

	//settings
	src       string
	dist      string
	exts      string
	lang      string
	translate bool
)

func main() {
	flag.StringVar(&src, `src`, `.`, `分析目录`)
	flag.StringVar(&dist, `dist`, `./messages`, `messages文件保存目录`)
	flag.StringVar(&exts, `exts`, `go|html`, `正则表达式`)
	flag.StringVar(&lang, `default`, `zh-cn`, `默认语言`)
	flag.BoolVar(&translate, `translate`, false, `是否自动翻译`)
	flag.Parse()

	data := map[string][]string{} //键保存待翻译的文本，值保存出现待翻译文本的文件名
	regexes := []*regexp.Regexp{reFunc, reFunc1, reFunc2}
	reExt := regexp.MustCompile(`\.(` + exts + `)$`)
	var err error
	src, err = filepath.Abs(src)
	if err != nil {
		log.Println(err)
		return
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
		os.MkdirAll(dist, 0666)
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
			text, err := t(key, destLang)
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
			err = ioutil.WriteFile(ppath, b, 0666)
		}
	}
	if err != nil {
		log.Println(err)
		return
	}
	//com.Dump(data)
}
