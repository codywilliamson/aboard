package trello

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const baseURL = "https://api.trello.com/1"

type Client struct {
	apiKey string
	token  string
	http   *http.Client
}

type Card struct {
	ID       string
	Name     string
	Desc     string
	URL      string
	ShortURL string
	IDList   string
	ListName string
}

type Board struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type List struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type cardResponse struct {
	ID       string `json:"id"`
	IDList   string `json:"idList"`
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	URL      string `json:"url"`
	ShortURL string `json:"shortUrl"`
}

func NewClient(apiKey, token string) *Client {
	return &Client{
		apiKey: apiKey,
		token:  token,
		http: &http.Client{
			Timeout: 12 * time.Second,
		},
	}
}

func (c *Client) CanAuth() bool {
	return c.apiKey != "" && c.token != ""
}

func (c *Client) Boards(ctx context.Context) ([]Board, error) {
	if !c.CanAuth() {
		return nil, errors.New("missing TRELLO_API_KEY or TRELLO_API_TOKEN")
	}

	u, err := url.Parse(baseURL + "/members/me/boards")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("key", c.apiKey)
	q.Set("token", c.token)
	q.Set("fields", "id,name")
	q.Set("filter", "open")
	u.RawQuery = q.Encode()

	var boards []Board
	if err := c.getJSON(ctx, u.String(), &boards); err != nil {
		return nil, err
	}
	return boards, nil
}

func (c *Client) CardsForBoard(ctx context.Context, boardID string) ([]Card, error) {
	if boardID == "" {
		return nil, errors.New("board id is required")
	}
	if !c.CanAuth() {
		return nil, errors.New("missing TRELLO_API_KEY or TRELLO_API_TOKEN")
	}

	lists, err := c.ListsForBoard(ctx, boardID)
	if err != nil {
		return nil, err
	}
	listByID := make(map[string]string, len(lists))
	for _, l := range lists {
		listByID[l.ID] = l.Name
	}

	u, err := url.Parse(baseURL + "/boards/" + boardID + "/cards")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("key", c.apiKey)
	q.Set("token", c.token)
	q.Set("fields", "id,name,desc,idList,url,shortUrl")
	q.Set("filter", "open")
	u.RawQuery = q.Encode()

	var rawCards []cardResponse
	if err := c.getJSON(ctx, u.String(), &rawCards); err != nil {
		return nil, err
	}

	cards := make([]Card, 0, len(rawCards))
	for _, rc := range rawCards {
		cards = append(cards, Card{
			ID:       rc.ID,
			IDList:   rc.IDList,
			Name:     rc.Name,
			Desc:     rc.Desc,
			URL:      rc.URL,
			ShortURL: rc.ShortURL,
			ListName: listByID[rc.IDList],
		})
	}

	return cards, nil
}

func (c *Client) ListsForBoard(ctx context.Context, boardID string) ([]List, error) {
	u, err := url.Parse(baseURL + "/boards/" + boardID + "/lists")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("key", c.apiKey)
	q.Set("token", c.token)
	q.Set("fields", "id,name")
	q.Set("filter", "open")
	u.RawQuery = q.Encode()

	var lists []List
	if err := c.getJSON(ctx, u.String(), &lists); err != nil {
		return nil, err
	}
	return lists, nil
}

func (c *Client) MoveCard(ctx context.Context, cardID, listID string) error {
	return c.putForm(ctx, "/cards/"+cardID, url.Values{"idList": {listID}})
}

func (c *Client) UpdateCard(ctx context.Context, cardID, name, desc string) error {
	vals := url.Values{}
	if name != "" {
		vals.Set("name", name)
	}
	if desc != "" {
		vals.Set("desc", desc)
	}
	return c.putForm(ctx, "/cards/"+cardID, vals)
}

func (c *Client) AddComment(ctx context.Context, cardID, text string) error {
	return c.postForm(ctx, "/cards/"+cardID+"/actions/comments", url.Values{"text": {text}}, nil)
}

func (c *Client) ArchiveCard(ctx context.Context, cardID string) error {
	return c.putForm(ctx, "/cards/"+cardID, url.Values{"closed": {"true"}})
}

func (c *Client) CreateCard(ctx context.Context, listID, name string) (*Card, error) {
	var raw cardResponse
	err := c.postForm(ctx, "/lists/"+listID+"/cards", url.Values{"name": {name}}, &raw)
	if err != nil {
		return nil, err
	}
	return &Card{ID: raw.ID, Name: raw.Name, IDList: raw.IDList, URL: raw.URL, ShortURL: raw.ShortURL}, nil
}

func (c *Client) CreateList(ctx context.Context, boardID, name string) (*List, error) {
	var raw List
	err := c.postForm(ctx, "/boards/"+boardID+"/lists", url.Values{"name": {name}}, &raw)
	if err != nil {
		return nil, err
	}
	return &raw, nil
}

func (c *Client) ArchiveList(ctx context.Context, listID string) error {
	return c.putForm(ctx, "/lists/"+listID+"/closed", url.Values{"value": {"true"}})
}

func (c *Client) getJSON(ctx context.Context, endpoint string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("trello returned %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return err
	}
	return nil
}

func (c *Client) putForm(ctx context.Context, path string, vals url.Values) error {
	vals.Set("key", c.apiKey)
	vals.Set("token", c.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, baseURL+path, strings.NewReader(vals.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("trello returned %s", resp.Status)
	}
	return nil
}

func (c *Client) postForm(ctx context.Context, path string, vals url.Values, target any) error {
	vals.Set("key", c.apiKey)
	vals.Set("token", c.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+path, strings.NewReader(vals.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("trello returned %s", resp.Status)
	}
	if target != nil {
		return json.NewDecoder(resp.Body).Decode(target)
	}
	return nil
}
