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

type ChunkError struct {
	What string
}

type SingleChunk struct {
	Data []byte
	Val string
	Offset int
}
func (e ChunkError) Error() string {
	return fmt.Sprintf("Chunk error: %v\n", e.What)
}

func New(low int, hi int, chunkSize int) (*Chunks, error) {
	var result *Chunks
	if chunkSize == 0 {
		return nil, ChunkError{
			What: "Chunk size cannot be 0",
		}
	}
	result = &Chunks{
		low:       low,
		hi:        hi,
		chunkSize: chunkSize,
		next:      "",
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
func (c *Chunks) Next() bool {
	if c.low > c.hi {
		c.next = ""
		return false
	} else {
		prev_low := c.low
		c.low += min(c.chunkSize, c.hi-c.low+1)
		c.next = fmt.Sprintf("bytes=%v-%v", prev_low, c.low)
		c.currOffset = prev_low
		c.currLength = c.low - prev_low
		return true
	}
}
func (c *Chunks) Get() (int, int, string) {
	return c.currOffset, c.currLength, c.next
}
