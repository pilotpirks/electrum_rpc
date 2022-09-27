package electrum_rpc

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type (
	// Client defines the JSON-RPC client structure.
	Client struct {
		address        string
		username       string
		password       string
		walletPassword string
		httpClient     *http.Client
		Debug          bool
	}

	// rpcRequest is the JSON-RPC request structure
	rpcRequest struct {
		ID      int64       `json:"id"`
		Method  string      `json:"method"`
		Params  interface{} `json:"params,omitempty"`
		JSONRPC string      `json:"jsonrpc"`
	}

	// rpcResponse is the JSON-RPC response structure.
	rpcResponse struct {
		ID     int64           `json:"id"`
		Result json.RawMessage `json:"result,omitempty"`
		Err    interface{}     `json:"error,omitempty"`
	}

	// JSONDate defines a custom time alias for json unmarshalling.
	JSONDate time.Time

	Info struct {
		AutoConnect      bool   `json:"auto_connect"`
		BlockchainHeight uint64 `json:"blockchain_height"`
		Connected        bool   `json:"connected"`
		DefaultWallet    string `json:"default_wallet"`
		FeePerKb         uint64 `json:"fee_per_kb"`
		Path             string `json:"path"`
		Server           string `json:"server"`
		ServerHeight     uint64 `json:"server_height"`
		SpvNodes         uint64 `json:"spv_nodes"`
		Version          string `json:"version"`
	}

	Create struct {
		Msg  string `json:"msg"`
		Path string `json:"path"`
		Seed string `json:"seed"`
	}

	// Balance represents a response to getbalance, getaddressbalance.
	Balance struct {
		Unconfirmed float64 `json:"unconfirmed"`
		Confirmed   float64 `json:"confirmed"`
		Unmatured   float64 `json:"unmatured"`
	}

	// AddressHistory represents a response to getaddresshistory.
	AddressHistory struct {
		TxHash string `json:"tx_hash"`
		Height int64
	}

	// Merkle represents a response to getmerkle.
	Merkle struct {
		Position    int      `json:"pos"`
		Merkle      []string `json:"merkle"`
		BlockHeight uint64   `json:"block_height"`
	}

	// UTXO represents all unspent transaction outputs (getaddressunspent).
	UTXO struct {
		Value  uint64 // value in satoshis
		TxHash string `json:"tx_hash"`
		TxPos  int    `json:"tx_pos"`
		Height uint64 `json:"height"`
	}

	// Unspent represents the unspent output returns from listunspent.
	Unspent struct {
		Address     string `json:"address"`
		Coinbase    bool   `json:"coinbase"`
		Height      uint64 `json:"height"`
		PrevoutHash string `json:"prevout_hash"`
		PrevoutN    int    `json:"prevout_n"`
		Value       string `json:"value"`
	}

	// Address represents the structure returned from listaddresses.
	Address struct {
		Address string
		Label   string
		Balance string
	}

	MultiSig struct {
		Address      string `json:"address"`
		RedeemScript string `json:"redeemScript"`
	}

	// FeeType defines a structure for getfeerate method.
	FeeType struct {
		// FeeMethod represents the estimation method to use: static, eta, mempool.
		FeeMethod string `json:"fee_method"`

		// FeeLevel represents a float between 0.0 and 1.0, representing fee slider position.
		FeeLevel float64 `json:"fee_level"`
	}

	// PayToN defines a structure for payto and paytomany methods.
	PayToN struct {
		// Transaction fee (absolute, in BTC)
		Fee float64 `json:"fee"`

		//  Transaction fee rate (in sat/byte)
		FeeRate int `json:"feerate"`

		//  Source address (must be a wallet address; use sweep to spend from non-wallet address).
		FromAddr string `json:"from_addr"`

		// Source coins (must be in wallet; use sweep to spend from non-wallet address)
		// FromCoins string `json:"from_coins"`

		// Change address. Default is a spare address, or the source address if it's not in the wallet
		// ChangeAddr string `json:"change_addr"`

		// Set locktime block number
		// Locktime int `json:"locktime"`
	}

	// ------------------------------------------------------------------------------------------

	// OnchainHistory defines a structure for onchain_history method.
	OnchainHistory struct {
		Summary            OnchainHistorySummary       `json:"summary"`
		OnchainTransaction []OnchainHistoryTransaction `json:"transactions"`
	}

	// OnchainHistorySummary defines a structure for onchain_history summary method.
	BeginEnd struct {
		BTCBalance  string `json:"BTC_balance"`
		BlockHeight int64  `json:"block_height"`
		Date        string `json:"date"` // 2021-09-07 10:53
	}

	Flow struct {
		BTCIncoming string `json:"BTC_incoming"`
		BTCOutgoing string `json:"BTC_outgoing"`
	}

	OnchainHistorySummary struct {
		Begin BeginEnd `json:"begin"`
		End   BeginEnd `json:"end"`
		Flow  Flow     `json:"flow"`
	}

	// OnchainHistoryTransaction defines a structure for onchain_history transactions method.
	OnchainHistoryTransaction struct {
		Balance            string                   `json:"bc_balance,omitempty"`
		Value              string                   `json:"bc_value,omitempty"`
		Confirmations      uint64                   `json:"confirmations,omitempty"`
		Date               string                   `json:"date,omitempty"`
		Fee                string                   `json:"fee,omitempty"`
		FeeSat             uint64                   `json:"fee_sat,omitempty"`
		Height             int64                    `json:"height,omitempty"`
		Incoming           bool                     `json:"incoming,omitempty"`
		Label              string                   `json:"label,omitempty"`
		MonotonicTimestamp uint64                   `json:"monotonic_timestamp,omitempty"`
		Timestamp          uint64                   `json:"timestamp,omitempty"`
		TxID               string                   `json:"txid,omitempty"`
		TxposInBlock       int                      `json:"txpos_in_block,omitempty"`
		Outputs            []map[string]interface{} `json:"outputs,omitempty"` //[ map[address value], ... ]
	}

	// ------------------------------------------------------------------------------------------

	// DeserializedTransaction represents the structure returned from deserialize method.
	DeserializedTransaction struct {
		Partial   bool                 `json:"partial"`
		Version   int                  `json:"version"`
		SegwitSer bool                 `json:"segwit_ser"`
		Inputs    []DeserializedInput  `json:"inputs"`
		Outputs   []DeserializedOutput `json:"outputs"`
		LockTime  uint64               `json:"lockTime"`
	}

	// DeserializedInput represents the structure returned from deserialize inputs.
	DeserializedInput struct {
		Address     *string `json:"address"`
		NumSig      uint64  `json:"num_sig"`
		PrevoutHash string  `json:"prevout_hash"`
		PrevoutN    int     `json:"prevout_n"`
		ScriptSig   string  `json:"scriptSig"`
		Sequence    uint64  `json:"sequence"`
		Type        string  `json:"type"`
		Witness     string  `json:"witness"`
	}

	// DeserializedOutput represents the structure returned from deserialize outputs.
	DeserializedOutput struct {
		Address      string `json:"address"`
		PrevoutN     int    `json:"prevout_n"`
		ScriptPubKey string `json:"scriptPubKey"`
		Type         int    `json:"type"`
		// Value represented in sats
		Value uint64 `json:"value"`
	}

	// Server represents the structure returned from getservers method.
	Server struct {
		Address string `json:"address"`
		Pruning string `json:"pruning"`
		S       string `json:"s"`
		T       string `json:"t"`
		Version string `json:"version"`
	}

	// NotifyRequest represents the structure used by electrum notify POST call.
	NotifyRequest struct {
		Address string  `json:"address"`
		Status  *string `json:"status"`
	}
)

const (
	// connection timeout
	timeout = 60
	// COIN represents 1 BTC in sats
	COIN = 100000000
)

// GetValue returns the output value as float64.
func (o *DeserializedOutput) GetValue() float64 {
	return float64(o.Value) / float64(COIN)
}

// New return a new JSON-RPC client.
func New(host string, port int, username, password string, useSSL bool) (*Client, error) {
	if host == "" {
		return nil, fmt.Errorf("missing host")
	}

	if username == "" {
		return nil, fmt.Errorf("missing username")
	}

	if password == "" {
		return nil, fmt.Errorf("missing password")
	}

	var protocol string
	httpClient := &http.Client{Timeout: timeout * time.Second}

	if useSSL {
		protocol = "https://"
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	} else {
		protocol = "http://"
	}

	return &Client{
		address:    fmt.Sprintf("%s%s:%d", protocol, host, port),
		username:   username,
		password:   password,
		httpClient: httpClient,
	}, nil
}

// request creates a new JSON-RPC request.
func (c *Client) request(method string, params interface{}) (response rpcResponse, err error) {

	// Prepare JSON request payload.
	j, err := json.Marshal(rpcRequest{time.Now().UnixNano(), method, params, "2.0"})
	if err != nil {
		return
	}
	if c.Debug {
		log.Printf("TX > %+v\n", string(j))
	}

	// Send POST request.
	req, err := http.NewRequest("POST", c.address, bytes.NewBuffer(j))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json;charset=utf-8")
	req.Header.Add("Accept", "application/json")

	// Check Authentication
	if c.username != "" || c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if c.Debug {
		log.Printf("RX < %+v\n", string(data))
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return
	}

	if res.StatusCode != 200 {
		err = fmt.Errorf("HTTP error: %d - Error: %v", res.StatusCode, response.Err)
		return
	}

	return
}

// error handles any error returned by the request method.
func (c *Client) error(err error, r *rpcResponse) error {
	if err != nil {
		return err
	}
	if r.Err != nil {
		responseError := r.Err.(map[string]interface{})
		return fmt.Errorf("(%v) %s", responseError["code"], responseError["message"])
	}
	return nil
}

// UnmarshalJSON implements custom unmarshal for json date.
func (j *JSONDate) UnmarshalJSON(b []byte) error {
	if j == nil {
		return nil
	}

	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("2006-01-02 15:04", s)
	if err != nil {
		return err
	}
	*j = JSONDate(t)
	return nil
}

// MarshalJSON returns a JSON version of JSONDate.
func (j JSONDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(j)
}

// Format prints your json date
func (j JSONDate) Format(s string) string {
	t := time.Time(j)
	return t.Format(s)
}

// AddRequest creates a payment request, using the first unused address of the wallet.
// The address will be considered as used after this operation.
// If no payment is received, the address will be considered as unused if the payment request is deleted from the wallet.
//
// configuration variables:
//
//	(set with SetConfig/GetConfig)
//
//	requests_dir          directory where a bip70 file will be written.
//	ssl_privkey           Path to your SSL private key, needed to sign the
//	                      request.
//	ssl_chain             Chain of SSL certificates, needed for signed requests.
//	                      Put your certificate at the top and the root CA at the
//	                      end
//	url_rewrite           Parameters passed to str.replace(), in order to create
//	                      the r= part of bitcoin: URIs. Example:
//	                      "('file:///var/www/','https://electrum.org/')"
//
// TODO: this method is not yet implemented.
func (c *Client) AddRequest(amount string) (err error) {
	return fmt.Errorf("not yet implemented")
}

// AddTransaction adds a transaction to the wallet history.
// TODO: this method is not yet implemented.
func (c *Client) AddTransaction(tx string) (err error) {
	return fmt.Errorf("not yet implemented")
}

// Broadcast broadcasts a transaction to the network.
func (c *Client) Broadcast(hex string) (txID string, err error) {
	r, err := c.request("broadcast", []interface{}{hex})
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &txID)
	if err != nil {
		return
	}

	return
}

// Bump the Fee for an unconfirmed Transaction
func (c *Client) BumpFee(tx string, new_fee_rate float64) (result bool, err error) {
	params := map[string]interface{}{
		"tx":           tx,           // Serialized transaction (hexadecimal)
		"new_fee_rate": new_fee_rate, // [int, float]
	}

	r, err := c.request("bumpfee", params)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &result)
	return
}

// Create a new wallet
func (c *Client) Create(path, password string, args ...string) (result Create, err error) {
	params := map[string]interface{}{
		"wallet": path,
	}

	if password != "" {
		params["password"] = password
	}

	for _, arg := range args {
		if arg == "encrypt_file" {
			params[arg] = true
		}
		if arg == "forgetconfig" {
			params[arg] = true
		}
	}

	r, err := c.request("create", params)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &result)
	return
}

// LoadWallet loads a new wallet on daemon with specified password.
func (c *Client) LoadWallet(path, password string) (result bool, err error) {
	params := map[string]interface{}{
		"wallet_path": path,
	}

	if password != "" {
		params["password"] = password
	}

	r, err := c.request("load_wallet", params)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &result)
	return
}

// CloseWallet closes the current open wallet.
func (c *Client) CloseWallet(path string) (result bool, err error) {
	params := map[string]interface{}{
		"wallet_path": path,
	}

	r, err := c.request("close_wallet", params)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &result)
	return
}

// CreateNewAddress creates a new receiving address, beyond the gap limit of the wallet.
func (c *Client) CreateNewAddress(path string) (address string, err error) {
	params := map[string]interface{}{
		"wallet": path,
	}
	r, err := c.request("createnewaddress", params)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &address)
	return
}

// Create multisig address
func (c *Client) CreateMultiSig(num int, pabkeys []string) (result MultiSig, err error) {
	params := map[string]interface{}{
		"num":     num,
		"pabkeys": pabkeys,
	}

	r, err := c.request("createmultisig", params)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &result)
	return
}

// Deserialize deserializes an hexadecimal serialized transaction.
func (c *Client) Deserialize(hex string) (result DeserializedTransaction, err error) {
	r, err := c.request("deserialize", []interface{}{hex})
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &result)
	return
}

// GetAddressBalance returns the balance of any address.
// Note: This is a walletless server query, results are not checked by SPV.
func (c *Client) GetAddressBalance(address string) (balance Balance, err error) {
	r, err := c.request("getaddressbalance", []interface{}{address})
	if err = c.error(err, &r); err != nil {
		return
	}

	var b map[string]interface{}
	err = json.Unmarshal(r.Result, &b)
	if err != nil {
		return
	}

	// Cast strings to float64
	if b["unconfirmed"] != nil {
		if v, err := strconv.ParseFloat(b["unconfirmed"].(string), 64); err == nil {
			balance.Unconfirmed = v
		}
	}

	if b["confirmed"] != nil {
		if v, err := strconv.ParseFloat(b["confirmed"].(string), 64); err == nil {
			balance.Confirmed = v
		}
	}

	return
}

// GetAddressHistory returns the transaction history of any address.
// Note: This is a walletless server query, results are not checked by SPV.
func (c *Client) GetAddressHistory(address string) (history []AddressHistory, err error) {
	r, err := c.request("getaddresshistory", []interface{}{address})
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &history)
	return
}

// GetAddressUnspent returns the UTXO list of any address.
// Note: This is a walletless server query, results are not checked by SPV.
func (c *Client) GetAddressUnspent(address string) (utxo []UTXO, err error) {
	r, err := c.request("getaddressunspent", []interface{}{address})
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &utxo)
	return
}

// GetBalance returns the balance of your wallet. TODO: test
func (c *Client) GetBalance(walletPath string) (balance Balance, err error) {
	params := map[string]interface{}{
		"wallet": walletPath,
	}

	r, err := c.request("getbalance", params)
	if err = c.error(err, &r); err != nil {
		return
	}

	var b map[string]interface{}
	err = json.Unmarshal(r.Result, &b)
	if err != nil {
		return
	}

	// Cast strings to float64
	if b["unconfirmed"] != nil {
		if v, err := strconv.ParseFloat(b["unconfirmed"].(string), 64); err == nil {
			balance.Unconfirmed = v
		}
	}

	if b["confirmed"] != nil {
		if v, err := strconv.ParseFloat(b["confirmed"].(string), 64); err == nil {
			balance.Confirmed = v
		}
	}

	if b["unmatured"] != nil {
		if v, err := strconv.ParseFloat(b["unmatured"].(string), 64); err == nil {
			balance.Unmatured = v
		}
	}
	return
}

// GetConfig returns a configuration variable.
func (c *Client) GetConfig(key string) (value interface{}, err error) {
	r, err := c.request("getconfig", []interface{}{key})
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &value)
	return
}

// GetInfo returns informations about running daemon.
func (c *Client) GetInfo() (info Info, err error) {
	r, err := c.request("getinfo", nil)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &info)
	return
}

// GetFeeRate returns current suggested fee rate (in sat/kvByte),
// according to config settings or supplied parameters.
func (c *Client) GetFeeRate(feeType FeeType) (fee int, err error) {
	args := make(map[string]interface{}, 0)

	if feeType.FeeMethod != "" {
		args["fee_method"] = feeType.FeeMethod
	}

	if feeType.FeeLevel != 0 {
		args["fee_level"] = feeType.FeeLevel
	}

	r, err := c.request("getfeerate", args)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &fee)
	return
}

// GetMasterPrivate returns your wallet's master private key.
func (c *Client) GetMasterPrivate() (privKey string, err error) {
	params := map[string]interface{}{}

	if c.walletPassword != "" {
		params["password"] = c.walletPassword
	}

	r, err := c.request("getmasterprivate", params)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &privKey)
	return
}

// GetMerkle gets merkle branch of a transaction included in a block.
// Electrum uses this to verify transactions (Simple Payment Verification).
func (c *Client) GetMerkle(txID string, height uint64) (merkle Merkle, err error) {
	r, err := c.request("getmerkle", []interface{}{txID, height})
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &merkle)
	return
}

// GetMasterPublicKey returns your wallet's master public key.
func (c *Client) GetMasterPublicKey() (mpk string, err error) {
	r, err := c.request("getmpk", nil)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &mpk)
	return
}

// GetPrivateKeys Get private keys of addresses.
func (c *Client) GetPrivateKeys(addresses ...string) (privKeys []string, err error) {
	params := map[string]interface{}{
		"address": addresses,
	}

	if c.walletPassword != "" {
		params["password"] = c.walletPassword
	}

	r, err := c.request("getprivatekeys", params)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &privKeys)
	return
}

// GetPubKeys returns the public keys for a wallet address.
func (c *Client) GetPubKeys(address string) (pubkeys []string, err error) {
	r, err := c.request("getpubkeys", []interface{}{address})
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &pubkeys)
	return
}

// GetSeed prints the generation seed of your wallet.
func (c *Client) GetSeed() (seed string, err error) {
	params := map[string]interface{}{}

	if c.walletPassword != "" {
		params["password"] = c.walletPassword
	}

	r, err := c.request("getseed", params)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &seed)
	return
}

// GetServers returns the list of available servers.
func (c *Client) GetServers() (servers []Server, err error) {
	r, err := c.request("getservers", nil)
	if err = c.error(err, &r); err != nil {
		return
	}

	var res map[string]Server

	err = json.Unmarshal(r.Result, &res)
	if err != nil {
		return
	}

	for k, v := range res {
		v.Address = k
		servers = append(servers, v)
	}

	return
}

// GetTransaction retrieves a transaction.
func (c *Client) GetTransaction(txID string) (transaction string, err error) {
	r, err := c.request("gettransaction", []interface{}{txID})
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &transaction)
	return
}

// GetUnusedAddress returns the first unused address of the wallet, or None if all addresses are used.
// An address is considered as used if it has received a transaction, or if it is used in a payment request.
func (c *Client) GetUnusedAddress() (address string, err error) {
	r, err := c.request("getunusedaddress", nil)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &address)
	return
}

// Onchain History returns the transaction history of your wallet.
func (c *Client) OnchainHistory(walletPath string, args ...string) (history OnchainHistory, err error) {
	params := map[string]interface{}{
		"wallet": walletPath,
	}

	for _, arg := range args {
		if arg == "show_addresses" {
			params[arg] = true
		}
		if arg == "show_fiat" {
			params[arg] = true
		}
	}

	r, err := c.request("onchain_history", params)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &history)
	return
}

// IsSynchronized returns wallet synchronization status.
func (c *Client) IsSynchronized() (result bool, err error) {
	r, err := c.request("is_synchronized", nil)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &result)
	return
}

// IsMine checks if address is in wallet. Return true if and only address is in wallet
func (c *Client) IsMine(address string) (result bool, err error) {
	r, err := c.request("ismine", []interface{}{address})
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &result)
	return
}

// ListAddresses returns a list of all addresses in your wallet. Use optional arguments to filter the results.
//
//	receiving       Show only receiving addresses
//	change          Show only change addresses
//	frozen          Show only frozen addresses
//	unused          Show only unused addresses
//	funded          Show only funded addresses
//	labels          Show the labels of listed addresses
//	balance         Show the balances of listed addresses
func (c *Client) ListAddresses(args ...string) (addresses []Address, err error) {
	params := make(map[string]bool, 0)
	var showLabels bool
	var showBalance bool

	// if showLabels || showBalance we expect an array of arrays,
	// otherwise the returned value is an array of strings.
	for _, arg := range args {
		if arg == "labels" {
			showLabels = true
		}

		if arg == "balance" {
			showBalance = true
		}

		params[arg] = true
	}

	r, err := c.request("listaddresses", params)
	if err = c.error(err, &r); err != nil {
		return
	}

	var j []interface{}
	err = json.Unmarshal(r.Result, &j)
	if err != nil {
		return
	}

	for _, record := range j {
		if showLabels || showBalance {
			r := record.([]interface{})
			address := Address{
				Address: r[0].(string),
			}

			if showBalance {
				address.Balance = r[1].(string)
			}

			if showLabels && !showBalance {
				address.Label = r[1].(string)
			}
			if showLabels && showBalance {
				address.Label = r[2].(string)
			}

			addresses = append(addresses, address)
		} else {
			addresses = append(addresses, Address{
				Address: record.(string),
			})
		}
	}

	return
}

// ListUnspent returns a list of unspent transaction outputs in your wallet.
func (c *Client) ListUnspent() (addresses []Unspent, err error) {
	r, err := c.request("listunspent", nil)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &addresses)
	return
}

// MakeSeed creates and returns a new seed.
func (c *Client) MakeSeed() (seed string, err error) {
	r, err := c.request("make_seed", nil)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &seed)
	return
}

// Notify watch an address, every time the address changes, an http POST is sent to the URL.
func (c *Client) Notify(address, url string) (result bool, err error) {
	params := map[string]interface{}{
		"address": address,
		"URL":     url,
	}

	r, err := c.request("notify", params)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &result)
	return
}

// Password changes the wallet password.
func (c *Client) Password(password, newpassword string) (result bool, err error) {
	params := map[string]interface{}{
		"password":     password,
		"new_password": newpassword,
	}

	r, err := c.request("password", params)
	if err = c.error(err, &r); err != nil {
		return
	}

	var res map[string]interface{}

	err = json.Unmarshal(r.Result, &res)

	result = res["password"].(bool)
	return
}

// PayTo creates a transaction.
func (c *Client) PayTo(destination, amount string, p PayToN, args ...string) (result string, err error) {
	params := map[string]interface{}{
		"destination": destination,
		"amount":      amount,
	}
	if c.walletPassword != "" {
		params["password"] = c.walletPassword
	}

	if p.FeeRate > 0 {
		params["feerate"] = p.FeeRate
	}

	if p.Fee > 0 {
		params["fee"] = p.Fee
	}

	if p.FromAddr != "" {
		params["from_addr"] = p.FromAddr
	}

	for _, arg := range args {
		if arg == "nocheck" {
			params[arg] = true
		}
		if arg == "rbf" {
			params[arg] = true
		}
		if arg == "addtransaction" {
			params[arg] = true
		}
		if arg == "unsigned" {
			params[arg] = true
		}
		if arg == "forgetconfig" {
			params[arg] = true
		}
	}

	r, err := c.request("payto", params)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &result)
	return
}

// PayToMany Create a multi-output transaction [["address", amount], ...]
func (c *Client) PayToMany(outputs [][]interface{}, p PayToN, args ...string) (result string, err error) {
	params := map[string]interface{}{
		"outputs": outputs,
	}
	if c.walletPassword != "" {
		params["password"] = c.walletPassword
	}

	if p.FeeRate > 0 {
		params["feerate"] = p.FeeRate
	}

	if p.Fee > 0 {
		params["fee"] = p.Fee
	}

	if p.FromAddr != "" {
		params["from_addr"] = p.FromAddr
	}

	for _, arg := range args {
		if arg == "nocheck" {
			params[arg] = true
		}
		if arg == "rbf" {
			params[arg] = true
		}
		if arg == "addtransaction" {
			params[arg] = true
		}
		if arg == "unsigned" {
			params[arg] = true
		}
		if arg == "forgetconfig" {
			params[arg] = true
		}
	}

	r, err := c.request("paytomany", params)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &result)
	return
}

// SetConfig sets a configuration variable. 'value' may be a string or a Python expression.
func (c *Client) SetConfig(key string, value interface{}) (result bool, err error) {
	r, err := c.request("setconfig", []interface{}{key, value})
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &result)
	return
}

// SetWalletPassword sets the password used on password required JSON-RPC calls.
func (c *Client) SetWalletPassword(password string) {
	c.walletPassword = password
}

// SignMessage signs a message with a key.
func (c *Client) SignMessage(address, message string) (signature string, err error) {
	params := map[string]interface{}{
		"address": address,
		"message": message,
	}

	if c.walletPassword != "" {
		params["password"] = c.walletPassword
	}

	r, err := c.request("signmessage", params)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &signature)
	return
}

// SignTransaction sign a transaction.
// The wallet keys will be used unless a private key is provided.
func (c *Client) SignTransaction(hex, privkey string) (err error) {
	params := map[string]interface{}{
		"tx": hex,
	}

	if c.walletPassword != "" {
		params["password"] = c.walletPassword
	}

	if privkey != "" {
		params["privkey"] = privkey
	}

	r, err := c.request("signtransaction", params)
	if err = c.error(err, &r); err != nil {
		return
	}

	return
}

// ValidateAddress checks that an address is valid.
func (c *Client) ValidateAddress(address string) (valid bool, err error) {
	r, err := c.request("validateaddress", map[string]interface{}{"address": address})
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &valid)
	return
}

// VerifyMessage verify a signature.
func (c *Client) VerifyMessage(address, message, signature string) (valid bool, err error) {
	params := map[string]interface{}{
		"address":   address,
		"signature": signature,
		"message":   message,
	}

	r, err := c.request("verifymessage", params)
	if err = c.error(err, &r); err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &valid)
	return
}

// Version returns current version.
func (c *Client) Version() (version string, err error) {
	r, err := c.request("version", nil)
	if err = c.error(err, &r); err != nil {
		return
	}
	err = json.Unmarshal(r.Result, &version)
	return
}

/*
    add_lightning_request
    add_peer
    add_request         Create a payment request, using the first unused address of the wallet
    + addtransaction      Add a transaction to the wallet history
    + broadcast           Broadcast a transaction to the network
	+ bumpfee				Bump the Fee for an unconfirmed Transaction
    changegaplimit      Change the gap limit of the wallet
    clear_invoices      Remove all invoices
    clear_ln_blacklist
    clear_requests      Remove all payment requests
    close_channel
    + close_wallet        Close wallet
    commands            List of commands
    convert_xkey        Convert xtype of a master key
    + create              Create a new wallet
    + createmultisig      Create multisig address
    + createnewaddress    Create a new receiving address, beyond the gap limit of the wallet
    decode_invoice
    decrypt             Decrypt a message encrypted with a public key
    + deserialize         Deserialize a serialized transaction
    dumpgraph
    dumpprivkeys        Deprecated
    enable_htlc_settle
    encrypt             Encrypt a message with a public key
    export_channel_backup
    freeze              Freeze address
    freeze_utxo         Freeze a UTXO so that the wallet will not spend it
    get                 Return item from wallet storage
    get_channel_ctx     return the current commitment transaction of a channel
    get_ssl_domain      Check and return the SSL domain set in ssl_keyfile and ssl_certfile
    get_tx_status       Returns some information regarding the tx
    get_watchtower_ctn  return the local watchtower's ctn of channel
    + getaddressbalance   Return the balance of any address
    + getaddresshistory   Return the transaction history of any address
    + getaddressunspent   Returns the UTXO list of any address
    getalias            Retrieve alias
    + getbalance          Return the balance of your wallet
    + getconfig           Return a configuration variable
    + getfeerate          Return current suggested fee rate (in sat/kvByte), according to config settings or supplied parameters
    + getinfo             network info
    + getmasterprivate    Get master private key
    + getmerkle           Get Merkle branch of a transaction included in a block
    getminacceptablegap Returns the minimum value for gap limit that would be sufficient to discover all known addresses in the wallet
    + getmpk              Get master public key
    getprivatekeyforpath Get private key corresponding to derivation path (address index)
    + getprivatekeys      Get private keys of addresses
    + getpubkeys          Return the public keys for a wallet address
    getrequest          Return a payment request
    + getseed             Get seed phrase
    + getservers          Return the list of known servers (candidates for connecting)
    + gettransaction      Retrieve a transaction
    + getunusedaddress    Returns the first unused address of the wallet, or None if all addresses are used
    help
    import_channel_backup
    importprivkey       Import a private key
    inject_fees
    + is_synchronized     return wallet synchronization status
    + ismine              Check if address is in wallet
    lightning_history   lightning history
    list_channels
    list_invoices
    list_peers
    list_requests       List the payment requests you made
    list_wallets        List wallets open in daemon
    + listaddresses       List wallet addresses
    listcontacts        Show your list of contacts
    + listunspent         List unspent outputs
    lnpay
    + load_wallet         Open wallet in daemon
    + make_seed           Create a seed
    nodeid
    normal_swap         Normal submarine swap: send on-chain BTC, receive on Lightning Note that your funds will be locked for 24h if you do not have enough incoming capacity
    + notify              Watch an address
    + onchain_history     Wallet onchain history
    open_channel
    + password            Change wallet password
    + payto               Create a transaction
    + paytomany           Create a multi-output transaction
    removelocaltx       Remove a 'local' transaction from the wallet, and its dependent transactions
    request_force_close Requests the remote to force close a channel
    reset_liquidity_hints
    restore             Restore a wallet from text
    reverse_swap        Reverse submarine swap: send on Lightning, receive on-chain
    rmrequest           Remove a payment request
    searchcontacts      Search through contacts, return matching entries
    serialize           Create a transaction from json inputs
    + setconfig           Set a configuration variable
    setlabel            Assign a label to an item
    + signmessage         Sign a message with a key
    signrequest         Sign payment request with an OpenAlias
    + signtransaction     Sign a transaction
    signtransaction_with_privkey	Sign a transaction
    stop                Stop daemon
    sweep               Sweep private keys
    unfreeze            Unfreeze address
    unfreeze_utxo       Unfreeze a UTXO so that the wallet might spend it
    + validateaddress     Check that an address is valid
    + verifymessage       Verify a signature
    + version             Return the version of Electrum
*/
