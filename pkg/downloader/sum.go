package downloader

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/pterm/pterm"
)

type SumError struct {
	Reference string
	Data      string
}

func (c *SumError) Error() string {
	return fmt.Sprintf("Error!!! Sha256 mismatch\nReference: %v\nData: %v", c.Reference, c.Data)
}

type SumType int

const (
	Sha1 SumType = iota
	Sha224
	Sha256
	Sha384
	Sha512
	Sha512224
	Sha512256
	Md5
)

type CheckSum struct {
	Sum string
	SumType
}

func NewSum(sumType SumType, sum string) CheckSum {
	return CheckSum{
		Sum:     sum,
		SumType: sumType,
	}
}

func (c *CheckSum) Check(data []byte) error {
	pterm.Debug.Println(pterm.Bold.Sprint(pterm.FgMagenta.Sprint("Comparing SHA256 sums")))
	computed, err := func() ([]byte, error) {
		switch c.SumType {
		case Sha1:
			sum := sha1.Sum(data)
			return sum[:sha1.Size], nil
		case Sha224:
			sum := sha256.Sum224(data)
			return sum[:sha256.Size224], nil
		case Sha256:
			sum := sha256.Sum256(data)
			return sum[:sha256.Size], nil
		case Sha384:
			sum := sha512.Sum384(data)
			return sum[:sha512.Size384], nil
		case Sha512:
			sum := sha512.Sum512(data)
			return sum[:sha512.Size], nil
		case Sha512224:
			sum := sha512.Sum512_224(data)
			return sum[:sha512.Size224], nil
		case Sha512256:
			sum := sha512.Sum512_256(data)
			return sum[:sha512.Size256], nil
		case Md5:
			sum := md5.Sum(data)
			return sum[:md5.Size], nil
		default:
			return nil, errors.New("unsupported sum type")
		}
	}()
	if err != nil {
		return err
	}
	refstring := hex.EncodeToString(computed)
	byted, err := hex.DecodeString(c.Sum)
	if err != nil {
		return err
	}
	if bytes.Compare(computed, byted) == 0 {
		return nil
	}
	return &SumError{
		Reference: c.Sum,
		Data:      refstring,
	}

}
