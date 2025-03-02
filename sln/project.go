package sln

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Project struct {
	ProjectDir          string
	ProjectPath         string
	XMlName             xml.Name              `xml:"Project"`
	ItemGroup           []ItemGroup           `xml:"ItemGroup"`
	ItemDefinitionGroup []ItemDefinitionGroup `xml:"ItemDefinitionGroup"`
}
type ItemGroup struct {
	XMLName                  xml.Name               `xml:"ItemGroup"`
	Label                    string                 `xml:"Label,attr"`
	ProjectConfigurationList []ProjectConfiguration `xml:"ProjectConfiguration"`
	ClCompileSrc             []ClCompileSrc         `xml:"ClCompile"`
}

type ProjectConfiguration struct {
	XMLName       xml.Name `xml:"ProjectConfiguration"`
	Include       string   `xml:"Include,attr"`
	Configuration string   `xml:"Configuration"`
	Platform      string   `xml:"Platform"`
}

type ItemDefinitionGroup struct {
	XMLName   xml.Name  `xml:"ItemDefinitionGroup"`
	Condition string    `xml:"Condition,attr"`
	ClCompile ClCompile `xml:"ClCompile"`
}

type ClCompile struct {
	XMLName                      xml.Name `xml:"ClCompile"`
	AdditionalIncludeDirectories string   `xml:"AdditionalIncludeDirectories"`
	PreprocessorDefinitions      string   `xml:"PreprocessorDefinitions"`
	LanguageStandard            string   `xml:"LanguageStandard"`
	ConformanceMode             string   `xml:"ConformanceMode"`
}

type ClCompileSrc struct {
	XMLName xml.Name `xml:"ClCompile"`
	Include string   `xml:"Include,attr"`
}

type CompileCommand struct {
	Dir  string `json:"directory"`
	Cmd  string `json:"command"`
	File string `json:"file"`
}

var badInclude = []string{
	";%(AdditionalIncludeDirectories)",
	"%(AdditionalIncludeDirectories);",
}
var badDef = []string{
	";%(PreprocessorDefinitions)",
	"%(PreprocessorDefinitions);",
}

func NewProject(path string) (Project, error) {
	var pro Project
	var err error

	pro.ProjectPath, err = filepath.Abs(path)
	if err != nil {
		return pro, err
	}
	pro.ProjectDir = filepath.Dir(pro.ProjectPath)

	f, err := os.Open(path)
	if err != nil {
		return Project{}, err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	err = xml.Unmarshal([]byte(data), &pro)
	if err != nil {
		return pro, err
	}
	return pro, nil
}

//return include, definition, cppStandard, error
func (pro *Project) FindConfig(conf string) (string, string, string, error) {
	var cfgList []ProjectConfiguration
	for _, v := range pro.ItemGroup {
		if len(v.ProjectConfigurationList) > 0 {
			cfgList = v.ProjectConfigurationList
			break
		}
	}
	fmt.Fprintln(os.Stderr, cfgList)
	if len(cfgList) == 0 {
		return "", "", "", errors.New(pro.ProjectPath + ":not found " + conf)
	}
	found := false
	for _, v := range cfgList {
		if v.Include == conf {
			found = true
			break
		}
	}
	if !found {
		return "", "", "", errors.New(pro.ProjectPath + ":not found " + conf)
	}
	for _, v := range pro.ItemDefinitionGroup {
		if strings.Contains(v.Condition, conf) {
			cl := v.ClCompile

			vlist := strings.Split(conf, "|")
			configuration := vlist[0]
			platform := vlist[1]

			willReplaceEnv := map[string]string{
				"$(ProjectDir)":        pro.ProjectDir,
				"$(Configuration)":     configuration,
				"$(ConfigurationName)": configuration,
				"$(Platform)":          platform,
			}
			for _, v := range os.Environ() {
				kv := strings.Split(v, "=")
				willReplaceEnv[fmt.Sprintf("$(%s)", kv[0])] = kv[1]
			}

			include := cl.AdditionalIncludeDirectories
			def := cl.PreprocessorDefinitions
			for k, v := range willReplaceEnv {
				if strings.Contains(include, k) {
					include = strings.Replace(include, k, v, -1)
				}
			}

			re := regexp.MustCompile(`\$\(.+\)`)
			badEnv := re.FindAllString(include, -1)
			if len(badEnv) > 0 {
				//fmt.Fprintf(os.Stderr, "%s:bad env[%v]\n", pro.ProjectPath, badEnv[:])
				//for _, v := range badEnv {
				//	include = strings.Replace(include, v, "", -1)
				//}
			}

			// 处理 C++ 标准
			cppStdFlag := getCppStandardFlag(cl.LanguageStandard, cl.ConformanceMode)

			return include, def, cppStdFlag, nil
		}
	}
	return "", "", "", errors.New("not found " + conf)
}

func (pro *Project) FindSourceFiles() []string {
	var fileList []string
	for _, v := range pro.ItemGroup {
		for _, inc := range v.ClCompileSrc {
			fileList = append(fileList, inc.Include)
		}
	}
	return fileList
}

func RemoveBadInclude(include string) string {
	for _, bad := range badInclude {
		include = strings.Replace(include, bad, ";.", -1)
	}
	return include
}

func RemoveBadDefinition(def string) string {
	for _, bad := range badDef {
		def = strings.Replace(def, bad, "", -1)
	}
	return def
}

// 添加一个辅助函数来处理 C++ 标准
func getCppStandardFlag(langStd string, conformance string) string {
	// 只处理语言标准，移除 conformance 相关的默认设置
	// 目前使用 clang-cl.exe ，所以参数格式为 /std:c++20 而非 -std=c++20
	switch langStd {
	case "stdcpplatest":
		return "/std:c++20"
	case "stdcpp20":
		return "/std:c++20"
	case "stdcpp17":
		return "/std:c++17"
	case "stdcpp14":
		return "/std:c++14"
	case "stdcpp11":
		return "/std:c++11"
	}
	return "/std:c++20"
}
