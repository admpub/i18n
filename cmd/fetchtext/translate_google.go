package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

func tk() string {
	url := `http://translate.google.cn`
	resp, e := http.Get(url)
	if e != nil {
		log.Println(e)
		return ``
	}
	defer resp.Body.Close()
	b, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		log.Println(e)
		return ``
	}
	rs := reTKK.FindSubmatch(b)
	if len(rs) > 3 {
		n1, e := strconv.Atoi(string(rs[1]))
		if e != nil {
			log.Println(e)
			return ``
		}
		n2, e := strconv.Atoi(string(rs[2]))
		if e != nil {
			log.Println(e)
			return ``
		}
		return string(rs[2]) + `.` + strconv.Itoa(n1+n2)
	}
	return ``
}

func t(text string, distLang string) string {
	if !translate || lang == distLang {
		return text
	}
	//TODO: Automatic translation
	url := `http://translate.google.cn/translate_a/single?client=gtx&sl=` + lang + `&tl=` + distLang + `&dt=t&q=` + url.QueryEscape(text)
	url += `&ie=UTF-8&oe=UTF-8`
	//url += `&tk=` + tk()
	resp, e := http.Get(url)
	if e != nil {
		log.Println(e)
		return text
	}
	defer resp.Body.Close()
	b, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		log.Println(e)
		return text
	}
	r := []interface{}{}
	e = json.Unmarshal(b, &r)
	if e != nil {
		log.Println(e)
		log.Println(string(b))
		return text
	}
	if len(r) == 3 {
		if v, y := r[0].([]interface{}); y {
			if len(v) > 0 {
				v, y = v[0].([]interface{})
				if y && len(v) > 0 {
					if val, ok := v[0].(string); ok {
						log.Printf("[ %s -> %s ] %s -> %s\n", lang, distLang, text, val)
						return val
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
	return text
}
