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

package adapter

import (
	"fmt"

	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/log"
	"github.com/pkg/errors"
)

func NewResponseSendAdapter() PulseConveyorAdapterTaskSink {
	return NewAdapterWithQueue(NewResponseSender())
}

type ResponseSenderTask struct {
	Future core.Future
	Result core.Reply
}

type ResponseSender struct{}

func NewResponseSender() Worker {
	return &ResponseSender{}
}

func (sr *ResponseSender) Process(adapterID uint32, task AdapterTask, cancelInfo *cancelInfoT) {
	payload, ok := task.taskPayload.(ResponseSenderTask)
	var msg interface{}

	if !ok {
		msg = errors.Errorf("[ PushTask ] Incorrect payload type: %T", task.taskPayload)
		task.respSink.PushResponse(adapterID, task.elementID, task.handlerID, msg)
		return
	}

	done := make(chan bool, 1)
	go func(payload ResponseSenderTask) {
		res := payload.Result
		f := payload.Future
		f.SetResult(res)
		done <- true
	}(payload)

	select {
	case <-cancelInfo.cancel:
		log.Info("[ SimpleWaitAdapter.doWork ] Cancel. Return Nil as Response")
		msg = nil
	case <-cancelInfo.flush:
		log.Info("[ SimpleWaitAdapter.doWork ] Flush. DON'T Return Response")
		return
	case <-done:
		msg = fmt.Sprintf("Response was send successfully")
	}

	log.Info("[ SimpleWaitAdapter.doWork ] ", msg)

	task.respSink.PushResponse(adapterID, task.elementID, task.handlerID, msg)
}
