# BlockChain Tutorial

## Usage

```bash
    # prints blockchain
    go run main.go print

    # initialize blockchain
    go run main.go init -address ADDRESS

    # send coins
    go run main.go send -from FROM -to TO -amount AMOUNT

    # get balance
    go run main.go getbalance -address ADDRESS
    
    # create wallet
    go run main.go createwallet
    
    # list addresses
    go run main.go listaddresses
```