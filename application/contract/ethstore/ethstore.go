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
	PublicKey  string
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
func New(publicKey string) (*EthStore, error) {
	return &EthStore{
		PublicKey:  publicKey,
		EthAddrMap: make(map[string]StoreElem),
	}, nil
}

var INSATTR_GetPublicKey_API = true

func (ethStore *EthStore) GetPublicKey() (string, error) {
	return ethStore.PublicKey, nil
}

func (ethStore *EthStore) verifySig(method string, params []byte, seed []byte, sign []byte) error {
	args, err := core.MarshalArgs(method, params, seed)
	if err != nil {
		return fmt.Errorf("[ verifySig ] Can't MarshalArgs: %s", err.Error())
	}
	key, err := ethStore.GetPublicKey()
	if err != nil {
		return fmt.Errorf("[ verifySig ]: %s", err.Error())
	}

	publicKey, err := foundation.ImportPublicKey(key)
	if err != nil {
		return fmt.Errorf("[ verifySig ] Invalid public key")
	}

	verified := foundation.Verify(args, sign, publicKey)
	if !verified {
		return fmt.Errorf("[ verifySig ] Incorrect signature")
	}
	return nil
}

var INSATTR_Call_API = true

// Call method for authorized calls
func (ethStore *EthStore) Call(rootDomain core.RecordRef, method string, params []byte, seed []byte, sign []byte) (interface{}, error) {

	if err := ethStore.verifySig(method, params, seed, sign); err != nil {
		return nil, fmt.Errorf("[ Call ]: %s", err.Error())
	}

	switch method {
	case "SaveToMap":
		return ethStore.saveToMap(params)
	}

	return nil, &foundation.Error{S: "Unknown method"}
}

// SaveToMap create new key with value in map
func (ethStore *EthStore) saveToMap(params []byte) (interface{}, error) {

	var ethAddr, balanceStr, ethTxHash string
	if err := signer.UnmarshalParams(params, &ethAddr, &balanceStr, &ethTxHash); err != nil {
		return nil, fmt.Errorf("[ saveToMap ]: %s", err.Error())
	}

	balance, err := strconv.Atoi(balanceStr)
	if err != nil {
		return nil, fmt.Errorf("[ saveToMap ]: %s", err.Error())
	}

	ethStore.EthAddrMap[ethAddr] =
		StoreElem{
			EthAddr:   ethAddr,
			Balance:   uint(balance),
			EthTxHash: ethTxHash,
			Marker:    false,
		}

	return nil, nil
}

// VerifyEthBalance activate Eth balance
func (ethStore *EthStore) VerifyEthBalance(params []byte) (uint, error) {

	var ethAddr, accountRefStr string
	if err := signer.UnmarshalParams(params, &ethAddr, &accountRefStr); err != nil {
		return 0, fmt.Errorf("[ VerifyEthBalance ]: %s", err.Error())
	}

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
