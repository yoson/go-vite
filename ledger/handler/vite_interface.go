package handler

import (
	"github.com/asaskevich/EventBus"
	"github.com/vitelabs/go-vite/consensus"
	"github.com/vitelabs/go-vite/miner"
	protoInterface "github.com/vitelabs/go-vite/protocols/interfaces"
	"github.com/vitelabs/go-vite/wallet"
)

type Vite interface {
	Pm() protoInterface.ProtocolManager
	WalletManager() *wallet.Manager
	Miner() *miner.Miner
	Verifier() consensus.Verifier
	EventBus() EventBus.Bus
}
