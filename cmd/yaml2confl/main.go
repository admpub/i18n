package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/admpub/confl"
	"github.com/admpub/i18n"
	"gopkg.in/yaml.v3"
)

func init() {
	log.SetFlags(0)

	flag.Usage = usage
	flag.Parse()
}

func usage() {
	log.Printf("Usage: %s name.yaml [ file2 ... ]\n",
		filepath.Base(os.Args[0]))
	flag.PrintDefaults()

	os.Exit(1)
}

func main() {
	if flag.NArg() < 1 {
		flag.Usage()
	}
	for _, f := range flag.Args() {
		fi, err := os.Stat(f)
		if err != nil {
			log.Fatalln(err)
		}
		if fi.IsDir() {
			f, err = filepath.Abs(f)
			if err != nil {
				log.Fatalln(err)
			}
			saveAs := filepath.Join(f, `confl`)
			os.MkdirAll(saveAs, os.ModePerm)
			err = filepath.Walk(f, func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					if info.Name() == `confl` {
						return filepath.SkipDir
					}
					return nil
				}
				save := strings.TrimPrefix(path, f)
				save = filepath.Join(saveAs, save)
				os.MkdirAll(filepath.Dir(save), os.ModePerm)
				return convFile(path, save)
			})
		} else {
			saveAs := filepath.Join(filepath.Dir(f), `confl`)
			os.MkdirAll(saveAs, os.ModePerm)
			err = convFile(f, filepath.Join(saveAs, filepath.Base(f)))
		}
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func convFile(f string, saveAs string) error {
	var tmp i18n.TranslatorRules
	b, err := os.ReadFile(f)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}
	b, err = confl.Marshal(tmp)
	if err != nil {
		return err
	}
	log.Println(`conversion file:`, f)
	return os.WriteFile(saveAs, b, os.ModePerm)
}
