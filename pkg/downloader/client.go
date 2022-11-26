package downloader

import (
	"encoding/json"
	"errors"
	"net"

	"github.com/valyala/fasthttp"
)

type Client struct {
	*fasthttp.Client
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
	return &Client{Client: cl}
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

// type Response struct {
// 	Headers map[string]string
// 	Body    []byte
// 	laddr   net.Addr
// 	raddr   net.Addr
// }
//
// func fromResponse(response *fasthttp.Response) (*Response, error) {
// 	heads := make(map[string]string)
// 	response.Header.VisitAll(func(key, value []byte) {
// 		heads[string(key)] = string(value)
// 	})
// 	var body []byte
// 	bw := response.BodyWriter()
// 	_, err := bw.Write(body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	laddr := fromAddr(response.LocalAddr())
// 	raddr := fromAddr(response.RemoteAddr())
// 	return &Response{
// 		Headers: heads,
// 		Body:    body,
// 		laddr:   laddr,
// 		raddr:   raddr,
// 	}, nil
// }

var respPool = func() chan *Response {
	c := make(chan *Response, 100)
	for i := 0; i < 100; i++ {
		c <- &Response{fasthttp.AcquireResponse()}
	}
	return c
}()

var reqPool = func() chan *Request {
	c := make(chan *Request, 100)
	for i := 0; i < 100; i++ {
		c <- &Request{fasthttp.AcquireRequest()}
	}
	return c
}()

type Response struct {
	*fasthttp.Response
}

type Request struct {
	*fasthttp.Request
}

func (r *Response) Json(res any) error {
	return json.Unmarshal(r.Body(), res)
}

func AcquireResponse() *Response {
	return <-respPool
}

func ReleaseResponse(response *Response) {
	fasthttp.ReleaseResponse(response.Response)
	response.Response = fasthttp.AcquireResponse()
	respPool <- response
}

func AcquireRequest() *Request {
	return <-reqPool
}

func ReleaseRequest(req *Request) {
	fasthttp.ReleaseRequest(req.Request)
	req.Request = fasthttp.AcquireRequest()
	reqPool <- req
}

func (c *Client) Head(url string) (*Response, error) {
	req := AcquireRequest()
	defer ReleaseRequest(req)
	req.SetRequestURI(url)
	req.Header.SetMethod("HEAD")
	resp := AcquireResponse()
	err := c.DoRedirects(req.Request, resp.Response, 30)
	if err != nil {
		return nil, err
	}
	// res, err := fromResponse(resp)
	// if err != nil {
	// 	return nil, err
	// }
	return resp, nil
}

func (c *Client) GetLength(url string) (int, error) {
	resp, err := c.Head(url)
	if err != nil {
		return 0, err
	}
	res := resp.Header.ContentLength()
	if res <= 0 {
		resp, err := c.GetRange(url, "bytes=0-0")
		if err != nil {
			return 0, err
		}
		res = resp.Header.ContentLength()
		if res <= 0 {
			return 0, errors.New("get length failed")
		}
	}
	ReleaseResponse(resp)
	return res, nil
}

func (c *Client) GetRange(url, val string) (*Response, error) {
	req := AcquireRequest()
	defer ReleaseRequest(req)
	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	req.Header.Set("RANGE", val)
	resp := AcquireResponse()
	err := c.DoRedirects(req.Request, resp.Response, 30)
	if err != nil {
		return nil, err
	}
	fasthttp.ReleaseRequest(req.Request)
	return resp, nil
}

func (c *Client) Get(url string) (*Response, error) {
	req := AcquireRequest()
	defer ReleaseRequest(req)
	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	resp := AcquireResponse()
	err := c.DoRedirects(req.Request, resp.Response, 30)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func Get(url string) (*Response, error) {
	return DefaultClient.Get(url)
}
