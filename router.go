/**

 */
package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

// *********************** variable ********************************************

// NetworkHost for holding the host
var NetworkHost = "http://localhost:8545" // Ganache host

var client *ethclient.Client // for client to access globally

// *********************** structs *********************************************

// for overall ganache statistics
type sysInfo struct {
	NumBlock                string
	NetworkID               *big.Int
	PendingTransactionCount uint
	SuggestedGasPrice       *big.Int
	BlockDetails            []blockInfo
	AccountDetails          []accountInfo
}

// for block details
type blockInfo struct {
	Block           string
	BlockHash       string
	BlockNonce      uint64
	Transactions    int
	Transactionhash string
	GasUsed         uint64
	MinedOn         time.Time
	Difficulty      *big.Int
	Size            common.StorageSize
	Gaslimit        uint64
	ParentHash      string
	UncleHash       string
}

// for ganache Default Account Details
type accountInfo struct {
	AccAddress  string
	AccBalance  string
	AccTXNCount uint64
	AccIndex    int
}

// for ganache Default Account Details
type accDetails struct {
	AccAddress  string
	AccBalance  string
	AccTXNCount uint64
}

// for transaction details
type txDetails struct {
	TxHash        string
	TxGas         uint64
	TxGasPrice    uint64
	TxNonce       uint64
	TxToAddress   string
	TxFromAddress string
	TxData        string
	TxValue       *big.Int
}

// for transaction details
type txPages struct {
	BlockHash         string
	BlockNumber       *big.Int
	Totaltransactions int
	TxDetails         []txDetails
}

// for error logs
type txLogs struct {
	Status   uint64
	Log      string
	ErrorMsg error
	Host     string
}

// *********************** Utility ******************************************
func weiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(params.Ether))
}

// *********************** block details ***************************************

/*
	blockInDetails function: fetches the block details based on hash

*/
func blockInDetails(w http.ResponseWriter, r *http.Request) {
	/* local variables */
	var blockHash common.Hash

	// parsing the request
	for _, qs := range r.URL.Query() {
		blockHash = common.HexToHash(qs[0])
	}

	// client request for the block
	blockDetails, blockByHashErr := client.BlockByHash(context.Background(), blockHash)
	kickBack(blockByHashErr,
		"Reason: `@BlockByHash` failed. Couldn't able to fetch block.")

	// block creation time
	creationTime := time.Unix(int64(blockDetails.Time()), 0)

	// loading data for rendering
	data := blockInfo{
		Block:        blockDetails.Number().String(),
		BlockHash:    blockDetails.Hash().Hex(),
		BlockNonce:   blockDetails.Nonce(),
		Transactions: len(blockDetails.Transactions()),
		GasUsed:      blockDetails.GasUsed(),
		MinedOn:      creationTime,
		Difficulty:   blockDetails.Difficulty(),
		Size:         blockDetails.Size(),
		Gaslimit:     blockDetails.GasLimit(),
		ParentHash:   blockDetails.ParentHash().String(),
		UncleHash:    blockDetails.UncleHash().String(),
	}

	// render
	tmpl := template.Must(template.ParseFiles("template/blockDetails.html"))
	tmpl.Execute(w, data)
}

// *********************** blockshomepage **************************************

/*
	blockPage function: fetches the block details based on number for the block
	page

*/
func blockPage(w http.ResponseWriter, bn *big.Int) blockInfo {

	// getting block based on given number
	block, _ := client.BlockByNumber(context.Background(), bn)
	// kickBack(w, r, blockByNumberErr,
	// 	"Reason: `@BlockByNumber` failed. Couldn't able to fetch block.")

	// block creation time
	creationTime := time.Unix(int64(block.Time()), 0)
	var tempTxn string
	_ = tempTxn
	// getting transaction details
	for _, tx := range block.Transactions() {
		tempTxn = tx.Hash().String()
	}
	// loading data for rendering
	blockData := blockInfo{
		Block:           bn.String(),
		BlockHash:       block.Hash().String(),
		BlockNonce:      block.Nonce(),
		Transactions:    len(block.Transactions()),
		Transactionhash: tempTxn,
		GasUsed:         block.GasUsed(),
		MinedOn:         creationTime,
	}
	return blockData
}

// *********************** homepage **************************************

/*
	accountsBalance function: fetches the account details and their balance

*/
func getAccountDetails(account common.Address, itr int) accountInfo {

	// load all the block details
	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		log.Fatal(err)
	}

	balanceETH := weiToEther(balance)
	//fmt.Println(balanceETH)
	// Here it fetches the latest block for the connected client (i.e., ganache)
	numBlock, headerByNumberErr := client.HeaderByNumber(context.Background(), nil)
	kickBack(headerByNumberErr, "Reason:`@HeaderByNumber` failed. Make sure GANACHE runs @ localhost")
	nonce, _ := client.NonceAt(context.Background(), account, numBlock.Number)
	//fmt.Println(state)
	// loading account data for rendering
	accountData := accountInfo{
		AccAddress:  account.String(),
		AccBalance:  balanceETH.String() + " ETH",
		AccTXNCount: nonce,
		AccIndex:    itr,
	}

	return accountData
}

/*
	accountsBalance function: fetches the account details and their balance

*/
func showBalanceInfo(w http.ResponseWriter, r *http.Request) {
	var qss string

	// parsing the request
	for _, qs := range r.URL.Query() {
		qss = qs[0]
	}

	// load all the block details
	balance, err := client.BalanceAt(context.Background(), common.BytesToAddress(common.FromHex(qss)), nil)
	if err != nil {
		log.Fatal(err)
	}

	balanceETH := weiToEther(balance)
	//fmt.Println(balanceETH)
	// Here it fetches the latest block for the connected client (i.e., ganache)
	numBlock, headerByNumberErr := client.HeaderByNumber(context.Background(), nil)
	kickBack(headerByNumberErr, "Reason:`@HeaderByNumber` failed. Make sure GANACHE runs @ localhost")
	nonce, _ := client.NonceAt(context.Background(), common.BytesToAddress(common.FromHex(qss)), numBlock.Number)
	//fmt.Println(state)
	// loading account data for rendering
	accountData := accDetails{
		AccAddress:  qss,
		AccBalance:  balanceETH.String() + " ETH",
		AccTXNCount: nonce,
	}

	// render
	tmpl := template.Must(template.ParseFiles("template/checkBalance.html"))
	tmpl.Execute(w, accountData)

}

// *********************** txpage **********************************************

/*
	txPage function: provide the complete transaction details based on the
	block number or block hash.

*/
func txPage(w http.ResponseWriter, r *http.Request) {
	/* local variables */
	var qss string
	var block *types.Block
	var listTxDetails []txDetails
	var err error
	var toAddress string
	var execStatus bool
	var log txLogs

	// parsing the request
	for _, qs := range r.URL.Query() {
		qss = qs[0]
	}
	bn, strConvErr := strconv.Atoi(qss) // converting string into number to pass in client call

	// has to accept either number or hash, so validating
	// TODO: what if some other happens, has to validate the err
	if strConvErr != nil {
		hash := common.HexToHash(qss)

		// getting block with hash
		block, err = client.BlockByHash(context.Background(), hash)

		// check whether block number exists or not
		if err != nil {
			execStatus = true
			log = txLogs{
				Status:   404,
				Log:      "Block with given hash is not available in the network",
				ErrorMsg: err,
				Host:     "homepage",
			}
		} else {
			execStatus = false
		}

	} else {

		// getting block with number
		block, err = client.BlockByNumber(context.Background(), big.NewInt(int64(bn)))

		// check whether block hash exists or not
		if err != nil {
			execStatus = true
			log = txLogs{
				Status:   404,
				Log:      "Block with given number is not available in the network",
				ErrorMsg: err,
				Host:     "homepage",
			}
		} else {
			execStatus = false
		}
	}

	// based on block availability execute
	if execStatus {
		// render
		tmpl := template.Must(template.ParseFiles("template/404.html"))
		tmpl.Execute(w, log)

	} else {

		// getting transaction details
		for _, tx := range block.Transactions() {
			// check for toAddress
			receipt, _ := client.TransactionReceipt(context.Background(), tx.Hash())
			if tx.To() == nil {
				toAddress = receipt.ContractAddress.Hex() + " [CONTRACT CREATION]"
			} else {
				toAddress = tx.To().Hex()
			}

			signer := types.NewEIP155Signer(tx.ChainId())
			sender, _ := signer.Sender(tx)

			dt := txDetails{
				TxHash:        tx.Hash().Hex(),
				TxGas:         tx.Gas(),
				TxGasPrice:    tx.GasPrice().Uint64(),
				TxNonce:       tx.Nonce(),
				TxToAddress:   toAddress,
				TxFromAddress: sender.Hex(),
				TxData:        hex.EncodeToString(tx.Data()),
				TxValue:       tx.Value(),
			}
			// since transaction are multiple, loading it into an array
			listTxDetails = append(listTxDetails, dt)
		}

		// updating final data into struct for rendering
		data := txPages{
			BlockNumber:       block.Number(),
			BlockHash:         block.Hash().Hex(),
			Totaltransactions: 1,
			TxDetails:         listTxDetails,
		}

		// render
		tmpl := template.Must(template.ParseFiles("template/txPage.html"))
		tmpl.Execute(w, data)
	}
}

// /*
// 	txDetailsPage function: provide the complete transaction details based on the
// 	transaction hash.

// */
func txDetailsPage(w http.ResponseWriter, r *http.Request) {
	/* local variables */
	var qss string
	var tx *types.Transaction
	var listTxDetails []txDetails
	var err error
	var toAddress string

	var execStatus bool
	var log txLogs
	var receipt *types.Receipt
	// parsing the request
	for _, qs := range r.URL.Query() {
		qss = qs[0]
	}
	_, strConvErr := strconv.Atoi(qss) // converting string into number to pass in client call

	// has to accept either number or hash, so validating
	// TODO: what if some other happens, has to validate the err
	if strConvErr != nil {
		hash := common.HexToHash(qss)

		// getting txn with hash
		tx, _, err = client.TransactionByHash(context.Background(), hash)
		receipt, _ = client.TransactionReceipt(context.Background(), hash)

		//		fmt.Println("Value : ", tx.Value())
		//		fmt.Println("contract address : ", receipt.ContractAddress)

		// check whether block number exists or not
		if err != nil {
			execStatus = true
			log = txLogs{
				Status:   404,
				Log:      "Txn with given hash is not available in the network",
				ErrorMsg: err,
				Host:     "homepage",
			}
		} else {
			execStatus = false
		}

	}

	// based on block availability execute
	if execStatus || err != nil {
		// render
		tmpl := template.Must(template.ParseFiles("template/404.html"))
		tmpl.Execute(w, log)

	} else {

		// getting transaction details

		// check for toAddress
		if tx.To() == nil {
			toAddress = receipt.ContractAddress.Hex() + " [CONTRACT CREATION]"
		} else {
			toAddress = tx.To().Hex()
		}

		signer := types.NewEIP155Signer(tx.ChainId())
		sender, _ := signer.Sender(tx)

		dt := txDetails{
			TxHash:        tx.Hash().Hex(),
			TxGas:         tx.Gas(),
			TxGasPrice:    tx.GasPrice().Uint64(),
			TxNonce:       tx.Nonce(),
			TxToAddress:   toAddress,
			TxFromAddress: sender.Hex(),
			TxData:        hex.EncodeToString(tx.Data()),
			TxValue:       tx.Value(),
		}
		// since transaction are multiple, loading it into an array
		listTxDetails = append(listTxDetails, dt)

		// updating final data into struct for rendering
		data := txPages{
			BlockNumber:       receipt.BlockNumber,
			BlockHash:         receipt.BlockHash.Hex(),
			Totaltransactions: 1,
			TxDetails:         listTxDetails,
		}

		// render
		tmpl := template.Must(template.ParseFiles("template/txPage.html"))
		tmpl.Execute(w, data)
	}
}

// *********************** txDetails *******************************************

// *********************** On Account of failure *******************************

/*
	kickBack function: kickback to 404 if any invalid request or failure happens

*/
func kickBackErr(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("template/404.html"))
	tmpl.Execute(w, nil)
}
func kickBack(err error, msg string) {
	if err != nil {
		fmt.Println("/******** ERROR ********************************************/")
		fmt.Printf("Error: %v", err)
		fmt.Printf("Reason: %v", msg)
		panic(err)
	}
}

// *********************** homepage ********************************************

/*
	homePage function: serves the content for the main home page.

*/
func homePage(w http.ResponseWriter, r *http.Request) {
	/* local variables */
	var _blockdetails []blockInfo     // to hold the blockNumber
	var _accountDetails []accountInfo // to hold the blockNumber

	var clientErr error

	// parsing the request
	for _, qs := range r.URL.Query() {
		NetworkHost = qs[0]
	}

	// updating the client
	client, clientErr = ethclient.Dial(NetworkHost)

	if clientErr != nil {
		log := txLogs{
			Status:   404,
			Log:      "Host provided is invalid. Client Error",
			ErrorMsg: clientErr,
		}
		tmpl := template.Must(template.ParseFiles("template/404.html"))
		tmpl.Execute(w, log)
	} else {
		// Here it fetches the latest block for the connected client (i.e., ganache)
		numBlock, headerByNumberErr := client.HeaderByNumber(context.Background(), nil)
		kickBack(headerByNumberErr, "Reason:`@HeaderByNumber` failed. Make sure GANACHE runs @ localhost")
		// Here it fetches the NetworkID for the connected client (i.e., ganache)
		networkID, networkIDErr := client.NetworkID(context.Background())
		kickBack(networkIDErr, "Reason: `@NetworkID` failed. Make sure GANACHE runs @ localhost")
		// Here it fetches the pending transaction for the connected client (i.e., ganache)
		pendingTxCount, _ := client.PendingTransactionCount(context.Background())
		// Here it fetches the suggested gas price for the connected client (i.e., ganache)
		suggestedGasPrice, suggestGasPriceError := client.SuggestGasPrice(context.Background())
		kickBack(suggestGasPriceError, "Reason: `@SuggestGasPrice` failed. Couldn't able to fetch Suggested Gas Price")

		// Here it fetches only the lasted 5 block for the home page
		for x := numBlock.Number.Int64(); x > (numBlock.Number.Int64() - 5); x-- {
			if x < 1 {
				// Todo : break here to overcome negativity
				break
			} else {
				// load all the block details
				_blockdetails = append(_blockdetails, blockPage(w, big.NewInt(x)))
			}
		}

		clientGanache := newClient(NetworkHost)
		var accounts []string

		err := clientGanache.call("eth_accounts", &accounts)

		if err != nil {
			log.Fatal(err)
		}
		//fmt.Println(accounts)
		// getting account details
		itr := 0
		for _, acc := range accounts {
			// check for toAddress
			account := common.HexToAddress(acc)
			_accountDetails = append(_accountDetails, getAccountDetails(account, itr))
			itr += 1
		}
		// data: values to be rendered
		data := sysInfo{
			NumBlock:                numBlock.Number.String(),
			NetworkID:               networkID,
			PendingTransactionCount: pendingTxCount,
			SuggestedGasPrice:       suggestedGasPrice,
			BlockDetails:            _blockdetails,
			AccountDetails:          _accountDetails,
		}

		// mux render
		tmpl := template.Must(template.ParseFiles("template/index.html"))
		tmpl.Execute(w, data)
	}
}

// *********************** welcome page ****************************************

/*
	welcomePage function: serves the welcome page.

*/
func welcomePage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("template/welcome.html"))
	tmpl.Execute(w, nil)
}

// *********************** main ************************************************
/*
	main: Main Handler, handles all the incoming request and maps for a route.

*/
func main() {

	fmt.Println("!!!!INITIALIZING SERVER!!!!")
	// mux router
	gorilla := mux.NewRouter()

	// network client activation
	client, _ = ethclient.Dial(NetworkHost)

	// for the static file handling, all the assets files will be loaded into the static folder
	staticFileHandler := http.FileServer(http.Dir("static"))

	// routes the all the static accessing url to the static folder
	gorilla.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticFileHandler))

	// controller
	gorilla.HandleFunc("/homepage", homePage)
	gorilla.HandleFunc("/txpage", txPage)
	gorilla.HandleFunc("/txinfo", txDetailsPage)
	gorilla.HandleFunc("/blockdetails", blockInDetails)
	gorilla.HandleFunc("/accInfo", showBalanceInfo)
	gorilla.HandleFunc("/", welcomePage)

	// http server
	// Note: Here gorilla is like passing our own server handler into net/http, by default its false
	srv := &http.Server{
		Handler: gorilla,
		Addr:    "127.0.0.1:5051",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Println("!!!! SERVER STARTED at ADDRESS : 127.0.0.1:5051 !!!!")
	log.Fatal(srv.ListenAndServe())
}
