package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"

	"github.com/webx-top/restyclient"
)

var reGoogleTKK = regexp.MustCompile(`(?i)TKK\=eval\('\(\(function\(\)\{var\s+a\\x3d(-?\d+);var\s+b\\x3d(-?\d+);return\s+(\d+)\+`)

func googleTK() (string, error) {
	url := `http://translate.google.cn`
	req := restyclient.Classic()
	resp, e := req.Get(url)
	if e != nil {
		return ``, e
	}
	if !resp.IsSuccess() {
		return ``, fmt.Errorf("[%v][%s] %s", resp.StatusCode(), resp.Status(), resp.Body())
	}
	rs := reGoogleTKK.FindSubmatch(resp.Body())
	if len(rs) > 3 {
		n1, e := strconv.Atoi(string(rs[1]))
		if e != nil {
			return ``, e
		}
		n2, e := strconv.Atoi(string(rs[2]))
		if e != nil {
			return ``, e
		}
		return string(rs[2]) + `.` + strconv.Itoa(n1+n2), nil
	}
	return ``, nil
}

func googleTranslate(text string, destLang string) (string, error) {
	//TODO: Automatic translation
	//http://translate.google.cn/translate_a/single?client=gtx&sl=zh-cn&tl=en&dt=t&q=中国&ie=UTF-8&oe=UTF-8
	url := `http://translate.google.cn/translate_a/single?client=gtx&sl=` + lang + `&tl=` + destLang + `&dt=t&q=` + url.QueryEscape(text)
	url += `&ie=UTF-8&oe=UTF-8`
	/*
		tk, err := tk()
		if err != nil {
			return ``, err
		}
		url += `&tk=` + tk
	*/
	req := restyclient.Classic()
	resp, e := req.Get(url)
	if e != nil {
		return text, e
	}
	if !resp.IsSuccess() {
		return text, fmt.Errorf("[%v][%s] %s", resp.StatusCode(), resp.Status(), resp.Body())
	}
	r := []interface{}{}
	e = json.Unmarshal(resp.Body(), &r)
	if e != nil {
		return text, e
	}
	if len(r) == 3 {
		if v, y := r[0].([]interface{}); y {
			if len(v) > 0 {
				v, y = v[0].([]interface{})
				if y && len(v) > 0 {
					if val, ok := v[0].(string); ok {
						log.Printf("[ %s -> %s ] %s -> %s\n", lang, destLang, text, val)
						return val, nil
					}
					log.Printf(`Google Translate: r[0][0][0] => %T`, v[0])
					log.Println()
				} else {
					log.Printf(`Google Translate: r[0][0] => %T`, v)
					log.Println()
				}
			}
		} else {
			log.Printf(`Google Translate: r[0] => %T`, r[0])
			log.Println()
		}
	}
	log.Println(`Google Translate:`, string(resp.Body()))
	return text, nil
}
