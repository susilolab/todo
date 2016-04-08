package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func IntToBytes(id int64) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, id)
	if err != nil {
		return nil, fmt.Errorf("intToByte: %s", err)
	}
	return buf.Bytes(), nil
}
