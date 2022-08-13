//go:build unix && !(darwin && arm64)

package downloader

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/gnet/v2"
	"github.com/pterm/pterm"
	"github.com/valyala/fasthttp"
)

const DefaultDialTimeout = 3 * time.Second

const DefaultDNSCacheDuration = time.Minute

func ProxyDialer(proxy string, timeout time.Duration) fasthttp.DialFunc {
	var auth string
	if strings.Contains(proxy, "@") {
		split := strings.Split(proxy, "@")
		auth = base64.StdEncoding.EncodeToString([]byte(split[0]))
		proxy = split[1]
	}
	return func(addr string) (net.Conn, error) {
		var conn net.Conn
		var err error
		if timeout == 0 {
			conn, err = Dial(proxy)
		} else {
			conn, err = DialTimeout(proxy, timeout)
		}
		if err != nil {
			return nil, err
		}

		req := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", addr, addr)
		if auth != "" {
			req += "Proxy-Authorization: Basic " + auth + "\r\n"
		}
		req += "\r\n"

		if _, err := conn.Write([]byte(req)); err != nil {
			return nil, err
		}

		res := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(res)

		res.SkipBody = true

		if err := res.Read(bufio.NewReader(conn)); err != nil {
			err1 := conn.Close()
			if err1 != nil {
				msg := fmt.Sprintf("%s | %s", err, err1)
				return nil, errors.New(msg)
			}
			return nil, err
		}
		if res.Header.StatusCode() != 200 {
			err = conn.Close()
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("could not connect to proxy: %s status code: %d", proxy, res.Header.StatusCode())
		}
		return conn, nil
	}
}

var defaultDial fasthttp.DialFunc = Dial

// var DefaultClient = &fasthttp.Client{
// 	Name:                     "manic-client",
// 	NoDefaultUserAgentHeader: true,
// 	Dial:                     defaultDial,
// 	DialDualStack:            true,
// 	ConfigureClient:          nil,
// }

func Dial(addr string) (net.Conn, error) {
	return defaultDialer.Dial(addr)
}

func DialTimeout(addr string, timeout time.Duration) (net.Conn, error) {
	return defaultDialer.DialTimeout(addr, timeout)
}

// DialDualStack dials the given TCP addr using both tcp4 and tcp6.
//
// This function has the following additional features comparing to net.Dial:
//
//   - It reduces load on DNS resolver by caching resolved TCP addressed
//     for DNSCacheDuration.
//   - It dials all the resolved TCP addresses in round-robin manner until
//     connection is established. This may be useful if certain addresses
//     are temporarily unreachable.
//   - It returns ErrDialTimeout if connection cannot be established during
//     DefaultDialTimeout seconds. Use DialDualStackTimeout for custom dial
//     timeout.
//
// This dialer is intended for custom code wrapping before passing
// to Client.Dial or HostClient.Dial.
//
// For instance, per-host counters and/or limits may be implemented
// by such wrappers.
//
// The addr passed to the function must contain port. Example addr values:
//
//   - foobar.baz:443
//   - foo.bar:80
//   - aaa.com:8080
func DialDualStack(addr string) (net.Conn, error) {
	return defaultDialer.dial(addr, true, DefaultDialTimeout)
}

// DialDualStackTimeout dials the given TCP addr using both tcp4 and tcp6
// using the given timeout.
//
// This function has the following additional features comparing to net.Dial:
//
//   - It reduces load on DNS resolver by caching resolved TCP addressed
//     for DNSCacheDuration.
//   - It dials all the resolved TCP addresses in round-robin manner until
//     connection is established. This may be useful if certain addresses
//     are temporarily unreachable.
//
// This dialer is intended for custom code wrapping before passing
// to Client.Dial or HostClient.Dial.
//
// For instance, per-host counters and/or limits may be implemented
// by such wrappers.
//
// The addr passed to the function must contain port. Example addr values:
//
//   - foobar.baz:443
//   - foo.bar:80
//   - aaa.com:8080
func DialDualStackTimeout(addr string, timeout time.Duration) (net.Conn, error) {
	return defaultDialer.dial(addr, true, timeout)
}

type tcpAddrEntry struct {
	addrs       []net.TCPAddr
	addrsIdx    uint32
	pending     int32
	resolveTime time.Time
}

var DefaultOptions = &DialerOptions{
	Conn:             1000,
	LocalAddr:        nil,
	DNSCacheDuration: DefaultDNSCacheDuration,
}

var defaultDialer = func() *GNetDialer {
	res, err := NewDialer(DefaultOptions)
	if err != nil {
		pterm.Fatal.Println(err)
	}
	return res
}()

var GNetResolver = func(cl *gnet.Client) *net.Resolver {
	return &net.Resolver{
		PreferGo:     true,
		StrictErrors: false,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			conn, err := defaultDialer.DialContext(ctx, network, address)
			err = checkDialError(ctx, err)
			if err != nil {
				return nil, err
			}
			return cl.Enroll(conn)
		},
	}
}

func checkDialError(ctx context.Context, err error) error {
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded || err == ErrDialTimeout {
			return ErrDialTimeout
		}
		return err
	}
	return nil
}

type DialerOptions struct {
	Conn             int
	LocalAddr        *net.TCPAddr
	ProxyFunc        ProxyFunc
	DNSCacheDuration time.Duration
}

type GNetDialer struct {
	Concurrency      int
	LocalAddr        *net.TCPAddr
	Resolver         *net.Resolver
	DNSCacheDuration time.Duration
	tcpAddrsMap      sync.Map
	concurrencyCh    chan struct{}
	once             sync.Once
	cl               *gnet.Client
}

func NewDialer(opts *DialerOptions, GNetOpts ...gnet.Option) (*GNetDialer, error) {
	cl, err := gnet.NewClient(&gnet.BuiltinEventEngine{}, GNetOpts...)
	if err != nil {
		return nil, err
	}
	res := GNetResolver(cl)
	if opts == nil {
		opts = DefaultOptions
	}
	dialer := new(GNetDialer)
	dialer.cl = cl
	dialer.Concurrency = opts.Conn
	dialer.LocalAddr = opts.LocalAddr
	dialer.DNSCacheDuration = opts.DNSCacheDuration
	dialer.Resolver = res
	dialer.tcpAddrsMap = sync.Map{}
	dialer.once = sync.Once{}
	dialer.concurrencyCh = make(chan struct{})
	return dialer, nil
}

// Dial dials the given TCP addr using tcp4.
//
// This function has the following additional features comparing to net.Dial:
//
//   - It reduces load on DNS resolver by caching resolved TCP addressed
//     for DNSCacheDuration.
//   - It dials all the resolved TCP addresses in round-robin manner until
//     connection is established. This may be useful if certain addresses
//     are temporarily unreachable.
//   - It returns ErrDialTimeout if connection cannot be established during
//     DefaultDialTimeout seconds. Use DialTimeout for customizing dial timeout.
//
// This dialer is intended for custom code wrapping before passing
// to Client.Dial or HostClient.Dial.
//
// For instance, per-host counters and/or limits may be implemented
// by such wrappers.
//
// The addr passed to the function must contain port. Example addr values:
//
//   - foobar.baz:443
//   - foo.bar:80
//   - aaa.com:8080
func (d *GNetDialer) Dial(addr string) (net.Conn, error) {
	return d.dial(addr, false, DefaultDialTimeout)
}

// DialTimeout dials the given TCP addr using tcp4 using the given timeout.
//
// This function has the following additional features comparing to net.Dial:
//
//   - It reduces load on DNS resolver by caching resolved TCP addressed
//     for DNSCacheDuration.
//   - It dials all the resolved TCP addresses in round-robin manner until
//     connection is established. This may be useful if certain addresses
//     are temporarily unreachable.
//
// This dialer is intended for custom code wrapping before passing
// to Client.Dial or HostClient.Dial.
//
// For instance, per-host counters and/or limits may be implemented
// by such wrappers.
//
// The addr passed to the function must contain port. Example addr values:
//
//   - foobar.baz:443
//   - foo.bar:80
//   - aaa.com:8080
func (d *GNetDialer) DialTimeout(addr string, timeout time.Duration) (net.Conn, error) {
	return d.dial(addr, false, timeout)
}

// DialDualStack dials the given TCP addr using both tcp4 and tcp6.
//
// This function has the following additional features comparing to net.Dial:
//
//   - It reduces load on DNS resolver by caching resolved TCP addressed
//     for DNSCacheDuration.
//   - It dials all the resolved TCP addresses in round-robin manner until
//     connection is established. This may be useful if certain addresses
//     are temporarily unreachable.
//   - It returns ErrDialTimeout if connection cannot be established during
//     DefaultDialTimeout seconds. Use DialDualStackTimeout for custom dial
//     timeout.
//
// This dialer is intended for custom code wrapping before passing
// to Client.Dial or HostClient.Dial.
//
// For instance, per-host counters and/or limits may be implemented
// by such wrappers.
//
// The addr passed to the function must contain port. Example addr values:
//
//   - foobar.baz:443
//   - foo.bar:80
//   - aaa.com:8080
func (d *GNetDialer) DialDualStack(addr string) (net.Conn, error) {
	return d.dial(addr, true, DefaultDialTimeout)
}

// DialDualStackTimeout dials the given TCP addr using both tcp4 and tcp6
// using the given timeout.
//
// This function has the following additional features comparing to net.Dial:
//
//   - It reduces load on DNS resolver by caching resolved TCP addressed
//     for DNSCacheDuration.
//   - It dials all the resolved TCP addresses in round-robin manner until
//     connection is established. This may be useful if certain addresses
//     are temporarily unreachable.
//
// This dialer is intended for custom code wrapping before passing
// to Client.Dial or HostClient.Dial.
//
// For instance, per-host counters and/or limits may be implemented
// by such wrappers.
//
// The addr passed to the function must contain port. Example addr values:
//
//   - foobar.baz:443
//   - foo.bar:80
//   - aaa.com:8080
func (d *GNetDialer) DialDualStackTimeout(addr string, timeout time.Duration) (net.Conn, error) {
	return d.dial(addr, true, timeout)
}

// ErrDialTimeout is returned when TCP dialing is timed out.
var ErrDialTimeout = errors.New("dialing to the given TCP address timed out")

func (d *GNetDialer) dial(addr string, dualStack bool, timeout time.Duration) (net.Conn, error) {
	d.once.Do(func() {
		if d.Concurrency > 0 {
			d.concurrencyCh = make(chan struct{}, d.Concurrency)
		}

		if d.DNSCacheDuration == 0 {
			d.DNSCacheDuration = DefaultDNSCacheDuration
		}

		go d.tcpAddrsClean()
	})

	addrs, idx, err := d.getTCPAddrs(addr, dualStack)
	if err != nil {
		return nil, err
	}
	network := "tcp4"
	if dualStack {
		network = "tcp"
	}

	var conn net.Conn
	n := uint32(len(addrs))
	deadline := time.Now().Add(timeout)
	for n > 0 {
		conn, err = d.tryDial(network, &addrs[idx%n], deadline, d.concurrencyCh)
		if err == nil {
			return conn, nil
		}
		if err == ErrDialTimeout {
			return nil, err
		}
		idx++
		n--
	}
	return nil, err
}

func (d *GNetDialer) tryDial(network string, addr *net.TCPAddr, deadline time.Time, concurrencyCh chan struct{}) (net.Conn, error) {
	timeout := -time.Since(deadline)
	if timeout <= 0 {
		return nil, ErrDialTimeout
	}

	if concurrencyCh != nil {
		select {
		case concurrencyCh <- struct{}{}:
		default:
			tc := fasthttp.AcquireTimer(timeout)
			isTimeout := false
			select {
			case concurrencyCh <- struct{}{}:
			case <-tc.C:
				isTimeout = true
			}
			fasthttp.ReleaseTimer(tc)
			if isTimeout {
				return nil, ErrDialTimeout
			}
		}
		defer func() { <-concurrencyCh }()
	}

	ctx, cancelCtx := context.WithDeadline(context.Background(), deadline)
	defer cancelCtx()
	conn, err := d.DialContext(ctx, network, addr.String())
	if err != nil && (ctx.Err() == context.DeadlineExceeded || err == ErrDialTimeout) {
		return nil, ErrDialTimeout
	}
	return conn, err
}

func (d *GNetDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	dialer := net.Dialer{}
	if d.LocalAddr != nil {
		dialer.LocalAddr = d.LocalAddr
	}
	conn, err := dialer.DialContext(ctx, network, addr)
	if err != nil && ctx.Err() == context.DeadlineExceeded {
		return nil, ErrDialTimeout
	} else if err != nil {
		return nil, err
	}
	return d.cl.Enroll(conn)
}

func (d *GNetDialer) tcpAddrsClean() {
	expireDuration := 2 * d.DNSCacheDuration
	for {
		time.Sleep(time.Second)
		t := time.Now()
		d.tcpAddrsMap.Range(func(k, v interface{}) bool {
			if e, ok := v.(*tcpAddrEntry); ok && t.Sub(e.resolveTime) > expireDuration {
				d.tcpAddrsMap.Delete(k)
			}
			return true
		})

	}
}

func (d *GNetDialer) getTCPAddrs(addr string, dualStack bool) ([]net.TCPAddr, uint32, error) {
	item, exist := d.tcpAddrsMap.Load(addr)
	e, ok := item.(*tcpAddrEntry)
	if exist && ok && e != nil && time.Since(e.resolveTime) > d.DNSCacheDuration {
		// Only let one goroutine re-resolve at a time.
		if atomic.SwapInt32(&e.pending, 1) == 0 {
			e = nil
		}
	}

	if e == nil {
		addrs, err := resolveTCPAddrs(addr, dualStack, d.Resolver)
		if err != nil {
			item, exist := d.tcpAddrsMap.Load(addr)
			e, ok = item.(*tcpAddrEntry)
			if exist && ok && e != nil {
				// Set pending to 0 so another goroutine can retry.
				atomic.StoreInt32(&e.pending, 0)
			}
			return nil, 0, err
		}

		e = &tcpAddrEntry{
			addrs:       addrs,
			resolveTime: time.Now(),
		}
		d.tcpAddrsMap.Store(addr, e)
	}

	idx := atomic.AddUint32(&e.addrsIdx, 1)
	return e.addrs, idx, nil
}

func resolveTCPAddrs(addr string, dualStack bool, resolver *net.Resolver) ([]net.TCPAddr, error) {
	host, portS, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(portS)
	if err != nil {
		return nil, err
	}

	if resolver == nil {
		resolver = net.DefaultResolver
	}

	ctx := context.Background()
	ipaddrs, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}

	n := len(ipaddrs)
	addrs := make([]net.TCPAddr, 0, n)
	for i := 0; i < n; i++ {
		ip := ipaddrs[i]
		if !dualStack && ip.IP.To4() == nil {
			continue
		}
		addrs = append(addrs, net.TCPAddr{
			IP:   ip.IP,
			Port: port,
			Zone: ip.Zone,
		})
	}
	if len(addrs) == 0 {
		return nil, errNoDNSEntries
	}
	return addrs, nil
}

var errNoDNSEntries = errors.New("couldn't find DNS entries for the given domain. Try using DialDualStack")
