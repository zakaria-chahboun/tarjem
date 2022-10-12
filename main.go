package main

import (
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/zakaria-chahboun/cute"
)

type Message struct {
	Code      string
	Status    string
	Variables map[string]string // {"name":"string", "age":"int"}
	Templates map[string]string // {"en":"hello", "ar":"marhaba"}
}

const (
	template_parameter_regex = `\{(\w+)\}` // (e.g) {name} ot or {post1} or {user_name}

	date_format     = "2006-01-02"          // YYYY-MM-DD
	time_format     = "15:04:05"            // hh:mm:ss
	datetime_format = "2006-01-02 15:04:05" // YYYY-MM-DD hh:mm:ss
)

var variable_types = []string{"int", "float", "string", "date", "time", "datetime"}

/* the location of messages.toml file */
var toml_path = "./messages.toml"

//go:embed messages.toml
var toml_init_data []byte

//go:embed template.tmpl
var template_data []byte

/* command line arguments */
var (
	Init       *bool
	ExportName *string
)

/* handle the command line arguments */
func init() {
	ExportName = flag.String("export", "./messages.go", "Name of the exported file without specifying the (.go) extension.")
	Init = flag.Bool("init", false, "Create the messages.toml file.")
	help := flag.Bool("h", false, "help")

	flag.Parse()

	/* flag = -init */
	if *Init {
		err := createTomlInitFile()
		cute.Check("Create Toml Init File", err)
		cute.Println("Created", toml_path)
		os.Exit(0)
	}
	/* flag = -h */
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}
	/* flag = -export */
	ext := filepath.Ext(*ExportName)
	if ext != ".go" {
		*ExportName += ".go"
	}

	/* check if messages.toml exist */
	_, err := os.Stat(toml_path)
	if err != nil {
		cute.SetMessageColor(cute.ColorBrightBlue)
		cute.Printlines(
			"oops!",
			`"messages.toml" file not found!`,
			"try: genmessage -init",
			"help: genmessage -h",
		)
		os.Exit(1)
	}
}

func main() {
	// functions to be used inside the template file
	my_funcs := template.FuncMap{
		"rename_code":    renameCode,
		"rename_type":    renameType,
		"title_case":     strings.Title,
		"join":           strings.Join,
		"trim":           strings.TrimSpace,
		"is_blank":       isBlank,
		"correct_params": correctParameters,
		"replace_params": replaceTemplateParamWithSymbol,
	}

	// parse template file
	t := template.Must(template.New("template.tmpl").Funcs(my_funcs).Parse(string(template_data)))

	// load messages file
	messages, err := LoadTranslationsFromTOML(toml_path)
	cute.Check("load translations from messages.toml file", err)

	// parse messages
	err = parse(messages)
	cute.Check("parse messages.toml file", err)

	// send data to template
	var blob bytes.Buffer
	err = t.Execute(&blob, struct {
		Messages       []Message
		Statuses       []string
		Langs          []string
		DateFormat     string
		TimeFormat     string
		DateTimeFormat string
	}{
		Messages:       messages,
		Statuses:       getUniqueStatuses(messages), // unique
		Langs:          getUniqueLangs(messages),    // unique
		DateFormat:     date_format,
		TimeFormat:     time_format,
		DateTimeFormat: date_format,
	})
	cute.Check("handle template", err)

	// go format
	page, err := format.Source(blob.Bytes())
	cute.Check("go format", err)

	// save template with values as a go file (.go)
	err = saveToFile(*ExportName, page)
	cute.Check(fmt.Sprintf("export %s file", *ExportName), err)

	// done
	cute.Println("done!")
}

/* rename code: (e.g) "user_is_out" to "UserIsOut" */
func renameCode(text string) (s string) {
	s = strings.ReplaceAll(text, "_", " ")
	s = strings.Title(strings.ToLower(s))
	s = strings.ReplaceAll(s, " ", "")
	return
}

/* rename type: (e.g) "date" to "time.Time" */
func renameType(text string) (s string) {
	switch text {
	case "date", "time", "datetime":
		s = "time.Time"
	default:
		s = text
	}
	return
}

/* join map: (e.g) {a:1, b:2} to "a, b" */
func joinkeys(m map[string]string) (s string) {
	for k := range m {
		s += k + " ,"
	}
	return
}

/* unique slice: (e.g) ["a", "b" , "a"] to ["a", "b"] */
func unique(duplist []string) (list []string) {
	allKeys := make(map[string]bool)
	for _, item := range duplist {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return
}

/* check empty or space */
func isBlank(s string) bool {
	if strings.TrimSpace(s) == "" {
		return true
	}
	return false
}

/* get unique statuses from messages */
func getUniqueStatuses(ms []Message) (statuses []string) {
	for _, v := range ms {
		if strings.TrimSpace(v.Status) != "" {
			v.Status = strings.TrimSpace(v.Status)
			statuses = append(statuses, v.Status)
		}
	}
	statuses = unique(statuses)
	return
}

/* get unique langs from messages.templates */
func getUniqueLangs(ms []Message) (langs []string) {
	for _, v := range ms {
		for k := range v.Templates {
			langs = append(langs, k)
		}
	}
	langs = unique(langs)
	return
}

/* get all parameters of a template in order (not unique) */
func getTemplateParamaters(template string) (params []string) {
	re1 := regexp.MustCompile(template_parameter_regex) // Prepare our regex
	result_slice := re1.FindAllStringSubmatch(template, -1)
	for _, v := range result_slice {
		params = append(params, v[1])
	}
	return
}

/* replace all parameters of a template with symbols: (e.g) "User {name} has {age} years old" to "User %s has %d years old" */
func replaceTemplateParamWithSymbol(template string, variables map[string]string) (new_template string) {
	params := getTemplateParamaters(template)
	new_template = template
	for _, param := range params {
		for variable, _type := range variables {
			if param == variable {
				switch _type {
				case "int", "int32", "int64":
					new_template = strings.ReplaceAll(new_template, "{"+param+"}", "%d")
				case "string":
					new_template = strings.ReplaceAll(new_template, "{"+param+"}", "%s")
				case "date", "time", "datatime":
					new_template = strings.ReplaceAll(new_template, "{"+param+"}", "%s")
				case "float", "float32", "float64":
					new_template = strings.ReplaceAll(new_template, "{"+param+"}", "%.2f")
				default:
					new_template = strings.ReplaceAll(template, "{"+param+"}", "%v")
				}
			}
		}
	}
	return
}

/* correct parameters of a template: (e.g) "birthday" to "birthday.Format(datetime_format)" */
func correctParameters(template string, variables map[string]string) (correct_params []string) {
	params := getTemplateParamaters(template)
	correct_params = params
	for i, param := range params {
		for variable, _type := range variables {
			if param == variable {
				switch _type {
				case "date":
					correct_params[i] = param + `.Format("` + date_format + `")`
				case "time":
					correct_params[i] = param + `.Format("` + time_format + `")`
				case "datetime":
					correct_params[i] = param + `.Format("` + datetime_format + `")`
				}
			}
		}
	}
	return
}

/* export the final go file */
func saveToFile(path string, data []byte) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		return
	}
	err = f.Sync()
	if err != nil {
		return
	}
	return
}

/* init the messages.toml file */
func createTomlInitFile() (err error) {
	f, err := os.Create(toml_path)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = f.Write(toml_init_data)
	if err != nil {
		return
	}
	err = f.Sync()
	if err != nil {
		return
	}
	return
}

/* parse messages */
func parse(messages []Message) error {
	langs := getUniqueLangs(messages)

	// parsing
	for i, m := range messages {
		// in case code is empty
		if m.Code == "" {
			return errors.New(fmt.Sprintf("you forget to add the Code in messages[%v]", i))
		}
		// in case Code is duplicated
		var counter = 0
		for _, m2 := range messages {
			c1 := strings.TrimSpace(m.Code)
			c2 := strings.TrimSpace(m2.Code)
			if c1 == c2 {
				counter++
			}
		}
		if counter >= 2 {
			return errors.New(fmt.Sprintf("you have duplicating 'Code=%v'", m.Code))
		}
		// in case a message doesn't have any langs
		if len(m.Templates) == 0 {
			return errors.New(fmt.Sprintf("no translations in 'Code=%v'", m.Code))
		}
		// calculate the count of langs in a message
		counter = 0
		for k := range m.Templates {
			for _, l := range langs {
				if k == l {
					counter++
				}
			}
		}
		// in case a lang is missing
		if counter != len(langs) {
			return errors.New(fmt.Sprintf("in 'Code=%v' you miss to implement some languages: %v", m.Code, langs))
		}
	}

	// parse the number/type variables and compare it with parameters in message templates

	// parsing
	for _, m := range messages {
		// check variables types
		for name, _type := range m.Variables {
			var counter = 0
			for _, t := range variable_types {
				if _type == t {
					counter++
					break
				}
			}
			if counter == 0 {
				return errors.New(fmt.Sprintf("in 'Code=%v' you put '%v=%v', only these types are allowed: %v", m.Code, name, _type, variable_types))
			}
		}
		// check variables/parameters count
		for lang, template := range m.Templates {
			params := unique(getTemplateParamaters(template))
			// in case the variable doesn't exist in template
			for variable := range m.Variables {
				var counter = 0
				for _, param := range params {
					if variable == param {
						counter++
						break
					}
				}
				if counter == 0 {
					return errors.New(fmt.Sprintf("in 'Code=%v' you miss to add the variable '%v' to template '%v'", m.Code, variable, lang))
				}
			}
			// in case the parameter doesn't exist in variables
			for _, param := range params {
				var counter = 0
				for variable := range m.Variables {
					if param == variable {
						counter++
						break
					}
				}
				if counter == 0 {
					return errors.New(fmt.Sprintf("in 'Code=%v' you miss the to add parameter '{%v}' to variables", m.Code, param))
				}
			}
		}
	}
	return nil
}

/* load translations from toml file and cast it to '[]Message' struct */
func LoadTranslationsFromTOML(path string) (Messages []Message, err error) {
	// read file
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	type Root struct {
		Messages []Message
	}
	root := &Root{}

	// decode it
	err = toml.Unmarshal(data, root)
	Messages = root.Messages
	return
}
