# otc-price

this project is a simple OTC price maker that generates a price based on the real price and some random values

- connects to binance websocket to get the real price
- generates a price based on the real price and some random values
- writes the price to a file
- reads the price from the file and plots it

### Server 

run `go run main.go` to start the server


### Chart 

run `python plot.py` to see the chart
