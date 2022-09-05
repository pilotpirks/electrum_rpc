##Go [Electrum](https://electrum.org/) JSON-RPC client library.

```
go get github.com/pilotpirks/electrum_rpc
```

Initialize Elecrum settings
```
electrum setconfig rpcport 7777
electrum setconfig rpcuser rpc_username
electrum setconfig rpcpassword rpc_password
```

Example
```
package main

import (
	wallet "github.com/pilotpirks/electrum_rpc"
)

const (
	walletPath =  ""
	wltPassword =  ""
	rpcHost =  ""
	rpcPort =  ""
	rpcUser =  ""
	rpcPassword =  ""
)

func  main() {
	if wlt, err := wallet.New(rpcHost, rpcPort, rpcUser, rpcPassword, false); err !=  nil {
		log.Fatal("Initialize BTC wallet error: ", err.Error())
	}

	wlt.SetWalletPassword(wltPassword)
	if _, err := wlt.LoadWallet(walletPath, wltPassword); err !=  nil {
		log.Fatal("LoadWallet():: error", err.Error())
	}

	info, _ := wlt.GetInfo()
	log.Printf("%+v\n", info)
}
```
