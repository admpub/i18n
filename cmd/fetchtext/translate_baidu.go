package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/webx-top/com"
)

/*
{
    "from": "zh",
    "to": "en",
    "trans_result": [
        {
            "src": "中国",
            "dst": "China"
        }
    ]
}
*/

type baiduTransResult struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}
type baiduResponse struct {
	From    string              `json:"from"`
	To      string              `json:"to"`
	Results []*baiduTransResult `json:"trans_result"`
}

func baiduTranslate(text string, destLang string) (string, error) {
	values := url.Values{
		`q`:     []string{text},
		`from`:  []string{strings.SplitN(lang, `-`, 2)[0]},
		`to`:    []string{strings.SplitN(destLang, `-`, 2)[0]},
		`appid`: []string{translatorParsedConfig[`appid`]},
		`salt`:  []string{com.RandomAlphanumeric(16)},
	}
	sign := com.Md5(translatorParsedConfig[`appid`] + values.Get(`q`) + values.Get(`salt`) + translatorParsedConfig[`secret`]) //  appid+q+salt+密钥
	values.Add(`sign`, sign)
	url := `https://fanyi-api.baidu.com/api/trans/vip/translate?` + values.Encode()
	resp, e := http.Get(url)
	if e != nil {
		return text, e
	}
	defer resp.Body.Close()
	b, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return text, e
	}
	if resp.StatusCode != 200 {
		return text, fmt.Errorf("[%v][%v] %v", resp.StatusCode, resp.Status, string(b))
	}
	r := &baiduResponse{}
	err := json.Unmarshal(b, r)
	if err != nil {
		return text, err
	}
	//echo.Dump(r)
	for _, v := range r.Results {
		return v.Dst, nil
	}
	return text, nil
}
