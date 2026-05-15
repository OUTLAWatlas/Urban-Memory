// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package main

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// UrbanLedgerMetaData contains all meta data concerning the UrbanLedger contract.
var UrbanLedgerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"layerType\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"year\",\"type\":\"uint16\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"sha256Hash\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"HashCommitted\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"admin\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_layerType\",\"type\":\"string\"},{\"internalType\":\"uint16\",\"name\":\"_year\",\"type\":\"uint16\"},{\"internalType\":\"string\",\"name\":\"_sha256Hash\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_sourceRef\",\"type\":\"string\"}],\"name\":\"commitLayerHash\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"name\":\"layerRegistry\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"sha256Hash\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"sourceRef\",\"type\":\"string\"},{\"internalType\":\"bool\",\"name\":\"exists\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_layerType\",\"type\":\"string\"},{\"internalType\":\"uint16\",\"name\":\"_year\",\"type\":\"uint16\"}],\"name\":\"verifyLayer\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// UrbanLedgerABI is the input ABI used to generate the binding from.
// Deprecated: Use UrbanLedgerMetaData.ABI instead.
var UrbanLedgerABI = UrbanLedgerMetaData.ABI

// UrbanLedger is an auto generated Go binding around an Ethereum contract.
type UrbanLedger struct {
	UrbanLedgerCaller     // Read-only binding to the contract
	UrbanLedgerTransactor // Write-only binding to the contract
	UrbanLedgerFilterer   // Log filterer for contract events
}

// UrbanLedgerCaller is an auto generated read-only Go binding around an Ethereum contract.
type UrbanLedgerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UrbanLedgerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type UrbanLedgerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UrbanLedgerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type UrbanLedgerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UrbanLedgerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type UrbanLedgerSession struct {
	Contract     *UrbanLedger      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// UrbanLedgerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type UrbanLedgerCallerSession struct {
	Contract *UrbanLedgerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// UrbanLedgerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type UrbanLedgerTransactorSession struct {
	Contract     *UrbanLedgerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// UrbanLedgerRaw is an auto generated low-level Go binding around an Ethereum contract.
type UrbanLedgerRaw struct {
	Contract *UrbanLedger // Generic contract binding to access the raw methods on
}

// UrbanLedgerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type UrbanLedgerCallerRaw struct {
	Contract *UrbanLedgerCaller // Generic read-only contract binding to access the raw methods on
}

// UrbanLedgerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type UrbanLedgerTransactorRaw struct {
	Contract *UrbanLedgerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewUrbanLedger creates a new instance of UrbanLedger, bound to a specific deployed contract.
func NewUrbanLedger(address common.Address, backend bind.ContractBackend) (*UrbanLedger, error) {
	contract, err := bindUrbanLedger(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &UrbanLedger{UrbanLedgerCaller: UrbanLedgerCaller{contract: contract}, UrbanLedgerTransactor: UrbanLedgerTransactor{contract: contract}, UrbanLedgerFilterer: UrbanLedgerFilterer{contract: contract}}, nil
}

// NewUrbanLedgerCaller creates a new read-only instance of UrbanLedger, bound to a specific deployed contract.
func NewUrbanLedgerCaller(address common.Address, caller bind.ContractCaller) (*UrbanLedgerCaller, error) {
	contract, err := bindUrbanLedger(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &UrbanLedgerCaller{contract: contract}, nil
}

// NewUrbanLedgerTransactor creates a new write-only instance of UrbanLedger, bound to a specific deployed contract.
func NewUrbanLedgerTransactor(address common.Address, transactor bind.ContractTransactor) (*UrbanLedgerTransactor, error) {
	contract, err := bindUrbanLedger(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &UrbanLedgerTransactor{contract: contract}, nil
}

// NewUrbanLedgerFilterer creates a new log filterer instance of UrbanLedger, bound to a specific deployed contract.
func NewUrbanLedgerFilterer(address common.Address, filterer bind.ContractFilterer) (*UrbanLedgerFilterer, error) {
	contract, err := bindUrbanLedger(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &UrbanLedgerFilterer{contract: contract}, nil
}

// bindUrbanLedger binds a generic wrapper to an already deployed contract.
func bindUrbanLedger(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := UrbanLedgerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_UrbanLedger *UrbanLedgerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _UrbanLedger.Contract.UrbanLedgerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_UrbanLedger *UrbanLedgerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UrbanLedger.Contract.UrbanLedgerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_UrbanLedger *UrbanLedgerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _UrbanLedger.Contract.UrbanLedgerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_UrbanLedger *UrbanLedgerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _UrbanLedger.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_UrbanLedger *UrbanLedgerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UrbanLedger.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_UrbanLedger *UrbanLedgerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _UrbanLedger.Contract.contract.Transact(opts, method, params...)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_UrbanLedger *UrbanLedgerCaller) Admin(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _UrbanLedger.contract.Call(opts, &out, "admin")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_UrbanLedger *UrbanLedgerSession) Admin() (common.Address, error) {
	return _UrbanLedger.Contract.Admin(&_UrbanLedger.CallOpts)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_UrbanLedger *UrbanLedgerCallerSession) Admin() (common.Address, error) {
	return _UrbanLedger.Contract.Admin(&_UrbanLedger.CallOpts)
}

// LayerRegistry is a free data retrieval call binding the contract method 0x96f77030.
//
// Solidity: function layerRegistry(string , uint16 ) view returns(string sha256Hash, uint256 timestamp, string sourceRef, bool exists)
func (_UrbanLedger *UrbanLedgerCaller) LayerRegistry(opts *bind.CallOpts, arg0 string, arg1 uint16) (struct {
	Sha256Hash string
	Timestamp  *big.Int
	SourceRef  string
	Exists     bool
}, error) {
	var out []interface{}
	err := _UrbanLedger.contract.Call(opts, &out, "layerRegistry", arg0, arg1)

	outstruct := new(struct {
		Sha256Hash string
		Timestamp  *big.Int
		SourceRef  string
		Exists     bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Sha256Hash = *abi.ConvertType(out[0], new(string)).(*string)
	outstruct.Timestamp = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.SourceRef = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.Exists = *abi.ConvertType(out[3], new(bool)).(*bool)

	return *outstruct, err

}

// LayerRegistry is a free data retrieval call binding the contract method 0x96f77030.
//
// Solidity: function layerRegistry(string , uint16 ) view returns(string sha256Hash, uint256 timestamp, string sourceRef, bool exists)
func (_UrbanLedger *UrbanLedgerSession) LayerRegistry(arg0 string, arg1 uint16) (struct {
	Sha256Hash string
	Timestamp  *big.Int
	SourceRef  string
	Exists     bool
}, error) {
	return _UrbanLedger.Contract.LayerRegistry(&_UrbanLedger.CallOpts, arg0, arg1)
}

// LayerRegistry is a free data retrieval call binding the contract method 0x96f77030.
//
// Solidity: function layerRegistry(string , uint16 ) view returns(string sha256Hash, uint256 timestamp, string sourceRef, bool exists)
func (_UrbanLedger *UrbanLedgerCallerSession) LayerRegistry(arg0 string, arg1 uint16) (struct {
	Sha256Hash string
	Timestamp  *big.Int
	SourceRef  string
	Exists     bool
}, error) {
	return _UrbanLedger.Contract.LayerRegistry(&_UrbanLedger.CallOpts, arg0, arg1)
}

// VerifyLayer is a free data retrieval call binding the contract method 0xfd6ca9fa.
//
// Solidity: function verifyLayer(string _layerType, uint16 _year) view returns(string, uint256, string)
func (_UrbanLedger *UrbanLedgerCaller) VerifyLayer(opts *bind.CallOpts, _layerType string, _year uint16) (string, *big.Int, string, error) {
	var out []interface{}
	err := _UrbanLedger.contract.Call(opts, &out, "verifyLayer", _layerType, _year)

	if err != nil {
		return *new(string), *new(*big.Int), *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	out2 := *abi.ConvertType(out[2], new(string)).(*string)

	return out0, out1, out2, err

}

// VerifyLayer is a free data retrieval call binding the contract method 0xfd6ca9fa.
//
// Solidity: function verifyLayer(string _layerType, uint16 _year) view returns(string, uint256, string)
func (_UrbanLedger *UrbanLedgerSession) VerifyLayer(_layerType string, _year uint16) (string, *big.Int, string, error) {
	return _UrbanLedger.Contract.VerifyLayer(&_UrbanLedger.CallOpts, _layerType, _year)
}

// VerifyLayer is a free data retrieval call binding the contract method 0xfd6ca9fa.
//
// Solidity: function verifyLayer(string _layerType, uint16 _year) view returns(string, uint256, string)
func (_UrbanLedger *UrbanLedgerCallerSession) VerifyLayer(_layerType string, _year uint16) (string, *big.Int, string, error) {
	return _UrbanLedger.Contract.VerifyLayer(&_UrbanLedger.CallOpts, _layerType, _year)
}

// CommitLayerHash is a paid mutator transaction binding the contract method 0x1b8b7ce7.
//
// Solidity: function commitLayerHash(string _layerType, uint16 _year, string _sha256Hash, string _sourceRef) returns()
func (_UrbanLedger *UrbanLedgerTransactor) CommitLayerHash(opts *bind.TransactOpts, _layerType string, _year uint16, _sha256Hash string, _sourceRef string) (*types.Transaction, error) {
	return _UrbanLedger.contract.Transact(opts, "commitLayerHash", _layerType, _year, _sha256Hash, _sourceRef)
}

// CommitLayerHash is a paid mutator transaction binding the contract method 0x1b8b7ce7.
//
// Solidity: function commitLayerHash(string _layerType, uint16 _year, string _sha256Hash, string _sourceRef) returns()
func (_UrbanLedger *UrbanLedgerSession) CommitLayerHash(_layerType string, _year uint16, _sha256Hash string, _sourceRef string) (*types.Transaction, error) {
	return _UrbanLedger.Contract.CommitLayerHash(&_UrbanLedger.TransactOpts, _layerType, _year, _sha256Hash, _sourceRef)
}

// CommitLayerHash is a paid mutator transaction binding the contract method 0x1b8b7ce7.
//
// Solidity: function commitLayerHash(string _layerType, uint16 _year, string _sha256Hash, string _sourceRef) returns()
func (_UrbanLedger *UrbanLedgerTransactorSession) CommitLayerHash(_layerType string, _year uint16, _sha256Hash string, _sourceRef string) (*types.Transaction, error) {
	return _UrbanLedger.Contract.CommitLayerHash(&_UrbanLedger.TransactOpts, _layerType, _year, _sha256Hash, _sourceRef)
}

// UrbanLedgerHashCommittedIterator is returned from FilterHashCommitted and is used to iterate over the raw logs and unpacked data for HashCommitted events raised by the UrbanLedger contract.
type UrbanLedgerHashCommittedIterator struct {
	Event *UrbanLedgerHashCommitted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *UrbanLedgerHashCommittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UrbanLedgerHashCommitted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(UrbanLedgerHashCommitted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *UrbanLedgerHashCommittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UrbanLedgerHashCommittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UrbanLedgerHashCommitted represents a HashCommitted event raised by the UrbanLedger contract.
type UrbanLedgerHashCommitted struct {
	LayerType  string
	Year       uint16
	Sha256Hash string
	Timestamp  *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterHashCommitted is a free log retrieval operation binding the contract event 0xc3221f60a3b76fb3baddf397ed6d8d20608a76a7c2da4fb9bde46a324e7a01f1.
//
// Solidity: event HashCommitted(string layerType, uint16 year, string sha256Hash, uint256 timestamp)
func (_UrbanLedger *UrbanLedgerFilterer) FilterHashCommitted(opts *bind.FilterOpts) (*UrbanLedgerHashCommittedIterator, error) {

	logs, sub, err := _UrbanLedger.contract.FilterLogs(opts, "HashCommitted")
	if err != nil {
		return nil, err
	}
	return &UrbanLedgerHashCommittedIterator{contract: _UrbanLedger.contract, event: "HashCommitted", logs: logs, sub: sub}, nil
}

// WatchHashCommitted is a free log subscription operation binding the contract event 0xc3221f60a3b76fb3baddf397ed6d8d20608a76a7c2da4fb9bde46a324e7a01f1.
//
// Solidity: event HashCommitted(string layerType, uint16 year, string sha256Hash, uint256 timestamp)
func (_UrbanLedger *UrbanLedgerFilterer) WatchHashCommitted(opts *bind.WatchOpts, sink chan<- *UrbanLedgerHashCommitted) (event.Subscription, error) {

	logs, sub, err := _UrbanLedger.contract.WatchLogs(opts, "HashCommitted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UrbanLedgerHashCommitted)
				if err := _UrbanLedger.contract.UnpackLog(event, "HashCommitted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseHashCommitted is a log parse operation binding the contract event 0xc3221f60a3b76fb3baddf397ed6d8d20608a76a7c2da4fb9bde46a324e7a01f1.
//
// Solidity: event HashCommitted(string layerType, uint16 year, string sha256Hash, uint256 timestamp)
func (_UrbanLedger *UrbanLedgerFilterer) ParseHashCommitted(log types.Log) (*UrbanLedgerHashCommitted, error) {
	event := new(UrbanLedgerHashCommitted)
	if err := _UrbanLedger.contract.UnpackLog(event, "HashCommitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
