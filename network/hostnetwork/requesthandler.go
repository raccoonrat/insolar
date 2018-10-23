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

package hostnetwork

import (
	"time"

	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/log"
	"github.com/insolar/insolar/network/hostnetwork/host"
	"github.com/insolar/insolar/network/hostnetwork/hosthandler"
	"github.com/insolar/insolar/network/hostnetwork/packet"
	"github.com/insolar/insolar/network/hostnetwork/transport"
	"github.com/insolar/insolar/version"
	"github.com/jbenet/go-base58"
	"github.com/pkg/errors"
)

// RelayRequest sends relay request to target.
func RelayRequest(hostHandler hosthandler.HostHandler, command, targetID string) error {
	ctx, err := NewContextBuilder(hostHandler).SetDefaultHost().Build()
	if err != nil {
		return err
	}
	var typedCommand packet.CommandType
	targetHost, exist, err := hostHandler.FindHost(ctx, targetID)
	if err != nil {
		return err
	}
	if !exist {
		err = errors.New("RelayRequest: target for relay request not found")
		return err
	}

	switch command {
	case "start":
		typedCommand = packet.StartRelay
	case "stop":
		typedCommand = packet.StopRelay
	default:
		err = errors.New("RelayRequest: unknown command")
		return err
	}
	builder := packet.NewBuilder(hostHandler.HtFromCtx(ctx).Origin)
	request := builder.Type(packet.TypeRelay).
		Receiver(targetHost).
		Request(&packet.RequestRelay{Command: typedCommand}).
		Build()
	future, err := hostHandler.SendRequest(request)

	if err != nil {
		return err
	}

	return checkResponse(hostHandler, future, targetID, request)
}

// CheckOriginRequest send a request to check target host originality
func CheckOriginRequest(hostHandler hosthandler.HostHandler, targetID string) error {
	ctx, err := NewContextBuilder(hostHandler).SetDefaultHost().Build()
	if err != nil {
		return err
	}
	targetHost, exist, err := hostHandler.FindHost(ctx, targetID)
	if err != nil {
		return err
	}
	if !exist {
		err = errors.New("CheckOriginRequest: target for relay request not found")
		return err
	}

	builder := packet.NewBuilder(hostHandler.HtFromCtx(ctx).Origin)
	request := builder.Type(packet.TypeCheckOrigin).
		Receiver(targetHost).
		Request(&packet.RequestCheckOrigin{}).
		Build()
	future, err := hostHandler.SendRequest(request)

	if err != nil {
		log.Debugln(err.Error())
		return err
	}

	return checkResponse(hostHandler, future, targetID, request)
}

// AuthenticationRequest sends an authentication request.
func AuthenticationRequest(hostHandler hosthandler.HostHandler, command, targetID string) error {
	ctx, err := NewContextBuilder(hostHandler).SetDefaultHost().Build()
	if err != nil {
		return err
	}
	targetHost, exist, err := hostHandler.FindHost(ctx, targetID)
	if err != nil {
		return err
	}
	if !exist {
		err = errors.New("AuthenticationRequest: target for auth request not found")
		return err
	}

	origin := hostHandler.HtFromCtx(ctx).Origin
	var authCommand packet.CommandType
	switch command {
	case "begin":
		authCommand = packet.BeginAuthentication
	case "revoke":
		authCommand = packet.RevokeAuthentication
	default:
		err = errors.New("AuthenticationRequest: unknown command")
		return err
	}
	builder := packet.NewBuilder(origin)
	request := builder.Type(packet.TypeAuthentication).
		Receiver(targetHost).
		Request(&packet.RequestAuthentication{Command: authCommand}).
		Build()
	future, err := hostHandler.SendRequest(request)

	if err != nil {
		log.Debugln(err.Error())
		return err
	}

	return checkResponse(hostHandler, future, targetID, request)
}

// ObtainIPRequest is request to self IP obtaining.
func ObtainIPRequest(hostHandler hosthandler.HostHandler, targetID string) error {
	ctx, err := NewContextBuilder(hostHandler).SetDefaultHost().Build()
	if err != nil {
		return err
	}
	targetHost, exist, err := hostHandler.FindHost(ctx, targetID)
	if err != nil {
		return err
	}
	if !exist {
		err = errors.New("ObtainIPRequest: target for relay request not found")
		return err
	}

	origin := hostHandler.HtFromCtx(ctx).Origin
	builder := packet.NewBuilder(origin)
	request := builder.Type(packet.TypeObtainIP).
		Receiver(targetHost).
		Request(&packet.RequestObtainIP{}).
		Build()

	future, err := hostHandler.SendRequest(request)

	if err != nil {
		log.Debugln(err.Error())
		return err
	}

	return checkResponse(hostHandler, future, targetID, request)
}

// RelayOwnershipRequest sends a relay ownership request.
func RelayOwnershipRequest(hostHandler hosthandler.HostHandler, targetID string) error {
	ctx, err := NewContextBuilder(hostHandler).SetDefaultHost().Build()
	if err != nil {
		return err
	}
	targetHost, exist, err := hostHandler.FindHost(ctx, targetID)
	if err != nil {
		return err
	}
	if !exist {
		err = errors.New("relayOwnershipRequest: target for relay request not found")
		return err
	}

	builder := packet.NewBuilder(hostHandler.HtFromCtx(ctx).Origin)
	request := builder.Type(packet.TypeRelayOwnership).
		Receiver(targetHost).
		Request(&packet.RequestRelayOwnership{Ready: true}).
		Build()
	future, err := hostHandler.SendRequest(request)

	if err != nil {
		return err
	}

	return checkResponse(hostHandler, future, targetID, request)
}

// CascadeSendMessage sends a message to the next cascade layer.
func CascadeSendMessage(hostHandler hosthandler.HostHandler, data core.Cascade, targetID string, method string, args [][]byte) error {
	ctx, err := NewContextBuilder(hostHandler).SetDefaultHost().Build()
	if err != nil {
		return err
	}
	targetHost, exist, err := hostHandler.FindHost(ctx, targetID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("cascadeSendMessage: couldn't find a target host")
	}

	request := packet.NewBuilder(hostHandler.HtFromCtx(ctx).Origin).Receiver(targetHost).Type(packet.TypeCascadeSend).
		Request(&packet.RequestCascadeSend{
			Data: data,
			RPC: packet.RequestDataRPC{
				Method: method,
				Args:   args,
			},
		}).Build()

	future, err := hostHandler.SendRequest(request)
	if err != nil {
		return err
	}

	return checkResponse(hostHandler, future, targetID, request)
}

func GetNonceRequest(hostHandler hosthandler.HostHandler, targetID string) ([]*core.ActiveNode, error) {
	ctx, err := NewContextBuilder(hostHandler).SetDefaultHost().Build()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build a context")
	}
	targetHost, exist, err := hostHandler.FindHost(ctx, targetID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find a target host")
	}
	if !exist {
		return nil, errors.Wrap(err, "couldn't find a target host")
	}

	sender := hostHandler.HtFromCtx(ctx).Origin
	nonce, err := sendNonceRequest(hostHandler, sender, targetHost)
	if err != nil {
		return nil, errors.Wrap(err, "failed getting nonce from discovery node")
	}
	log.Debugf("got nonce from discovery node: %s", base58.Encode(nonce))
	signedNonce, err := hostHandler.GetNetworkCommonFacade().GetSignHandler().SignNonce(nonce)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign nonce from discovery node")
	}
	log.Debugf("signed nonce: %s", base58.Encode(signedNonce))
	result, err := sendCheckSignedNonceRequest(hostHandler, sender, targetHost, signedNonce)
	if err != nil {
		return nil, errors.Wrap(err, "failed checking signed nonce on discovery node")
	}
	return result, nil
}

func sendNonceRequest(hostHandler hosthandler.HostHandler, sender *host.Host, receiver *host.Host) ([]byte, error) {
	log.Debug("Started getting nonce request to discovery node")

	request := packet.NewBuilder(sender).
		Receiver(receiver).Type(packet.TypeGetNonce).
		Request(&packet.RequestGetNonce{NodeID: hostHandler.GetNodeID()}).
		Build()

	future, err := hostHandler.SendRequest(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send an authorization request")
	}
	rsp, err := future.GetResult(hostHandler.GetPacketTimeout())
	if err != nil {
		return nil, errors.Wrap(err, "checkResponse error")
	}
	response := rsp.Data.(*packet.ResponseGetNonce)
	err = handleCheckPublicKeyResponse(hostHandler, response)
	if err != nil {
		return nil, errors.Wrap(err, "public key check failed on discovery node")
	}
	return response.Nonce, nil
}

func sendCheckSignedNonceRequest(hostHandler hosthandler.HostHandler, sender *host.Host,
	receiver *host.Host, nonce []byte) ([]*core.ActiveNode, error) {

	log.Debug("Started request to discovery node to check signed nonce and add to unsync list")

	// TODO: get role from certificate
	// TODO: get public key from certificate
	request := packet.NewBuilder(sender).Type(packet.TypeCheckSignedNonce).
		Receiver(receiver).
		Request(&packet.RequestCheckSignedNonce{
			Signed:    nonce,
			NodeID:    hostHandler.GetNodeID(),
			NodeRoles: []core.NodeRole{core.RoleUnknown},
			Version:   version.Version,
			// PublicKey: ???
		}).
		Build()

	future, err := hostHandler.SendRequest(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send an authorization request")
	}
	rsp, err := future.GetResult( /*hostHandler.GetPacketTimeout()*/ time.Second * 40)
	if err != nil {
		return nil, errors.Wrap(err, "checkResponse error")
	}

	responseSignedNonce := rsp.Data.(*packet.ResponseCheckSignedNonce)
	if responseSignedNonce.Error != "" {
		return nil, errors.New(responseSignedNonce.Error)
	}
	return responseSignedNonce.ActiveNodes, nil
}

// ResendPulseToKnownHosts resends received pulse to all known hosts
func ResendPulseToKnownHosts(hostHandler hosthandler.HostHandler, hosts []host.Host, pulse *packet.RequestPulse) {
	for _, host1 := range hosts {
		err := sendPulse(hostHandler, &host1, pulse)
		if err != nil {
			log.Debugf("error resending pulse to host %s: %s", host1.ID, err.Error())
		}
	}
}

func sendPulse(hostHandler hosthandler.HostHandler, host *host.Host, pulse *packet.RequestPulse) error {
	ctx, err := NewContextBuilder(hostHandler).SetDefaultHost().Build()
	if err != nil {
		return errors.Wrap(err, "failed to send pulse")
	}
	request := packet.NewBuilder(hostHandler.HtFromCtx(ctx).Origin).Receiver(host).
		Type(packet.TypePulse).Request(pulse).Build()

	future, err := hostHandler.SendRequest(request)
	if err != nil {
		return errors.Wrap(err, "failed to send pulse")
	}
	return checkResponse(hostHandler, future, "", request)
}

func checkNodePrivRequest(hostHandler hosthandler.HostHandler, targetID string) error {
	ctx, err := NewContextBuilder(hostHandler).SetDefaultHost().Build()
	if err != nil {
		return err
	}
	targetHost, exist, err := hostHandler.FindHost(ctx, targetID)
	if err != nil {
		return err
	}
	if !exist {
		err = errors.New("checkNodePrivRequest: target for check node privileges request not found")
		return err
	}

	origin := hostHandler.HtFromCtx(ctx).Origin
	builder := packet.NewBuilder(origin)
	request := builder.Type(packet.TypeCheckNodePriv).Receiver(targetHost).Request(&packet.RequestCheckNodePriv{RoleKey: "test string"}).Build()
	future, err := hostHandler.SendRequest(request)

	if err != nil {
		return errors.Wrap(err, "Failed to SendRequest")
	}

	return checkResponse(hostHandler, future, targetID, request)
}

func knownOuterHostsRequest(hostHandler hosthandler.HostHandler, targetID string, hosts int) error {
	ctx, err := NewContextBuilder(hostHandler).SetDefaultHost().Build()
	if err != nil {
		return err
	}
	targetHost, exist, err := hostHandler.FindHost(ctx, targetID)
	if err != nil {
		return errors.Wrap(err, "Failed to FindHost")
	}
	if !exist {
		err = errors.New("knownOuterHostsRequest: target for relay request not found")
		return err
	}

	builder := packet.NewBuilder(hostHandler.HtFromCtx(ctx).Origin)
	request := builder.Type(packet.TypeKnownOuterHosts).
		Receiver(targetHost).
		Request(&packet.RequestKnownOuterHosts{
			ID:         hostHandler.HtFromCtx(ctx).Origin.ID.String(),
			OuterHosts: hosts},
		).
		Build()
	future, err := hostHandler.SendRequest(request)

	if err != nil {
		return errors.Wrap(err, "Failed to SendRequest")
	}

	return checkResponse(hostHandler, future, targetID, request)
}

// SendRelayOwnership send a relay ownership request.
func SendRelayOwnership(hostHandler hosthandler.HostHandler, subnetIDs []string) {
	for _, id1 := range subnetIDs {
		err := RelayOwnershipRequest(hostHandler, id1)
		log.Errorln(err.Error())
	}
}

func sendRelayedRequest(hostHandler hosthandler.HostHandler, request *packet.Packet) {
	_, err := hostHandler.SendRequest(request)
	if err != nil {
		log.Debugln(err)
	}
}

func sendDisconnectRequest(hostHandler hosthandler.HostHandler, target *host.Host) error {
	ctx, err := NewContextBuilder(hostHandler).SetDefaultHost().Build()
	if err != nil {
		return err
	}

	builder := packet.NewBuilder(hostHandler.HtFromCtx(ctx).Origin)
	request := builder.Type(packet.TypeDisconnect).
		Receiver(target).
		Request(&packet.RequestDisconnect{}).
		Build()

	future, err := hostHandler.SendRequest(request)

	if err != nil {
		return errors.Wrap(err, "Failed to send disconnect request")
	}

	return checkResponse(hostHandler, future, target.ID.String(), request)
}

func checkResponse(hostHandler hosthandler.HostHandler, future transport.Future, targetID string, request *packet.Packet) error {
	var err error
	rsp, err := future.GetResult(hostHandler.GetPacketTimeout())
	if err != nil {
		return errors.Wrap(err, "checkResponse error")
	}
	switch request.Type {
	case packet.TypeKnownOuterHosts:
		response := rsp.Data.(*packet.ResponseKnownOuterHosts)
		err = handleKnownOuterHosts(hostHandler, response, targetID)
	case packet.TypeCheckOrigin:
		response := rsp.Data.(*packet.ResponseCheckOrigin)
		handleCheckOriginResponse(hostHandler, response, targetID)
	case packet.TypeAuthentication:
		response := rsp.Data.(*packet.ResponseAuthentication)
		err = handleAuthResponse(hostHandler, response, targetID)
	case packet.TypeObtainIP:
		response := rsp.Data.(*packet.ResponseObtainIP)
		err = handleObtainIPResponse(hostHandler, response, targetID)
	case packet.TypeRelayOwnership:
		response := rsp.Data.(*packet.ResponseRelayOwnership)
		handleRelayOwnership(hostHandler, response, targetID)
	case packet.TypeCheckNodePriv:
		response := rsp.Data.(*packet.ResponseCheckNodePriv)
		err = handleCheckNodePrivResponse(hostHandler, response)
	case packet.TypeRelay:
		response := rsp.Data.(*packet.ResponseRelay)
		err = handleRelayResponse(hostHandler, response, targetID)
	case packet.TypeCascadeSend:
		response := rsp.Data.(*packet.ResponseCascadeSend)
		if !response.Success {
			err = errors.New(response.Error)
		}
	case packet.TypePulse:
		response := rsp.Data.(*packet.ResponsePulse)
		if !response.Success {
			err = errors.New(response.Error)
		}
	case packet.TypeDisconnect:
		response := rsp.Data.(*packet.ResponseDisconnect)
		if (response.Error == nil) && response.Disconnected {
			// TODO: be a disconnected sad node
		} else {
			return response.Error
		}
	}
	return err
}
