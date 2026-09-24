package httpfetch

import (
	"context"
	"errors"
)

// Getter is the part of Client an image adapter uses (the cloud and burnt-area
// layers), so its tests can answer in its place.
type Getter interface {
	GetAccepting(ctx context.Context, rawURL, validator, accept string) (Response, error)
}

// ErrNotChanged is a source answering not modified to a request that asked
// nothing conditional, which leaves no body to use.
var ErrNotChanged = errors.New("the service answered not modified to an unconditional request")

// Fresh fetches rawURL unconditionally, asking for accept; it answers the body.
func Fresh(ctx context.Context, get Getter, rawURL, accept string) ([]byte, error) {
	resp, err := get.GetAccepting(ctx, rawURL, "", accept)
	if err != nil {
		return nil, err
	}
	if resp.NotModified {
		return nil, ErrNotChanged
	}
	return resp.Body, nil
}
