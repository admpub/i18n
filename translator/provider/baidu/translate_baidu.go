package translator

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/admpub/i18n/translator"
	"github.com/webx-top/com"
	"github.com/webx-top/restyclient"
)

func init() {
	translator.RegisterTranslator(`baidu`, baiduTranslate)
}

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

func baiduTranslate(tcfg translator.Config, text string, destLang string) (string, error) {
	time.Sleep(time.Second) // 接口频率限制：1次/秒
	values := url.Values{
		`q`:     []string{text},
		`from`:  []string{strings.SplitN(tcfg.Lang, `-`, 2)[0]},
		`to`:    []string{strings.SplitN(destLang, `-`, 2)[0]},
		`appid`: []string{tcfg.TranslatorParsedConfig[`appid`]},
		`salt`:  []string{com.RandomAlphanumeric(16)},
	}
	sign := com.Md5(tcfg.TranslatorParsedConfig[`appid`] + values.Get(`q`) + values.Get(`salt`) + tcfg.TranslatorParsedConfig[`secret`]) //  appid+q+salt+密钥
	values.Add(`sign`, sign)
	url := `https://fanyi-api.baidu.com/api/trans/vip/translate?` + values.Encode()
	req := restyclient.Classic()
	resp, e := req.Get(url)
	if e != nil {
		return text, e
	}
	if !resp.IsSuccess() {
		return text, fmt.Errorf("[%v][%s] %s", resp.StatusCode(), resp.Status(), resp.Body())
	}
	r := &baiduResponse{}
	err := json.Unmarshal(resp.Body(), r)
	if err != nil {
		return text, err
	}
	//com.Dump(r)
	for _, v := range r.Results {
		return v.Dst, nil
	}
	return text, nil
}
