package nodemanager

import (
	"bytes"
	"encoding/hex"
	"github.com/vitelabs/go-vite/chain"
	"github.com/vitelabs/go-vite/common/helper"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/trie"
	"github.com/vitelabs/go-vite/vm/abi"
	cabi "github.com/vitelabs/go-vite/vm/contracts/abi"
	"github.com/vitelabs/go-vite/vm/util"
	"math/big"
	"strings"
)

var (
	STORAGE_KEY_BALANCE = []byte("$balance")
	STORAGE_KEY_CODE    = []byte("$code")
)

func isBalanceOrCode(key []byte) bool {
	return bytes.HasPrefix(key, STORAGE_KEY_CODE) || bytes.HasPrefix(key, STORAGE_KEY_BALANCE)
}

var registerNameMap = make(map[string]string)

func exportContractBalanceAndStorage(m map[types.Address]*big.Int, sm map[types.Address]map[string]string, lm map[types.Address]ledger.VmLogList, addr types.Address, balance *big.Int, trie *trie.Trie, c chain.Chain) (map[types.Address]*big.Int, map[types.Address]map[string]string, map[types.Address]ledger.VmLogList, error) {
	if addr == types.AddressRegister {
		m, sm, lm = exportRegisterBalanceAndStorage(m, sm, lm, trie)
		return m, sm, lm, nil
	} else if addr == types.AddressPledge {
		m, sm, lm = exportPledgeBalanceAndStorage(m, sm, lm, trie)
		return m, sm, lm, nil
	} else if addr == types.AddressMintage {
		m, sm, lm = exportMintageBalanceAndStorage(m, sm, lm, trie)
		return m, sm, lm, nil
	} else if addr == types.AddressVote {
		m, sm, lm = exportVoteBalanceAndStorage(m, sm, lm, trie)
		return m, sm, lm, nil
	} else if addr == types.AddressConsensusGroup {
		m, sm, lm = exportConsensusGroupBalanceAndStorage(m, sm, lm, trie)
		return m, sm, lm, nil
	} else {
		// for other contract, return to creator
		responseBlock, err := c.GetAccountBlockByHeight(&addr, 1)
		if err != nil {
			return m, sm, lm, err
		}
		requestBlock, err := c.GetAccountBlockByHash(&responseBlock.FromBlockHash)
		if err != nil {
			return m, sm, lm, err
		}
		m = updateBalance(m, requestBlock.AccountAddress, new(big.Int).Add(requestBlock.Fee, balance))
		return m, sm, lm, err
	}
}

var (
	newPledgeAmount    = new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18))
	refundPledgeAmount = new(big.Int).Mul(big.NewInt(500000), big.NewInt(1e18))
	newWithdrawHeight  = uint64(7776000)
	rewardTime         = int64(1)
)

func exportRegisterBalanceAndStorage(m map[types.Address]*big.Int, sm map[types.Address]map[string]string, lm map[types.Address]ledger.VmLogList, trie *trie.Trie) (map[types.Address]*big.Int, map[types.Address]map[string]string, map[types.Address]ledger.VmLogList) {
	iter := trie.NewIterator(nil)
	for {
		key, value, ok := iter.Next()
		if !ok {
			break
		}
		if isBalanceOrCode(key) || len(value) == 0 {
			continue
		}
		if cabi.IsRegisterKey(key) {
			old := new(types.Registration)
			cabi.ABIRegister.UnpackVariable(old, cabi.VariableNameRegistration, value)
			if !old.IsActive() {
				continue
			}
			registerNameMap[old.Name] = ""
			gid := cabi.GetGidFromRegisterKey(key)
			newValue, err := ABIConsensusGroupNew.PackVariable("registration", old.Name, old.NodeAddr, old.PledgeAddr, newPledgeAmount, newWithdrawHeight, rewardTime, int64(0), []types.Address{old.NodeAddr})
			if err != nil {
				panic(err)
			}
			sm = updateStorage(sm, types.AddressConsensusGroup, append(gid.Bytes(), types.DataHash([]byte(old.Name)).Bytes()[types.GidSize:]...), newValue)
			newHisNameValue, err := ABIConsensusGroupNew.PackVariable("hisName", old.Name)
			sm = updateStorage(sm, types.AddressConsensusGroup, append(old.NodeAddr.Bytes(), gid.Bytes()...), newHisNameValue)
			m = updateBalance(m, old.PledgeAddr, refundPledgeAmount)
		}
	}
	return m, sm, lm
}

func exportPledgeBalanceAndStorage(m map[types.Address]*big.Int, sm map[types.Address]map[string]string, lm map[types.Address]ledger.VmLogList, trie *trie.Trie) (map[types.Address]*big.Int, map[types.Address]map[string]string, map[types.Address]ledger.VmLogList) {
	iter := trie.NewIterator(nil)
	for {
		key, value, ok := iter.Next()
		if !ok {
			break
		}
		if isBalanceOrCode(key) || len(value) == 0 {
			continue
		}
		if cabi.IsPledgeKey(key) {
			old := new(cabi.PledgeInfo)
			if err := cabi.ABIPledge.UnpackVariable(old, cabi.VariableNamePledgeInfo, value); err == nil {
				newValue, err := ABIPledgeNew.PackVariable("pledgeInfo", old.Amount, uint64(1))
				if err != nil {
					panic(err)
				}
				sm = updateStorage(sm, types.AddressPledge, key, newValue)
			}
		} else {
			sm = updateStorage(sm, types.AddressPledge, key, value)
		}
	}
	return m, sm, lm
}

var mintageFee = new(big.Int).Mul(big.NewInt(1e3), big.NewInt(1e18))

func exportMintageBalanceAndStorage(m map[types.Address]*big.Int, sm map[types.Address]map[string]string, lm map[types.Address]ledger.VmLogList, trie *trie.Trie) (map[types.Address]*big.Int, map[types.Address]map[string]string, map[types.Address]ledger.VmLogList) {
	iter := trie.NewIterator(nil)
	for {
		key, value, ok := iter.Next()
		if !ok {
			break
		}
		if isBalanceOrCode(key) || len(value) == 0 {
			continue
		}
		if !cabi.IsMintageKey(key) {
			continue
		}
		tokenId := cabi.GetTokenIdFromMintageKey(key)
		old, _ := cabi.ParseTokenInfo(value)
		if tokenId == ledger.ViteTokenId {
			newValue, err := ABIMintageNew.PackVariable("tokenInfo", old.TokenName, old.TokenSymbol, old.TotalSupply, old.Decimals, types.AddressConsensusGroup, old.PledgeAmount, old.WithdrawHeight, old.PledgeAddr, true, helper.Tt256m1, false)
			if err != nil {
				panic(err)
			}
			sm = updateStorage(sm, types.AddressMintage, tokenId.Bytes(), newValue)
		} else {
			if old.MaxSupply == nil {
				old.MaxSupply = big.NewInt(0)
			}
			newValue, err := ABIMintageNew.PackVariable("tokenInfo", old.TokenName, old.TokenSymbol, old.TotalSupply, old.Decimals, old.Owner, old.PledgeAmount, old.WithdrawHeight, old.PledgeAddr, old.IsReIssuable, old.MaxSupply, old.OwnerBurnOnly)
			if err != nil {
				panic(err)
			}
			sm = updateStorage(sm, types.AddressMintage, tokenId.Bytes(), newValue)
		}
		lm = appendLog(lm, types.AddressMintage, util.NewLog(ABIMintageNew, "mint", tokenId))
	}
	return m, sm, lm
}

var voteKeyPrefix = []byte{0}

func exportVoteBalanceAndStorage(m map[types.Address]*big.Int, sm map[types.Address]map[string]string, lm map[types.Address]ledger.VmLogList, trie *trie.Trie) (map[types.Address]*big.Int, map[types.Address]map[string]string, map[types.Address]ledger.VmLogList) {
	iterator := trie.NewIterator(nil)
	for {
		key, value, ok := iterator.Next()
		if !ok {
			break
		}
		if isBalanceOrCode(key) || len(value) == 0 {
			continue
		}
		voterAddr := cabi.GetAddrFromVoteKey(key)
		sm = updateStorage(sm, types.AddressConsensusGroup, helper.JoinBytes(voteKeyPrefix, types.SNAPSHOT_GID.Bytes(), voterAddr.Bytes()), value)
	}
	return m, sm, lm
}

var groupInfoKeyPrefix = []byte{1}

func exportConsensusGroupBalanceAndStorage(m map[types.Address]*big.Int, sm map[types.Address]map[string]string, lm map[types.Address]ledger.VmLogList, trie *trie.Trie) (map[types.Address]*big.Int, map[types.Address]map[string]string, map[types.Address]ledger.VmLogList) {
	iterator := trie.NewIterator(nil)
	for {
		key, value, ok := iterator.Next()
		if !ok {
			break
		}
		if isBalanceOrCode(key) || len(value) == 0 {
			continue
		}
		old := new(types.ConsensusGroupInfo)
		oldParam := new(cabi.VariableConditionRegisterOfPledge)
		cabi.ABIConsensusGroup.UnpackVariable(oldParam, cabi.VariableNameConditionRegisterOfPledge, old.RegisterConditionParam)
		newParam, err := ABIConsensusGroupNew.PackVariable("registerOfPledge", newPledgeAmount, oldParam.PledgeToken, oldParam.PledgeHeight)
		if err != nil {
			panic(err)
		}
		cabi.ABIConsensusGroup.UnpackVariable(old, cabi.VariableNameConsensusGroupInfo, value)
		newValue, err := ABIConsensusGroupNew.PackVariable(
			"consensusGroupInfo",
			old.NodeCount,
			old.Interval,
			old.PerCount,
			old.RandCount,
			old.RandRank,
			old.CountingTokenId,
			old.RegisterConditionId,
			newParam,
			old.VoteConditionId,
			old.VoteConditionParam,
			old.Owner,
			old.PledgeAmount,
			old.WithdrawHeight)
		if err != nil {
			panic(err)
		}
		gid := cabi.GetGidFromConsensusGroupKey(key)
		sm = updateStorage(sm, types.AddressConsensusGroup, append(groupInfoKeyPrefix, gid.Bytes()...), newValue)
	}
	return m, sm, lm
}

func updateBalance(m map[types.Address]*big.Int, addr types.Address, balance *big.Int) map[types.Address]*big.Int {
	if v, ok := m[addr]; ok {
		v = v.Add(v, balance)
		m[addr] = v
	} else {
		m[addr] = balance
	}
	return m
}

func updateStorage(sm map[types.Address]map[string]string, addr types.Address, key []byte, value []byte) map[types.Address]map[string]string {
	keyStr := hex.EncodeToString(key)
	valueStr := hex.EncodeToString(value)
	if m, ok := sm[addr]; !ok {
		sm[addr] = make(map[string]string)
		sm[addr][keyStr] = valueStr
		return sm
	} else if v, ok := m[keyStr]; ok {
		panic("update storage failed, duplicate key " + keyStr + " in addr " + addr.String() + ", origin value " + v + ", new value " + valueStr)
	} else {
		sm[addr][keyStr] = valueStr
		return sm
	}
}

func appendLog(lm map[types.Address]ledger.VmLogList, addr types.Address, l *ledger.VmLog) map[types.Address]ledger.VmLogList {
	list := lm[addr]
	lm[addr] = append(list, l)
	return lm
}

func filterContractStorageMap(sm map[types.Address]map[string]string) map[types.Address]map[string]string {
	for k, v := range sm[types.AddressVote] {
		nodeName := new(string)
		data, _ := hex.DecodeString(v)
		ABIConsensusGroupNew.UnpackVariable(nodeName, "voteStatus", data)
		if _, ok := registerNameMap[*nodeName]; !ok {
			delete(sm[types.AddressVote], k)
		}
	}
	return sm
}

const (
	jsonMintage = `
	[
		{"type":"function","name":"CancelMintPledge","inputs":[{"name":"tokenId","type":"tokenId"}]},
		{"type":"function","name":"Mint","inputs":[{"name":"isReIssuable","type":"bool"},{"name":"tokenId","type":"tokenId"},{"name":"tokenName","type":"string"},{"name":"tokenSymbol","type":"string"},{"name":"totalSupply","type":"uint256"},{"name":"decimals","type":"uint8"},{"name":"maxSupply","type":"uint256"},{"name":"ownerBurnOnly","type":"bool"}]},
		{"type":"function","name":"Issue","inputs":[{"name":"tokenId","type":"tokenId"},{"name":"amount","type":"uint256"},{"name":"beneficial","type":"address"}]},
		{"type":"function","name":"Burn","inputs":[]},
		{"type":"function","name":"TransferOwner","inputs":[{"name":"tokenId","type":"tokenId"},{"name":"newOwner","type":"address"}]},
		{"type":"function","name":"ChangeTokenType","inputs":[{"name":"tokenId","type":"tokenId"}]},
		{"type":"variable","name":"mintage","inputs":[{"name":"tokenName","type":"string"},{"name":"tokenSymbol","type":"string"},{"name":"totalSupply","type":"uint256"},{"name":"decimals","type":"uint8"},{"name":"owner","type":"address"},{"name":"pledgeAmount","type":"uint256"},{"name":"withdrawHeight","type":"uint64"}]},
		{"type":"variable","name":"tokenInfo","inputs":[{"name":"tokenName","type":"string"},{"name":"tokenSymbol","type":"string"},{"name":"totalSupply","type":"uint256"},{"name":"decimals","type":"uint8"},{"name":"owner","type":"address"},{"name":"pledgeAmount","type":"uint256"},{"name":"withdrawHeight","type":"uint64"},{"name":"pledgeAddr","type":"address"},{"name":"isReIssuable","type":"bool"},{"name":"maxSupply","type":"uint256"},{"name":"ownerBurnOnly","type":"bool"}]},
		{"type":"event","name":"mint","inputs":[{"name":"tokenId","type":"tokenId","indexed":true}]},
		{"type":"event","name":"issue","inputs":[{"name":"tokenId","type":"tokenId","indexed":true}]},
		{"type":"event","name":"burn","inputs":[{"name":"tokenId","type":"tokenId","indexed":true},{"name":"address","type":"address"},{"name":"amount","type":"uint256"}]},
		{"type":"event","name":"transferOwner","inputs":[{"name":"tokenId","type":"tokenId","indexed":true},{"name":"owner","type":"address"}]},
		{"type":"event","name":"changeTokenType","inputs":[{"name":"tokenId","type":"tokenId","indexed":true}]}
	]`
	jsonPledge = `
	[
		{"type":"function","name":"Pledge", "inputs":[{"name":"beneficial","type":"address"}]},
		{"type":"function","name":"CancelPledge","inputs":[{"name":"beneficial","type":"address"},{"name":"amount","type":"uint256"}]},
		{"type":"variable","name":"pledgeInfo","inputs":[{"name":"amount","type":"uint256"},{"name":"withdrawHeight","type":"uint64"}]},
		{"type":"variable","name":"pledgeBeneficial","inputs":[{"name":"amount","type":"uint256"}]}
	]`
	jsonConsensusGroup = `
	[
		{"type":"function","name":"CreateConsensusGroup", "inputs":[{"name":"gid","type":"gid"},{"name":"nodeCount","type":"uint8"},{"name":"interval","type":"int64"},{"name":"perCount","type":"int64"},{"name":"randCount","type":"uint8"},{"name":"randRank","type":"uint8"},{"name":"countingTokenId","type":"tokenId"},{"name":"registerConditionId","type":"uint8"},{"name":"registerConditionParam","type":"bytes"},{"name":"voteConditionId","type":"uint8"},{"name":"voteConditionParam","type":"bytes"}]},
		{"type":"function","name":"CancelConsensusGroup", "inputs":[{"name":"gid","type":"gid"}]},
		{"type":"function","name":"ReCreateConsensusGroup", "inputs":[{"name":"gid","type":"gid"}]},
		{"type":"variable","name":"consensusGroupInfo","inputs":[{"name":"nodeCount","type":"uint8"},{"name":"interval","type":"int64"},{"name":"perCount","type":"int64"},{"name":"randCount","type":"uint8"},{"name":"randRank","type":"uint8"},{"name":"countingTokenId","type":"tokenId"},{"name":"registerConditionId","type":"uint8"},{"name":"registerConditionParam","type":"bytes"},{"name":"voteConditionId","type":"uint8"},{"name":"voteConditionParam","type":"bytes"},{"name":"owner","type":"address"},{"name":"pledgeAmount","type":"uint256"},{"name":"withdrawHeight","type":"uint64"}]},
		{"type":"variable","name":"registerOfPledge","inputs":[{"name":"pledgeAmount","type":"uint256"},{"name":"pledgeToken","type":"tokenId"},{"name":"pledgeHeight","type":"uint64"}]},
		
		{"type":"function","name":"Register", "inputs":[{"name":"gid","type":"gid"},{"name":"name","type":"string"},{"name":"nodeAddr","type":"address"}]},
		{"type":"function","name":"UpdateRegistration", "inputs":[{"name":"gid","type":"gid"},{"Name":"name","type":"string"},{"name":"nodeAddr","type":"address"}]},
		{"type":"function","name":"CancelRegister","inputs":[{"name":"gid","type":"gid"}, {"name":"name","type":"string"}]},
		{"type":"function","name":"Reward","inputs":[{"name":"gid","type":"gid"},{"name":"name","type":"string"},{"name":"beneficialAddr","type":"address"}]},
		{"type":"variable","name":"registration","inputs":[{"name":"name","type":"string"},{"name":"nodeAddr","type":"address"},{"name":"pledgeAddr","type":"address"},{"name":"amount","type":"uint256"},{"name":"withdrawHeight","type":"uint64"},{"name":"rewardTime","type":"int64"},{"name":"cancelTime","type":"int64"},{"name":"hisAddrList","type":"address[]"}]},
		{"type":"variable","name":"hisName","inputs":[{"name":"name","type":"string"}]},
		
		{"type":"function","name":"Vote", "inputs":[{"name":"gid","type":"gid"},{"name":"nodeName","type":"string"}]},
		{"type":"function","name":"CancelVote","inputs":[{"name":"gid","type":"gid"}]},
		{"type":"variable","name":"voteStatus","inputs":[{"name":"nodeName","type":"string"}]}
	]`
)

var (
	ABIMintageNew, _        = abi.JSONToABIContract(strings.NewReader(jsonMintage))
	ABIPledgeNew, _         = abi.JSONToABIContract(strings.NewReader(jsonPledge))
	ABIConsensusGroupNew, _ = abi.JSONToABIContract(strings.NewReader(jsonConsensusGroup))
)
