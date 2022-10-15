package main

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"go/format"
	"os"
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
	template_placeholder_regex = `\{(\w+)\}` // (e.g) {name} ot or {post1} or {user_name}

	date_format     = "2006-01-02"          // YYYY-MM-DD
	time_format     = "15:04:05"            // hh:mm:ss
	datetime_format = "2006-01-02 15:04:05" // YYYY-MM-DD hh:mm:ss
)

var (
	// allowed variables
	variable_types = []string{"int", "float", "string", "date", "time", "datetime"}
	// location of messages.toml file
	toml_file_path = "./messages.toml"

	//go:embed messages.toml
	toml_file_init_data []byte

	//go:embed template-messages.tmpl
	template_messages_data []byte

	//go:embed template-error-handling.tmpl
	template_error_handling_data []byte

	// exported file names
	exported_messages_path      = "gen.messages.go"
	exported_error_hanling_path = "gen.errors.go"
)

/* handle the command line arguments */
func init() {
	// count args
	if len(os.Args) > 2 {
		cute.SetTitleColor(cute.ColorBrightYellow)
		cute.SetMessageColor(cute.ColorBrightYellow)
		cute.Printlines("oops!", "too many arguments!", "try: genmessage help")
		os.Exit(1)
	}

	// no args: export
	if len(os.Args) == 1 {
		_, err := os.Stat(toml_file_path)
		if err != nil {
			cute.SetTitleColor(cute.ColorBrightYellow)
			cute.SetMessageColor(cute.ColorBrightYellow)
			cute.Printlines(
				"oops!",
				`"messages.toml" file not found!`,
				"_______________________________",
				"try: genmessage init",
				"help: genmessage help",
				"visit: https://github.com/zakaria-chahboun/genmessage",
			)
			os.Exit(1)
		}
		return // means go to main
	}

	/*
		version
		Note: always check "git tag --sort=-version:refname | head -n 1"
	*/
	version := "v1.0.5"

	// prepare args
	var arg = os.Args[1]
	var argmap = map[string]string{}
	argmap["init"] = "create 'messages.toml' file"
	argmap["clear"] = "Remove generated files"
	argmap["version"] = "Version: " + version
	argmap["help"] = "Get help"

	// init
	if arg == "init" {
		err := createTomlInitFile()
		cute.Check("Init", err)

		cute.SetTitleColor(cute.ColorBrightBlue)
		cute.SetMessageColor(cute.ColorBrightBlue)
		cute.Println("'messages.toml' was created")
		os.Exit(0)
	}
	// clear
	if arg == "clear" {
		var queue []any
		// check if file exists, then append it to queue
		_, err := os.Stat(exported_messages_path)
		if err == nil {
			queue = append(queue, exported_messages_path)
		}
		_, err = os.Stat(exported_error_hanling_path)
		if err == nil {
			queue = append(queue, exported_error_hanling_path)
		}
		// no file exists?
		if len(queue) == 0 {
			cute.SetTitleColor(cute.ColorBrightYellow)
			cute.SetMessageColor(cute.ColorBrightYellow)
			cute.Println("No exported files to remove.")
		} else {
			// remove files in queue
			for _, name := range queue {
				os.Remove(name.(string))
			}
			cute.SetTitleColor(cute.ColorBrightBlue)
			cute.SetMessageColor(cute.ColorBrightBlue)
			cute.Printlines("Remove", queue...)
		}
		os.Exit(0)
	}
	// version
	if arg == "version" {
		cute.SetTitleColor(cute.ColorBrightBlue)
		cute.SetMessageColor(cute.ColorBrightBlue)
		cute.Println("Version", version)
		os.Exit(0)
	}
	// help
	if arg == "help" {
		var list []any
		for k, v := range argmap {
			list = append(list, fmt.Sprintln(k, ":", v))
		}
		cute.SetTitleColor(cute.ColorBrightBlue)
		cute.SetMessageColor(cute.ColorBrightBlue)
		cute.Printlines("Help", list...)
		os.Exit(0)
	}
	// no arg match?
	cute.SetTitleColor(cute.ColorBrightYellow)
	cute.SetMessageColor(cute.ColorBrightYellow)
	cute.Println("oops!", "try to get help: genmessage help")
	os.Exit(1)

}

func main() {
	// functions to be used inside the template files
	my_funcs := template.FuncMap{
		"rename_code":          renameCode,
		"rename_type":          renameType,
		"title_case":           strings.Title,
		"join":                 strings.Join,
		"trim":                 strings.TrimSpace,
		"is_blank":             isBlank,
		"is_contains_date":     isContainsDate,
		"correct_placeholders": correctPlaceholders,
		"replace_placeholders": replacePlaceholdersWithSymbol,
	}

	// parse template files
	t_messages := template.Must(template.New("template-messages.tmpl").Funcs(my_funcs).Parse(string(template_messages_data)))
	t_error_handling := template.Must(template.New("template-error-handling.tmpl").Funcs(my_funcs).Parse(string(template_error_handling_data)))

	// load messages.toml file
	all_messages, err := LoadTranslationsFromTOML(toml_file_path)
	cute.Check("load translations from messages.toml file", err)

	// parse messages
	err = parse(all_messages)
	cute.Check("parse messages.toml file", err)

	// split all messages into: "messages" and "error-handling"
	var messages []Message
	var error_handling []Message
	for _, m := range all_messages {
		if isCodeError(m.Code) {
			error_handling = append(error_handling, m)
		} else {
			messages = append(messages, m)
		}
	}

	// send data to messages template
	var messages_blob bytes.Buffer
	err = t_messages.Execute(&messages_blob, struct {
		Messages             []Message
		UniqueLangs          []string
		UniqueVariablesTypes []string
		DateFormat           string
		TimeFormat           string
		DateTimeFormat       string
	}{
		Messages:             messages,
		UniqueLangs:          getUniqueLangs(all_messages), // all_messages not messages!
		UniqueVariablesTypes: getUniqueVariablesTypes(messages),
		DateFormat:           date_format,
		TimeFormat:           time_format,
		DateTimeFormat:       date_format,
	})
	cute.Check("parse (messages) template", err)

	// send data to error handling template if exists
	var error_handling_blob bytes.Buffer
	if len(error_handling) > 0 {
		err = t_error_handling.Execute(&error_handling_blob, struct {
			Messages             []Message
			UniqueStatuses       []string
			UniqueVariablesTypes []string
			DateFormat           string
			TimeFormat           string
			DateTimeFormat       string
		}{
			Messages:             error_handling,
			UniqueStatuses:       getUniqueStatuses(error_handling),
			UniqueVariablesTypes: getUniqueVariablesTypes(error_handling),
			DateFormat:           date_format,
			TimeFormat:           time_format,
			DateTimeFormat:       date_format,
		})
		cute.Check("parse (error handling messages) template", err)
	}

	// final process for: messages
	// go format
	page, err := format.Source(messages_blob.Bytes())
	cute.Check("go format", err)

	// save template with values as a go file (.go)
	err = saveToFile(exported_messages_path, page)
	cute.Check(fmt.Sprintf("export %s file", exported_messages_path), err)

	// final process for: error handling
	if error_handling_blob.String() != "" {
		// go format
		page, err := format.Source(error_handling_blob.Bytes())
		cute.Check("go format", err)

		// save template with values as a go file (.go)
		err = saveToFile(exported_error_hanling_path, page)
		cute.Check(fmt.Sprintf("export %s file", exported_error_hanling_path), err)
	}

	// done
	cute.SetTitleColor(cute.ColorBrightGreen)
	cute.Println("files generated successfully!")
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
func joinMapKeys(m map[string]string) (s string) {
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

/* check type is date or not */
func isDate(t string) (ok bool) {
	switch t {
	case "date", "time", "datetime":
		ok = true
	default:
		ok = false
	}
	return
}

/* check if variables type contains date */
func isContainsDate(types []string) bool {
	for _, v := range types {
		if isDate(v) {
			return true
		}
	}
	return false
}

/* is error type or not (begin with "err or error" e.g "err_no_user") */
func isCodeError(s string) bool {
	s = strings.ToLower(s)
	if strings.HasPrefix(s, "err") || strings.HasPrefix(s, "error") {
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

/* get unique variables types from messages.Variables */
func getUniqueVariablesTypes(ms []Message) (types []string) {
	for _, m := range ms {
		for _, v := range m.Variables {
			types = append(types, v)
		}
	}
	types = unique(types)
	return
}

/* get all placeholders of a template in order (not unique) */
func getPlaceholders(template string) (placeholders []string) {
	re1 := regexp.MustCompile(template_placeholder_regex) // Prepare our regex
	result_slice := re1.FindAllStringSubmatch(template, -1)
	for _, v := range result_slice {
		placeholders = append(placeholders, v[1])
	}
	return
}

/* replace all placeholders of a template with symbols: (e.g) "User {name} has {age} years old" to "User %s has %d years old" */
func replacePlaceholdersWithSymbol(template string, variables map[string]string) (new_template string) {
	placeholders := getPlaceholders(template)
	new_template = template
	for _, placeh := range placeholders {
		for variable, _type := range variables {
			if placeh == variable {
				switch _type {
				case "int", "int32", "int64":
					new_template = strings.ReplaceAll(new_template, "{"+placeh+"}", "%d")
				case "string":
					new_template = strings.ReplaceAll(new_template, "{"+placeh+"}", "%s")
				case "date", "time", "datatime":
					new_template = strings.ReplaceAll(new_template, "{"+placeh+"}", "%s")
				case "float", "float32", "float64":
					new_template = strings.ReplaceAll(new_template, "{"+placeh+"}", "%.2f")
				default:
					new_template = strings.ReplaceAll(template, "{"+placeh+"}", "%v")
				}
			}
		}
	}
	return
}

/* correct placeholders of a template: (e.g) "birthday" to "birthday.Format(datetime_format)" */
func correctPlaceholders(template string, variables map[string]string) (correct_placeholders []string) {
	placeholders := getPlaceholders(template)
	correct_placeholders = placeholders
	for i, placeh := range placeholders {
		for variable, _type := range variables {
			if placeh == variable {
				switch _type {
				case "date":
					correct_placeholders[i] = placeh + `.Format("` + date_format + `")`
				case "time":
					correct_placeholders[i] = placeh + `.Format("` + time_format + `")`
				case "datetime":
					correct_placeholders[i] = placeh + `.Format("` + datetime_format + `")`
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
	f, err := os.Create(toml_file_path)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = f.Write(toml_file_init_data)
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
		// code has bad variabe name
		ok, _ := regexp.MatchString("^[_a-zA-Z]\\w*$", m.Code)
		if !ok {
			return errors.New(fmt.Sprintf("bad code name! '%v' rules: no spaces, just letters or '_' in the beginning", m.Code))
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
			return errors.New(fmt.Sprintf("you have duplicate 'Code=%v'", m.Code))
		}
		// in case a message doesn't have any langs
		if len(m.Templates) == 0 {
			return errors.New(fmt.Sprintf("no translations in 'Code=%v'", m.Code))
		}
		// in case a lang is missing
		counter = 0
		for k := range m.Templates {
			for _, l := range langs {
				if k == l {
					counter++
				}
			}
		}
		if counter != len(langs) {
			return errors.New(fmt.Sprintf("in 'Code=%v' you miss to implement some languages: %v", m.Code, langs))
		}
		// Status exist, But has bad variabe name
		if m.Status != "" {
			ok, _ = regexp.MatchString("^[_a-zA-Z]\\w*$", m.Status)
			if !ok {
				return errors.New(fmt.Sprintf("bad status name! '%v' rules: no spaces, just letters or '_' in the beginning", m.Status))
			}
		}
	}

	// parse the number/type variables and compare it with placeholders in message templates

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
		// check variables/placeholders count
		for lang, template := range m.Templates {
			placeholders := unique(getPlaceholders(template))
			// in case the variable doesn't exist in template
			for variable := range m.Variables {
				var counter = 0
				for _, placeh := range placeholders {
					if variable == placeh {
						counter++
						break
					}
				}
				if counter == 0 {
					return errors.New(fmt.Sprintf("in 'Code=%v' you have an unused variable '%v' in '%v' template", m.Code, variable, lang))
				}
			}
			// in case the placeholder doesn't exist in variables
			for _, placeh := range placeholders {
				var counter = 0
				for variable := range m.Variables {
					if placeh == variable {
						counter++
						break
					}
				}
				if counter == 0 {
					return errors.New(fmt.Sprintf("in 'Code=%v' you miss to add the placeholder '{%v}' in variables list", m.Code, placeh))
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
