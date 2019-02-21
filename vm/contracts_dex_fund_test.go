package vm

import (
	"bytes"
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/vitelabs/go-vite/common/helper"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/vm/contracts"
	"github.com/vitelabs/go-vite/vm/contracts/abi"
	cabi "github.com/vitelabs/go-vite/vm/contracts/abi"
	"github.com/vitelabs/go-vite/vm/contracts/dex"
	dexproto "github.com/vitelabs/go-vite/vm/contracts/dex/proto"
	"github.com/vitelabs/go-vite/vm/util"
	"math/big"
	"testing"
	"time"
)


type tokenInfo struct {
	tokenId types.TokenTypeId
	decimals int32
}

var (
	ETH = tokenInfo{types.TokenTypeId{'E', 'T', 'H', ' ', ' ', 'T', 'O', 'K', 'E', 'N'}, 12} //tradeToken
	VITE = tokenInfo{types.TokenTypeId{'V', 'I', 'T', 'E', ' ', 'T', 'O', 'K', 'E', 'N'}, 14} //quoteToken
)

func TestDexFund(t *testing.T) {
	db := initDexFundDatabase()
	userAddress, _ := types.BytesToAddress([]byte("12345678901234567890"))
	depositCash(db, userAddress, 3000, VITE.tokenId)
	innerTestDepositAndWithdraw(t, db, userAddress)
	innerTestFundNewOrder(t, db, userAddress)
	innerTestSettleOrder(t, db, userAddress)
}

func innerTestDepositAndWithdraw(t *testing.T, db *testDatabase, userAddress types.Address) {
	var err error
	registerToken(db, VITE)
	//deposit
	depositMethod := contracts.MethodDexFundUserDeposit{}

	depositSendAccBlock := &ledger.AccountBlock{}
	depositSendAccBlock.AccountAddress = userAddress

	depositSendAccBlock.TokenId = ETH.tokenId
	depositSendAccBlock.Amount = big.NewInt(100)

	depositSendAccBlock.Data, _ = contracts.ABIDexFund.PackMethod(contracts.MethodNameDexFundUserDeposit)
	err = depositMethod.DoSend(db, depositSendAccBlock)
	assert.True(t, err != nil)
	assert.True(t, bytes.Equal([]byte(err.Error()), []byte("token is invalid")))

	depositSendAccBlock.TokenId = VITE.tokenId
	depositSendAccBlock.Amount = big.NewInt(3000)

	depositSendAccBlock.Data, err = contracts.ABIDexFund.PackMethod(contracts.MethodNameDexFundUserDeposit)
	err = depositMethod.DoSend(db, depositSendAccBlock)
	assert.True(t, err == nil)
	assert.True(t, bytes.Equal(depositSendAccBlock.TokenId.Bytes(), VITE.tokenId.Bytes()))
	assert.Equal(t, depositSendAccBlock.Amount.Uint64(), uint64(3000))

	depositReceiveBlock := &ledger.AccountBlock{}

	_, err = depositMethod.DoReceive(db, depositReceiveBlock, depositSendAccBlock)
	assert.True(t, err == nil)

	dexFund, _ := dex.GetUserFundFromStorage(db, userAddress)
	assert.Equal(t, 1, len(dexFund.Accounts))
	acc := dexFund.Accounts[0]
	assert.True(t, bytes.Equal(acc.Token, VITE.tokenId.Bytes()))
	assert.True(t, CheckBigEqualToInt(3000, acc.Available))
	assert.True(t, CheckBigEqualToInt(0, acc.Locked))

	//withdraw
	withdrawMethod := contracts.MethodDexFundUserWithdraw{}

	withdrawSendAccBlock := &ledger.AccountBlock{}
	withdrawSendAccBlock.AccountAddress = userAddress
	withdrawSendAccBlock.Data, err = contracts.ABIDexFund.PackMethod(contracts.MethodNameDexFundUserWithdraw, VITE.tokenId, big.NewInt(200))
	err = withdrawMethod.DoSend(db, withdrawSendAccBlock)
	assert.True(t, err == nil)

	withdrawReceiveBlock := &ledger.AccountBlock{}
	now := time.Now()
	withdrawReceiveBlock.Timestamp = &now

	withdrawSendAccBlock.Data, err = contracts.ABIDexFund.PackMethod(contracts.MethodNameDexFundUserWithdraw, VITE.tokenId, big.NewInt(200))
	appendedBlocks, _ := withdrawMethod.DoReceive(db, withdrawReceiveBlock, withdrawSendAccBlock)

	dexFund, _ = dex.GetUserFundFromStorage(db, userAddress)
	acc = dexFund.Accounts[0]
	assert.True(t, CheckBigEqualToInt(2800, acc.Available))
	assert.Equal(t, 1, len(appendedBlocks))

}

func innerTestFundNewOrder(t *testing.T, db *testDatabase, userAddress types.Address) {
	registerToken(db, ETH)

	method := contracts.MethodDexFundNewOrder{}

	senderAccBlock := &ledger.AccountBlock{}
	senderAccBlock.AccountAddress = userAddress
	senderAccBlock.Data, _ = contracts.ABIDexFund.PackMethod(contracts.MethodNameDexFundNewOrder, orderIdBytesFromInt(1), VITE.tokenId.Bytes(), ETH.tokenId.Bytes(), true, uint32(dex.Limited), "0.3", big.NewInt(2000))
	//fmt.Printf("PackMethod err for send %s\n", err.Error())
	err := method.DoSend(db, senderAccBlock)
	assert.True(t, err == nil)

	param := new(dex.ParamDexFundNewOrder)
	err = contracts.ABIDexFund.UnpackMethod(param, contracts.MethodNameDexFundNewOrder, senderAccBlock.Data)
	assert.True(t, err == nil)
	//fmt.Printf("UnpackMethod err for send %s\n", err.Error())

	receiveBlock := &ledger.AccountBlock{}
	now := time.Now()
	receiveBlock.Timestamp = &now
	receiveBlock.SnapshotHash = types.DataHash([]byte{10, 1})

	var appendedBlocks []*contracts.SendBlock
	appendedBlocks, err = method.DoReceive(db, receiveBlock, senderAccBlock)
	assert.True(t, err == nil)

	dexFund, _ := dex.GetUserFundFromStorage(db, userAddress)
	acc := dexFund.Accounts[0]

	assert.True(t, CheckBigEqualToInt(800, acc.Available))
	assert.True(t, CheckBigEqualToInt(2000, acc.Locked))
	assert.Equal(t, 1, len(appendedBlocks))

	param1 := new(dex.ParamDexSerializedData)
	err = contracts.ABIDexTrade.UnpackMethod(param1, contracts.MethodNameDexTradeNewOrder, appendedBlocks[0].Data)
	order1 := &dexproto.Order{}
	proto.Unmarshal(param1.Data, order1)
	assert.True(t, CheckBigEqualToInt(6, order1.Amount))
	assert.Equal(t, order1.Status, int32(dex.Pending))
}

func innerTestSettleOrder(t *testing.T, db *testDatabase, userAddress types.Address) {
	method := contracts.MethodDexFundSettleOrders{}

	senderAccBlock := &ledger.AccountBlock{}
	senderAccBlock.AccountAddress = types.AddressDexTrade

	viteFundSettle := &dexproto.FundSettle{}
	viteFundSettle.Token = VITE.tokenId.Bytes()
	viteFundSettle.ReduceLocked = big.NewInt(1000).Bytes()
	viteFundSettle.ReleaseLocked = big.NewInt(100).Bytes()

	ethFundSettle := &dexproto.FundSettle{}
	ethFundSettle.Token = ETH.tokenId.Bytes()
	ethFundSettle.IncAvailable = big.NewInt(30).Bytes()

	fundAction := &dexproto.UserFundSettle{}
	fundAction.Address = userAddress.Bytes()
	fundAction.FundSettles = append(fundAction.FundSettles, viteFundSettle, ethFundSettle)

	feeAction := dexproto.FeeSettle{}
	feeAction.Token = ETH.tokenId.Bytes()
	userFeeSettle := &dexproto.UserFeeSettle{}
	userFeeSettle.Address = userAddress.Bytes()
	userFeeSettle.Amount = big.NewInt(15).Bytes()
	feeAction.UserFeeSettles = append(feeAction.UserFeeSettles, userFeeSettle)

	actions := dexproto.SettleActions{}
	actions.FundActions = append(actions.FundActions, fundAction)
	actions.FeeActions = append(actions.FeeActions, &feeAction)
	data, _ := proto.Marshal(&actions)

	senderAccBlock.Data, _ = contracts.ABIDexFund.PackMethod(contracts.MethodNameDexFundSettleOrders, data)
	err := method.DoSend(db, senderAccBlock)
	//fmt.Printf("err %s\n", err.Error())
	assert.True(t, err == nil)

	receiveBlock := &ledger.AccountBlock{}
	now := time.Now()
	receiveBlock.Timestamp = &now
	receiveBlock.SnapshotHash = types.DataHash([]byte{10, 1})
	_, err = method.DoReceive(db, receiveBlock, senderAccBlock)
	assert.True(t, err == nil)
	//fmt.Printf("receive err %s\n", err.Error())
	dexFund, _ := dex.GetUserFundFromStorage(db, userAddress)
	assert.Equal(t, 2, len(dexFund.Accounts))
	var ethAcc, viteAcc *dexproto.Account
	acc := dexFund.Accounts[0]
	if bytes.Equal(acc.Token, ETH.tokenId.Bytes()) {
		ethAcc = dexFund.Accounts[0]
		viteAcc = dexFund.Accounts[1]
	} else {
		ethAcc = dexFund.Accounts[1]
		viteAcc = dexFund.Accounts[0]
	}
	assert.True(t, CheckBigEqualToInt(0, ethAcc.Locked))
	assert.True(t, CheckBigEqualToInt(30, ethAcc.Available))
	assert.True(t, CheckBigEqualToInt(900, viteAcc.Locked))
	assert.True(t, CheckBigEqualToInt(900, viteAcc.Available))

	dexFee, err := dex.GetCurrentFeeSumFromStorage(db) // initDexFundDatabase snapshotBlock Height
	assert.Equal(t, 1, len(dexFee.Fees))
	feeAcc := dexFee.Fees[0]
	assert.True(t, CheckBigEqualToInt(15, feeAcc.Amount))
}

func initDexFundDatabase() *testDatabase {
	db := NewNoDatabase()
	db.addr = types.AddressDexFund
	t1 := time.Unix(1536214502, 0)
	snapshot1 := &ledger.SnapshotBlock{Height: 123, Timestamp: &t1, Hash: types.DataHash([]byte{10, 1})}
	db.snapshotBlockList = append(db.snapshotBlockList, snapshot1)

	db.storageMap[types.AddressConsensusGroup] = make(map[string][]byte)
	consensusGroupKey, _ := types.BytesToHash(abi.GetConsensusGroupKey(types.SNAPSHOT_GID))
	consensusGroupData, _ := abi.ABIConsensusGroup.PackVariable(abi.VariableNameConsensusGroupInfo,
		uint8(25),
		int64(1),
		int64(3),
		uint8(2),
		uint8(50),
		ledger.ViteTokenId,
		uint8(1),
		helper.JoinBytes(helper.LeftPadBytes(new(big.Int).Mul(big.NewInt(1e6), util.AttovPerVite).Bytes(), helper.WordSize), helper.LeftPadBytes(ledger.ViteTokenId.Bytes(), helper.WordSize), helper.LeftPadBytes(big.NewInt(3600*24*90).Bytes(), helper.WordSize)),
		uint8(1),
		[]byte{},
		db.addr,
		big.NewInt(0),
		uint64(1))
	db.storageMap[types.AddressConsensusGroup][string(consensusGroupKey.Bytes())] = consensusGroupData

	return db
}

func depositCash(db *testDatabase, address types.Address, amount uint64, token types.TokenTypeId) {
	if _, ok := db.balanceMap[address]; !ok {
		db.balanceMap[address] = make(map[types.TokenTypeId]*big.Int)
	}
	db.balanceMap[address][token] = big.NewInt(0).SetUint64(amount)
}

func registerToken(db *testDatabase, token tokenInfo) {
	tokenName := string(token.tokenId.Bytes()[0:4])
	tokenSymbol := string(token.tokenId.Bytes()[5:10])
	decimals := uint8(token.decimals)
	tokenData, _ := abi.ABIMintage.PackVariable(abi.VariableNameMintage, tokenName, tokenSymbol, big.NewInt(1e16), decimals, ledger.GenesisAccountAddress, big.NewInt(0), uint64(0))
	if _, ok := db.storageMap[types.AddressMintage]; !ok {
		db.storageMap[types.AddressMintage] = make(map[string][]byte)
	}
	mintageKey := string(cabi.GetMintageKey(token.tokenId))
	db.storageMap[types.AddressMintage][mintageKey] = tokenData
}

func CheckBigEqualToInt(expected int, value []byte) bool {
	return new(big.Int).SetUint64(uint64(expected)).Cmp(new(big.Int).SetBytes(value)) == 0
}

func orderIdBytesFromInt(v int) []byte {
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, uint32(v))
	return append([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, bs...)
}