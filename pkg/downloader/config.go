package downloader

import (
	"net/url"
	"time"
)

type Config struct {
	Client   *Client
	Length   int
	progress bool
	sum      *CheckSum
}

func (c *Config) Sum(sum *CheckSum) *Config {
	c.sum = sum
	return c
}

func NewConfig() *Config {
	return &Config{Client: DefaultClient}
}

func (c *Config) WithProxy(proxy *url.URL) *Config {
	if c.Client == nil {
		c.Client = DefaultClient
	}
	c.Client.cl.Dial = ProxyDialer(proxy.String(), time.Second*10)
	return c
}

func (c *Config) WithTimeout(timeout time.Duration) *Config {
	if c.Client == nil {
		c.Client = DefaultClient
	}
	c.Client.cl.WriteTimeout = timeout
	c.Client.cl.ReadTimeout = timeout
	c.Client.cl.MaxConnWaitTimeout = timeout
	return c
}

func (c *Config) Progress(progress bool) *Config {
	c.progress = progress
	return c
}
