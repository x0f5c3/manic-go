package chunk

import (
	"fmt"
)

type Chunks struct {
	low        int
	hi         int
	chunkSize  int
	next       string
	currOffset int
	currLength int
}

type Error struct {
	What string
}

type SingleChunk struct {
	Data   []byte
	Val    string
	Length int
	Offset int
}

func (e *Error) Error() string {
	return fmt.Sprintf("Chunk error: %v\n", e.What)
}

func New(low int, hi int, chunkSize int) (*Chunks, error) {
	var result *Chunks
	if chunkSize == 0 {
		return nil, &Error{
			What: "Chunk size cannot be 0",
		}
	}
	result = &Chunks{
		low:        low,
		hi:         hi,
		chunkSize:  chunkSize,
		next:       "",
		currLength: 0,
	}
	return result, nil

}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type CollectedChunks []SingleChunk

func (c *Chunks) Next() bool {
	if c.low > c.hi {
		c.next = ""
		return false
	} else {
		prevLow := c.low
		c.low += min(c.chunkSize, c.hi-c.low+1)
		c.next = fmt.Sprintf("bytes=%v-%v", prevLow, c.low)
		c.currOffset = prevLow
		c.currLength = c.low - prevLow
		return true
	}
}

func (c *Chunks) Collect() CollectedChunks {
	var res []SingleChunk
	for c.Next() {
		res = append(res, c.Get())
	}
	return res
}

func (c *Chunks) Get() SingleChunk {
	return SingleChunk{
		Val:    c.next,
		Length: c.currLength,
		Offset: c.currOffset,
	}
}
