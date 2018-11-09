package api

import (
	"errors"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/pow"
	"github.com/vitelabs/go-vite/pow/remote"
	"math/big"
)

type Pow struct {
}

func (p Pow) GetPowNonce(difficulty string, data types.Hash) ([]byte, error) {
	log.Info("GetPowNonce")

	if pow.VMTestParamEnabled {
		log.Info("use defaultTarget to calc")
		return pow.GetPowNonce(nil, data)
	}

	realDifficulty, ok := new(big.Int).SetString(difficulty, 10)
	if !ok {
		return nil, ErrStrToBigInt
	}

	nonce, e := remote.GenerateWork(realDifficulty, data)
	if e != nil {
		return nil, e
	}
	return nonce, nil
}

func (p Pow) CancelPow(data types.Hash) error {
	if err := remote.CancelWork(data.Bytes()); err != nil {
		return errors.New("pow cancel failed")
	}
	return nil
}
