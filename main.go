package main

import (
    "encoding/json"
	"flag"
	"fmt"
	"os"
	"vs_export/sln"
	"io/ioutil"
    "path/filepath"
)

func main() {
	path := flag.String("s", "", "sln file path (可选，若不指定则在build/目录下自动查找)")
	configuration := flag.String("c", "Debug|x64", "Configuration, [configuration|platform], default Debug|x64")
	flag.Parse()

	slnPath := *path
	if slnPath == "" {
		// 在build目录下查找sln文件
		buildDir := "build"
		files, err := os.ReadDir(buildDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "查找sln文件时出错: %v\n", err)
			os.Exit(1)
		}

		for _, file := range files {
			if !file.IsDir() && filepath.Ext(file.Name()) == ".sln" {
				slnPath = filepath.Join(buildDir, file.Name())
				break
			}
		}

		if slnPath == "" {
			fmt.Fprintln(os.Stderr, "在build/目录下未找到.sln文件")
			os.Exit(1)
		}
	}

	solution, err := sln.NewSln(slnPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	cmdList, err := solution.CompileCommandsJson(*configuration)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	js, err := json.Marshal(cmdList)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%s\n", js[:])
	ioutil.WriteFile("compile_commands.json", js[:], 0644)
}

func usage() {
	var echo = `Usage: %s [-s <path>] -c <configuration>

Where:
            -s   path                        sln filename (可选，若不指定则在build/目录下自动查找)
            -c   configuration               project configuration,eg Debug|x64.
                                             default Debug|x64
	`
	echo = fmt.Sprintf(echo, filepath.Base(os.Args[0]))
    fmt.Println(echo)
}
