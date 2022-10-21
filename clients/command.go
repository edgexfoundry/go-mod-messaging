//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	"github.com/edgexfoundry/go-mod-messaging/v2/messaging"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

const (
	QueryRequestTopicPrefix   = "QueryRequestTopicPrefix"
	QueryResponseTopic        = "QueryResponseTopic"
	CommandRequestTopicPrefix = "CommandRequestTopicPrefix"
	CommandResponseTopic      = "CommandResponseTopic"
)

type CommandClient struct {
	messageBus      messaging.MessageClient
	queryMessages   chan types.MessageEnvelope
	queryErrors     chan error
	commandMessages chan types.MessageEnvelope
	commandErrors   chan error
	topics          map[string]string
	timeout         time.Duration
}

func NewCommandClient(messageBus messaging.MessageClient, topics map[string]string, timeout time.Duration) (interfaces.CommandClient, error) {
	client := &CommandClient{
		messageBus: messageBus,
		topics:     topics,
		timeout:    timeout,
	}

	queryResponseTopic, ok := topics[QueryResponseTopic]
	if ok {
		queryMessages := make(chan types.MessageEnvelope, 1)
		queryErrors := make(chan error)
		queryTopics := []types.TopicChannel{
			{
				Topic:    queryResponseTopic,
				Messages: queryMessages,
			},
		}
		err := messageBus.Subscribe(queryTopics, queryErrors)
		if err != nil {
			return nil, err
		}

		client.queryMessages = queryMessages
		client.queryErrors = queryErrors
	}

	commandResponseTopic, ok := topics[CommandResponseTopic]
	if ok {
		commandMessages := make(chan types.MessageEnvelope, 1)
		commandErrors := make(chan error)
		commandTopics := []types.TopicChannel{
			{
				Topic:    commandResponseTopic,
				Messages: commandMessages,
			},
		}
		err := messageBus.Subscribe(commandTopics, commandErrors)
		if err != nil {
			return nil, err
		}

		client.commandMessages = commandMessages
		client.commandErrors = commandErrors
	}

	return client, nil
}

func (c *CommandClient) AllDeviceCoreCommands(ctx context.Context, offset int, limit int) (responses.MultiDeviceCoreCommandsResponse, edgexErr.EdgeX) {
	if c.queryMessages == nil {
		return responses.MultiDeviceCoreCommandsResponse{}, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "commandquery request/response topics not provided", nil)
	}

	queryParams := map[string]string{common.Offset: strconv.Itoa(offset), common.Limit: strconv.Itoa(limit)}
	requestEnvelope := types.NewMessageEnvelopeForRequest(nil, queryParams)
	requestTopic := strings.Join([]string{c.topics[QueryRequestTopicPrefix], common.All}, "/")
	err := c.messageBus.Publish(requestEnvelope, requestTopic)
	if err != nil {
		return responses.MultiDeviceCoreCommandsResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
	}

	for {
		select {
		case <-ctx.Done():
			return responses.MultiDeviceCoreCommandsResponse{}, nil
		case <-time.After(c.timeout):
			return responses.MultiDeviceCoreCommandsResponse{}, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "timed out waiting response", nil)
		case err = <-c.queryErrors:
			return responses.MultiDeviceCoreCommandsResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
		case responseEnvelope := <-c.queryMessages:
			if responseEnvelope.RequestID != requestEnvelope.RequestID {
				continue
			}
			if responseEnvelope.ErrorCode == 1 {
				return responses.MultiDeviceCoreCommandsResponse{}, edgexErr.NewCommonEdgeXWrapper(errors.New(string(responseEnvelope.Payload)))
			}

			var res responses.MultiDeviceCoreCommandsResponse
			err = json.Unmarshal(responseEnvelope.Payload, &res)
			if err != nil {
				return responses.MultiDeviceCoreCommandsResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
			}

			return res, nil
		}
	}
}

func (c *CommandClient) DeviceCoreCommandsByDeviceName(ctx context.Context, deviceName string) (responses.DeviceCoreCommandResponse, edgexErr.EdgeX) {
	if c.queryMessages == nil {
		return responses.DeviceCoreCommandResponse{}, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "commandquery request/response topics not provided", nil)
	}

	requestEnvelope := types.NewMessageEnvelopeForRequest(nil, nil)
	requestTopic := strings.Join([]string{c.topics[QueryRequestTopicPrefix], deviceName}, "/")
	err := c.messageBus.Publish(requestEnvelope, requestTopic)
	if err != nil {
		return responses.DeviceCoreCommandResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
	}

	for {
		select {
		case <-ctx.Done():
			return responses.DeviceCoreCommandResponse{}, nil
		case <-time.After(c.timeout):
			return responses.DeviceCoreCommandResponse{}, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "timed out waiting response", nil)
		case err = <-c.queryErrors:
			return responses.DeviceCoreCommandResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
		case responseEnvelope := <-c.queryMessages:
			if responseEnvelope.RequestID != requestEnvelope.RequestID {
				continue
			}
			if responseEnvelope.ErrorCode == 1 {
				return responses.DeviceCoreCommandResponse{}, edgexErr.NewCommonEdgeXWrapper(errors.New(string(responseEnvelope.Payload)))
			}

			var res responses.DeviceCoreCommandResponse
			err = json.Unmarshal(responseEnvelope.Payload, &res)
			if err != nil {
				return responses.DeviceCoreCommandResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
			}

			return res, nil
		}
	}
}

func (c *CommandClient) IssueGetCommandByName(ctx context.Context, deviceName string, commandName string, dsPushEvent string, dsReturnEvent string) (*responses.EventResponse, edgexErr.EdgeX) {
	if c.commandMessages == nil {
		return nil, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "command request/response topics not provided", nil)
	}

	queryParams := map[string]string{common.PushEvent: dsPushEvent, common.ReturnEvent: dsReturnEvent}
	return c.IssueGetCommandByNameWithQueryParams(ctx, deviceName, commandName, queryParams)
}

func (c *CommandClient) IssueGetCommandByNameWithQueryParams(ctx context.Context, deviceName string, commandName string, queryParams map[string]string) (*responses.EventResponse, edgexErr.EdgeX) {
	if c.commandMessages == nil {
		return nil, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "command request/response topics not provided", nil)
	}

	requestEnvelope := types.NewMessageEnvelopeForRequest(nil, queryParams)
	requestTopic := strings.Join([]string{c.topics[CommandRequestTopicPrefix], deviceName, commandName, "get"}, "/")
	err := c.messageBus.Publish(requestEnvelope, requestTopic)
	if err != nil {
		return nil, edgexErr.NewCommonEdgeXWrapper(err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil, nil
		case <-time.After(c.timeout):
			return nil, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "timed out waiting response", nil)
		case err = <-c.commandErrors:
			return nil, edgexErr.NewCommonEdgeXWrapper(err)
		case responseEnvelope := <-c.commandMessages:
			if responseEnvelope.RequestID != requestEnvelope.RequestID {
				continue
			}
			if responseEnvelope.ErrorCode == 1 {
				return nil, edgexErr.NewCommonEdgeXWrapper(errors.New(string(responseEnvelope.Payload)))
			}

			var res responses.EventResponse
			returnEvent, ok := queryParams[common.ReturnEvent]
			if ok && returnEvent == common.ValueNo {
				res.ApiVersion = common.ApiVersion
				res.RequestId = responseEnvelope.RequestID
				res.StatusCode = http.StatusOK
			} else {
				err = json.Unmarshal(responseEnvelope.Payload, &res)
				if err != nil {
					return nil, edgexErr.NewCommonEdgeXWrapper(err)
				}
			}

			return &res, nil
		}
	}
}

func (c *CommandClient) IssueSetCommandByName(ctx context.Context, deviceName string, commandName string, settings map[string]string) (commonDTO.BaseResponse, edgexErr.EdgeX) {
	if c.commandMessages == nil {
		return commonDTO.BaseResponse{}, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "command request/response topics not provided", nil)
	}

	payloadBytes, err := json.Marshal(settings)
	if err != nil {
		return commonDTO.BaseResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
	}

	requestEnvelope := types.NewMessageEnvelopeForRequest(payloadBytes, nil)
	requestTopic := strings.Join([]string{c.topics[CommandRequestTopicPrefix], deviceName, commandName, "set"}, "/")
	err = c.messageBus.Publish(requestEnvelope, requestTopic)
	if err != nil {
		return commonDTO.BaseResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
	}

	for {
		select {
		case <-ctx.Done():
			return commonDTO.BaseResponse{}, nil
		case <-time.After(c.timeout):
			return commonDTO.BaseResponse{}, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "timed out waiting response", nil)
		case err = <-c.commandErrors:
			return commonDTO.BaseResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
		case responseEnvelope := <-c.commandMessages:
			if responseEnvelope.RequestID != requestEnvelope.RequestID {
				continue
			}
			if responseEnvelope.ErrorCode == 1 {
				return commonDTO.BaseResponse{}, edgexErr.NewCommonEdgeXWrapper(errors.New(string(responseEnvelope.Payload)))
			}

			res := commonDTO.NewBaseResponse(responseEnvelope.RequestID, "", http.StatusOK)
			return res, nil
		}
	}
}

func (c *CommandClient) IssueSetCommandByNameWithObject(ctx context.Context, deviceName string, commandName string, settings map[string]any) (commonDTO.BaseResponse, edgexErr.EdgeX) {
	if c.commandMessages == nil {
		return commonDTO.BaseResponse{}, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "command request/response topics not provided", nil)
	}

	payloadBytes, err := json.Marshal(settings)
	if err != nil {
		return commonDTO.BaseResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
	}

	requestEnvelope := types.NewMessageEnvelopeForRequest(payloadBytes, nil)
	requestTopic := strings.Join([]string{c.topics[CommandRequestTopicPrefix], deviceName, commandName, "set"}, "/")
	err = c.messageBus.Publish(requestEnvelope, requestTopic)
	if err != nil {
		return commonDTO.BaseResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
	}

	for {
		select {
		case <-ctx.Done():
			return commonDTO.BaseResponse{}, nil
		case <-time.After(c.timeout):
			return commonDTO.BaseResponse{}, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "timed out waiting response", nil)
		case err = <-c.commandErrors:
			return commonDTO.BaseResponse{}, edgexErr.NewCommonEdgeXWrapper(err)
		case responseEnvelope := <-c.commandMessages:
			if responseEnvelope.RequestID != requestEnvelope.RequestID {
				continue
			}
			if responseEnvelope.ErrorCode == 1 {
				return commonDTO.BaseResponse{}, edgexErr.NewCommonEdgeXWrapper(errors.New(string(responseEnvelope.Payload)))
			}

			res := commonDTO.NewBaseResponse(responseEnvelope.RequestID, "", http.StatusOK)
			return res, nil
		}
	}
}
