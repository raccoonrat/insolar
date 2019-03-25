/*
 *    Copyright 2019 Insolar Technologies
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package ethstore

import (
	"fmt"
	"github.com/insolar/insolar/application/contract/member/signer"
	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/logicrunner/goplugin/foundation"
	"strconv"
)

// EthStore
type EthStore struct {
	foundation.BaseContract
	EthAddrMap map[string]StoreElem
}

type StoreElem struct {
	EthHash    string
	EthAddr    string
	Balance    uint
	EthTxHash  string
	AccountRef *core.RecordRef
	Marker     bool
}

// New creates EthStore
func New() (*EthStore, error) {
	return &EthStore{
		EthAddrMap: make(map[string]StoreElem),
	}, nil
}

var INSATTR_Call_API = true

// Call method for authorized calls
func (ethStore *EthStore) Call(rootDomain core.RecordRef, method string, params []byte, seed []byte, sign []byte) (interface{}, error) {

	switch method {
	case "SaveToMap":
		return ethStore.saveToMap(params)
	}

	return nil, &foundation.Error{S: "[ EthStore Call ] Unknown method"}
}

// SaveToMap create new key with value in map
func (ethStore *EthStore) saveToMap(params []byte) (interface{}, error) {

	type inputRequest struct {
		EthAddr    string `ethAddr`
		BalanceStr string `balanceStr`
		EthTxHash  string `ethTxHash`
	}

	inputJSON := new(inputRequest)

	if err := signer.UnmarshalParams(params, &inputJSON); err != nil {
		return nil, fmt.Errorf("[ saveToMap ]: %s", err.Error())
	}

	balance, err := strconv.Atoi(inputJSON.BalanceStr)
	if err != nil {
		return nil, fmt.Errorf("[ saveToMap ]: %s", err.Error())
	}

	if _, ok := ethStore.EthAddrMap[inputJSON.EthAddr]; ok {
		return nil, fmt.Errorf("[ saveToMap ]: element is already exist")
	}
	// проверить есть ли такой мембер

	ethStore.EthAddrMap[inputJSON.EthAddr] =
		StoreElem{
			EthAddr:   inputJSON.EthAddr,
			Balance:   uint(balance),
			EthTxHash: inputJSON.EthTxHash,
			Marker:    false,
		}

	return nil, nil
}

// VerifyEthBalance activate Eth balance
func (ethStore *EthStore) VerifyEthBalance(ethAddr string, accountRefStr string) (uint, error) {

	accountRef, err := core.NewRefFromBase58(accountRefStr)
	if err != nil {
		return 0, fmt.Errorf("[ VerifyEthBalance ] Failed to parse 'to' param: %s", err.Error())
	}

	if storeElem, ok := ethStore.EthAddrMap[ethAddr]; ok {
		if !storeElem.Marker {
			storeElem.Marker = true
			storeElem.AccountRef = accountRef
			ethStore.EthAddrMap[ethAddr] = storeElem
			return storeElem.Balance, nil
		} else {
			return 0, fmt.Errorf("[ VerifyEthBalance ] This ethereum hash has already been used.")
		}
	} else {
		return 0, fmt.Errorf("[ VerifyEthBalance ] No record for this ethereum hash.")
	}
}

// GetEthList return all map
func (ethStore *EthStore) GetEthList() ([]StoreElem, error) {
	result := make([]StoreElem, len(ethStore.EthAddrMap))
	i := 0

	for _, storeElem := range ethStore.EthAddrMap {
		result[i] = storeElem
		i++
	}

	return result[:], nil
}
