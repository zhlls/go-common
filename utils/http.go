package utils

import (
	"crypto/tls"
	"net"

	//"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	netURL "net/url"
	"strings"
	"time"
)

type HTTP struct {
	URL             string            `toml:"urls"`
	Method          string            `toml:"method"`
	Body            string            `toml:"body"`
	Path            string            `toml:"path"`
	ContentEncoding string            `toml:"content_encoding"`
	Parameters      map[string]string `json:"parameters"`

	Headers map[string]string `toml:"headers"`

	// HTTP Basic Auth Credentials
	Username string `toml:"username"`
	Password string `toml:"password"`
	//tls.ClientConfig

	Timeout time.Duration `toml:"timeout"`

	Client *http.Client
}

func (h *HTTP) Gather() ([]byte, error) {
	if h.Client == nil {
		h.Client = &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
					Proxy: http.ProxyFromEnvironment,
					DialContext: (&net.Dialer{
						Timeout:   30 * time.Second,
						KeepAlive: 30 * time.Second,
						DualStack: true,
					}).DialContext,
					MaxIdleConns:          50,
					IdleConnTimeout:       90 * time.Second,
					TLSHandshakeTimeout:   10 * time.Second,
					ExpectContinueTimeout: 1 * time.Second,
					MaxIdleConnsPerHost:   20,
					MaxConnsPerHost:       20,
				},
			Timeout: h.Timeout,
		}
	}

	data, err := h.gatherURL()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (h *HTTP) gatherURL() ([]byte, error) {
	body, err := makeRequestBodyReader(h.ContentEncoding, h.Body)
	if err != nil {
		return nil, err
	}

	u, err := netURL.ParseRequestURI(h.URL)
	if err != nil {
		return nil, err
	}

	if h.Path != "" {
		u.Path = h.Path
	}
	url := u.String()

	request, err := http.NewRequest(h.Method, url, body)
	if err != nil {
		return nil, err
	}

	if h.Parameters != nil {
		q := request.URL.Query()
		for k, v := range h.Parameters {
			q.Add(k, v)
		}
		request.URL.RawQuery = q.Encode()
	}

	if h.ContentEncoding == "gzip" {
		request.Header.Set("Content-Encoding", "gzip")
	}

	for k, v := range h.Headers {
		if strings.ToLower(k) == "host" {
			request.Host = v
		} else {
			request.Header.Add(k, v)
		}
	}

	if h.Username != "" || h.Password != "" {
		request.SetBasicAuth(h.Username, h.Password)
	}

	resp, err := h.Client.Do(request)

	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf(string(data))
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func makeRequestBodyReader(contentEncoding, body string) (io.Reader, error) {
	var (
		err    error
		reader io.Reader = strings.NewReader(body)
	)

	if contentEncoding == "gzip" {
		reader, err = CompressWithGzip(reader)
		if err != nil {
			return nil, err
		}
	}
	return reader, nil
}
