<img src="https://raw.githubusercontent.com/zakaria-chahboun/ZakiQtProjects/master/IMAGE1.png">

Support me to be an independent open-source programmer 💟

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/U7U3FQ2JA)

# Tarjem

Tarjem is a powerful CLI tool for managing translations in your projects. It simplifies the process of internationalization by generating type-safe code from YAML translation files.

> *tarjem ترجم is an arabic word means 'translate'*

## Features

- Tarjem uses `fmt.Sprintf` directly for precise formatting. This approach skips _templates_, which makes it faster and more efficient.
- Initialize translation files with a simple command
- Export translations to various programming languages
- Support for multiple natural languages
- Type-safe translation functions
- Date and time formatting support
- Easy integration with existing projects

## Current Language Support

As of now, Tarjem supports code generation for:

- Go

## Upcoming Language Support

We're actively working on expanding Tarjem's capabilities. In the near future, we plan to add support for:

- [ ] JavaScript
- [ ] Python
- [ ] Dart
- [ ] C
- [ ] C++
- [ ] Rust
- [ ] Zig

Stay tuned for updates!

## Installation

To install Tarjem, use the following command:

```console
go install github.com/yourusername/tarjem@latest
```

## Usage

### Initializing Translations

To create a new `translations.yaml` file:

```console
tarjem init
```

Use the `--force` flag to overwrite an existing file:

```console
tarjem init --force
```

### Exporting Translations

To generate code from your translations:

```console
tarjem export --lang go
```

Optionally, specify a package name (for Go):

```console
tarjem export --lang go --package mypackage
```

#### We provide good pacing messages when error happing, examples 

![missing placeholder](/screenshot/parse_1.png)

![not user variables in translation](/screenshot/parse_2.png)

![missing translations](/screenshot/parse_3.png)

![unsupported type](/screenshot/parse_4.png)

### Clearing Generated Files

To remove the generated translation file:

```console
tarjem clear
```

### Translation File Format

The `translations.yaml` file should follow this structure:

```yaml
welcome:
  translations:
    english: "Welcome!"
    arabic: "أهلاً وسهلاً!"

item_status:
  variables:
    name: string
    count: int
  translations:
    english: "{name} is available with {count} in stock."
    arabic: "العنصر {name} متاح بكمية {count}."
```

### Generated Code Usage (Go Example)

After exporting, you can use the generated functions in your Go code:

```go
import "yourproject/tarjem"

func main() {

	// Set Arabic as langues
	tarjem.SetCurrentLang(tarjem.LangArabic)
    
	// Print the translations
    fmt.Println(tarjem.Welcome()) // Output: أهلا وسهلا!
	fmt.Println(tarjem.ItemStatus("تفاح", 5)) // Output: العنصر تفاح متاح بكمية 5.

	// Set English as langues
    tarjem.SetCurrentLang(tarjem.LangEnglish)


	// Print the translations
    fmt.Println(tarjem.Welcome()) // Output: Welcome!
    fmt.Println(tarjem.ItemStatus("Apple", 5)) // Output: Apple is available with 5 in stock.
}
```

## Supported Variable Types

* `string`
* `int`
* `float`
* `date`
* `time`
* `datetime`

## Language Field Consistency in Translations

When defining translation fields, you have flexibility in naming the language keys. You can use any format you prefer:

```yaml
# Option 1
welcome:
  translations:
    arabic: "أهلاً وسهلاً!"
    english: "Welcome!"

# Option 2
welcome:
  translations:
    ar: "أهلاً وسهلاً!"
    en: "Welcome!"

# Option 3
welcome:
  translations:
    lang1: "أهلاً وسهلاً!"
    lang2: "Welcome!"
```

> [!IMPORTANT]
> However, the language keys must be consistent across all translation entries.


## Contribute 🌻

Feel free to contribute or propose a feature or share your idea with us!

## License

This project is licensed under the MIT License - see the LICENSE file for details.

-----
Follow me on X: [@zaki_chahboun](https://x.com/Zaki_Chahboun)