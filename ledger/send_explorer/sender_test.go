package send_explorer

import (
	"github.com/vitelabs/go-vite/common"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/crypto/ed25519"
	"github.com/vitelabs/go-vite/ledger"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

var (
	wg sync.WaitGroup
)

func getSender() *Sender {
	filename := filepath.Join(common.DefaultDataDir(), "ledger_mq_test")
	os.Remove(filename)

	sender := NewSender([]string{"118.25.228.148:9092"}, filename)
	return sender
}

func TestSender(t *testing.T) {
	getSender()
	//for i := 0; i < 50000; i++ {
	//	testInsertAccountBlock(t, sender)
	//	testInsertSnapshotBlock(t, sender)
	//	testDeleteAccountBlocks(t, sender)
	//	testDeleteSnapshotBlocks(t, sender)
	//}

	wg.Add(1)
	wg.Wait()
}

func testInsertAccountBlock(t *testing.T, sender *Sender) {
	publicKey, _ := ed25519.HexToPublicKey("7194af5b7032cb470c41b313e2675e2c3ba3377e66617247012b8d638552fb17")
	address, _ := types.HexToAddress("vite_e03a928782f6a7a6d458bd3da2a20604c6177949bf928db76c")
	toAddress, _ := types.HexToAddress("vite_098dfae02679a4ca05a4c8bf5dd00a8757f0c622bfccce7d68")
	prevHash, _ := types.HexToHash("9642414cca2cfdadd736fe3bb168bcfa06a373bf65870f86f44301fc97fc7429")
	hash, _ := types.HexToHash("7cf17fb81125ed207095a0fa076be3e333ab95c7279e5eab675b2e809367790f")
	sHash, _ := types.HexToHash("9990ede4df31383aec521d4827fb40610f22cb21e7f63e749868730748effff4")

	signature := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	sender.InsertAccountBlock(&ledger.AccountBlock{
		Meta: &ledger.AccountBlockMeta{
			AccountId: big.NewInt(123),
			Height:    big.NewInt(234),
			Status:    2,
		},
		AccountAddress:    &address,
		PublicKey:         publicKey,
		To:                &toAddress,
		From:              nil,
		FromHash:          nil,
		PrevHash:          &prevHash,
		Hash:              &hash,
		Balance:           big.NewInt(10000000),
		Amount:            big.NewInt(10),
		Timestamp:         uint64(123456789012),
		TokenId:           &ledger.MockViteTokenId,
		Data:              "hAHAASDF",
		SnapshotTimestamp: &sHash,
		Signature:         signature,
		Nounce:            []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
		Difficulty:        []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
		FAmount:           big.NewInt(123),
	})
}

func testInsertSnapshotBlock(t *testing.T, sender *Sender) {
	prevHash, _ := types.HexToHash("9642414cca2cfdadd736fe3bb168bcfa06a373bf65870f86f44301fc97fc7429")

	hash, _ := types.HexToHash("7cf17fb81125ed207095a0fa076be3e333ab95c7279e5eab675b2e809367790f")
	hash2, _ := types.HexToHash("17350791ecf0c0e142457db7ce488e5cb0be9105b066b69b0b0c4b81777a6ad3")
	hash3, _ := types.HexToHash("59a118ec0a6259a8cfda52b1aa9db26185ecb8a0c1e161a3a5418d1b569807dc")

	address, _ := types.HexToAddress("vite_e03a928782f6a7a6d458bd3da2a20604c6177949bf928db76c")
	signature := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	publicKey, _ := ed25519.HexToPublicKey("7194af5b7032cb470c41b313e2675e2c3ba3377e66617247012b8d638552fb17")

	result := sender.InsertSnapshotBlock(&ledger.SnapshotBlock{
		PrevHash: &prevHash,
		Hash:     &hash,
		Height:   big.NewInt(12),
		Producer: &address,
		Snapshot: map[string]*ledger.SnapshotItem{
			"ABCDEF": {
				AccountBlockHash:   &hash2,
				AccountBlockHeight: big.NewInt(12),
			},
			"GHIJKJS": {
				AccountBlockHash:   &hash3,
				AccountBlockHeight: big.NewInt(13),
			},
		},
		Signature: signature,
		Timestamp: uint64(123456789012),
		Amount:    big.NewInt(10),
		PublicKey: publicKey,
	})
	t.Log(result)
}

func testDeleteAccountBlocks(t *testing.T, sender *Sender) {
	hash, _ := types.HexToHash("7cf17fb81125ed207095a0fa076be3e333ab95c7279e5eab675b2e809367790f")
	hash2, _ := types.HexToHash("17350791ecf0c0e142457db7ce488e5cb0be9105b066b69b0b0c4b81777a6ad3")
	hash3, _ := types.HexToHash("59a118ec0a6259a8cfda52b1aa9db26185ecb8a0c1e161a3a5418d1b569807dc")
	hashList := []*types.Hash{&hash, &hash2, &hash3}

	result := sender.DeleteAccountBlocks(hashList)
	t.Log(result)
}

func testDeleteSnapshotBlocks(t *testing.T, sender *Sender) {
	hash, _ := types.HexToHash("7cf17fb81125ed207095a0fa076be3e333ab95c7279e5eab675b2e809367790f")
	hash2, _ := types.HexToHash("17350791ecf0c0e142457db7ce488e5cb0be9105b066b69b0b0c4b81777a6ad3")
	hash3, _ := types.HexToHash("59a118ec0a6259a8cfda52b1aa9db26185ecb8a0c1e161a3a5418d1b569807dc")
	hashList := []*types.Hash{&hash, &hash2, &hash3}

	result := sender.DeleteSnapshotBlocks(hashList)
	t.Log(result)
}
