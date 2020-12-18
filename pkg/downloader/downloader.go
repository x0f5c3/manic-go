package downloader

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"bytes"
	"encoding/hex"
	"net/url"
	"sync"

	"github.com/i582/cfmt"
	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
	"github.com/x0f5c3/manic-go/pkg/chunk"
	"path"
	"runtime"
)

type SumError struct {
	Reference []byte
	Data      []byte
}

type FileNameError struct{}

func (e *FileNameError) Error() string {
	return cfmt.Sprintln("{{Error: No filename in the url, you probably provided a url pointing to a directory, not a file}}::red|blink")
}

type ProgressWait struct {
	bar      []*mpb.Bar
	progress *mpb.Progress
}

type File struct {
	Url      string
	FileName string
	Data     *[]byte
	Sha      string
	Client   *http.Client
	Length   int
}

func New(url, sha string, client *http.Client) (*File, error) {
	data := make([]byte, 0)
	file := File{
		Url:      url,
		FileName: "",
		Data:     &data,
		Sha:      sha,
		Client:   client,
		Length:   0,
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
	cfmt.Println("{{Comparing SHA256 sums}}::magenta|bold")
	sum := sha256.Sum256(*c.Data)
	byted, err := hex.DecodeString(c.Sha)
	if err != nil {
		return err
	}
	if bytes.Compare(sum[:32], byted) == 0 {
		cfmt.Printf("{{Successfully downloaded file: %s\n}}::green|bold", c.FileName)
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

func aggregate(ch chan chunk.Chunk, multi []chan chunk.Chunk, wg *sync.WaitGroup, pb ...*ProgressWait) {
	defer close(ch)
	for _, c := range multi {
		dat := <-c
		ch <- dat
	}
	pb[0].progress.Wait()
}

func (c *File) startWorkers(chans []chan chunk.Chunk, chnk chunk.ChunkIter, progress bool, wg *sync.WaitGroup, pb ...*ProgressWait) {
	final := make(chan chunk.Chunk)
	for _, ch := range chans {
		some := chnk.Next()
		if some {
			off, val := chnk.Get()
			length := chnk.GetLength()
			if progress && len(pb) != 0 {
				wg.Add(1)
				go c.DownloadChunkProgress(val, ch, off, length, wg, pb[0].bar[0])
			} else {
				go c.DownloadChunk(val, ch, off)
			}
		}
	}
	go aggregate(final, chans, wg, pb[0])
	c.dataPut(final, pb[0].bar[1])

}

func (c *File) dataPut(ch chan chunk.Chunk, pb ...*mpb.Bar) {
	for dat := range ch {
		startPos := dat.Offset
		for _, val := range dat.Data {
			(*c.Data)[startPos] = val
			startPos++
			pb[0].Increment()
		}
	}
	cfmt.Printf("{{Downloaded %v MB\n}}::green|blink", c.Length/1000000)
}

func (c *File) Download(workers, threads int) error {
	c.GetLength()
	var wg sync.WaitGroup
	var chans []chan chunk.Chunk
	if c.Length%workers == 0 {
		chans = makeChannels(workers)
	} else {
		chans = makeChannels(workers + 1)
	}
	dat := make([]byte, c.Length)
	c.Data = &dat
	runtime.GOMAXPROCS(threads)
	p := mpb.New(mpb.WithWidth(64))

	name := cfmt.Sprintf("{{Downloading %s}}::magenta|blink", c.FileName)
	bar := p.AddBar(int64(c.Length),
		mpb.BarStyle("╢▌▌░╟"),
		mpb.PrependDecorators(
			decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
		),
		mpb.AppendDecorators(
			decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: 4}),
			decor.AverageSpeed(decor.UnitKB, "  % .2f  ", decor.WC{W: 5}),
			decor.Percentage(),
		),
	)
	nameCopy := cfmt.Sprint("{{Copying}}::magenta|blink")
	copyBar := p.AddBar(int64(c.Length),
		mpb.BarStyle("╢▌▌░╟"),
		mpb.PrependDecorators(
			decor.Name(nameCopy, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
		),
		mpb.AppendDecorators(
			decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: 4}),
			decor.Percentage(),
		),
	)

	pbWait := ProgressWait{
		bar:      []*mpb.Bar{bar, copyBar},
		progress: p,
	}
	chnk, err := chunk.New(0, c.Length-1, c.Length/workers, c.Length)
	if err != nil {
		return err
	}

	c.startWorkers(chans, chnk, true, &wg, &pbWait)
	wg.Wait()
	return nil
}

func (c *File) DownloadAndVerify(workers, threads int) error {
	err := c.Download(workers, threads)
	if err != nil {
		return err
	}
	shaErr := c.CompareSha()
	if shaErr != nil {
		return shaErr
	}
	return nil
}

func (c *File) DownloadChunkProgress(val string, ch chan chunk.Chunk, off int, length int, wg *sync.WaitGroup, pb ...*mpb.Bar) error {
	// defer close(ch)
	defer wg.Done()
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
	bodyReader := pb[0].ProxyReader(resp.Body)
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return err
	}
	chnk := chunk.Chunk{
		Data:   body,
		Offset: off,
		Length: length,
	}
	ch <- chnk
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
