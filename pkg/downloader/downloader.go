package downloader

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/go-multierror"
	"github.com/pterm/pterm"

	"github.com/x0f5c3/manic-go/pkg/chunk"
)

type FileNameError string

type ProxyFunc func(r *http.Request) (*url.URL, error)

func (e *FileNameError) Error() string {
	return fmt.Sprintf("Error: No filename in the url %s, you probably provided a url pointing to a directory, not a file", e)
}

type ToDownload struct {
	Url      string
	FileName string
	sum      *CheckSum
	finished *DownloadedFile
	Length   int
	Chunks   chunk.CollectedChunks
	bar      *pterm.ProgressbarPrinter
}

type Buffer struct {
	Chunks chunk.CollectedChunks
	Buf    []byte
	// ctx       context.Context
	// wg        *multierror.Group
	// chunkChan chan downloadResult
	// errFormat multierror.ErrorFormatFunc
	// pb        *pterm.ProgressbarPrinter
}

// func NewBuffer(ctx context.Context, chunks chunk.CollectedChunks, progress bool) *Buffer {
// 	wg := &multierror.Group{}
// 	var pb *pterm.ProgressbarPrinter
// 	var err error
// 	if progress {
// 		pb, err = pterm.DefaultProgressbar.WithTotal(len(chunks)).WithTitle("Writing chunks").Start()
// 		if err != nil {
// 			pterm.Error.Println(err)
// 			pb = nil
// 		}
// 	} else {
// 		pb = nil
// 	}
// 	chunkChan := make(chan downloadResult)
// 	errFormat := func(e []error) string {
// 		res := ""
// 		for _, v := range e {
// 			res += pterm.Error.Sprintln(v)
// 		}
// 		return res
// 	}
// 	buf := &Buffer{
// 		Chunks:    chunks,
// 		Buf:       make([]byte, 0),
// 		ctx:       ctx,
// 		chunkChan: chunkChan,
// 		errFormat: errFormat,
// 		wg:        wg,
// 		pb:        pb,
// 	}
// 	wg.Go(func() error {
// 		internalWg := &multierror.Group{}
// 		internalCtx, cancel := context.WithCancel(ctx)
// 		go func() {
// 			_ = internalWg.Wait()
// 			cancel()
// 		}()
// 		defer cancel()
// 		for {
// 			select {
// 			case <-internalCtx.Done():
// 				err := internalWg.Wait()
// 				err.ErrorFormat = buf.errFormat
// 				return err.ErrorOrNil()
// 			case res := <-chunkChan:
// 				if res.Err != nil {
// 					return res.Err
// 				}
// 				internalWg.Go(func() error {
// 					n, err := buf.WriteAt(res.Chunk.Data, int64(res.Chunk.Offset))
// 					if pb != nil {
// 						pb.Increment()
// 					}
// 					if err != nil {
// 						return err
// 					}
// 					if n < len(res.Chunk.Data) {
// 						return io.ErrShortWrite
// 					}
// 					return nil
// 				})
// 			}
// 		}
// 	})
// 	return buf
// }

func (b *Buffer) Bytes() []byte {
	return b.Buf
}

func (b *Buffer) WriteAt(p []byte, off int64) (n int, err error) {
	if len(b.Buf) < int(off)+len(p) || cap(b.Buf) < int(off)+len(p) {
		newBuf := make([]byte, int(off)+len(p))
		copy(newBuf, b.Buf)
		b.Buf = newBuf
	}
	n = copy(b.Buf[off:], p)
	if n < len(p) {
		err = io.ErrShortWrite
	}
	return
}

type File struct {
	Url       string
	FileName  string
	Sha       string
	Client    *Client
	Length    int
	Chunks    chunk.Chunks
	chunkChan chan chunk.SingleChunk
	bar       *pterm.ProgressbarPrinter
	queue     []ToDownload
}

type DownloadedFile struct {
	Url, FileName, Path string
	Data                *Buffer
	saved               bool
	sum                 *CheckSum
}

func (c *DownloadedFile) Verify() error {
	if c.saved && c.Path != "" && c.sum != nil {
		data, err := os.ReadFile(c.Path)
		if err != nil {
			return err
		}
		return c.sum.Check(data)
	} else if c.sum != nil {
		return c.sum.Check(c.Data.Bytes())
	} else {
		return nil
	}
}

func (c *DownloadedFile) Save(path string) error {
	bar, err := pterm.DefaultProgressbar.WithTotal(len(c.Data.Buf)).WithShowPercentage(true).WithShowCount(false).WithTitle("Saving file").Start()
	if err != nil {
		return err
	}
	fPath := filepath.Join(path, c.FileName)
	f, err := os.Create(fPath)
	if err != nil {
		return err
	}
	errFormat := func(e []error) string {
		res := ""
		for _, v := range e {
			res += pterm.Error.Sprintln(v)
		}
		return res
	}
	wg := &multierror.Group{}
	for _, v := range c.Data.Chunks {
		v := v
		wg.Go(func() error {
			n, err := f.WriteAt(v.Data, int64(v.Offset))
			if err != nil {
				return err
			}
			bar.Add(n)
			return nil
		})
	}
	result := wg.Wait()
	if result != nil {
		result.ErrorFormat = errFormat
		err = result.ErrorOrNil()
		if err != nil {
			return err
		}
	}
	_, err = bar.Stop()
	if err != nil {
		pterm.Error.Println(err)
	}
	c.saved = true
	c.Path = fPath
	return nil
}

func New(url string, sha string, client *Client, len *int) (*File, error) {
	if client == nil {
		client = NewClient()
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
	file := File{
		Url:      url,
		FileName: "",
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

func GetLength(url string, client *Client) (int, error) {
	resp, err := client.Head(url)
	if err != nil {
		return 0, err
	}
	rawString := resp.Header.ContentLength()
	if rawString == 0 {
		resp, err := client.GetRange(url, "bytes=0-0")
		if err != nil {
			return 0, err
		}
		rawString = resp.Header.ContentLength()
		if rawString == 0 {
			return 0, errors.New("can't retrieve length")
		}
	}
	return rawString, nil

}

func (c *File) GetFilename() error {
	u, err := url.Parse(c.Url)
	if err != nil {
		return err
	}
	fileName := path.Base(u.Path)
	if fileName == "" {
		fErr := FileNameError(c.Url)
		return &fErr
	}
	c.FileName = fileName
	return nil
}

func (c *File) DownloadChunk(chunk *chunk.SingleChunk) (*chunk.SingleChunk, error) {
	resp, err := c.Client.GetRange(c.Url, chunk.Val)
	defer ReleaseResponse(resp)
	if err != nil {
		return nil, err
	} else {
		// res, err := ioutil.ReadAll(&resp.body)
		// if err != nil {
		// 	return nil, err
		// } else {
		copy(chunk.Data, resp.Body())
		if c.bar != nil {
			c.bar.Add(len(chunk.Data))
		}
		return chunk, nil
	}
}

type downloadResult struct {
	Chunk *chunk.SingleChunk
	Err   error
}

func (c *File) downloadInner(workers, threads int) (*DownloadedFile, error) {
	runtime.GOMAXPROCS(threads)
	chnk, err := chunk.New(0, c.Length-1, c.Length/workers)
	if err != nil {
		return nil, err
	}
	collChunks := chnk.Collect()
	resChan := make(chan downloadResult, len(collChunks))
	wg := &multierror.Group{}
	for _, v := range collChunks {
		wg.Go(func() error {
			res, err := c.DownloadChunk(&v)
			if err != nil {
				return err
			}
			resChan <- downloadResult{
				Chunk: res,
				Err:   err,
			}
			return nil
		})
	}

	err2 := wg.Wait().ErrorOrNil()
	if err2 != nil {
		return nil, err2
	}
	close(resChan)
	if c.bar != nil {
		_, err = c.bar.Stop()
		if err != nil {
			return nil, err
		}
	}
	resBuf := &Buffer{Chunks: collChunks, Buf: make([]byte, c.Length+1)}
	for v := range resChan {
		if v.Err != nil {
			return nil, err
		}
		startPos := v.Chunk.Offset
		_, err := resBuf.WriteAt(v.Chunk.Data, int64(startPos))
		if err != nil {
			return nil, err
		}
	}
	var sum *CheckSum
	if c.Sha != "" {
		sum = &CheckSum{
			Sum:     c.Sha,
			SumType: Sha256,
		}
	} else {
		sum = nil
	}
	res := DownloadedFile{
		Url:      c.Url,
		FileName: c.FileName,
		sum:      sum,
		Data:     resBuf,
	}
	return &res, nil
}

func (c *File) DownloadWithProgress(workers, threads int) (*DownloadedFile, error) {
	name := pterm.Blink.Sprint(pterm.FgMagenta.Sprintf("Downloading %s", c.FileName))
	bar, err := pterm.DefaultProgressbar.WithTotal(c.Length).WithShowCount(false).WithShowPercentage(true).WithTitle(name).Start()
	if err != nil {
		return nil, err
	}
	c.bar = bar
	res, err := c.downloadInner(workers, threads)
	if err != nil {
		return nil, err
	}
	// _, err = c.bar.Stop()
	// if err != nil {
	// 	pterm.Error.Println(err)
	// }
	return res, nil
}
func (c *File) Download(workers, threads int, progress bool) (*DownloadedFile, error) {
	var err error
	var res *DownloadedFile
	if progress {
		res, err = c.DownloadWithProgress(workers, threads)
	} else {
		res, err = c.downloadInner(workers, threads)
	}
	if err != nil {
		return nil, err
	}
	if res.sum != nil {
		shaErr := res.Verify()
		if shaErr != nil {
			return nil, shaErr
		}
	}
	return res, nil
}
