<img src="https://raw.githubusercontent.com/zakaria-chahboun/ZakiQtProjects/master/IMAGE1.png">

# Tarjem

Tarjem is a powerful CLI tool for managing translations in your projects. It simplifies the process of internationalization by generating type-safe code from YAML translation files.

*tarjem ترجم is an arabic word means 'translate'*

![tarjem_mascot](/screenshot/tarjem_mascot.png)


## Features

- Tarjem uses `fmt.Sprintf` directly for precise formatting. This approach skips _templates_, which makes it faster and more efficient.
- Initialize translation files with a simple command
- Export translations to various programming languages
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
go install github.com/zakaria-chahboun/tarjem@latest
```

Or you can directly download the latest version for each OS from the [releases](https://github.com/zakaria-chahboun/tarjem/releases) page.

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

Tarjem provides clear error messages for various issues during translation, example:

![missing placeholder](/screenshot/parse_1.png)

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

order_status:
  variables:
    order_id: string
    delivery_time: datetime
  translations:
    english: "Your order {order_id} was placed on {delivery_time}."
    arabic: "تم تقديم طلبك {order_id} في {delivery_time}."
```

### Generated Code Usage (Go Example)

After exporting, you can use the generated functions in your Go code:

```go
import (
	"fmt"
	"time"
	"yourproject/tarjem"
)

func main() {

	// Set Arabic as language
	tarjem.SetCurrentLang(tarjem.LangArabic)
    
	// Print the translations
	fmt.Println(tarjem.Welcome()) // Output: أهلاً وسهلاً!
	fmt.Println(tarjem.OrderStatus("12345", time.Now())) // Output: تم تقديم طلبك 12345 في 2024-08-25 14:45:00.

	// Set English as language
	tarjem.SetCurrentLang(tarjem.LangEnglish)

	// Print the translations 
	fmt.Println(tarjem.Welcome()) // Output: Welcome!
	fmt.Println(tarjem.OrderStatus("12345", time.Now())) // Output: Your order 12345 was placed on 2024-08-25 14:45:00.
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

Support me to be an independent open-source programmer 💟

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/U7U3FQ2JA)

## License

This project is licensed under the MIT License - see the LICENSE file for details.

-----
Follow me on X: [@zaki_chahboun](https://x.com/Zaki_Chahboun)