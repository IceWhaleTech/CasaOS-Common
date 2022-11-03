// methods in this package automatically include a context with a timeout, in order to solve the problem of hanging requests and to avoid goroutine leaks
package http

import (
	"bytes"
	"context"
	"net/http"
	"time"
)

func Do(requestFunc func(ctx context.Context) (*http.Request, error), timeout time.Duration) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	request, err := requestFunc(ctx)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func Get(url string, timeout time.Duration) (*http.Response, error) {
	return Do(func(ctx context.Context) (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	}, timeout)
}

func Post(url string, body []byte, timeout time.Duration) (*http.Response, error) {
	return Do(func(ctx context.Context) (*http.Request, error) {
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
		if err != nil {
			return nil, err
		}

		request.Header.Set("Content-Type", "application/json")

		return request, nil
	}, timeout)
}

func Put(url string, body []byte, timeout time.Duration) (*http.Response, error) {
	return Do(func(ctx context.Context) (*http.Request, error) {
		request, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(body))
		if err != nil {
			return nil, err
		}

		request.Header.Set("Content-Type", "application/json")

		return request, nil
	}, timeout)
}

func Delete(url string, body []byte, timeout time.Duration) (*http.Response, error) {
	return Do(func(ctx context.Context) (*http.Request, error) {
		request, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, bytes.NewBuffer(body))
		if err != nil {
			return nil, err
		}

		request.Header.Set("Content-Type", "application/json")

		return request, nil
	}, timeout)
}
