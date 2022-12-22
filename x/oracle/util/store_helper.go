package util

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/store"
)

func GetFirstValueInRange[T any](storeObj store.KVStore, keyStart []byte, keyEnd []byte, reverseIterate bool, parseValue func([]byte) (T, error)) (T, error) {
	iterator := makeIterator(storeObj, keyStart, keyEnd, reverseIterate)
	defer iterator.Close()

	if !iterator.Valid() {
		var blankValue T
		return blankValue, errors.New("no values in range")
	}

	return parseValue(iterator.Value())
}

func RemoveFirstValueInRange(storeObj store.KVStore, keyStart []byte, keyEnd []byte, reverseIterate bool) error {
	iterator := makeIterator(storeObj, keyStart, keyEnd, reverseIterate)
	defer iterator.Close()

	if !iterator.Valid() {
		return errors.New("no values in range")
	}

	storeObj.Delete(iterator.Key())
	return nil
}

func makeIterator(storeObj store.KVStore, keyStart []byte, keyEnd []byte, reverse bool) store.Iterator {
	if reverse {
		return storeObj.ReverseIterator(keyStart, keyEnd)
	}
	return storeObj.Iterator(keyStart, keyEnd)
}
