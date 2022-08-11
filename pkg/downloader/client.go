package downloader

import (
	"net"

	"github.com/valyala/fasthttp"
)

type Client struct {
	cl *fasthttp.Client
}

var DefaultClient = NewClient()

func NewClient() *Client {
	cl := &fasthttp.Client{
		Name:                          "manic-client",
		NoDefaultUserAgentHeader:      true,
		DisableHeaderNamesNormalizing: false,
		DisablePathNormalizing:        false,
		Dial:                          defaultDial,
		DialDualStack:                 true,
	}
	return &Client{cl: cl}
}

type Addr struct {
	network string
	addr    string
}

func (a Addr) Network() string {
	return a.network
}

func (a Addr) String() string {
	return a.addr
}

func fromAddr(addr net.Addr) Addr {
	return Addr{
		network: addr.Network(),
		addr:    addr.String(),
	}
}

type Response struct {
	headers map[string]string
	body    []byte
	laddr   net.Addr
	raddr   net.Addr
}

func fromResponse(response *fasthttp.Response) (*Response, error) {
	heads := make(map[string]string)
	response.Header.VisitAll(func(key, value []byte) {
		heads[string(key)] = string(value)
	})
	var body []byte
	bw := response.BodyWriter()
	_, err := bw.Write(body)
	if err != nil {
		return nil, err
	}
	laddr := fromAddr(response.LocalAddr())
	raddr := fromAddr(response.RemoteAddr())
	return &Response{
		headers: heads,
		body:    body,
		laddr:   laddr,
		raddr:   raddr,
	}, nil
}

func (c *Client) Head(url string) (*Response, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(url)
	req.Header.SetMethod("HEAD")
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	err := c.cl.DoRedirects(req, resp, 30)
	if err != nil {
		return nil, err
	}
	res, err := fromResponse(resp)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) GetRange(url, val string) (*Response, error) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	req.Header.Set("RANGE", val)
	resp := fasthttp.AcquireResponse()
	err := c.cl.DoRedirects(req, resp, 30)
	if err != nil {
		return nil, err
	}
	fasthttp.ReleaseRequest(req)
	res, err := fromResponse(resp)
	if err != nil {
		return nil, err
	}
	fasthttp.ReleaseResponse(resp)
	return res, nil
}

func (c *Client) Get(url string) (*Response, error) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	resp := fasthttp.AcquireResponse()
	err := c.cl.DoRedirects(req, resp, 30)
	if err != nil {
		return nil, err
	}
	fasthttp.ReleaseRequest(req)
	res, err := fromResponse(resp)
	if err != nil {
		return nil, err
	}
	fasthttp.ReleaseResponse(resp)
	return res, nil
}

func Get(url string) (*Response, error) {
	return DefaultClient.Get(url)
}
