package sdk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/thk-im/thk-im-base-server/conf"
	baseErrorx "github.com/thk-im/thk-im-base-server/errorx"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
)

type (
	SfuApi interface {
		// CreateSession 创建sfu会话
		CreateSession() (*dto.NewSessionResponse, error)
		// NewTracks 创建Track 推流/拉流使用该接口
		NewTracks(sessionId string, req *dto.TracksRequest) (*dto.TracksResponse, error)
		// CloseTracks 关闭track 结束推流/拉流使用该接口
		CloseTracks(sessionId string, req *dto.CloseTracksRequest) (*dto.CloseTracksResponse, error)
		// UpdateTracks 更新track 重新推流/重新订阅使用该接口
		UpdateTracks(sessionId string, req *dto.UpdateTracksRequest) (*dto.UpdateTracksResponse, error)
		// Renegotiate 协商 TracksResponse返回RequiresImmediateRenegotiation为true需要调用该接口
		Renegotiate(sessionId string, req *dto.RenegotiateRequest) (*dto.RenegotiateResponse, error)
		// GetSessionState 查询Session状态
		GetSessionState(sessionId string) (*dto.GetSessionStateResponse, error)
	}

	defaultSfuApi struct {
		endpoint string
		appId    string
		token    string
		logger   *logrus.Entry
		client   *resty.Client
	}
)

func (d defaultSfuApi) CreateSession() (*dto.NewSessionResponse, error) {
	url := fmt.Sprintf("%s/apps/%s/sessions/new", d.endpoint, d.appId)

	res := &dto.NewSessionResponse{}
	resp, err := d.newRequest().
		SetResult(res).
		Post(url)

	if err != nil {
		d.logger.Errorf("CreateSession error: %v", err)
		return nil, err
	}
	d.logger.Tracef("CreateSession response: %d, %s", resp.StatusCode(), string(resp.Body()))

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		errRes := baseErrorx.NewErrorXFromResp(resp)
		d.logger.Errorf("CreateSession failed: %v", errRes)
		return nil, errRes
	}

	err = json.Unmarshal(resp.Body(), &res)
	return res, err
}

func (d defaultSfuApi) NewTracks(sessionId string, req *dto.TracksRequest) (*dto.TracksResponse, error) {
	url := fmt.Sprintf("%s/apps/%s/sessions/%s/tracks/new", d.endpoint, d.appId, sessionId)

	res := &dto.TracksResponse{}
	resp, err := d.newRequest().
		SetBody(req).
		SetResult(res).
		Post(url)

	if err != nil {
		d.logger.Errorf("NewTracks error: %v", err)
		return nil, err
	}

	d.logger.Tracef("NewTracks response: %d, %s", resp.StatusCode(), string(resp.Body()))

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		errRes := baseErrorx.NewErrorXFromResp(resp)
		d.logger.Errorf("NewTracks failed: %v", errRes)
		return nil, errRes
	}

	err = json.Unmarshal(resp.Body(), &res)
	return res, err
}

func (d defaultSfuApi) CloseTracks(sessionId string, req *dto.CloseTracksRequest) (*dto.CloseTracksResponse, error) {
	url := fmt.Sprintf("%s/apps/%s/sessions/%s/tracks/close", d.endpoint, d.appId, sessionId)

	res := &dto.CloseTracksResponse{}
	resp, err := d.newRequest().
		SetBody(req).
		SetResult(res).
		Put(url)

	if err != nil {
		d.logger.Errorf("CloseTracks error: %v", err)
		return nil, err
	}

	d.logger.Tracef("CloseTracks response: %d, %s", resp.StatusCode(), string(resp.Body()))

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		errRes := baseErrorx.NewErrorXFromResp(resp)
		d.logger.Errorf("CloseTracks failed: %v", errRes)
		return nil, errRes
	}

	err = json.Unmarshal(resp.Body(), &res)
	return res, err
}

func (d defaultSfuApi) UpdateTracks(sessionId string, req *dto.UpdateTracksRequest) (*dto.UpdateTracksResponse, error) {
	url := fmt.Sprintf("%s/apps/%s/sessions/%s/tracks/update", d.endpoint, d.appId, sessionId)

	res := &dto.UpdateTracksResponse{}
	resp, err := d.newRequest().
		SetBody(req).
		SetResult(res).
		Put(url)

	if err != nil {
		d.logger.Errorf("UpdateTracks error: %v", err)
		return nil, err
	}

	d.logger.Tracef("UpdateTracks response: %d, %s", resp.StatusCode(), string(resp.Body()))

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		errRes := baseErrorx.NewErrorXFromResp(resp)
		d.logger.Errorf("UpdateTracks failed: %v", errRes)
		return nil, errRes
	}

	err = json.Unmarshal(resp.Body(), &res)
	return res, err
}

func (d defaultSfuApi) Renegotiate(sessionId string, req *dto.RenegotiateRequest) (*dto.RenegotiateResponse, error) {
	url := fmt.Sprintf("%s/apps/%s/sessions/%s/renegotiate", d.endpoint, d.appId, sessionId)

	res := &dto.RenegotiateResponse{}
	resp, err := d.newRequest().
		SetBody(req).
		SetResult(res).
		Put(url)

	if err != nil {
		d.logger.Errorf("Renegotiate error: %v", err)
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		errRes := baseErrorx.NewErrorXFromResp(resp)
		d.logger.Errorf("Renegotiate failed: %v", errRes)
		return nil, errRes
	}

	d.logger.Tracef("Renegotiate response: %d, %s", resp.StatusCode(), string(resp.Body()))

	err = json.Unmarshal(resp.Body(), &res)
	return res, err
}

func (d defaultSfuApi) GetSessionState(sessionId string) (*dto.GetSessionStateResponse, error) {
	url := fmt.Sprintf("%s/apps/%s/sessions/%s", d.endpoint, d.appId, sessionId)

	res, err := d.newRequest().
		SetHeader("Content-Type", "application/json").
		Get(url)

	if err != nil {
		d.logger.Errorf("GetSessionState request err: %v", err)
		return nil, err
	}

	d.logger.Tracef("GetSessionState response: %d, %s", res.StatusCode(), res.String())

	if res.StatusCode() != http.StatusOK && res.StatusCode() != http.StatusGone {
		return nil, baseErrorx.NewErrorXFromResp(res)
	}

	var result dto.GetSessionStateResponse
	if err = json.Unmarshal(res.Body(), &result); err != nil {
		d.logger.Errorf("GetSessionState unmarshal err: %v", err)
		return nil, err
	}

	return &result, nil
}

func (d defaultSfuApi) newRequest() *resty.Request {
	return d.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+d.token)
}

func NewSfuApi(sdk conf.Sdk, logger *logrus.Entry) SfuApi {
	return &defaultSfuApi{
		endpoint: sdk.Endpoint,
		token:    os.Getenv("CloudflareSFUToken"),
		appId:    os.Getenv("CloudflareSFUAppId"),
		logger:   logger.WithField("sdk", sdk.Name),
		client: resty.New().
			SetTransport(&http.Transport{
				MaxIdleConns:    10,
				MaxConnsPerHost: 10,
				IdleConnTimeout: 30 * time.Second,
			}).
			SetTimeout(30 * time.Second),
	}
}
