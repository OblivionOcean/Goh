package main

import (
	"flag"
	"os"
	"path"
)

func main() {
	dest := flag.String("dest", "", "generated golang files dir, it will be the same with source if not set")
	ext := flag.String("ext", ".html", "source file extensions, comma splitted if many")
	packname := flag.String("pkg", "template", "the generated template package name, default is template")
	src := flag.String("src", "./", "the html template file or directory")
	flag.Parse()
	files, err := os.ReadDir(*src)
	if err != nil {
		panic(err.Error())
	}
	for i:=0;i<len(files);i++ {
		if files[i].IsDir() {
			continue
		}
		if path.Ext(files[i].Name()) == *ext {
			g := &Generator{
				Dest: *dest,
				PackageName: *packname,
			}
			g.New(files[i].Name())
		}
	}
}