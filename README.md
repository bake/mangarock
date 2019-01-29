# MangaRock

[![GoDoc](https://godoc.org/github.com/bakerolls/mangarock?status.svg)](https://godoc.org/github.com/bakerolls/mangarock)
[![Go Report Card](https://goreportcard.com/badge/github.com/bakerolls/mangarock)](https://goreportcard.com/report/github.com/bakerolls/mangarock)

A Go client for the MangaRock API.

```go
func main() {
	m, err := c.Manga("mrs-serie-100177863")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(m.Name)
	// Fushigi Neko no Kyuu-chan
}
```
