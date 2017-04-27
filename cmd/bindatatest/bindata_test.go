package i18n

import (
	"net/http"
	"testing"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/webx-top/echo/middleware/language"
)

/*
go get github.com/jteeuwen/go-bindata/...
go get github.com/elazarl/go-bindata-assetfs/...
go-bindata-assetfs -tags bindata data/...
*/
var langFSFunc = func(dir string) http.FileSystem {
	return &assetfs.AssetFS{
		Asset:     Asset,
		AssetDir:  AssetDir,
		AssetInfo: AssetInfo,
		Prefix:    dir,
	}
}

func TestOpen(t *testing.T) {
	c := &language.Config{
		Default:      `en`,
		Fallback:     `en`,
		AllList:      []string{`en`, `fr`},
		RulesPath:    []string{`data/rules`},
		MessagesPath: []string{`data/messages`},
	}
	language.New(c.SetFSFunc(langFSFunc))
}
