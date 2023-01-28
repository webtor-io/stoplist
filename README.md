# stoplist

Generates stoplist from YAML-format

Supports:
1. Regexp patterns
2. AND/OR conditions
3. References

## Installation

```
go get 'github.com/webtor-io/stoplist'
```

## Sample usage

```go
package main

import (
	"log"
	"github.com/webtor-io/stoplist"
)

var rules = `
main:
- abra
- cadabra
`

func main() error {
    r, err := stoplist.NewRuleFromYaml([]byte(rules))
    if err != nil {
        return err
    }
    res := r.Check("abra cadabra")
    log.Println(res)
    return nil
}
```

## Rules

1. YAML must have `main` key - it is starting point
2. Every key must have an array of rules and all these rules will be 
used with OR condition, like:
```yaml
main:
- rule1
- rule2
- ...
```
3. All keys can be referenced with curly brackets, like:
```yaml
rule:
- something
main:
- {rule}
```
4. Regular expression must be enclosed with slashes, like:
```yaml
main:
- /\d+/
```
5. OR condition is represented with pipe `|`, like:
```yaml
main:
- aaa|bbb|ccc
```
5. AND condition is represented with plus `+`, like:
```yaml
main:
- aaa+bbb+ccc
```


