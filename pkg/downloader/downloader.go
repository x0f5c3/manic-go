package downloader

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"bytes"
	"encoding/hex"

	"github.com/x0f5c3/manic-go/pkg/chunk"
)

type SumError struct {
	Reference []byte
	Data      []byte
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
		Data:   make([]byte, 0),
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
	byted, err := hex.DecodeString(c.Sha)
	if err != nil {
		return nil
	}
	if bytes.Compare(sum[:32], byted) == 0 {
		return nil
	}
	fmt.Println("Len:", len(byted))
	return &SumError{
		Reference: byted,
		Data:      sum[:32],
	}
}

func makeChannels(count int) []chan chunk.Chunk {
	var res []chan chunk.Chunk
	for i := 0; i < count; i++ {
		ch := make(chan chunk.Chunk)
		res = append(res, ch)
	}
	return res
}
func (c *File) Download(workers int) error {
	c.GetLength()
	res := make([]byte, c.Length)
	chnk, err := chunk.New(0, c.Length-1, c.Length/workers)
	if err != nil {
		return err
	}
	var chans []chan chunk.Chunk
	if c.Length%workers == 0 {
		chans = makeChannels(workers)
	} else {
		chans = makeChannels(workers + 1)
	}
	for _, ch := range chans {
		some := chnk.Next()
		if some {
			off, val := chnk.Get()
			go c.DownloadChunk(val, ch, off)
		}
	}
	arrArr := make(map[int]chunk.Chunk)
	for i, ch := range chans {
		for data := range ch {
			arrArr[i] = data
		}
	}
	for _, val := range arrArr {
		startPos := val.Offset
		for _, dat := range val.Data {
			res[startPos] = dat
			startPos++
		}
	}
	for _, final := range res {
		c.Data = append(c.Data, final)
	}

	return nil
}
func (c *File) DownloadAndVerify(workers int) error {
	err := c.Download(workers)
	if err != nil {
		return err
	}
	shaErr := c.CompareSha()
	if shaErr != nil {
		return shaErr
	}
	return nil
}

func (c *File) DownloadChunk(val string, ch chan chunk.Chunk, off int) error {
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
	chnk := chunk.Chunk{
		Data:   body,
		Offset: off,
	}
	ch <- chnk
	close(ch)
	return nil
}
