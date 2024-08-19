package socket

import (
	"encoding/json"
	"io"
)

func SendMessage(writer io.Writer, msg interface{}) (err error) {
	bt, err := json.Marshal(msg)
	return Write(writer, bt)
}

func Write(writer io.Writer, buf []byte) error {
	total := len(buf)
	for pos := 0; pos < total; {
		n, err := writer.Write(buf[pos:])
		if err != nil {
			return err
		}
		pos += n
	}
	return nil
}
