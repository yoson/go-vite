package mobile

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/vitelabs/go-vite/rpc"
)

type Client struct {
	c *rpc.Client
}

func Dial(rawurl string) (*Client, error) {
	return DialContext(context.Background(), rawurl)
}

func DialContext(ctx context.Context, rawurl string) (*Client, error) {
	c, err := rpc.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	return NewClient(c), nil
}

func NewClient(c *rpc.Client) *Client {
	return &Client{c}
}

func (vc *Client) Close() {
	vc.c.Close()
}

func (vc *Client) GetBlocksByAccAddr(addr string, index int, count int) (string, error) {
	return vc.rawMsgCall("ledger_getBlocksByAccAddr", addr, index, count)
}

func (vc *Client) GetBlocksByHash(addr string, startHash string, count int) (string, error) {
	if startHash == "" {
		return vc.rawMsgCall("ledger_getBlocksByHash", addr, nil, count)
	}
	return vc.rawMsgCall("ledger_getBlocksByHash", addr, startHash, count)

}

func (vc *Client) GetAccountByAccAddr(addr string) (string, error) {
	return vc.rawMsgCall("ledger_getAccountByAccAddr", addr)
}

func (vc *Client) GetOnroadAccountByAccAddr(addr string) (string, error) {
	return vc.rawMsgCall("onroad_getAccountOnroadInfo", addr)
}

func (vc *Client) GetOnroadBlocksByAddress(addr string, index int, count int) (string, error) {
	return vc.rawMsgCall("onroad_getOnroadBlocksByAddress", addr, index, count)
}

func (vc *Client) GetLatestBlock(addr string) (string, error) {
	return vc.rawMsgCall("ledger_getLatestBlock", addr)
}

func (vc *Client) GetPledgeData(addr string) (string, error) {
	return vc.stringCall("pledge_getPledgeData", addr)
}

func (vc *Client) GetPledgeQuota(addr string) (string, error) {
	return vc.rawMsgCall("pledge_getPledgeQuota", addr)
}

func (vc *Client) GetPledgeList(addr string, index int, count int) (string, error) {
	return vc.rawMsgCall("pledge_getPledgeList", addr, index, count)
}

func (vc *Client) GetFittestSnapshotHash(addr string, sendBlockHash string) (string, error) {
	if sendBlockHash == "" {
		return vc.stringCall("ledger_getFittestSnapshotHash", addr)
	}
	return vc.stringCall("ledger_getFittestSnapshotHash", addr, sendBlockHash)
}

func (vc *Client) GetPowNonce(difficulty string, data string) (string, error) {
	return vc.stringCall("pow_getPowNonce", difficulty, data)
}

func (vc *Client) GetSnapshotChainHeight() (string, error) {
	return vc.stringCall("ledger_getSnapshotChainHeight")
}

func (vc *Client) GetTokenMintage(tokenId string) (string, error) {
	return vc.rawMsgCall("ledger_getCancelVoteData", tokenId)
}

func (vc *Client) GetCandidateList(gid string) (string, error) {
	return vc.rawMsgCall("register_getCandidateList", gid)
}

func (vc *Client) GetVoteData(gid string, name string) (string, error) {
	return vc.stringCall("vote_getVoteData", gid, name)
}

func (vc *Client) GetVoteInfo(gid string, addr string) (string, error) {
	return vc.rawMsgCall("vote_getVoteInfo", gid, addr)
}

func (vc *Client) GetCancelVoteData(gid string) (string, error) {
	return vc.stringCall("vote_getCancelVoteData", gid)
}

func (vc *Client) CalcPoWDifficulty(accBlock string) (string, error) {
	var js json.RawMessage = []byte(accBlock)
	return vc.stringCall("tx_calcPoWDifficulty", js)
}

func (vc *Client) SendRawTx(accBlock string) error {
	var js json.RawMessage = []byte(accBlock)
	err := vc.c.Call(nil, "tx_sendRawTx", js)
	return makeJsonError(err)
}

func (vc *Client) rawMsgCall(method string, args ...interface{}) (string, error) {
	info := json.RawMessage{}
	err := vc.c.Call(&info, method, args...)
	if err != nil {
		return "", makeJsonError(err)
	}
	return string(info), nil
}

func (vc *Client) stringCall(method string, args ...interface{}) (string, error) {
	info := ""
	err := vc.c.Call(&info, method, args...)
	if err != nil {
		return "", makeJsonError(err)
	}
	return info, nil
}

func makeJsonError(err error) error {
	if err == nil {
		return nil
	}
	if jr, ok := err.(*rpc.JsonError); ok {
		bytes, _ := json.Marshal(jr)
		return errors.New(string(bytes))
	}
	return err
}
