package multicall

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// see https://github.com/makerdao/multicall

const (
	ETHMainnet = iota
	Kovan
	Rinkeby
	Gorli
	Ropsten
	XDai
	Polygon
	Mumbai
	BSCMainnet
	BSCTestnet
)

var (
	V1 = map[int]common.Address{
		ETHMainnet: common.HexToAddress("0xeefba1e63905ef1d7acba5a8513c70307c1ce441"),
		Kovan:      common.HexToAddress("0x2cc8688c5f75e365aaeeb4ea8d6a480405a48d2a"),
		Rinkeby:    common.HexToAddress("0x42ad527de7d4e9d9d011ac45b31d8551f8fe9821"),
		Gorli:      common.HexToAddress("0x77dca2c955b15e9de4dbbcf1246b4b85b651e50e"),
		Ropsten:    common.HexToAddress("0x53c43764255c17bd724f74c4ef150724ac50a3ed"),
		XDai:       common.HexToAddress("0xb5b692a88bdfc81ca69dcb1d924f59f0413a602a"),
		Polygon:    common.HexToAddress("0x11ce4B23bD875D7F5C6a31084f55fDe1e9A87507"),
		Mumbai:     common.HexToAddress("0x08411ADd0b5AA8ee47563b146743C13b3556c9Cc"),
	}

	V2 = map[int]common.Address{
		ETHMainnet: common.HexToAddress("0x5ba1e12693dc8f9c48aad8770482f4739beed696"),
		Kovan:      common.HexToAddress("0x5ba1e12693dc8f9c48aad8770482f4739beed696"),
		Rinkeby:    common.HexToAddress("0x5ba1e12693dc8f9c48aad8770482f4739beed696"),
		Gorli:      common.HexToAddress("0x5ba1e12693dc8f9c48aad8770482f4739beed696"),
		Ropsten:    common.HexToAddress("0x5ba1e12693dc8f9c48aad8770482f4739beed696"),
		BSCMainnet: common.HexToAddress("0x41263cba59eb80dc200f3e2544eda4ed6a90e76c"),
		BSCTestnet: common.HexToAddress("0xae11C5B5f29A6a25e955F0CB8ddCc416f522AF5C"),
	}
)

type Call struct {
	Contract common.Address
	ABI      abi.ABI
	Method   string
	Params   []any
	Result   any
}

type Calls []Call

func NewCalls(calls ...Call) Calls {
	return calls
}

func (calls Calls) Warp(warpHandler func(Calls, []Multicall2Call) error) error {
	batch := make([]Multicall2Call, 0, len(calls))
	for i, call := range calls {
		encodeData, err := calls[i].ABI.Pack(call.Method, call.Params...)
		if err != nil {
			return err
		}

		batch = append(batch, Multicall2Call{
			Target:   call.Contract,
			CallData: encodeData,
		})
	}
	return warpHandler(calls, batch)
}

func (calls Calls) WarpHandlerForTryAggregate(contract *Multicall, opts *bind.CallOpts, requireSuccess bool) func(Calls, []Multicall2Call) error {
	return func(calls Calls, batch []Multicall2Call) error {
		results, err := contract.TryAggregate(opts, requireSuccess, batch)
		if err != nil {
			return err
		}

		for i, result := range results {
			if !result.Success {
				continue
			}

			if err := calls[i].ABI.UnpackIntoInterface(calls[i].Result, calls[i].Method, result.ReturnData); err != nil {
				return err
			}
		}

		return nil
	}
}

func (calls Calls) WarpHandlerForAggregate(contract *Multicall, opts *bind.CallOpts) func(Calls, []Multicall2Call) error {
	return func(calls Calls, batch []Multicall2Call) error {
		results, err := contract.Aggregate(opts, batch)
		if err != nil {
			return err
		}

		for i, returnData := range results.ReturnData {
			if err := calls[i].ABI.UnpackIntoInterface(calls[i].Result, calls[i].Method, returnData); err != nil {
				return err
			}
		}

		return nil
	}
}
