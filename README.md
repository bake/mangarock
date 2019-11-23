# MangaRock

[![GoDoc](https://godoc.org/github.com/bake/mangarock?status.svg)](https://godoc.org/github.com/bake/mangarock)
[![Go Report Card](https://goreportcard.com/badge/github.com/bake/mangarock)](https://goreportcard.com/report/github.com/bake/mangarock)

A Go client for the MangaRock API. Use [mri](https://github.com/bake/mri) to
decode images.

```go
func main() {
  c := mangarock.New()
  m, err := c.Manga("mrs-serie-100177863")
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(m.Name)
  // Fushigi Neko no Kyuu-chan
}
```
