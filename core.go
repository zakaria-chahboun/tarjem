package main

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
	"text/template"

	"github.com/zakaria-chahboun/cute"
)

var SupportedProgrammingLanguages = []string{"go"}

func ExportForProgrammingLanguage(pl string) {

	if !contains(SupportedProgrammingLanguages, strings.ToLower(pl)) {
		cute.Println("Unsupported language", "Only:", strings.Join(SupportedProgrammingLanguages, ","))
		return
	}

	// functions to be used inside the template files
	templateFuncs := template.FuncMap{
		"snakeCaseToCamelCase":          snakeCaseToCamelCase,
		"convertToGoType":               convertToGoType,
		"titleCase":                     strings.Title,
		"join":                          strings.Join,
		"trim":                          strings.TrimSpace,
		"isBlank":                       isBlank,
		"containsDateType":              containsDateType,
		"correctPlaceholders":           correctPlaceholders,
		"replacePlaceholdersWithFormat": replacePlaceholdersWithFormat,
	}

	// parse template files
	translationsTemplate := template.Must(template.New("translations.go.tmpl").Funcs(templateFuncs).Parse(string(TRANSLATIONS_TEMPLATE_DATA)))

	// load translations.yaml file
	translationsData, err := loadTranslationsFromFile(DEFAULT_TRANSLATIONS_FILE_PATH)
	cute.Check("load translations.yaml file", err)

	// parse messages
	err = parseMessages(translationsData)
	cute.Check("parsing translations.yaml file", err)

	// Bind data into translations template
	var compiledOutput bytes.Buffer
	err = translationsTemplate.Execute(&compiledOutput, struct {
		PackageName         string
		Messages            Messages
		UniqueLangs         []string
		UniqueVariableTypes []string
		DateFormat          string
		TimeFormat          string
		DateTimeFormat      string
	}{
		PackageName:         EXPORTED_PACKAGE_NAME,
		Messages:            translationsData,
		UniqueLangs:         getUniqueLangs(translationsData), // all_messages not messages!
		UniqueVariableTypes: getUniqueVariableTypes(translationsData),
		DateFormat:          DATE_FORMAT,
		TimeFormat:          TIME_FORMAT,
		DateTimeFormat:      DATETIME_FORMAT,
	})
	cute.Check("bind data in translations.go.tmpl file", err)

	// go format
	formattedOutput, err := format.Source(compiledOutput.Bytes())
	cute.Check("go format", err)

	// save template with values as a go file (.go)
	err = saveToFile(EXPORTED_TRANSLATIONS_FILE, formattedOutput)
	cute.Check(fmt.Sprintf("exporting %s file", EXPORTED_TRANSLATIONS_FILE), err)

	// done
	cute.SetTitleColor(cute.BrightGreen)
	cute.Println("translations generated successfully 🎉")
}
