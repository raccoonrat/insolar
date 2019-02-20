package ethstore

import (
	"fmt"
	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/logicrunner/goplugin/foundation"
)

// EthStore
type EthStore struct {
	foundation.BaseContract
	EthHashMap map[string]StoreElem
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
		EthHashMap: make(map[string]StoreElem),
	}, nil
}

// SaveToMap create new key with value in map
func (ethStore *EthStore) SaveToMap(EthHash string, EthAddr string, Balance uint, EthTxHash string) error {

	ethStore.EthHashMap[EthHash] =
		StoreElem{
			EthHash:   EthHash,
			EthAddr:   EthAddr,
			Balance:   Balance,
			EthTxHash: EthTxHash,
			Marker:    false,
		}

	return nil
}

// VerifyEthBalance activate Eth balance
func (ethStore *EthStore) VerifyEthBalance(EthHash string, AccountRef *core.RecordRef) (uint, error) {

	if storeElem, ok := ethStore.EthHashMap[EthHash]; ok {
		if !storeElem.Marker {
			storeElem.Marker = true
			ethStore.EthHashMap[EthHash] = storeElem
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
	result := [len(ethStore.EthHashMap)]StoreElem{}
	i := 0

	for _, storeElem := range ethStore.EthHashMap {
		result[i] = storeElem
		i++
	}

	return result[:], nil
}
