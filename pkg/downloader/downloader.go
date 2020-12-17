package downloader

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/x0f5c3/manic-go/pkg/chunk"
)

type SumError struct {
	Reference string
	Data      string
}

type File struct {
	Url    string
	Data   []byte
	Sha    string
	Client *http.Client
	Length int
}

func New(url, sha string, client *http.Client) File {
	return File{
		Url:    url,
		Data:   nil,
		Sha:    sha,
		Client: client,
		Length: 0,
	}
}
func (c *SumError) Error() string {
	return fmt.Sprintf("Error!!! Sha256 mismatch\nReference: %v\nData: %v\n", c.Reference, c.Data)
}

func (c *File) GetLength() error {
	resp, err := c.Client.Head(c.Url)
	if err != nil {
		return err
	}
	rawString := resp.Header.Get("Content-Length")
	parsed, err := strconv.Atoi(rawString)
	if err != nil {
		return err
	}
	c.Length = parsed
	return nil

}
func (c *File) CompareSha() error {
	sum := sha256.Sum256(c.Data)
	sumString := string(sum[:])
	if sumString == c.Sha {
		return nil
	}
	return &SumError{
		Reference: c.Sha,
		Data:      sumString,
	}
}

func (c *File) Download(workers int) error {
	c.GetLength()
	chnk, err := chunk.New(0, c.Length-1, c.Length/workers)
	if err != nil {
		return err
	}
	for chnk.Next() {
		val := chnk.Get()
		go c.DownloadChunk(val)
	}
	return nil
}

func (c *File) DownloadChunk(val string) error {
	req, err := http.NewRequest("GET", c.Url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Range", val)
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	for _, n := range body {
		c.Data = append(c.Data, n)
	}
	return nil
}
