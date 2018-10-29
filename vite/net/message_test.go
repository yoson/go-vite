package net

import (
	"fmt"
	"github.com/vitelabs/go-vite/chain"
	"github.com/vitelabs/go-vite/common"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/config"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/p2p"
	"github.com/vitelabs/go-vite/vite/net/message"
	"github.com/vitelabs/go-vite/vm_context"
	"math/big"
	"math/rand"
	net2 "net"
	"path/filepath"
	"testing"
	"time"
)

// mock chain
var chn chain.Chain

var accounts = make(map[types.Address][]types.Hash, 1000)

func getAccountList() []types.Address {
	list := make([]types.Address, len(accounts))
	i := 0
	for act := range accounts {
		list[i] = act
		i++
	}

	return list
}

func getChain() chain.Chain {
	if chn == nil {
		chn = chain.NewChain(&config.Config{
			DataDir: filepath.Join(common.HomeDir(), "Library/GVite/test"),
		})

		chn.Init()
		chn.Start()

		to := rand.Intn(1000)
		mockBlocks(chn, uint64(to))
	}

	return chn
}

func mockAccountBlock(chn chain.Chain, snapshotBlockHash types.Hash, addr1 *types.Address, addr2 *types.Address) ([]*vm_context.VmAccountBlock, []types.Address, error) {
	now := time.Now()

	if addr1 == nil {
		accountAddress, _, _ := types.CreateAddress()
		addr1 = &accountAddress
	}
	if addr2 == nil {
		accountAddress, _, _ := types.CreateAddress()
		addr2 = &accountAddress
	}

	vmContext, err := vm_context.NewVmContext(chn, nil, nil, addr1)
	if err != nil {
		return nil, nil, err
	}
	latestBlock, _ := chn.GetLatestAccountBlock(addr1)
	nextHeight := uint64(1)
	var prevHash types.Hash
	if latestBlock != nil {
		nextHeight = latestBlock.Height + 1
		prevHash = latestBlock.Hash
	}

	sendAmount := new(big.Int).Mul(big.NewInt(100), big.NewInt(1e9))
	var sendBlock = &ledger.AccountBlock{
		PrevHash:       prevHash,
		BlockType:      ledger.BlockTypeSendCall,
		AccountAddress: *addr1,
		ToAddress:      *addr2,
		Amount:         sendAmount,
		TokenId:        ledger.ViteTokenId,
		Height:         nextHeight,
		Fee:            big.NewInt(0),
		//PublicKey:      publicKey,
		SnapshotHash: snapshotBlockHash,
		Timestamp:    &now,
		Nonce:        []byte("test nonce test nonce"),
		Signature:    []byte("test signature test signature test signature"),
	}

	vmContext.AddBalance(&ledger.ViteTokenId, sendAmount)

	sendBlock.StateHash = *vmContext.GetStorageHash()
	sendBlock.Hash = sendBlock.ComputeHash()

	accounts[*addr1] = append(accounts[*addr1], sendBlock.Hash)

	return []*vm_context.VmAccountBlock{{
		AccountBlock: sendBlock,
		VmContext:    vmContext,
	}}, []types.Address{*addr1, *addr2}, nil
}

func mockSnapshotBlock(chn chain.Chain) (*ledger.SnapshotBlock, error) {
	latestBlock := chn.GetLatestSnapshotBlock()
	now := time.Now()
	snapshotBlock := &ledger.SnapshotBlock{
		Height:    latestBlock.Height + 1,
		PrevHash:  latestBlock.Hash,
		Timestamp: &now,
	}

	content := chn.GetNeedSnapshotContent()
	snapshotBlock.SnapshotContent = content

	trie, err := chn.GenStateTrie(latestBlock.StateHash, content)
	if err != nil {
		return nil, err
	}

	snapshotBlock.StateTrie = trie
	snapshotBlock.StateHash = *trie.Hash()
	snapshotBlock.Hash = snapshotBlock.ComputeHash()

	return snapshotBlock, err
}

func mockBlocks(chn chain.Chain, to uint64) {
	current := chn.GetLatestSnapshotBlock()
	if to <= current.Height {
		return
	}

	count := to - current.Height

	accountAddress1, _, _ := types.CreateAddress()
	accountAddress2, _, _ := types.CreateAddress()
	for i := uint64(0); i < count; i++ {
		snapshotBlock, _ := mockSnapshotBlock(chn)
		chn.InsertSnapshotBlock(snapshotBlock)

		for j := 0; j < 100; j++ {
			blocks, _, _ := mockAccountBlock(chn, snapshotBlock.Hash, &accountAddress1, &accountAddress2)
			chn.InsertAccountBlocks(blocks)
		}

		if (i+1)%100 == 0 {
			fmt.Printf("Make %d snapshot blocks.\n", i+1)
		}
	}
}

// mock peer
type mock_Peer struct {
}

func (m *mock_Peer) Report(err error) {
	panic("implement me")
}

func (m *mock_Peer) FileAddress() *net2.TCPAddr {
	panic("implement me")
}

func (m *mock_Peer) SetHead(head types.Hash, height uint64) {
	panic("implement me")
}

func (m *mock_Peer) SeeBlock(hash types.Hash) {
	panic("implement me")
}

func (m *mock_Peer) SendSubLedger(bs []*ledger.SnapshotBlock, abs []*ledger.AccountBlock, msgId uint64) (err error) {
	panic("implement me")
}

func (m *mock_Peer) SendSnapshotBlocks(bs []*ledger.SnapshotBlock, msgId uint64) (err error) {
	panic("implement me")
}

func (m *mock_Peer) SendNewSnapshotBlock(b *ledger.SnapshotBlock) (err error) {
	panic("implement me")
}

func (m *mock_Peer) SendNewAccountBlock(b *ledger.AccountBlock) (err error) {
	panic("implement me")
}

func (m *mock_Peer) SendAccountBlocks(bs []*ledger.AccountBlock, msgId uint64) (err error) {
	for _, block := range bs {
		fmt.Printf("accountblock %s/%d\n", block.Hash, block.Height)
	}
	return nil
}

func (m *mock_Peer) Send(code ViteCmd, msgId uint64, payload p2p.Serializable) (err error) {
	fmt.Printf("account exception %s\n", code)
	return nil
}

func (m *mock_Peer) RemoteAddr() *net2.TCPAddr {
	return nil
}

// mock GetAccountBlocksMsg
func chooseAccountHash() (types.Address, types.Hash) {
	actIndex := rand.Intn(len(accounts))
	acts := getAccountList()
	act := acts[actIndex]

	hashIndex := rand.Intn(len(accounts[act]))
	hash := accounts[act][hashIndex]

	return act, hash
}

func mockGetAccountBlocksMsg() *p2p.Msg {
	act, hash := chooseAccountHash()

	count := uint64(rand.Intn(100))
	ga := &message.GetAccountBlocks{
		Address: act,
		From: ledger.HashHeight{
			Hash: hash,
		},
		Count:   count,
		Forward: false,
	}

	fmt.Printf("mock getAccountBlocks %s/%d@%s\n", hash, count, act)

	payload, _ := ga.Serialize()

	return &p2p.Msg{
		CmdSet:  CmdSet,
		Cmd:     p2p.Cmd(GetAccountBlocksCode),
		Payload: payload,
	}
}

var gaHandler = getAccountBlocksHandler{
	chain: getChain(),
}

func TestGetAccountBlocksHandler_Handle(t *testing.T) {
	gaHandler.Handle(mockGetAccountBlocksMsg(), &mock_Peer{})
}
