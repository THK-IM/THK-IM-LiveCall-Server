package logic

import (
	"time"

	baseDto "github.com/thk-im/thk-im-base-server/dto"
	baseErr "github.com/thk-im/thk-im-base-server/errorx"
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/errorx"
	"github.com/thk-im/thk-im-livecall-server/pkg/sdk"
)

type StreamLogic struct {
	appCtx    *app.Context
	roomLogic *RoomLogic
}

func NewStreamLogic(appCtx *app.Context) *StreamLogic {
	return &StreamLogic{
		appCtx:    appCtx,
		roomLogic: NewRoomLogic(appCtx),
	}
}

func (l StreamLogic) api() sdk.SfuApi {
	if l.appCtx.Context.SdkMap["cloudflare_connect_api"] == nil {
		return nil
	}
	return l.appCtx.Context.SdkMap["cloudflare_connect_api"].(sdk.SfuApi)
}

func (l StreamLogic) PublishStream(req *dto.PublishStreamReq, claims baseDto.ThkClaims) (*dto.PublishStreamResp, error) {
	room, isMember, errMember := l.checkMember(req.RoomId, req.Uid, claims)
	if errMember != nil {
		l.appCtx.Logger().Error("PublishStream err, ", errMember)
		return nil, baseErr.ErrInternalServerError
	}

	if !isMember {
		l.appCtx.Logger().Error("PublishStream err, ", "not member", req.Uid)
		return nil, errorx.ErrNoPermission
	}

	resp, err := l.api().CreateSession()
	if err != nil {
		l.appCtx.Logger().Error("PublishStream err, ", err)
		return nil, baseErr.ErrInternalServerError
	}

	if resp == nil || resp.SessionID == "" {
		l.appCtx.Logger().Error("PublishStream resp err", resp)
		return nil, baseErr.ErrInternalServerError
	}

	tracks := dto.ParseTracksFromSDP(req.Sdp, resp.SessionID)

	tracksResp, errTracks := l.api().NewTracks(resp.SessionID, &dto.TracksRequest{
		SessionDescription: &dto.SessionDescription{
			Type: "offer",
			SDP:  req.Sdp,
		},
		Tracks: tracks,
	})

	if errTracks != nil {
		l.appCtx.Logger().Error("PublishStream err, ", err)
		return nil, baseErr.ErrInternalServerError
	}

	if tracksResp == nil || tracksResp.ErrorCode != "" || resp.SessionID == "" || tracksResp.SessionDescription == nil {
		l.appCtx.Logger().Error("PublishStream resp err", resp)
		return nil, baseErr.ErrInternalServerError
	}

	_ = l.roomLogic.OnUserJoinEvent(&dto.RoomUserJoinEvent{
		RoomId:    room.Id,
		UserId:    req.Uid,
		Timestamp: time.Now().UnixMilli(),
	}, claims)

	return &dto.PublishStreamResp{SessionId: resp.SessionID, Sdp: tracksResp.SessionDescription.SDP, Type: tracksResp.SessionDescription.Type}, nil
}

func (l StreamLogic) SubscribeStream(req *dto.SubscribeStreamReq, claims baseDto.ThkClaims) (*dto.SubscribeStreamResp, error) {
	room, isMember, errMember := l.checkMember(req.RoomId, req.Uid, claims)
	if errMember != nil {
		l.appCtx.Logger().Error("SubscribeStream err, ", errMember)
		return nil, baseErr.ErrInternalServerError
	}

	if !isMember {
		l.appCtx.Logger().Error("SubscribeStream err, ", "not member", req.Uid)
		return nil, errorx.ErrNoPermission
	}

	resp, err := l.api().CreateSession()
	if err != nil {
		l.appCtx.Logger().Error("SubscribeStream err, ", err)
		return nil, baseErr.ErrInternalServerError
	}

	if resp == nil || resp.SessionID == "" {
		l.appCtx.Logger().Error("SubscribeStream resp err", resp)
		return nil, baseErr.ErrInternalServerError
	}

	tracks := make([]dto.TrackObject, 0)
	if room.Mode == dto.ModeVideo || room.Mode == dto.ModeVideoRoom {
		tracks = append(tracks, dto.TrackObject{
			Location:  "remote",
			SessionID: req.SessionId,
			TrackName: "camera",
		})
	}
	tracks = append(tracks, dto.TrackObject{
		Location:  "remote",
		SessionID: req.SessionId,
		TrackName: "mic",
	})

	tracksResp, tracksError := l.api().NewTracks(resp.SessionID, &dto.TracksRequest{
		SessionDescription: &dto.SessionDescription{
			SDP:  req.Sdp,
			Type: "offer",
		},
		Tracks: tracks,
	})

	if tracksError != nil {
		l.appCtx.Logger().Error("SubscribeStream err, ", tracksError)
		return nil, baseErr.ErrInternalServerError
	}
	if tracksResp == nil || tracksResp.ErrorCode != "" || tracksResp.SessionDescription == nil {
		l.appCtx.Logger().Error("SubscribeStream resp err", tracksResp)
		return nil, baseErr.ErrInternalServerError
	}

	res := &dto.SubscribeStreamResp{
		Renegotiation: tracksResp.RequiresImmediateRenegotiation,
	}
	if tracksResp.SessionDescription != nil {
		res.Sdp = tracksResp.SessionDescription.SDP
		res.Type = tracksResp.SessionDescription.Type
	}
	return res, nil
}

func (l StreamLogic) UpdateStreamStatus(req *dto.StreamStatusUpdateReq, claims baseDto.ThkClaims) error {
	_, isMember, errMember := l.checkMember(req.RoomId, req.Uid, claims)
	if errMember != nil {
		l.appCtx.Logger().Error("UpdateStreamStatus err, ", errMember)
		return baseErr.ErrInternalServerError
	}

	if !isMember {
		l.appCtx.Logger().Error("UpdateStreamStatus err, ", "not member", req.Uid)
		return errorx.ErrNoPermission
	}

	if req.Status == "start" {
		return l.roomLogic.OnUserPushEvent(&dto.RoomUserPushStreamEvent{
			RoomId:    req.RoomId,
			UserId:    req.Uid,
			StreamKey: req.SessionId,
			Timestamp: time.Now().UnixMilli(),
		}, claims)
	} else if req.Status == "stop" {

	} else if req.Status == "ing" {

	}
	return nil
}

func (l StreamLogic) checkMember(roomId string, uId int64, claims baseDto.ThkClaims) (*dto.Room, bool, error) {
	room, errRoom := l.roomLogic.QueryRoom(roomId, claims)
	if errRoom != nil {
		l.appCtx.Logger().Error("checkMember err, ", errRoom)
		return nil, false, baseErr.ErrInternalServerError
	}

	if room == nil {
		l.appCtx.Logger().Error("checkMember room is nil, ", roomId)
		return nil, false, errorx.ErrRoomNotExisted
	}

	isMember := false
	for _, p := range room.Participants {
		if p.UId == uId {
			isMember = true
			break
		}
	}

	return room, isMember, nil
}
