// Package fetchtest answers in place of the network for the adapters' tests:
// one fixed response or error, with every request recorded.
package fetchtest

import (
	"context"

	"github.com/oernster/EarthNow/internal/infrastructure/httpfetch"
)

// Getter implements httpfetch.Getter with a fixed answer.
type Getter struct {
	Resp    httpfetch.Response
	Err     error
	Asked   []string
	Accepts []string
}

// Answering is a Getter whose every answer is body.
func Answering(body []byte) *Getter {
	return &Getter{Resp: httpfetch.Response{Body: body}}
}

// GetAccepting records the request and answers the fixed response.
func (g *Getter) GetAccepting(_ context.Context, rawURL, _, accept string) (httpfetch.Response, error) {
	g.Asked = append(g.Asked, rawURL)
	g.Accepts = append(g.Accepts, accept)
	return g.Resp, g.Err
}
