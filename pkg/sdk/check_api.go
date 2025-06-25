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
		CheckLiveCallCreate(req *dto.CheckLiveCallCreateReq, claims baseDto.ThkClaims) error
		CheckLiveCallJoin(req *dto.CheckLiveJoinReq, claims baseDto.ThkClaims) error
		CheckLiveCallInvite(req *dto.CheckLiveInviteReq, claims baseDto.ThkClaims) error
		CheckLiveCallStatus(req *dto.CheckLiveCallStatusReq, claims baseDto.ThkClaims) error
	}

	defaultCheckApi struct {
		endpoint string
		logger   *logrus.Entry
		client   *resty.Client
	}
)

func (d defaultCheckApi) CheckLiveCallCreate(req *dto.CheckLiveCallCreateReq, claims baseDto.ThkClaims) error {
	dataBytes, err := json.Marshal(req)
	if err != nil {
		d.logger.WithFields(logrus.Fields(claims)).Errorf("CheckLiveCallCreate: %v %v", req, err)
		return err
	}
	url := fmt.Sprintf("%s/live_call/check/create", d.endpoint)
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
		d.logger.Errorf("CheckLiveCallCreate %v %v", claims, errRequest)
		return errRequest
	}
	if res.StatusCode() != http.StatusOK {
		errRes := baseErrorx.NewErrorXFromResp(res)
		d.logger.Errorf("CheckLiveCallCreate: %v %v", claims, errRes)
		return errRes
	} else {
		return nil
	}
}

func (d defaultCheckApi) CheckLiveCallJoin(req *dto.CheckLiveJoinReq, claims baseDto.ThkClaims) error {
	dataBytes, err := json.Marshal(req)
	if err != nil {
		d.logger.WithFields(logrus.Fields(claims)).Errorf("CheckLiveCallJoin: %v %v", req, err)
		return err
	}
	url := fmt.Sprintf("%s/live_call/check/join", d.endpoint)
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
		d.logger.Errorf("CheckLiveCallJoin %v %v", claims, errRequest)
		return errRequest
	}
	if res.StatusCode() != http.StatusOK {
		errRes := baseErrorx.NewErrorXFromResp(res)
		d.logger.Errorf("CheckLiveCall: %v %v", claims, errRes)
		return errRes
	} else {
		return nil
	}
}

func (d defaultCheckApi) CheckLiveCallInvite(req *dto.CheckLiveInviteReq, claims baseDto.ThkClaims) error {
	dataBytes, err := json.Marshal(req)
	if err != nil {
		d.logger.WithFields(logrus.Fields(claims)).Errorf("CheckLiveCallInvite: %v %v", req, err)
		return err
	}
	url := fmt.Sprintf("%s/live_call/check/invite", d.endpoint)
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
		d.logger.Errorf("CheckLiveCallInvite %v %v", claims, errRequest)
		return errRequest
	}
	if res.StatusCode() != http.StatusOK {
		errRes := baseErrorx.NewErrorXFromResp(res)
		d.logger.Errorf("CheckLiveCall: %v %v", claims, errRes)
		return errRes
	} else {
		return nil
	}
}

func (d defaultCheckApi) CheckLiveCallStatus(req *dto.CheckLiveCallStatusReq, claims baseDto.ThkClaims) error {
	dataBytes, err := json.Marshal(req)
	if err != nil {
		d.logger.WithFields(logrus.Fields(claims)).Errorf("CheckLiveCallStatus: %v %v", req, err)
		return err
	}
	url := fmt.Sprintf("%s/live_call/check/status", d.endpoint)
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
		d.logger.Errorf("CheckLiveCallStatus %v %v", claims, errRequest)
		return errRequest
	}
	if res.StatusCode() != http.StatusOK {
		errRes := baseErrorx.NewErrorXFromResp(res)
		d.logger.Errorf("CheckLiveCallStatus: %v %v", claims, errRes)
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
