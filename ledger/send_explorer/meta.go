package send_explorer

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"sync"
)

type meta struct {
	offset  *big.Int
	total   *big.Int
	hasRead *big.Int

	currentFileIndex *big.Int
	filename         string
	file             *os.File

	lock sync.Mutex
}

func NewMeta(dirname string) *meta {
	metaFilename := filepath.Join(dirname, "meta")
	m := &meta{
		filename: metaFilename,
	}

	var f *os.File
	var ofErr error
	if _, err := os.Stat(metaFilename); os.IsNotExist(err) {
		f, ofErr = os.Create(metaFilename)
	} else {
		f, ofErr = os.OpenFile(metaFilename, os.O_RDWR|os.O_SYNC, 0644)
	}

	if ofErr != nil {
		senderLog.Crit(ofErr.Error(), "func", "meta.NewMeta")
		return nil
	}

	m.file = f
	if rErr := m.readFromDisk(); rErr != nil {
		senderLog.Crit(rErr.Error(), "func", "meta.NewMeta")
		return nil
	}

	return m
}
func (m *meta) RecoverMeta() error {
	return nil
}

func (m *meta) flush() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	jsonMap := make(map[string]string)
	jsonMap["offset"] = m.offset.String()
	jsonMap["total"] = m.total.String()
	jsonBytes, err := json.Marshal(jsonMap)

	if err != nil {
		senderLog.Error(err.Error(), "func", "meta.flush")
		return err
	}

	m.file.Truncate(0)
	m.file.Seek(0, 0)
	m.file.Write(jsonBytes)

	return nil
}

func (m *meta) readFromDisk() error {
	jsonMap := make(map[string]string)

	content, err := ioutil.ReadFile(m.filename)

	if err != nil {
		if os.IsNotExist(err) {
			jsonMap["offset"] = "0"
			jsonMap["hasRead"] = "0"
			jsonMap["total"] = "0"
		} else {
			senderLog.Error(err.Error(), "func", "meta.ReadFromDisk")
			return err
		}
	} else {
		json.Unmarshal([]byte(content), &jsonMap)

	}

	m.offset = &big.Int{}
	m.offset.SetString(jsonMap["offset"], 10)

	m.hasRead = &big.Int{}
	m.hasRead.SetString(jsonMap["offset"], 10)

	m.total = &big.Int{}
	m.total.SetString(jsonMap["total"], 10)

	m.currentFileIndex = getFileIndex(m.total)

	return nil
}
