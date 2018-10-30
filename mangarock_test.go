package mangarock_test

import (
	"testing"

	"git.192k.pw/bake/mangarock"
	"github.com/BakeRolls/httpcache"
	"github.com/BakeRolls/httpcache/diskcache"
)

var c *mangarock.Client

func init() {
	client := httpcache.New(diskcache.New("mangarock_test_data", 0)).Client()
	c = mangarock.New(mangarock.WithHTTPClient(client))
}

func TestLatest(t *testing.T) {
	mangas, err := c.Latest(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(mangas) == 0 {
		t.Fatal("no mangas returned")
	}
	if mangas[0].Author.ID == "" {
		t.Fatal("author id empty")
	}
}

func TestManga(t *testing.T) {
	id := "mrs-serie-100177863"
	name := "Fushigi Neko no Kyuu-chan"
	manga, err := c.Manga(id)
	if err != nil {
		t.Fatal(err)
	}
	if manga.Name != name {
		t.Fatalf("expected name to be %s, got %s", name, manga.Name)
	}
}

func TestSearch(t *testing.T) {
	query := "kyuu-chan"
	num := 1
	name := "Fushigi Neko no Kyuu-chan"
	author := "Nitori Sasami"
	mangas, err := c.Search(query)
	if err != nil {
		t.Fatal(err)
	}
	if len(mangas) != num {
		t.Fatalf("expected %d results, got %d", num, len(mangas))
	}
	if mangas[0].Name != name {
		t.Fatalf("expected name to be %s, got %s", name, mangas[0].Name)
	}
	if mangas[0].Author.Name != author {
		t.Fatalf("expected author to be %s, got %s", author, mangas[0].Author.Name)
	}
}

func TestChapter(t *testing.T) {
	mid := "mrs-serie-100177863"
	cid := "mrs-chapter-100177864"
	name := "Chapter 1: Snow"
	pages := 1
	chapter, err := c.Chapter(mid, cid)
	if err != nil {
		t.Fatal(err)
	}
	if chapter.Name != name {
		t.Fatalf("expected name to be %s, got %s", name, chapter.Name)
	}
	if len(chapter.Pages) != pages {
		t.Fatalf("expected number of pages to be %d, got %d", pages, len(chapter.Pages))
	}
}
