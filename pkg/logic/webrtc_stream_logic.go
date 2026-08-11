package logic

import (
	"fmt"
	"time"

	baseDto "github.com/thk-im/thk-im-base-server/dto"
	baseErr "github.com/thk-im/thk-im-base-server/errorx"
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/errorx"
	rtcDto "github.com/thk-im/thk-im-rtc-server/pkg/dto"
)

type WebRTCStreamLogic struct {
	appCtx    *app.Context
	roomLogic *RoomLogic
}

func NewWebRTCStreamLogic(appCtx *app.Context) *WebRTCStreamLogic {
	return &WebRTCStreamLogic{
		appCtx:    appCtx,
		roomLogic: NewRoomLogic(appCtx),
	}
}

func (l WebRTCStreamLogic) PublishStream(req *dto.PublishStreamReq, claims baseDto.ThkClaims) (*dto.PublishStreamResp, error) {
	room, isMember, errMember := l.checkMember(req.RoomId, req.Uid, claims)
	if errMember != nil {
		l.appCtx.Logger().Error("PublishStream err, ", errMember)
		return nil, baseErr.ErrInternalServerError
	}

	if !isMember {
		l.appCtx.Logger().Error("PublishStream err, ", "not member", req.Uid)
		return nil, errorx.ErrNoPermission
	}

	videoEnable := true
	if room.Mode == 2 || room.Mode == 4 {
		videoEnable = false
	}
	pubReq := &rtcDto.PublishReq{
		ChannelId:   room.Id,
		UId:         fmt.Sprintf("%d", req.Uid),
		OfferSdp:    req.Sdp,
		AudioEnable: true,
		VideoEnable: videoEnable,
	}
	pubResp, errPub := l.appCtx.WebRTCApi().Publish(pubReq, claims)

	if errPub != nil {
		l.appCtx.Logger().Error("PublishStream err, ", errPub)
		return nil, baseErr.ErrInternalServerError
	}

	if pubResp == nil {
		l.appCtx.Logger().Error("PublishStream resp err", pubResp)
		return nil, baseErr.ErrInternalServerError
	}

	_ = l.roomLogic.OnUserJoinEvent(&dto.RoomUserJoinEvent{
		RoomId:    room.Id,
		UserId:    req.Uid,
		Timestamp: time.Now().UnixMilli(),
	}, claims)

	return &dto.PublishStreamResp{SessionId: pubResp.StreamKey, Sdp: pubResp.AnswerSdp, Type: "answer"}, nil
}

func (l WebRTCStreamLogic) SubscribeStream(req *dto.SubscribeStreamReq, claims baseDto.ThkClaims) (*dto.SubscribeStreamResp, error) {
	room, isMember, errMember := l.checkMember(req.RoomId, req.Uid, claims)
	if errMember != nil {
		l.appCtx.Logger().Error("SubscribeStream err, ", errMember)
		return nil, baseErr.ErrInternalServerError
	}

	if !isMember {
		l.appCtx.Logger().Error("SubscribeStream err, ", "not member", req.Uid)
		return nil, errorx.ErrNoPermission
	}

	playReq := &rtcDto.PlayReq{
		ChannelId: room.Id,
		UId:       fmt.Sprintf("%d", req.Uid),
		OfferSdp:  req.Sdp,
		StreamKey: req.SessionId,
	}
	playResp, errPlay := l.appCtx.WebRTCApi().Play(playReq, claims)
	if errPlay != nil {
		l.appCtx.Logger().Error("SubscribeStream err, ", errPlay)
		return nil, baseErr.ErrInternalServerError
	}

	if playResp == nil {
		l.appCtx.Logger().Error("SubscribeStream resp err", playResp)
		return nil, baseErr.ErrInternalServerError
	}

	res := &dto.SubscribeStreamResp{
		Renegotiation: false,
		Sdp:           playResp.AnswerSdp,
		Type:          "answer",
	}
	return res, nil
}

func (l WebRTCStreamLogic) UpdateStreamStatus(req *dto.StreamStatusUpdateReq, claims baseDto.ThkClaims) error {
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

func (l WebRTCStreamLogic) checkMember(roomId string, uId int64, claims baseDto.ThkClaims) (*dto.Room, bool, error) {
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
