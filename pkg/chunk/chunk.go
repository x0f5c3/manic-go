package chunk

import (
	"fmt"
)

type ChunkIter struct {
	low        int
	hi         int
	chunkSize  int
	next       string
	currOffset int
}

type Chunk struct {
	Data   []byte
	Offset int
}
type ChunkError struct {
	What string
}

func (e ChunkError) Error() string {
	return fmt.Sprintf("Chunk error: %v\n", e.What)
}

func New(low int, hi int, chunkSize int) (ChunkIter, error) {
	var result *ChunkIter
	if chunkSize == 0 {
		return *result, ChunkError{
			What: "Chunk size cannot be 0",
		}
	}
	result = &ChunkIter{
		low:       low,
		hi:        hi,
		chunkSize: chunkSize,
		next:      "",
	}
	return *result, nil

}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func (c *ChunkIter) Next() bool {
	if c.low > c.hi {
		c.next = ""
		return false
	} else if c.low == c.hi {
		c.next = fmt.Sprintf("bytes=%v-", c.hi)
		return true
	} else {
		prev_low := c.low
		c.low += min(c.chunkSize, c.hi-c.low+1)
		c.next = fmt.Sprintf("bytes=%v-%v", prev_low, c.low-1)
		c.currOffset = prev_low
		return true
	}
}
func (c *ChunkIter) Get() (int, string) {
	return c.currOffset, c.next
}
