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

func exportContractBalanceAndStorage(m map[types.Address]*big.Int, g *Genesis, addr types.Address, balance *big.Int, trie *trie.Trie, c chain.Chain) (map[types.Address]*big.Int, *Genesis, error) {
	if addr == types.AddressRegister {
		m, g = exportRegisterBalanceAndStorage(m, g, trie)
		return m, g, nil
	} else if addr == types.AddressPledge {
		m, g = exportPledgeBalanceAndStorage(m, g, trie)
		return m, g, nil
	} else if addr == types.AddressMintage {
		m, g = exportMintageBalanceAndStorage(m, g, trie)
		return m, g, nil
	} else if addr == types.AddressVote {
		m, g = exportVoteBalanceAndStorage(m, g, trie)
		return m, g, nil
	} else if addr == types.AddressConsensusGroup {
		m, g = exportConsensusGroupBalanceAndStorage(m, g, trie)
		return m, g, nil
	} else {
		// for other contract, return to creator
		responseBlock, err := c.GetAccountBlockByHeight(&addr, 1)
		if err != nil {
			return m, g, err
		}
		requestBlock, err := c.GetAccountBlockByHash(&responseBlock.FromBlockHash)
		if err != nil {
			return m, g, err
		}
		m = updateBalance(m, requestBlock.AccountAddress, new(big.Int).Add(requestBlock.Fee, balance))
		return m, g, err
	}
}

var (
	newPledgeAmount    = new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18))
	refundPledgeAmount = new(big.Int).Mul(big.NewInt(400000), big.NewInt(1e18))
	newWithdrawHeight  = uint64(7776000)
	rewardTime         = int64(1)
)

func exportRegisterBalanceAndStorage(m map[types.Address]*big.Int, g *Genesis, trie *trie.Trie) (map[types.Address]*big.Int, *Genesis) {
	if g.ConsensusGroupInfo == nil {
		g.ConsensusGroupInfo = &ConsensusGroupContractInfo{}
	}
	g.ConsensusGroupInfo.RegistrationInfoMap = make(map[string]map[string]RegistrationInfo)
	g.ConsensusGroupInfo.HisNameMap = make(map[string]map[string]string)
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
			gidStr := gid.String()
			if _, ok := g.ConsensusGroupInfo.RegistrationInfoMap[gidStr]; !ok {
				g.ConsensusGroupInfo.RegistrationInfoMap[gidStr] = make(map[string]RegistrationInfo)
			}
			g.ConsensusGroupInfo.RegistrationInfoMap[gidStr][old.Name] = RegistrationInfo{
				old.NodeAddr, old.PledgeAddr, newPledgeAmount, newWithdrawHeight, rewardTime, int64(0), []types.Address{old.NodeAddr},
			}
			if _, ok := g.ConsensusGroupInfo.HisNameMap[gidStr]; !ok {
				g.ConsensusGroupInfo.HisNameMap[gidStr] = make(map[string]string)
			}
			g.ConsensusGroupInfo.HisNameMap[gidStr][old.NodeAddr.String()] = old.Name
		}
	}
	return m, g
}

func exportPledgeBalanceAndStorage(m map[types.Address]*big.Int, g *Genesis, trie *trie.Trie) (map[types.Address]*big.Int, *Genesis) {
	if g.PledgeInfo == nil {
		g.PledgeInfo = &PledgeContractInfo{}
	}
	g.PledgeInfo.PledgeInfoMap = make(map[string]PledgeInfo)
	g.PledgeInfo.PledgeBeneficialMap = make(map[string]*big.Int)
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
				g.PledgeInfo.PledgeInfoMap[cabi.GetPledgeAddrFromPledgeKey(key).String()] = PledgeInfo{old.Amount, 1, cabi.GetBeneficialFromPledgeKey(key)}
			}
		} else {
			amount := new(big.Int)
			if err := cabi.ABIPledge.UnpackVariable(amount, cabi.VariableNamePledgeBeneficial, value); err == nil {
				g.PledgeInfo.PledgeBeneficialMap[cabi.GetBeneficialFromPledgeBeneficialKey(key).String()] = amount
			}

		}
	}
	return m, g
}

var mintageFee = new(big.Int).Mul(big.NewInt(1e3), big.NewInt(1e18))

func exportMintageBalanceAndStorage(m map[types.Address]*big.Int, g *Genesis, trie *trie.Trie) (map[types.Address]*big.Int, *Genesis) {
	if g.MintageInfo == nil {
		g.MintageInfo = &MintageContractInfo{}
	}
	g.MintageInfo.TokenInfoMap = make(map[string]TokenInfo)
	g.MintageInfo.LogList = make([]GenesisVmLog, 0)
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
			g.MintageInfo.TokenInfoMap[tokenId.String()] = TokenInfo{old.TokenName, old.TokenSymbol, old.TotalSupply, old.Decimals, types.AddressConsensusGroup, old.PledgeAmount, old.PledgeAddr, old.WithdrawHeight, helper.Tt256m1, false, true}
		} else {
			if old.MaxSupply == nil {
				old.MaxSupply = big.NewInt(0)
			}
			g.MintageInfo.TokenInfoMap[tokenId.String()] = TokenInfo{old.TokenName, old.TokenSymbol, old.TotalSupply, old.Decimals, old.Owner, old.PledgeAmount, old.PledgeAddr, old.WithdrawHeight, old.MaxSupply, old.OwnerBurnOnly, old.IsReIssuable}
		}
		log := util.NewLog(ABIMintageNew, "mint", tokenId)
		g.MintageInfo.LogList = append(g.MintageInfo.LogList, GenesisVmLog{hex.EncodeToString(log.Data), log.Topics})
	}
	return m, g
}

var voteKeyPrefix = []byte{0}

func exportVoteBalanceAndStorage(m map[types.Address]*big.Int, g *Genesis, trie *trie.Trie) (map[types.Address]*big.Int, *Genesis) {
	if g.ConsensusGroupInfo == nil {
		g.ConsensusGroupInfo = &ConsensusGroupContractInfo{}
	}
	g.ConsensusGroupInfo.VoteStatusMap = make(map[string]map[string]string)
	g.ConsensusGroupInfo.VoteStatusMap[types.SNAPSHOT_GID.String()] = make(map[string]string)
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
		nodeName := new(string)
		if err := cabi.ABIVote.UnpackVariable(nodeName, cabi.VariableNameVoteStatus, value); err == nil {
			g.ConsensusGroupInfo.VoteStatusMap[types.SNAPSHOT_GID.String()][voterAddr.String()] = *nodeName
		}
	}
	return m, g
}

var groupInfoKeyPrefix = []byte{1}

func exportConsensusGroupBalanceAndStorage(m map[types.Address]*big.Int, g *Genesis, trie *trie.Trie) (map[types.Address]*big.Int, *Genesis) {
	if g.ConsensusGroupInfo == nil {
		g.ConsensusGroupInfo = &ConsensusGroupContractInfo{}
	}
	g.ConsensusGroupInfo.ConsensusGroupInfoMap = make(map[string]ConsensusGroupInfo)
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
		cabi.ABIConsensusGroup.UnpackVariable(old, cabi.VariableNameConsensusGroupInfo, value)
		if gid := cabi.GetGidFromConsensusGroupKey(key); gid == types.SNAPSHOT_GID {
			g.ConsensusGroupInfo.ConsensusGroupInfoMap[gid.String()] = ConsensusGroupInfo{
				old.NodeCount,
				old.Interval,
				old.PerCount,
				old.RandCount,
				old.RandRank,
				uint16(1),
				uint8(0),
				old.CountingTokenId,
				old.RegisterConditionId,
				RegisterConditionParam{newPledgeAmount, oldParam.PledgeToken, oldParam.PledgeHeight},
				old.VoteConditionId,
				VoteConditionParam{},
				old.Owner,
				old.PledgeAmount,
				old.WithdrawHeight}
		} else {
			g.ConsensusGroupInfo.ConsensusGroupInfoMap[gid.String()] = ConsensusGroupInfo{
				old.NodeCount,
				old.Interval,
				old.PerCount,
				old.RandCount,
				old.RandRank,
				uint16(48),
				uint8(1),
				old.CountingTokenId,
				old.RegisterConditionId,
				RegisterConditionParam{newPledgeAmount, oldParam.PledgeToken, oldParam.PledgeHeight},
				old.VoteConditionId,
				VoteConditionParam{},
				old.Owner,
				old.PledgeAmount,
				old.WithdrawHeight}
		}
	}
	return m, g
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
		{"type":"function","name":"CreateConsensusGroup", "inputs":[{"name":"gid","type":"gid"},{"name":"nodeCount","type":"uint8"},{"name":"interval","type":"int64"},{"name":"perCount","type":"int64"},{"name":"randCount","type":"uint8"},{"name":"randRank","type":"uint8"},{"name":"repeat","type":"uint16"},{"name":"checkLevel","type":"uint8"},{"name":"countingTokenId","type":"tokenId"},{"name":"registerConditionId","type":"uint8"},{"name":"registerConditionParam","type":"bytes"},{"name":"voteConditionId","type":"uint8"},{"name":"voteConditionParam","type":"bytes"}]},
		{"type":"function","name":"CancelConsensusGroup", "inputs":[{"name":"gid","type":"gid"}]},
		{"type":"function","name":"ReCreateConsensusGroup", "inputs":[{"name":"gid","type":"gid"}]},
		{"type":"variable","name":"consensusGroupInfo","inputs":[{"name":"nodeCount","type":"uint8"},{"name":"interval","type":"int64"},{"name":"perCount","type":"int64"},{"name":"randCount","type":"uint8"},{"name":"randRank","type":"uint8"},{"name":"repeat","type":"uint16"},{"name":"checkLevel","type":"uint8"},{"name":"countingTokenId","type":"tokenId"},{"name":"registerConditionId","type":"uint8"},{"name":"registerConditionParam","type":"bytes"},{"name":"voteConditionId","type":"uint8"},{"name":"voteConditionParam","type":"bytes"},{"name":"owner","type":"address"},{"name":"pledgeAmount","type":"uint256"},{"name":"withdrawHeight","type":"uint64"}]},
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
