package common

import (
	"bytes"
	"io"
	"net/http"
)

func IIf[T any](cond bool, trueVal T, falseVal T) T {
	if !cond {
		return falseVal
	}

	return trueVal
}

type RequestOpts struct {
	Params  map[string]string
	Headers map[string]string
}

type Response struct {
	*http.Response
}

func (r *Response) ReadBody() ([]byte, error) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func MakeRequest(method string, url string, data []byte, opts ...func(o *RequestOpts)) (*Response, error) {
	body := bytes.NewBuffer(data)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	o := RequestOpts{
		Params:  make(map[string]string),
		Headers: make(map[string]string),
	}

	for _, fn := range opts {
		fn(&o)
	}

	q := req.URL.Query()
	for k, v := range o.Params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	for k, v := range o.Headers {
		req.Header.Add(k, v)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return &Response{res}, nil
}
