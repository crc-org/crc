package main

import (
	"log"
	"os"
	"strings"
	"text/template"
)

var crcVersion = "unset" // version set at compile time

func main() {
	if len(os.Args) != 1 {
		log.Fatal("Split takes only one argument (the file to split)")
	}

	if err := generateWxsFromTemplate(); err != nil {
		log.Fatalf("Wxs generation failed: %s", err)
	}
}

func generateWxsFromTemplate() error {
	tmpl := template.New("product.wxs.template")
	tmpl.Funcs(template.FuncMap{
		"strjoin": strings.Join,
		"inc":     func(val int) int { return val + 1 },
	})
	tmpl, err := tmpl.ParseFiles("packaging/windows/product.wxs.template")
	if err != nil {
		return err
	}
	type templateData struct {
		Version string
	}
	tmplData := templateData{
		Version: crcVersion,
	}

	f, err := os.OpenFile("packaging/windows/product.wxs", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpl.Execute(f, tmplData)
}
