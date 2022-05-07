package downloader

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/pterm/pterm"
	"github.com/reugn/async"
	"github.com/x0f5c3/manic-go/pkg/chunk"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"strconv"
)

type SumError struct {
	Reference string
	Data      string
}

type FileNameError struct{}

func (e *FileNameError) Error() string {
	return fmt.Sprintln("Error: No filename in the url, you probably provided a url pointing to a directory, not a file")
}

type File struct {
	Url       string
	FileName  string
	Data      *[]byte
	Sha       string
	Client    *http.Client
	Length    int
	Chunks    chunk.Chunks
	chunkChan chan chunk.SingleChunk
	bar       *pterm.ProgressbarPrinter
}

func New(url string, sha string, client *http.Client, len *int) (*File, error) {
	if client == nil {
		client = http.DefaultClient
	}
	var length int
	if len == nil {
		var err error
		length, err = GetLength(url, client)
		if err != nil {
			return nil, err
		}
	} else {
		length = *len
	}
	data := make([]byte, length)
	file := File{
		Url:      url,
		FileName: "",
		Data:     &data,
		Sha:      sha,
		Client:   client,
		Length:   length,
	}
	err := file.GetFilename()
	if err != nil {
		return nil, err
	}
	return &file, nil
}
func (c *SumError) Error() string {
	return fmt.Sprintf("Error!!! Sha256 mismatch\nReference: %v\nData: %v\n", c.Reference, c.Data)
}

func (c *File) Save(path string) error {
	return ioutil.WriteFile(path, *c.Data, 0644)
}
func GetLength(url string, client *http.Client) (int, error) {
	resp, err := client.Head(url)
	if err != nil {
		return 0, err
	}
	rawString := resp.Header.Get("Content-Length")
	parsed, err := strconv.Atoi(rawString)
	if err != nil {
		return 0, err
	}
	return parsed, nil

}

func (c *File) GetFilename() error {
	u, err := url.Parse(c.Url)
	if err != nil {
		return err
	}
	fileName := path.Base(u.Path)
	if fileName == "" {
		return &FileNameError{}
	}
	c.FileName = fileName
	return nil
}

func (c *File) CompareSha() error {
	pterm.Debug.Println(pterm.Bold.Sprint(pterm.FgMagenta.Sprint("Comparing SHA256 sums")))
	sum := sha256.Sum256(*c.Data)
	byted, err := hex.DecodeString(c.Sha)
	refstring := hex.EncodeToString(sum[:32])
	if err != nil {
		return err
	}
	if bytes.Compare(sum[:32], byted[:32]) == 0 {
		pterm.Success.Println(pterm.Bold.Sprint(pterm.FgGreen.Sprintf("Successfully downloaded file: %s\n", c.FileName)))
		return nil
	}
	return &SumError{
		Reference: c.Sha,
		Data:      refstring,
	}
}

func (c *File) DownloadChunk(chunk chunk.SingleChunk) async.Future {
	promise := async.NewPromise()
	go func() {
		req, err := http.NewRequest("GET", c.Url, nil)
		if err != nil {
			promise.Failure(err)
		} else {
			req.Header.Add("RANGE", chunk.Val)
			resp, err := c.Client.Do(req)
			if err != nil {
				promise.Failure(err)
			} else {
				if c.bar != nil {
					res, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						promise.Failure(err)
					} else {
						chunk.Data = res
						c.bar.Add(len(res))
						promise.Success(chunk)
					}
				} else {
					res, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						promise.Failure(err)
					} else {
						chunk.Data = res
						promise.Success(chunk)
					}
				}
			}
		}
	}()
	return promise.Future()
}

func (c *File) downloadInner(workers, threads int) error {
	runtime.GOMAXPROCS(threads)
	chnk, err := chunk.New(0, c.Length-1, c.Length/workers)
	var promises []async.Future
	if err != nil {
		return err
	}
	for chnk.Next() {
		next := chnk.Get()
		promises = append(promises, c.DownloadChunk(next))
	}
	for _, fut := range promises {
		res, err := fut.Get()
		if err != nil {
			return err
		}
		convert := res.(chunk.SingleChunk)
		startPos := convert.Offset
		for _, dat := range convert.Data {
			(*c.Data)[startPos] = dat
			startPos++
		}
	}
	return nil
}

func (c *File) DownloadWithProgress(workers, threads int) error {
	name := pterm.Blink.Sprint(pterm.FgMagenta.Sprintf("Downloading %s", c.FileName))
	bar, err := pterm.DefaultProgressbar.WithTotal(c.Length).WithTitle(name).Start()
	if err != nil {
		return err
	}
	c.bar = bar
	err = c.downloadInner(workers, threads)
	if err != nil {
		return err
	}
	_, err = c.bar.Stop()
	return err
}
func (c *File) Download(workers, threads int, progress bool) error {
	var err error
	if progress {
		err = c.DownloadWithProgress(workers, threads)
	} else {
		err = c.downloadInner(workers, threads)
	}
	if err != nil {
		return err
	}
	if c.Sha != "" {
		shaErr := c.CompareSha()
		if shaErr != nil {
			return shaErr
		}
	}
	return nil
}
