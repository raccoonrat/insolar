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

package pulsar

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/gob"
	"errors"
	"math/big"
	"net"
	"net/rpc"

	"github.com/insolar/insolar/configuration"
	"github.com/insolar/insolar/log"
	"golang.org/x/crypto/sha3"
)

type Pulsar struct {
	Sock               net.Listener
	SockConnectionType configuration.ConnectionType
	RPCServer          *rpc.Server

	Neighbours map[string]*Neighbour
	PrivateKey *ecdsa.PrivateKey
}

// Creation new pulsar-node
func NewPulsar(configuration configuration.Pulsar, listener func(string, string) (net.Listener, error)) (*Pulsar, error) {
	// Listen for incoming connections.
	l, err := listener(configuration.ConnectionType.String(), configuration.ListenAddress)
	if err != nil {
		return nil, err
	}

	// Parse private key from config
	privateKey, err := ImportPrivateKey(configuration.PrivateKey)
	if err != nil {
		return nil, err
	}
	pulsar := &Pulsar{Sock: l, Neighbours: map[string]*Neighbour{}, SockConnectionType: configuration.ConnectionType}
	pulsar.PrivateKey = privateKey

	// Adding other pulsars
	for _, neighbour := range configuration.ListOfNeighbours {
		if len(neighbour.PublicKey) == 0 {
			continue
		}
		publicKey, err := ImportPublicKey(neighbour.PublicKey)
		if err != nil {
			continue
		}
		pulsar.Neighbours[neighbour.PublicKey] = &Neighbour{
			ConnectionType:    neighbour.ConnectionType,
			ConnectionAddress: neighbour.Address,
			PublicKey:         publicKey}
	}

	gob.Register(Payload{})
	gob.Register(HandshakePayload{})

	return pulsar, nil
}

func (pulsar *Pulsar) Start() {
	server := rpc.NewServer()

	err := server.RegisterName("Pulsar", &Handler{pulsar: pulsar})
	if err != nil {
		panic(err)
	}
	pulsar.RPCServer = server
	server.Accept(pulsar.Sock)
}

func (pulsar *Pulsar) Close() {
	for _, neighbour := range pulsar.Neighbours {
		if neighbour.OutgoingClient != nil {
			err := neighbour.OutgoingClient.Close()
			if err != nil {
				log.Error(err)
			}
		}
	}

	err := pulsar.Sock.Close()
	if err != nil {
		log.Error(err)
	}
}

func (pulsar *Pulsar) EstablishConnection(pubKey string) error {
	neighbour, err := pulsar.fetchNeighbour(pubKey)
	if err != nil {
		return err
	}
	if neighbour.OutgoingClient != nil {
		return nil
	}

	conn, err := net.Dial(neighbour.ConnectionType.String(), neighbour.ConnectionAddress)
	if err != nil {
		return err
	}

	clt := rpc.NewClient(conn)
	neighbour.OutgoingClient = clt
	generator := StandardEntropyGenerator{}
	convertedKey, err := ExportPublicKey(&pulsar.PrivateKey.PublicKey)
	if err != nil {
		return nil
	}
	var rep Payload
	message := Payload{PublicKey: convertedKey, Body: HandshakePayload{Entropy: generator.GenerateEntropy()}}
	message.Signature, err = singData(pulsar.PrivateKey, message.Body)
	if err != nil {
		return err
	}
	err = clt.Call(Handshake.String(), message, &rep)
	if err != nil {
		return err
	}

	result, err := checkSignature(&rep)
	if err != nil {
		return err
	}
	if !result {
		return errors.New("Signature check failed")
	}

	return nil
}

func (pulsar *Pulsar) fetchNeighbour(pubKey string) (*Neighbour, error) {
	neighbour, ok := pulsar.Neighbours[pubKey]
	if !ok {
		return nil, errors.New("Forbidden connection")
	}

	return neighbour, nil
}

func checkSignature(request *Payload) (bool, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(request.Body)
	if err != nil {
		return false, err
	}

	r := big.Int{}
	s := big.Int{}
	sigLen := len(request.Signature)
	r.SetBytes(request.Signature[:(sigLen / 2)])
	s.SetBytes(request.Signature[(sigLen / 2):])

	h := sha3.New256()
	_, err = h.Write(b.Bytes())
	if err != nil {
		return false, err
	}
	hash := h.Sum(nil)
	publicKey, err := ImportPublicKey(request.PublicKey)
	if err != nil {
		return false, err
	}

	return ecdsa.Verify(publicKey, hash, &r, &s), nil
}

func singData(privateKey *ecdsa.PrivateKey, data interface{}) ([]byte, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(data)
	if err != nil {
		return nil, err
	}

	h := sha3.New256()
	_, err = h.Write(b.Bytes())
	if err != nil {
		return nil, err
	}
	hash := h.Sum(nil)
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash)
	if err != nil {
		return nil, err
	}

	return append(r.Bytes(), s.Bytes()...), nil
}