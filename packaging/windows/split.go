package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const chunkSize = 1024 * 1024 * 1024 // 1GiB chunk size

var crcVersion = "unset" // version set at compile time

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Split takes only one argument (the file to split)")
	}

	parts, err := split(os.Args[1])
	if err != nil {
		for _, part := range parts {
			os.Remove(part)
		}
		log.Fatal(err.Error())
	}

	if err := generateWxsFromTemplate(filepath.Base(os.Args[1]), parts); err != nil {
		log.Fatalf("Wxs generation failed: %s", err)
	}
}

func generateWxsFromTemplate(bundleName string, parts []string) error {
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
		BundleName string
		Version    string
		Parts      []string
	}
	tmplData := templateData{
		BundleName: bundleName,
		Version:    crcVersion,
		Parts:      []string{},
	}
	for _, part := range parts {
		tmplData.Parts = append(tmplData.Parts, filepath.Base(part))
	}

	f, err := os.OpenFile("packaging/windows/product.wxs", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpl.Execute(f, tmplData)
}

func split(filePath string) ([]string, error) {
	splitFiles := []string{}
	bundle, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer bundle.Close()
	bundleName := filepath.Base(filePath)
	for i := 1; ; i++ {
		partFileName := fmt.Sprintf("%s.%d", bundleName, i)
		partFile, err := os.Create(filepath.Join(filepath.Dir(filePath), partFileName))
		if err != nil {
			return splitFiles, err
		}
		splitFiles = append(splitFiles, partFileName)
		defer partFile.Close()
		n, err := io.CopyN(partFile, bundle, chunkSize)
		fmt.Printf("Copied %d bytes from %s to %s\n", n, bundleName, partFileName)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return splitFiles, nil
			}
			return splitFiles, err
		}
		if err = partFile.Close(); err != nil {
			return splitFiles, err
		}
	}
}
