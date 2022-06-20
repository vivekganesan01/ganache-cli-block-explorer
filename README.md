# Ganache-CLI-Block-Explorer
Author : Vinay Awasthi

Thanks to Original Author: vivekganesan01@gmail.com
Go the https://github.com/vivekganesan01/ganache-cli-block-explorer and give a like

Ganache-cli-block-explorer is a web based block reader, which connects to your local ganache (powered by truffle) and explore the block details from the local blockchain network.

  - Explore all the transactions within the blocks
  - Gas limit
  - Pending transactions
  - Block mined details and much more ...


> Designed to help block chain learners to
> understand ganache cli in a better way.


### Tech

Ganache-CLI-Block-Explorer involves:

* [HTML CSS Bootstrap] - HTML enhanced for web apps!
* [Go lang ethereum library] - To communicate with ganache client

And of course Dillinger itself is open source with a [public repository][dill]
 on GitHub.

### Installation

`git pull the repository`

```go
go run router.go
```
`Verify the deployment by navigating to your server address in your preferred browser.`

```sh
127.0.0.1:5051
```
Note : This web application hosts in port `5051`, please make sure the port `5051` is not occupied.

```sh
Enter the ganache host and port in the welcome page Eg: http://127.0.0.1:8545, Good to Go.. Enjoy !
```
### Development

Want to contribute? Great!
Open your favourite Terminal and run these commands.

First thing:
```sh
 git checkout -b your-branch
 Make changes and create a pull request to `release` branch
```
Note: Checkout from `master`.

### Dependencies
- add go mod, open command prompt and execute the following commands
  * go mod init ganache-cli-block-explorer
  * go mod tidy

- install deps
  * go get github.com/ethereum/go-ethereum/common github.com/ethereum/go-ethereum/core/types github.com/ethereum/go-ethereum/ethclient github.com/gorilla/mux
