package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

var reGoogleTKK = regexp.MustCompile(`(?i)TKK\=eval\('\(\(function\(\)\{var\s+a\\x3d(-?\d+);var\s+b\\x3d(-?\d+);return\s+(\d+)\+`)

func googleTK() (string, error) {
	url := `http://translate.google.cn`
	resp, e := http.Get(url)
	if e != nil {
		return ``, e
	}
	defer resp.Body.Close()
	b, e := io.ReadAll(resp.Body)
	if e != nil {
		return ``, e
	}
	rs := reGoogleTKK.FindSubmatch(b)
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
	resp, e := http.Get(url)
	if e != nil {
		return text, e
	}
	defer resp.Body.Close()
	b, e := io.ReadAll(resp.Body)
	if e != nil {
		return text, e
	}
	if resp.StatusCode != 200 {
		return text, fmt.Errorf("[%v][%v] %v", resp.StatusCode, resp.Status, string(b))
	}
	r := []interface{}{}
	e = json.Unmarshal(b, &r)
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
	log.Println(`Google Translate:`, string(b))
	return text, nil
}
