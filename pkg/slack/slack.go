package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	httpClient *http.Client
	hook       string
}

func New(hook string) Client {
	return Client{
		httpClient: new(http.Client),
		hook:       hook,
	}
}

func (c Client) Send(ctx context.Context, text string) error {
	var msg bytes.Buffer
	err := json.NewEncoder(&msg).Encode(struct {
		LinkNames bool   `json:"link_names"`
		Text      string `json:"text"`
	}{
		LinkNames: true,
		Text:      text,
	})
	if err != nil {
		return fmt.Errorf("error while preparing the Slack message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.hook, &msg)
	if err != nil {
		return fmt.Errorf("error building the HTTP request: %w", err)
	}
	req.Header.Add("content-type", "application/JSON")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error while sending the Slack message: %w", err)
	}

	defer func() {
		if _, e := io.Copy(io.Discard, res.Body); e != nil && err == nil {
			err = e
		}
		if e := res.Body.Close(); e != nil && err == nil {
			err = e
		}
	}()

	switch res.StatusCode {
	case http.StatusOK, http.StatusNoContent, http.StatusAccepted:
	default:
		err = fmt.Errorf("unexpected status code %q", res.Status)
	}

	return err
}
