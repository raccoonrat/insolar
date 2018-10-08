/*
 *    Copyright 2018 Insolar
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

package hash

import (
	"hash"
	"io"

	"golang.org/x/crypto/sha3"
)

// Writer is the interface that wraps the WriteHash method.
//
// WriteHash should write all required for proper hashing data to io.Writer.
type Writer interface {
	WriteHash(io.Writer)
}

// SHA3hash224 returns SHA3 hash calculated on data received from Writer.
func SHA3hash224(hw ...Writer) []byte {
	h := sha3.New224()
	for _, w := range hw {
		w.WriteHash(h)
	}
	return h.Sum(nil)
}

// NewIDHash returns hash used for records ID generation.
func NewIDHash() hash.Hash {
	return sha3.New224()
}

// SHA3Bytes224 generates SHA3-224 hash for byte slice.
func SHA3Bytes224(b []byte) []byte {
	return SHA3hash224(hashableBytes(b))
}

// hashableBytes exists just to allow []byte implements hash.Writer
type hashableBytes []byte

func (b hashableBytes) WriteHash(w io.Writer) {
	_, err := w.Write(b)
	if err != nil {
		panic(err)
	}
}
