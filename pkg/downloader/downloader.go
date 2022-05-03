package downloader

import (
	"crypto/sha256"
	"fmt"
	"github.com/vbauerster/mpb/v7"
	"io/ioutil"
	"net/http"
	"strconv"

	"bytes"
	"encoding/hex"
	"github.com/i582/cfmt/cmd/cfmt"
	"github.com/reugn/async"
	"github.com/vbauerster/mpb/v7/decor"
	"github.com/x0f5c3/manic-go/pkg/chunk"
	"net/url"
	"path"
	"runtime"
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
	Url      string
	FileName string
	Data     *[]byte
	Sha      string
	Client   *http.Client
	Length   int
	Chunks   chunk.Chunks
	bar      *mpb.Bar
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
	_, _ = cfmt.Println("{{Comparing SHA256 sums}}::magenta|bold")
	sum := sha256.Sum256(*c.Data)
	byted, err := hex.DecodeString(c.Sha)
	refstring := hex.EncodeToString(sum[:32])
	if err != nil {
		return err
	}
	if bytes.Compare(sum[:32], byted[:32]) == 0 {
		_, _ = cfmt.Printf("{{Successfully downloaded file: %s\n}}::green|bold", c.FileName)
		return nil
	}
	fmt.Println("Len:", len(byted))
	return &SumError{
		Reference: c.Sha,
		Data:      refstring,
	}
}

func (c *File) DownloadChunk(chunk chunk.SingleChunk, bar *mpb.Bar) async.Future {
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
					if bar != nil {
						reader := bar.ProxyReader(resp.Body)
						res, err := ioutil.ReadAll(reader)
						if err != nil {
							promise.Failure(err)
						} else {
							chunk.Data = res
							c.bar.IncrBy(len(res))
							promise.Success(chunk)
						}
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

func (c *File) downloadInner(workers, threads int, progress *mpb.Progress) error {
	runtime.GOMAXPROCS(threads)
	chnk, err := chunk.New(0, c.Length-1, c.Length/workers)
	var promises []async.Future
	if err != nil {
		return err
	}
	for chnk.Next() {
		next := chnk.Get()
		if progress != nil {
			name := fmt.Sprintf("Chunk offset: %d", next.Offset)
			bar := progress.New(int64(next.Length),
				mpb.BarStyle().Lbound("╢").Filler(cfmt.Sprintf("{{▌}}::green")).Tip(cfmt.Sprintf("{{▌}}::green")).Padding("░").Rbound("╟"),
				mpb.PrependDecorators(
					decor.Name(name, decor.WC{W: 30, C: decor.DidentRight}),
				),
				mpb.AppendDecorators(
					decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: 4}),
					decor.AverageSpeed(decor.UnitKB, "  % .2f  ", decor.WC{W: 5}),
					decor.Percentage(),
				),
			)
			promises = append(promises, c.DownloadChunk(next, bar))
		} else {
			promises = append(promises, c.DownloadChunk(next, nil))
		}
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
	name := cfmt.Sprintf("{{Downloading %s}}::magenta|blink", c.FileName)
	p := mpb.New(mpb.WithWidth(64))
	bar := p.New(int64(c.Length),
		mpb.BarStyle().Lbound("╢").Filler(cfmt.Sprintf("{{▌}}::green")).Tip(cfmt.Sprintf("{{▌}}::green")).Padding("░").Rbound("╟"),
		mpb.PrependDecorators(
			decor.Name(name, decor.WC{W: 30, C: decor.DidentRight}),
		),
		mpb.AppendDecorators(
			decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: 4}),
			decor.AverageSpeed(decor.UnitKB, "  % .2f  ", decor.WC{W: 5}),
			decor.Percentage(),
		),
	)
	c.bar = bar
	err := c.downloadInner(workers, threads, p)
	if err != nil {
		return err
	}
	return nil
}
func (c *File) Download(workers, threads int, progress bool) error {
	var err error
	if progress {
		err = c.DownloadWithProgress(workers, threads)
	} else {
		err = c.downloadInner(workers, threads, nil)
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
