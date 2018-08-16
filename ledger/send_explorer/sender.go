package send_explorer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/log15"
	"gopkg.in/Shopify/sarama.v1"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var senderLog = log15.New("module", "ledger/send_explorer/sender")

const MSG_VERSION = "0.1"
const NUM_PER_FILE = 10000

func getFileIndex(total *big.Int) *big.Int {
	fileIndex := big.NewInt(0)
	fileIndex.Div(total, big.NewInt(NUM_PER_FILE))
	return fileIndex
}

type Sender struct {
	producer sarama.AsyncProducer
	dirname  string

	file *os.File
	meta *meta

	topic string

	needSendMessageList []string
	needSendMessageLock sync.Mutex

	writeLock sync.Mutex
}

// [Fixme]
var _sender *Sender

func GetSender(addr []string, dirname string, topic string) *Sender {
	if _sender == nil {
		_sender = NewSender(addr, dirname, topic)
	}

	return _sender
}
func NewSender(addr []string, dirname string, topic string) *Sender {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	producer, err := sarama.NewAsyncProducer(addr, config)

	if err != nil {
		senderLog.Crit(err.Error())
	}

	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		os.Mkdir(dirname, 0755)
	}

	s := &Sender{
		producer: producer,
		dirname:  dirname,
		meta:     NewMeta(dirname),

		topic: topic,
	}

	s.setFile(s.meta.currentFileIndex)

	s.startSender()
	return s
}
func (s *Sender) setFile(fileIndex *big.Int) {
	fileName := filepath.Join(s.dirname, "message."+fileIndex.String()+".log")
	f, ofErr := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0644)
	if ofErr != nil {
		senderLog.Crit(ofErr.Error())
	}

	s.meta.currentFileIndex = fileIndex
	s.file = f
}

func (s *Sender) resetFile(total *big.Int) {
	fileIndex := getFileIndex(total)
	if !bytes.Equal(fileIndex.Bytes(), s.meta.currentFileIndex.Bytes()) {
		s.setFile(fileIndex)
	}
}

func (s *Sender) writeMessage(message string) error {
	newTotal := &big.Int{}
	newTotal.Add(s.meta.total, big.NewInt(1))

	// try reset file
	s.resetFile(newTotal)
	_, err := s.file.WriteString(message)
	if err != nil {
		senderLog.Error(err.Error(), "func", "sender.WriteMessage")
		return err
	}

	s.meta.total = newTotal
	fErr := s.meta.flush()

	if fErr != nil {
		return fErr
	}

	return nil
}

func (s *Sender) appendNeedSend(message string) {
	s.needSendMessageLock.Lock()
	gap := &big.Int{}
	gap = gap.Sub(s.meta.total, s.meta.hasRead)

	if gap.Cmp(big.NewInt(1)) <= 0 {
		s.needSendMessageList = append(s.needSendMessageList, message)
	}

	s.meta.hasRead = s.meta.hasRead.Add(s.meta.hasRead, big.NewInt(1))
	s.needSendMessageLock.Unlock()
}

func (s *Sender) startSender() {
	senderLog.Info("start sender")
	go func() {
		for {
			s.needSendMessageLock.Lock()
			if len(s.needSendMessageList) <= 0 {
				if s.meta.total.Cmp(s.meta.hasRead) > 0 {
					s.readNextPage()
					s.needSendMessageLock.Unlock()
				} else {
					s.needSendMessageLock.Unlock()
					time.Sleep(2000 * time.Millisecond)
				}
				continue
			}

			needSendMessage := s.needSendMessageList[0]
			s.needSendMessageLock.Unlock()

			message := &sarama.ProducerMessage{Topic: s.topic, Value: sarama.StringEncoder(needSendMessage)}

			s.producer.Input() <- message
			select {
			// success
			case <-s.producer.Successes():
				s.needSendMessageLock.Lock()

				s.needSendMessageList = s.needSendMessageList[1:]
				s.meta.offset.Add(s.meta.offset, big.NewInt(1))
				s.meta.flush()

				s.needSendMessageLock.Unlock()

			// error
			case sendError := <-s.producer.Errors():
				senderLog.Error(sendError.Error())
				time.Sleep(2000 * time.Millisecond)
			}
		}
	}()
}

func (s *Sender) readNextPage() error {
	fileIndex := getFileIndex(s.meta.offset)
	fileName := filepath.Join(s.dirname, "message."+fileIndex.String()+".log")

	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	contentStr := string(content)
	messageList := strings.Split(contentStr, "\n")
	messageListLen := len(messageList)
	messageList = messageList[:messageListLen]

	currentFileOffset := &big.Int{}
	currentFileOffset.Mod(s.meta.offset, big.NewInt(NUM_PER_FILE))
	messageList = messageList[currentFileOffset.Int64():]

	s.meta.hasRead = &big.Int{}
	s.meta.hasRead.Add(s.meta.offset, big.NewInt(int64(len(messageList))))

	s.needSendMessageList = messageList
	return nil
}

func (s *Sender) pushMessage(msg *message) error {
	s.writeLock.Lock()
	defer s.writeLock.Unlock()

	msg.EventId = &big.Int{}
	msg.EventId.Add(s.meta.total, big.NewInt(1))

	result, strErr := msg.ToJson()

	if strErr != nil {
		senderLog.Error(strErr.Error())
		return strErr
	}

	messageStr := fmt.Sprintf("%s\n", result)
	wErr := s.writeMessage(messageStr)

	if wErr != nil {
		senderLog.Error(wErr.Error())
		return wErr
	}

	s.appendNeedSend(messageStr)

	return nil
}

func (s *Sender) InsertAccountBlock(block *ledger.AccountBlock) error {
	senderLog.Info("InsertAccountBlock")

	data, tjErr := block.ToJson(true)
	if tjErr != nil {
		senderLog.Error(tjErr.Error())
		return tjErr
	}

	return s.pushMessage(&message{
		MsgType: "InsertAccountBlock",
		Data:    string(data),
		Version: MSG_VERSION,
	})
}

func (s *Sender) InsertSnapshotBlock(block *ledger.SnapshotBlock) error {
	senderLog.Info("InsertSnapshotBlock")
	data, tjErr := block.ToJson()
	if tjErr != nil {
		senderLog.Error(tjErr.Error())
		return tjErr
	}

	return s.pushMessage(&message{
		MsgType: "InsertSnapshotBlock",
		Data:    string(data),
		Version: MSG_VERSION,
	})
}

func (s *Sender) DeleteAccountBlocks(hashList []*types.Hash) error {
	senderLog.Info("DeleteAccountBlocks")

	data := struct {
		HashList []string `json:"hashList"`
	}{
		HashList: make([]string, len(hashList)),
	}

	for index, hash := range hashList {
		data.HashList[index] = hash.String()
	}

	dataJson, jsonErr := json.Marshal(data)
	if jsonErr != nil {
		senderLog.Error(jsonErr.Error())
		return jsonErr
	}

	return s.pushMessage(&message{
		MsgType: "DeleteAccountBlocks",
		Data:    string(dataJson),
		Version: MSG_VERSION,
	})
}

func (s *Sender) DeleteSnapshotBlocks(hashList []*types.Hash) error {
	senderLog.Info("DeleteSnapshotBlocks")

	data := struct {
		HashList []string `json:"hashList"`
	}{
		HashList: make([]string, len(hashList)),
	}

	for index, hash := range hashList {
		data.HashList[index] = hash.String()
	}

	dataJson, jsonErr := json.Marshal(data)
	if jsonErr != nil {
		senderLog.Error(jsonErr.Error())
		return jsonErr
	}

	return s.pushMessage(&message{
		MsgType: "DeleteSnapshotBlocks",
		Data:    string(dataJson),
		Version: MSG_VERSION,
	})
}
