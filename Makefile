lint:
	golangci-lint run

gen:
	abigen --abi=multicall_abi.json --pkg=multicall --out=multicall_contract.go
