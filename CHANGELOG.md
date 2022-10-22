## Breaking changes (v1.1.0)

- Add `String() string` method to `Message` struct.
- Change functions names for error messages: from `Create...` to `Report...`

| normal message example | error handling message example|
|----------------|-------------------------|
| `func CreateLastDatePayBill(date time.Time) (m *Message)` | `func ReportErrUserAccessDenied() (m *MessageError)` |

- Change the default package name in generated files: from `package messages` to `package tarjem`

### You now can specify the package name:

```go
# e.g translations
tarjem -package translations
```
