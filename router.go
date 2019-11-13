/**

 */
package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/mux"
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
}

// for block details
type blockInfo struct {
	Block        string
	BlockHash    string
	BlockNonce   uint64
	Transactions int
	GasUsed      uint64
	MinedOn      time.Time
	Difficulty   *big.Int
	Size         common.StorageSize
	Gaslimit     uint64
	ParentHash   string
	UncleHash    string
}

// for transaction details
type txDetails struct {
	TxHash      string
	TxGas       uint64
	TxGasPrice  uint64
	TxNonce     uint64
	TxToAddress string
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

// *********************** dependency ******************************************

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

	// loading data for rendering
	blockData := blockInfo{
		Block:        bn.String(),
		BlockHash:    block.Hash().String(),
		BlockNonce:   block.Nonce(),
		Transactions: len(block.Transactions()),
		GasUsed:      block.GasUsed(),
		MinedOn:      creationTime,
	}
	return blockData
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
			if tx.To() == nil {
				toAddress = "Nil (No Tx)"
			} else {
				toAddress = tx.To().Hex()
			}

			dt := txDetails{
				TxHash:      tx.Hash().Hex(),
				TxGas:       tx.Gas(),
				TxGasPrice:  tx.GasPrice().Uint64(),
				TxNonce:     tx.Nonce(),
				TxToAddress: toAddress,
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

// *********************** txDetails *******************************************

/*
	txDetailPage function: provide the block transaction details based on the
	transaction hash.

*/
func txDetailPage(w http.ResponseWriter, r *http.Request) {
	/* local variables */
	var txHash common.Hash // to store transaction hash
	var log string         // to store logs message

	// parsing request
	for _, qs := range r.URL.Query() {
		txHash = common.HexToHash(qs[0])
	}

	// getting tx details
	receipt, transactionReceiptErr := client.TransactionReceipt(context.Background(), txHash)
	kickBack(transactionReceiptErr,
		"Reason:`@TransactionReceipt` failed. Please pass a valid transaction hash ...")

	// checking if any logs in transaction
	if len(receipt.Logs) <= 0 {
		log = "Null"
	} else {
		for _, unpackLogs := range receipt.Logs {
			log = string(unpackLogs.Data)
		}
	}

	// load data into struct for rendering
	data := txLogs{
		Status: receipt.Status,
		Log:    log,
	}

	// render
	tmpl := template.Must(template.ParseFiles("template/txDetails.html"))
	tmpl.Execute(w, data)
}

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
	var _blockdetails []blockInfo // to hold the blockNumber
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

		// data: values to be rendered
		data := sysInfo{
			NumBlock:                numBlock.Number.String(),
			NetworkID:               networkID,
			PendingTransactionCount: pendingTxCount,
			SuggestedGasPrice:       suggestedGasPrice,
			BlockDetails:            _blockdetails,
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
	gorilla.HandleFunc("/txdetails", txDetailPage)
	gorilla.HandleFunc("/blockdetails", blockInDetails)
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
	log.Fatal(srv.ListenAndServe())
}
