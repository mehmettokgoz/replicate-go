package replicate

import (
	"context"
	"encoding/json"
)

// Page represents a paginated response from Replicate's API.
type Page[T any] struct {
	Previous *string `json:"previous,omitempty"`
	Next     *string `json:"next,omitempty"`
	Results  []T     `json:"results"`

	rawJSON json.RawMessage `json:"-"`
}

func (p Page[T]) MarshalJSON() ([]byte, error) {
	if p.rawJSON != nil {
		return p.rawJSON, nil
	} else {
		type Alias Page[T]
		return json.Marshal(&struct{ *Alias }{Alias: (*Alias)(&p)})
	}
}

func (p *Page[T]) UnmarshalJSON(data []byte) error {
	p.rawJSON = data
	type Alias Page[T]
	alias := &struct{ *Alias }{Alias: (*Alias)(p)}
	return json.Unmarshal(data, alias)
}

// Paginate takes a Page and the Client request method, and iterates through pages of results.
func Paginate[T any](ctx context.Context, client *Client, initialPage *Page[T]) (<-chan []T, <-chan error) {
	resultsChan := make(chan []T)
	errChan := make(chan error)

	go func() {
		defer close(resultsChan)
		defer close(errChan)

		resultsChan <- initialPage.Results
		nextURL := initialPage.Next

		for nextURL != nil {
			page := &Page[T]{}
			err := client.fetch(ctx, "GET", *nextURL, nil, page)
			if err != nil {
				errChan <- err
				return
			}

			resultsChan <- page.Results

			nextURL = page.Next
		}
	}()

	return resultsChan, errChan
}
