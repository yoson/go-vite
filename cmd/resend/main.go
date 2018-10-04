package main

import (
	"flag"
	"github.com/vitelabs/go-vite"
	"github.com/vitelabs/go-vite/config"
	"github.com/vitelabs/go-vite/log15"
	"github.com/vitelabs/go-vite/vite"
	"github.com/vitelabs/go-vite/common/types"
	"fmt"
	"time"
	"github.com/vitelabs/go-vite/cmd/rpc_vite"
	"github.com/vitelabs/go-vite/crypto/ed25519"
	"encoding/hex"
)

func parseConfig() *config.Config {
	var globalConfig = config.GlobalConfig

	flag.StringVar(&globalConfig.Name, "name", globalConfig.Name, "boot name")
	flag.UintVar(&globalConfig.MaxPeers, "peers", globalConfig.MaxPeers, "max number of connections will be connected")
	flag.StringVar(&globalConfig.Addr, "addr", globalConfig.Addr, "will be listen by vite")
	flag.StringVar(&globalConfig.PrivateKey, "priv", globalConfig.PrivateKey, "hex encode of ed25519 privateKey, use for sign message")
	flag.StringVar(&globalConfig.DataDir, "dir", globalConfig.DataDir, "use for store all files")
	flag.UintVar(&globalConfig.NetID, "netid", globalConfig.NetID, "the network vite will connect")

	flag.Parse()

	globalConfig.P2P.Datadir = globalConfig.DataDir

	return globalConfig
}

func main() {

	govite.PrintBuildVersion()

	mainLog := log15.New("module", "gvite/main")

	parsedConfig := parseConfig()

	if s, e := parsedConfig.RunLogDirFile(); e == nil {
		log15.Root().SetHandler(
			log15.LvlFilterHandler(log15.LvlInfo, log15.Must.FileHandler(s, log15.TerminalFormat())),
		)
	}

	vnode, err := vite.New(parsedConfig)

	if err != nil {
		mainLog.Crit("Start vite failed.", "err", err)
	}

	addr,_:= types.HexToAddress("vite_098dfae02679a4ca05a4c8bf5dd00a8757f0c622bfccce7d68")
	go func() {
		time.Sleep(time.Second * 2)

		count := 100

		blocks, _:= vnode.Ledger().Ac().GetBlocksByAccAddr(&addr, 0,   1, count)

		publicKey, _ := hex.DecodeString("3af9a47a11140c681c2b2a85a4ce987fab0692589b2ce233bf7e174bd430177a")
		var genesisPublicKey = ed25519.PublicKey(publicKey)
		for i:=count -1;i>=0;i--  {
			blocks[i].PublicKey = genesisPublicKey
			fmt.Printf("%d: %+v\n", i, blocks[i])
			if blocks[i].Hash.String() == "0e7381d3e1a4c53eb252095fb5b6475c7b851f48619d5e5b7919c7a82f40ddf7" {
				fmt.Println("helloworldS")
			}

			vnode.Pm().SendMsg(nil, &protoTypes.Msg{
				Code:    protoTypes.AccountBlocksMsgCode,
				Payload: &protoTypes.AccountBlocksMsg{blocks[i]},
			})
			time.Sleep(time.Millisecond)

		}
	}()
	rpc_vite.StartIpcRpc(vnode, parsedConfig.DataDir)


}
