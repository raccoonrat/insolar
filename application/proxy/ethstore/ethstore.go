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
	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/logicrunner/goplugin/foundation"
	"github.com/insolar/insolar/logicrunner/goplugin/proxyctx"
)

type StoreElem struct {
	EthHash    string
	EthAddr    string
	Balance    uint
	EthTxHash  string
	AccountRef *core.RecordRef
	Marker     bool
}

// PrototypeReference to prototype of this contract
// error checking hides in generator
var PrototypeReference, _ = core.NewRefFromBase58("11113avYAQmWfocJSZqLuBrCwigYFDPMNeaLcXgQxNJ.11111111111111111111111111111111")

// EthStore holds proxy type
type EthStore struct {
	Reference core.RecordRef
	Prototype core.RecordRef
	Code      core.RecordRef
}

// ContractConstructorHolder holds logic with object construction
type ContractConstructorHolder struct {
	constructorName string
	argsSerialized  []byte
}

// AsChild saves object as child
func (r *ContractConstructorHolder) AsChild(objRef core.RecordRef) (*EthStore, error) {
	ref, err := proxyctx.Current.SaveAsChild(objRef, *PrototypeReference, r.constructorName, r.argsSerialized)
	if err != nil {
		return nil, err
	}
	return &EthStore{Reference: ref}, nil
}

// AsDelegate saves object as delegate
func (r *ContractConstructorHolder) AsDelegate(objRef core.RecordRef) (*EthStore, error) {
	ref, err := proxyctx.Current.SaveAsDelegate(objRef, *PrototypeReference, r.constructorName, r.argsSerialized)
	if err != nil {
		return nil, err
	}
	return &EthStore{Reference: ref}, nil
}

// GetObject returns proxy object
func GetObject(ref core.RecordRef) (r *EthStore) {
	return &EthStore{Reference: ref}
}

// GetPrototype returns reference to the prototype
func GetPrototype() core.RecordRef {
	return *PrototypeReference
}

// GetImplementationFrom returns proxy to delegate of given type
func GetImplementationFrom(object core.RecordRef) (*EthStore, error) {
	ref, err := proxyctx.Current.GetDelegate(object, *PrototypeReference)
	if err != nil {
		return nil, err
	}
	return GetObject(ref), nil
}

// New is constructor
func New() *ContractConstructorHolder {
	var args [0]interface{}

	var argsSerialized []byte
	err := proxyctx.Current.Serialize(args, &argsSerialized)
	if err != nil {
		panic(err)
	}

	return &ContractConstructorHolder{constructorName: "New", argsSerialized: argsSerialized}
}

// GetReference returns reference of the object
func (r *EthStore) GetReference() core.RecordRef {
	return r.Reference
}

// GetPrototype returns reference to the code
func (r *EthStore) GetPrototype() (core.RecordRef, error) {
	if r.Prototype.IsEmpty() {
		ret := [2]interface{}{}
		var ret0 core.RecordRef
		ret[0] = &ret0
		var ret1 *foundation.Error
		ret[1] = &ret1

		res, err := proxyctx.Current.RouteCall(r.Reference, true, "GetPrototype", make([]byte, 0), *PrototypeReference)
		if err != nil {
			return ret0, err
		}

		err = proxyctx.Current.Deserialize(res, &ret)
		if err != nil {
			return ret0, err
		}

		if ret1 != nil {
			return ret0, ret1
		}

		r.Prototype = ret0
	}

	return r.Prototype, nil

}

// GetCode returns reference to the code
func (r *EthStore) GetCode() (core.RecordRef, error) {
	if r.Code.IsEmpty() {
		ret := [2]interface{}{}
		var ret0 core.RecordRef
		ret[0] = &ret0
		var ret1 *foundation.Error
		ret[1] = &ret1

		res, err := proxyctx.Current.RouteCall(r.Reference, true, "GetCode", make([]byte, 0), *PrototypeReference)
		if err != nil {
			return ret0, err
		}

		err = proxyctx.Current.Deserialize(res, &ret)
		if err != nil {
			return ret0, err
		}

		if ret1 != nil {
			return ret0, ret1
		}

		r.Code = ret0
	}

	return r.Code, nil
}

// SaveToMap is proxy generated method
func (r *EthStore) SaveToMap(EthHash string, EthAddr string, Balance uint, EthTxHash string) error {
	var args [4]interface{}
	args[0] = EthHash
	args[1] = EthAddr
	args[2] = Balance
	args[3] = EthTxHash

	var argsSerialized []byte

	ret := [1]interface{}{}
	var ret0 *foundation.Error
	ret[0] = &ret0

	err := proxyctx.Current.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	res, err := proxyctx.Current.RouteCall(r.Reference, true, "SaveToMap", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	err = proxyctx.Current.Deserialize(res, &ret)
	if err != nil {
		return err
	}

	if ret0 != nil {
		return ret0
	}
	return nil
}

// SaveToMapNoWait is proxy generated method
func (r *EthStore) SaveToMapNoWait(EthHash string, EthAddr string, Balance uint, EthTxHash string) error {
	var args [4]interface{}
	args[0] = EthHash
	args[1] = EthAddr
	args[2] = Balance
	args[3] = EthTxHash

	var argsSerialized []byte

	err := proxyctx.Current.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	_, err = proxyctx.Current.RouteCall(r.Reference, false, "SaveToMap", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	return nil
}

// VerifyEthBalance is proxy generated method
func (r *EthStore) VerifyEthBalance(EthHash string, AccountRef *core.RecordRef) (uint, error) {
	var args [2]interface{}
	args[0] = EthHash
	args[1] = AccountRef

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 uint
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := proxyctx.Current.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := proxyctx.Current.RouteCall(r.Reference, true, "VerifyEthBalance", argsSerialized, *PrototypeReference)
	if err != nil {
		return ret0, err
	}

	err = proxyctx.Current.Deserialize(res, &ret)
	if err != nil {
		return ret0, err
	}

	if ret1 != nil {
		return ret0, ret1
	}
	return ret0, nil
}

// VerifyEthBalanceNoWait is proxy generated method
func (r *EthStore) VerifyEthBalanceNoWait(EthHash string, AccountRef *core.RecordRef) error {
	var args [2]interface{}
	args[0] = EthHash
	args[1] = AccountRef

	var argsSerialized []byte

	err := proxyctx.Current.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	_, err = proxyctx.Current.RouteCall(r.Reference, false, "VerifyEthBalance", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	return nil
}

// GetEthList is proxy generated method
func (r *EthStore) GetEthList() ([]StoreElem, error) {
	var args [0]interface{}

	var argsSerialized []byte

	ret := [2]interface{}{}
	var ret0 []StoreElem
	ret[0] = &ret0
	var ret1 *foundation.Error
	ret[1] = &ret1

	err := proxyctx.Current.Serialize(args, &argsSerialized)
	if err != nil {
		return ret0, err
	}

	res, err := proxyctx.Current.RouteCall(r.Reference, true, "GetEthList", argsSerialized, *PrototypeReference)
	if err != nil {
		return ret0, err
	}

	err = proxyctx.Current.Deserialize(res, &ret)
	if err != nil {
		return ret0, err
	}

	if ret1 != nil {
		return ret0, ret1
	}
	return ret0, nil
}

// GetEthListNoWait is proxy generated method
func (r *EthStore) GetEthListNoWait() error {
	var args [0]interface{}

	var argsSerialized []byte

	err := proxyctx.Current.Serialize(args, &argsSerialized)
	if err != nil {
		return err
	}

	_, err = proxyctx.Current.RouteCall(r.Reference, false, "GetEthList", argsSerialized, *PrototypeReference)
	if err != nil {
		return err
	}

	return nil
}
