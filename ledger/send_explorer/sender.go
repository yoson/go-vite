package send_explorer

import (
	"github.com/Shopify/sarama"
	"github.com/gin-gonic/gin/json"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/log15"
	"os"
	"strings"
)

var senderLog = log15.New("module", "ledger/send_explorer/sender")

type Sender struct {
	producer sarama.AsyncProducer
	file     *os.File
}

type message struct {
	MsgType string `json:"type"`
	Data    string `json:"data"`
}

func (m *message) String() (string, error) {
	result, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

func NewSender(addr []string, fileName string) *Sender {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	producer, err := sarama.NewAsyncProducer(addr, config)

	if err != nil {
		senderLog.Crit(err.Error())
	}

	f, ofErr := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if ofErr != nil {
		senderLog.Crit(ofErr.Error())
	}
	return &Sender{
		producer: producer,
		file:     f,
	}
}

func (s *Sender) InsertAccountBlock(block *ledger.AccountBlock) string {
	data, _ := block.ToJson()
	message := &message{
		MsgType: "InsertAccountBlock",
		Data:    string(data),
	}

	result, _ := message.String()
	return result
}

func (s *Sender) InsertSnapshotBlock(block *ledger.SnapshotBlock) string {
	data, _ := block.ToJson()
	message := &message{
		MsgType: "InsertSnapshotBlock",
		Data:    string(data),
	}

	result, _ := message.String()
	return result
}

func (s *Sender) DeleteAccountBlocks(hashList []*types.Hash) string {
	hashStringList := make([]string, len(hashList))
	for index, hash := range hashList {
		hashStringList[index] = hash.String()
	}

	message := &message{
		MsgType: "DeleteAccountBlocks",
		Data:    strings.Join(hashStringList, ","),
	}

	result, _ := message.String()

	return result
}

func (s *Sender) DeleteSnapshotBlocks(hashList []*types.Hash) string {
	hashStringList := make([]string, len(hashList))
	for index, hash := range hashList {
		hashStringList[index] = hash.String()
	}

	message := &message{
		MsgType: "DeleteSnapshotBlocks",
		Data:    strings.Join(hashStringList, ","),
	}

	result, _ := message.String()

	return result
}
