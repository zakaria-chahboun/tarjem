package main

import (
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/zakaria-chahboun/cute"
	"gopkg.in/yaml.v3"
)

type Message struct {
	Variables    map[string]string `yaml:"variables,omitempty"` // {"name":"string", "age":"int"}
	Translations map[string]string `yaml:"translations"`        // {"english":"hello", "arabic":"hala"}
}

type Messages map[string]Message

const (
	TEMPLATE_PLACEHOLDER_PATTERN = `\{(\w+)\}`           // e.g: {name} {phone} ..
	DATE_FORMAT                  = "2006-01-02"          // YYYY-MM-DD
	TIME_FORMAT                  = "15:04:05"            // hh:mm:ss
	DATETIME_FORMAT              = "2006-01-02 15:04:05" // YYYY-MM-DD hh:mm:ss
)

var (
	ALLOWED_VARIABLE_TYPES         = []string{"int", "float", "string", "date", "time", "datetime"}
	DEFAULT_TRANSLATIONS_FILE_PATH = "./translations.yaml"

	//go:embed translations.yaml
	DEFAULT_TRANSLATIONS_FILE_DATA []byte

	//go:embed templates/translations.go.tmpl
	TRANSLATIONS_TEMPLATE_DATA []byte

	EXPORTED_PACKAGE_NAME      = "tarjem" // default in the exported files in go
	EXPORTED_TRANSLATIONS_FILE = "translations.go"

	/*
		Note:
		check:
			git tag --sort=-version:refname | head -n 1
		or add it with:
			go build -ldflags "-X main.version=`git tag --sort=-version:refname | head -n 1`
	*/
	version = "v1.1.0"
)

func init() {
	initCmd.Flags().Bool("force", false, "Force initialization, overwriting existing translations.yaml files")

	exportCmd.Flags().String("lang", "", "Specify the language for exporting Go files")
	exportCmd.Flags().String("package", "", "Optional package name for exporting Go files into a directory")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(clearCmd)
	rootCmd.AddCommand(helpCmd)
	rootCmd.AddCommand(versionCmd)

}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Rename name: (e.g: pay_bills to PayBills)
func snakeCaseToCamelCase(text string) (s string) {
	s = strings.ReplaceAll(text, "_", " ")
	s = strings.Title(strings.ToLower(s))
	s = strings.ReplaceAll(s, " ", "")
	return
}

// Convert type: (e.g: date to time.Time)
func convertToGoType(typeName string) (s string) {
	switch typeName {
	case "date", "time", "datetime":
		return "time.Time"
	case "float":
		return "float64"
	default:
		return typeName
	}
}

// Unique slice: (e.g: ["a", "b" , "a"] to ["a", "b"])
func unique(dataList []string) (list []string) {
	allKeys := make(map[string]bool)
	for _, item := range dataList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return
}

// Check string is empty or space
func isBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

// Check type is date
func isDate(t string) (ok bool) {
	switch t {
	case "date", "time", "datetime":
		ok = true
	default:
		ok = false
	}
	return
}

// Check if variables type contains date
func containsDateType(types []string) bool {
	for _, v := range types {
		if isDate(v) {
			return true
		}
	}
	return false
}

// Get unique langs from messages.Translations
func getUniqueLangs(messages Messages) (langs []string) {
	for _, v := range messages {
		for k := range v.Translations {
			langs = append(langs, k)
		}
	}
	langs = unique(langs)
	return
}

// Get unique variables types from messages.Variables
func getUniqueVariableTypes(messages Messages) (types []string) {
	for _, message := range messages {
		for _, v := range message.Variables {
			types = append(types, v)
		}
	}
	types = unique(types)
	return
}

// Get all placeholders in translation
func getPlaceholders(translation string) (placeholders []string, err error) {
	// Compile the regular expression
	re, compileErr := regexp.Compile(TEMPLATE_PLACEHOLDER_PATTERN)
	if compileErr != nil {
		return nil, fmt.Errorf("failed to compile regex: %w", compileErr)
	}

	// Find all matches for the pattern in the translation string
	matches := re.FindAllStringSubmatch(translation, -1)
	if matches == nil {
		return placeholders, nil // No placeholders found
	}

	// Extract placeholders from matches
	for _, match := range matches {
		if len(match) > 1 {
			placeholders = append(placeholders, match[1])
		}
	}

	return placeholders, nil
}

// Replace all placeholders of a translation with string format: (e.g: "User {name} has {age} years old" to "User %s has %d years old"
func replacePlaceholdersWithFormat(translation string, variables map[string]string) (string, error) {
	placeholders, err := getPlaceholders(translation)
	if err != nil {
		return "", fmt.Errorf("error getting placeholders: %w", err)
	}

	for _, placeholder := range placeholders {
		for variable, vtype := range variables {
			if placeholder == variable {
				switch vtype {
				case "int", "int32", "int64":
					translation = strings.ReplaceAll(translation, "{"+placeholder+"}", "%d")
				case "string":
					translation = strings.ReplaceAll(translation, "{"+placeholder+"}", "%s")
				case "date", "time", "datatime":
					translation = strings.ReplaceAll(translation, "{"+placeholder+"}", "%s")
				case "float", "float32", "float64":
					translation = strings.ReplaceAll(translation, "{"+placeholder+"}", "%.2f")
				default:
					translation = strings.ReplaceAll(translation, "{"+placeholder+"}", "%v")
				}
			}
		}
	}
	return translation, nil
}

// Correct placeholders translation: (e.g: "birthday" to "birthday.Format(DATE_FORMAT)")
func correctPlaceholders(template string, variables map[string]string) (correct_placeholders []string) {
	placeholders, _ := getPlaceholders(template)
	correct_placeholders = placeholders
	for i, placeh := range placeholders {
		for variable, _type := range variables {
			if placeh == variable {
				switch _type {
				case "date":
					correct_placeholders[i] = placeh + `.Format(DATE_FORMAT)`
				case "time":
					correct_placeholders[i] = placeh + `.Format(TIME_FORMAT)`
				case "datetime":
					correct_placeholders[i] = placeh + `.Format(DATETIME_FORMAT)`
				}
			}
		}
	}
	return
}

// Save the final go file
func saveToFile(filePath string, data []byte) (err error) {
	f, err := os.Create(filePath)
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

// Init the translations.yaml file
func createInitTranslationFile(filePath string, data []byte) (err error) {
	f, err := os.Create(filePath)
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

// Parse messages
func parseMessages(messages Messages) error {

	uniqueLangs := getUniqueLangs(messages)

	// Compile the regular expression once
	var validNamePattern, err = regexp.Compile(`^[_a-zA-Z]\w*$`)
	if err != nil {
		return fmt.Errorf("failed to compile regex of validNamePattern: %w", err)
	}

	// parsing
	for name, message := range messages {
		// name must respect the regex
		ok := validNamePattern.MatchString(name)
		if !ok {
			return fmt.Errorf("invalid name '%v': rules - no spaces, only letters or '_' allowed at the beginning", name)
		}

		// Check for duplicated names
		nameCounts := make(map[string]int)
		for searchedName := range messages {
			nameCounts[searchedName]++
		}
		if count, exists := nameCounts[name]; exists && count > 1 {
			return fmt.Errorf("duplicate detected: '%v' appears %d times", name, count)
		}

		// Check if a message has translations
		if len(message.Translations) == 0 {
			return fmt.Errorf("no translations found for '%v'", name)
		}

		// Check for missing languages
		existingLangs := make(map[string]bool)
		for lang := range message.Translations {
			existingLangs[lang] = true
		}
		missingLangs := []string{}
		for _, lang := range uniqueLangs {
			if _, exists := existingLangs[lang]; !exists {
				missingLangs = append(missingLangs, lang)
			}
		}
		if len(missingLangs) > 0 {
			return fmt.Errorf("'%v' is missing translations for the following languages: '%v'", name, missingLangs)
		}

		// Check if all variable types in the message are valid
		for variable, vtype := range message.Variables {
			if !isValidType(vtype) {
				return fmt.Errorf("invalid type '%v' for variable '%v' in '%v'. Allowed types are: %v", vtype, variable, name, ALLOWED_VARIABLE_TYPES)

			}
		}
		// check variables/placeholders count
		for lang, translation := range message.Translations {
			placeholders, err := getPlaceholders(translation)
			if err != nil {
				return fmt.Errorf("error processing placeholders in translation '%v' for language '%v' : %w", name, lang, err)
			}
			uniquePlaceholders := unique(placeholders)

			// Check if all variables are used in the template
			for variable := range message.Variables {
				if !contains(uniquePlaceholders, variable) {
					return fmt.Errorf("variable '%v' not used in the '%v' translation for language '%v'", variable, name, lang)
				}
			}

			// Check if all placeholders are defined in variables
			for _, placeholder := range uniquePlaceholders {
				if _, exists := message.Variables[placeholder]; !exists {
					return fmt.Errorf("placeholder '{%v}' is missing from the variables list in the '%v' translation for language '%v'", placeholder, name, lang)
				}
			}
		}
	}

	return nil
}

// Load translations from yaml file and cast it to Message map
func loadTranslationsFromFile(filePath string) (messages Messages, err error) {
	// read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, &messages)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

// Function to check if a type is in the ALLOWED_VARIABLE_TYPES slice
func isValidType(vtype string) bool {
	for _, t := range ALLOWED_VARIABLE_TYPES {
		if vtype == t {
			return true
		}
	}
	return false
}

// Contains checks if a slice contains a specific string.
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// Info message in CLI when issue happened
func alert() {
	cute.SetTitleColor(cute.BrightYellow)
	cute.SetMessageColor(cute.BrightYellow)
	list := cute.NewList(cute.BrightYellow, "OOPS 😢!")
	list.Add(cute.BrightYellow, `"translations.yaml" file not found!`)
	list.Add(cute.BrightBlue, "try: tarjem init")
	list.Add(cute.BrightBlue, "help: tarjem help")
	list.Add(cute.BrightBlue, "visit: https://github.com/zakaria-chahboun/tarjem")
	list.Print()
}
