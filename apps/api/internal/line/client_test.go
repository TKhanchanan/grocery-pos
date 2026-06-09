package line

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestClientPush(t *testing.T) {
	var authorization, contentType, body string
	client := NewClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		authorization = req.Header.Get("Authorization")
		contentType = req.Header.Get("Content-Type")
		payload, _ := io.ReadAll(req.Body)
		body = string(payload)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
		}, nil
	})})

	err := client.Push(context.Background(), "token", "target", NewTextMessage("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if authorization != "Bearer token" {
		t.Errorf("authorization = %q", authorization)
	}
	if contentType != "application/json" {
		t.Errorf("content type = %q", contentType)
	}
	for _, want := range []string{`"to":"target"`, `"type":"text"`, `"text":"hello"`} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %s: %s", want, body)
		}
	}
}
