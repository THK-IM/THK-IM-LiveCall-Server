package sdk

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/thk-im/thk-im-base-server/conf"
	baseDto "github.com/thk-im/thk-im-base-server/dto"
	baseErrorx "github.com/thk-im/thk-im-base-server/errorx"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	"net/http"
	"time"
)

const (
	jsonContentType string = "application/json"
)

type (
	CheckApi interface {
		CheckCreateRoom(req *dto.CheckCreateRoomReq, claims baseDto.ThkClaims) error
		CheckJoinRoom(req *dto.CheckJoinRoomReq, claims baseDto.ThkClaims) error
	}

	defaultCheckApi struct {
		endpoint string
		logger   *logrus.Entry
		client   *resty.Client
	}
)

func (d defaultCheckApi) CheckJoinRoom(req *dto.CheckJoinRoomReq, claims baseDto.ThkClaims) error {
	dataBytes, err := json.Marshal(req)
	if err != nil {
		d.logger.WithFields(logrus.Fields(claims)).Errorf("CheckJoinRoom: %v %v", req, err)
		return err
	}
	url := fmt.Sprintf("%s/live_call/check/join_room", d.endpoint)
	request := d.client.R()
	for k, v := range claims {
		vs := v.(string)
		request.SetHeader(k, vs)
	}
	res, errRequest := request.
		SetHeader("Content-Type", jsonContentType).
		SetBody(dataBytes).
		Post(url)
	if errRequest != nil {
		d.logger.Errorf("CheckJoinRoom %v %v", claims, errRequest)
		return errRequest
	}
	if res.StatusCode() != http.StatusOK {
		errRes := baseErrorx.NewErrorXFromResp(res)
		d.logger.Errorf("CheckJoinRoom: %v %v", claims, errRes)
		return errRes
	} else {
		return nil
	}
}

func (d defaultCheckApi) CheckCreateRoom(req *dto.CheckCreateRoomReq, claims baseDto.ThkClaims) error {
	dataBytes, err := json.Marshal(req)
	if err != nil {
		d.logger.WithFields(logrus.Fields(claims)).Errorf("CheckCreateRoom: %v %v", req, err)
		return err
	}
	url := fmt.Sprintf("%s/live_call/check/create_room", d.endpoint)
	request := d.client.R()
	for k, v := range claims {
		vs := v.(string)
		request.SetHeader(k, vs)
	}
	res, errRequest := request.
		SetHeader("Content-Type", jsonContentType).
		SetBody(dataBytes).
		Post(url)
	if errRequest != nil {
		d.logger.Errorf("CheckCreateRoom %v %v", claims, errRequest)
		return errRequest
	}
	if res.StatusCode() != http.StatusOK {
		errRes := baseErrorx.NewErrorXFromResp(res)
		d.logger.Errorf("CheckCreateRoom: %v %v", claims, errRes)
		return errRes
	} else {
		return nil
	}
}

func NewCheckApi(sdk conf.Sdk, logger *logrus.Entry) CheckApi {
	return defaultCheckApi{
		endpoint: sdk.Endpoint,
		logger:   logger.WithField("rpc", sdk.Name),
		client: resty.New().
			SetTransport(&http.Transport{
				MaxIdleConns:    10,
				MaxConnsPerHost: 10,
				IdleConnTimeout: 30 * time.Second,
			}).
			SetTimeout(5 * time.Second).
			SetRetryCount(3).
			SetRetryWaitTime(15 * time.Second).
			SetRetryMaxWaitTime(5 * time.Second),
	}
}
