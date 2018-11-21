package mangarock

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

// Client is a MangaRock Client.
type Client struct {
	base   string
	client *http.Client
}

// WithBase sets the API base.
func WithBase(base string) func(*Client) {
	return func(mr *Client) {
		mr.base = base
	}
}

// WithHTTPClient makes the manga client use a given http.Client to make
// requests.
func WithHTTPClient(c *http.Client) func(*Client) {
	return func(mr *Client) {
		mr.client = c
	}
}

// New returns a new MangaRock Client.
func New(options ...func(*Client)) *Client {
	mr := &Client{
		base:   "https://api.mangarockhd.com/query/web401",
		client: &http.Client{},
	}
	for _, option := range options {
		option(mr)
	}
	return mr
}

// get sends a HTTP GET request.
func (c *Client) get(url string, query url.Values) (json.RawMessage, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not create get request")
	}
	req.URL.RawQuery = query.Encode()
	res, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "cound not get %v", url)
	}
	defer res.Body.Close()
	var resp response
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, errors.Wrap(err, "could not decode response")
	}
	if resp.Code != 0 {
		return nil, errors.Errorf("get response code %d", resp.Code)
	}
	return resp.Data, nil
}

// post sends a HTTP POST request.
func (c *Client) post(url string, body interface{}) (json.RawMessage, error) {
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		json.NewEncoder(pw).Encode(body)
	}()
	req, err := http.NewRequest(http.MethodPost, url, pr)
	if err != nil {
		return nil, errors.Wrap(err, "could not create post request")
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "cound not post to %v", url)
	}
	defer res.Body.Close()
	var resp response
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, errors.Wrap(err, "could not decode response")
	}
	if resp.Code != 0 {
		return nil, errors.Errorf("get response code %d", resp.Code)
	}
	return resp.Data, nil
}

type response struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

type Manga struct {
	ID              string    `json:"oid"`
	Name            string    `json:"name"`
	Author          Author    `json:"-"`
	Authors         []Author  `json:"authors"`
	AuthorIDs       []string  `json:"author_ids"`
	Genres          []string  `json:"genres"`
	Rank            int       `json:"rank"`
	UpdatedChapters int       `json:"updated_chapters"`
	NewChapters     []Chapter `json:"new_chapters"`
	Completed       bool      `json:"cmpleted"`
	Thumbnail       string    `json:"thumbnail"`
	Updated         time.Time `json:"updated_at"`
}

type MangaSingle struct {
	Manga
	Description string     `json:"description"`
	Chapters    []Chapter  `json:"chapters"`
	Categories  []Category `json:"rich_categories"`
	Cover       string     `json:"cover"`
	Artworks    []string   `json:"artworks"`
	Aliases     []string   `json:"alias"`
}

type Chapter struct {
	ID   string `json:"oid"`
	Name string `json:"name"`
	// Updated string `json:"updatedAt"`

	// Fields only available if requested as a single object
	Order int `json:"order"`

	// Fields available if requested as chapter
	Pages []string `json:"pages"`
}

type Category struct {
	ID   string `json:"oid"`
	Name string `json:"name"`
}

type Author struct {
	ID        string `json:"oid"`
	Name      string `json:"name"`
	Thumbnail string `json:"thumbnail"`

	// Only available if requested through a manga
	Role string `json:"role"`
}

// Latest returns the latest mangas. It only uses the manga IDs and requests a
// list like the one that would be returnd by a search. Fields like recently
// added chapters are missing, but authors are added.
func (c *Client) Latest(page int) ([]Manga, error) {
	res, err := c.get(c.base+"/mrs_latest", nil)
	if err != nil {
		return nil, err
	}
	var mangas []Manga
	if err := json.Unmarshal(res, &mangas); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal latest mangas")
	}
	ids := make([]string, len(mangas))
	for i, manga := range mangas {
		ids[i] = manga.ID
	}
	mangas, err = c.mangasByIDs(ids)
	if err != nil {
		return nil, errors.Wrap(err, "could not get latest mangas by ids")
	}
	return c.addAuthors(mangas)
}

// Search for mangas.
func (c *Client) Search(query string) ([]Manga, error) {
	body := struct {
		Type     string `json:"type"`
		Keywords string `json:"keywords"`
	}{"series", query}
	res, err := c.post(c.base+"/mrs_search", body)
	if err != nil {
		return nil, errors.Wrap(err, "could not execute search")
	}

	var ids []string
	if err := json.Unmarshal(res, &ids); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal manga IDs")
	}
	mangas, err := c.mangasByIDs(ids)
	if err != nil {
		return nil, errors.Wrap(err, "could not get mangas by ids")
	}
	return c.addAuthors(mangas)
}

// addAuthors adds authors to mangas based on their IDs.
func (c *Client) addAuthors(mangas []Manga) ([]Manga, error) {
	var ids []string
	for _, manga := range mangas {
		ids = append(ids, manga.AuthorIDs...)
	}
	authors, err := c.authorsByIDs(ids)
	if err != nil {
		return nil, errors.Wrap(err, "could not get authors by ids")
	}
	authorMap := map[string]Author{}
	for _, author := range authors {
		authorMap[author.ID] = author
	}

	for i, manga := range mangas {
		for _, id := range manga.AuthorIDs {
			mangas[i].Authors = append(mangas[i].Authors, authorMap[id])
		}
		if len(mangas[i].Authors) == 0 {
			continue
		}
		mangas[i].Author = mangas[i].Authors[0]
	}

	return mangas, nil
}

// mangasByIDs returns a list of mangas based on IDs. Can be used to unify
// manga results that are slightly different.
func (c *Client) mangasByIDs(ids []string) ([]Manga, error) {
	res, err := c.post("https://api.mangarockhd.com/meta", ids)
	if err != nil {
		return nil, errors.Wrap(err, "could not get meta data by manga ids")
	}
	var mangaMap map[string]Manga
	if err := json.Unmarshal(res, &mangaMap); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal mangas by ids")
	}
	var mangas []Manga
	for _, id := range ids {
		if manga, ok := mangaMap[id]; ok {
			mangas = append(mangas, manga)
		}
	}
	return mangas, nil
}

// authorsByIDs returns a slice of authors by their IDs.
func (c *Client) authorsByIDs(ids []string) ([]Author, error) {
	res, err := c.post("https://api.mangarockhd.com/meta", ids)
	if err != nil {
		return nil, errors.Wrap(err, "could not get meta data by author ids")
	}
	var authorMap map[string]Author
	if err := json.Unmarshal(res, &authorMap); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal authors by ids")
	}
	var authors []Author
	for _, author := range authorMap {
		authors = append(authors, author)
	}
	return authors, nil
}

// Manga returns a single manga. It may contain more fields than a regular one.
func (c *Client) Manga(id string) (MangaSingle, error) {
	res, err := c.post(c.base+"/info?oid="+id, nil)
	if err != nil {
		return MangaSingle{}, err
	}
	var manga MangaSingle
	if err := json.Unmarshal(res, &manga); err != nil {
		return MangaSingle{}, errors.Wrap(err, "could not unmarshal manga")
	}
	manga.Author = manga.Authors[0]
	return manga, nil
}

// Mangas returns a slice of mangas.
func (c *Client) Mangas(ids []string) ([]Manga, error) {
	mangas, err := c.mangasByIDs(ids)
	if err != nil {
		return nil, errors.Wrap(err, "could not get authors mangas")
	}
	return mangas, nil
}

// Chapter returns a chapter containing its images.
func (c *Client) Chapter(id, cid string) (Chapter, error) {
	manga, err := c.Manga(id)
	if err != nil {
		return Chapter{}, errors.Wrap(err, "could not get manga")
	}

	res, err := c.post(c.base+"/pages?oid="+cid, nil)
	if err != nil {
		return Chapter{}, errors.Wrap(err, "could not get pages")
	}
	var pages []string
	if err := json.Unmarshal(res, &pages); err != nil {
		return Chapter{}, errors.Wrap(err, "could not unmarhal pages")
	}

	for _, chapter := range manga.Chapters {
		if chapter.ID != cid {
			continue
		}
		chapter.Pages = pages
		return chapter, nil
	}
	return Chapter{}, errors.New("chapter not found")
}

// Author returns an author and their mangas.
func (c *Client) Author(id string) (Author, []Manga, error) {
	authors, err := c.authorsByIDs([]string{id})
	if err != nil {
		return Author{}, nil, errors.Wrap(err, "could not get authors meta data")
	}
	if len(authors) == 0 {
		return Author{}, nil, errors.Errorf("author with id %s not found", id)
	}

	res, err := c.get(c.base+"/mrs_serie_related_author", url.Values{"oid": []string{id}})
	if err != nil {
		return Author{}, nil, errors.Wrap(err, "could not get authors mangas")
	}
	var mangaIDStructs []struct {
		ID string `json:"oid"`
	}
	if err := json.Unmarshal(res, &mangaIDStructs); err != nil {
		return Author{}, nil, errors.Wrap(err, "could not unmarshal authors meta data")
	}
	var mangaIDs []string
	for _, manga := range mangaIDStructs {
		mangaIDs = append(mangaIDs, manga.ID)
	}
	mangas, err := c.mangasByIDs(mangaIDs)
	if err != nil {
		return Author{}, nil, errors.Wrap(err, "could not get authors mangas")
	}
	for i := range mangas {
		mangas[i].Author = authors[0]
	}
	return authors[0], mangas, nil
}
