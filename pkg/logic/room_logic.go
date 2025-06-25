package logic

import (
	baseDto "github.com/thk-im/thk-im-base-server/dto"
	baseErrorx "github.com/thk-im/thk-im-base-server/errorx"
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/errorx"
	msgDto "github.com/thk-im/thk-im-msgapi-server/pkg/dto"
	"time"
)

type RoomLogic struct {
	appCtx *app.Context
}

func NewRoomLogic(appCtx *app.Context) *RoomLogic {
	return &RoomLogic{appCtx: appCtx}
}

func (l RoomLogic) CreateRoom(req *dto.RoomCreateReq, claims baseDto.ThkClaims) (*dto.Room, error) {
	checkApi := l.appCtx.CheckApi()
	if checkApi != nil {
		checkReq := &dto.CheckLiveCallCreateReq{
			UId:      req.UId,
			RoomType: req.Mode,
		}
		errCheck := checkApi.CheckLiveCallCreate(checkReq, claims)
		if errCheck != nil {
			return nil, errCheck
		}
	}
	return l.appCtx.RoomService().CreateRoom(req)
}

func (l RoomLogic) QueryRoom(id string, claims baseDto.ThkClaims) (*dto.Room, error) {
	return l.appCtx.RoomService().FindRoomById(id)
}

func (l RoomLogic) JoinRoom(req *dto.RoomJoinReq, claims baseDto.ThkClaims) (*dto.Room, error) {
	room, err := l.appCtx.RoomService().RequestJoinRoom(req)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errorx.ErrRoomNotExisted
	}

	checkApi := l.appCtx.CheckApi()
	if checkApi != nil {
		checkReq := &dto.CheckLiveJoinReq{
			UId:  req.UId,
			Room: room,
		}
		errCheck := checkApi.CheckLiveCallJoin(checkReq, claims)
		if errCheck != nil {
			return nil, errCheck
		}
	}

	signal := dto.MakeAcceptRequestSignal(
		room.Id, "", req.UId, time.Now().UnixMilli(),
	)
	pushMessage := &msgDto.PushMessageReq{
		UIds:        []int64{room.OwnerId},
		Type:        l.appCtx.SignalType(),
		Body:        signal.JsonString(),
		OfflinePush: true,
	}
	_, errPush := l.appCtx.MsgApi().PushMessage(pushMessage, claims)
	return room, errPush
}

func (l RoomLogic) CallRoomMembers(req *dto.RoomCallReq, claims baseDto.ThkClaims) error {
	room, err := l.appCtx.RoomService().FindRoomById(req.RoomId)
	if err != nil {
		return err
	}
	if room == nil {
		return errorx.ErrRoomNotExisted
	}

	if len(req.Members) == 0 {
		return baseErrorx.ErrParamsError
	}

	signal := dto.MakeBeingRequestedSignal(
		room.Id, req.Members, room.Mode, req.Msg, req.UId, room.CreateTime,
		time.Now().UnixMilli()+req.Duration*1000,
	)
	pushMessage := &msgDto.PushMessageReq{
		UIds:        req.Members,
		Type:        l.appCtx.SignalType(),
		Body:        signal.JsonString(),
		OfflinePush: true,
	}
	_, errPush := l.appCtx.MsgApi().PushMessage(pushMessage, claims)
	return errPush
}

func (l RoomLogic) CancelCallRoomMembers(req *dto.CancelCallingReq, claims baseDto.ThkClaims) error {
	room, err := l.appCtx.RoomService().FindRoomById(req.RoomId)
	if err != nil {
		return err
	}
	if room == nil {
		return errorx.ErrRoomNotExisted
	}

	if len(req.Members) == 0 {
		return baseErrorx.ErrParamsError
	}

	signal := dto.MakeCancelRequestingSignal(
		room.Id, req.Msg, room.CreateTime, time.Now().UnixMilli(),
	)
	pushMessage := &msgDto.PushMessageReq{
		UIds:        req.Members,
		Type:        l.appCtx.SignalType(),
		Body:        signal.JsonString(),
		OfflinePush: true,
	}
	_, errPush := l.appCtx.MsgApi().PushMessage(pushMessage, claims)
	return errPush
}

func (l RoomLogic) InviteJoinRoom(req *dto.InviteJoinRoomReq, claims baseDto.ThkClaims) error {
	room, err := l.appCtx.RoomService().FindRoomById(req.RoomId)
	if err != nil {
		return err
	}
	if room == nil {
		return errorx.ErrRoomNotExisted
	}

	checkApi := l.appCtx.CheckApi()
	if checkApi != nil {
		checkReq := &dto.CheckLiveInviteReq{
			Room:       room,
			InviteUIds: req.InviteUIds,
			RequestUId: req.UId,
		}
		errCheck := checkApi.CheckLiveCallInvite(checkReq, claims)
		if errCheck != nil {
			return errCheck
		}
	}

	signal := dto.MakeBeingRequestedSignal(
		room.Id, req.InviteUIds, room.Mode, req.Msg, req.UId, room.CreateTime,
		time.Now().UnixMilli()+req.Duration*1000,
	)
	pushMessage := &msgDto.PushMessageReq{
		UIds:        req.InviteUIds,
		Type:        l.appCtx.SignalType(),
		Body:        signal.JsonString(),
		OfflinePush: true,
	}
	_, errPush := l.appCtx.MsgApi().PushMessage(pushMessage, claims)
	return errPush
}

func (l RoomLogic) RefuseJoinRoom(req *dto.RefuseJoinRoomReq, claims baseDto.ThkClaims) error {
	room, err := l.appCtx.RoomService().FindRoomById(req.RoomId)
	if err != nil {
		return err
	}
	if room == nil {
		return errorx.ErrRoomNotExisted
	}
	members := make([]int64, 0)
	for _, p := range room.Participants {
		if p.UId != req.UId {
			members = append(members, p.UId)
		}
	}
	signal := dto.MakeRejectRequestSignal(
		room.Id, req.Msg, req.UId, time.Now().UnixMilli(),
	)
	pushMessage := &msgDto.PushMessageReq{
		UIds:        members,
		Type:        l.appCtx.SignalType(),
		Body:        signal.JsonString(),
		OfflinePush: true,
	}
	_, errPush := l.appCtx.MsgApi().PushMessage(pushMessage, claims)
	return errPush
}

func (l RoomLogic) RoomMemberLeave(req *dto.RoomMemberLeaveReq, claims baseDto.ThkClaims) error {
	room, err := l.appCtx.RoomService().FindRoomById(req.RoomId)
	if err != nil {
		return err
	}
	if room == nil {
		return nil
	}
	members := make([]int64, 0)
	for _, p := range room.Participants {
		members = append(members, p.UId)
	}
	if len(members) > 0 {
		signal := dto.MakeHangupSignal(
			room.Id, req.Msg, req.UId, time.Now().UnixMilli(),
		)
		pushMessage := &msgDto.PushMessageReq{
			UIds:        members,
			Type:        l.appCtx.SignalType(),
			Body:        signal.JsonString(),
			OfflinePush: true,
		}
		_, errPush := l.appCtx.MsgApi().PushMessage(pushMessage, claims)
		return errPush
	}
	return nil
}

func (l RoomLogic) KickoffRoomMember(req *dto.KickoffMemberReq, claims baseDto.ThkClaims) error {
	room, err := l.appCtx.RoomService().FindRoomById(req.RoomId)
	if err != nil {
		return err
	}
	if room == nil {
		return errorx.ErrRoomNotExisted
	}
	hasPermission := room.OwnerId == req.UId
	if !hasPermission {
		return errorx.ErrNoPermission
	}
	members := make([]int64, 0)
	for _, p := range room.Participants {
		members = append(members, p.UId)
	}
	signal := dto.MakeKickMemberSignal(
		room.Id, req.Msg, req.UId, time.Now().UnixMilli(), req.KickoffUIds,
	)
	pushMessage := &msgDto.PushMessageReq{
		UIds:        members,
		Type:        l.appCtx.SignalType(),
		Body:        signal.JsonString(),
		OfflinePush: true,
	}
	_, errPush := l.appCtx.MsgApi().PushMessage(pushMessage, claims)
	return errPush
}

func (l RoomLogic) DeleteRoom(req *dto.RoomDelReq, claims baseDto.ThkClaims) error {
	room, errRoom := l.appCtx.RoomService().FindRoomById(req.RoomId)
	if errRoom != nil {
		return errRoom
	}
	if room == nil {
		return errorx.ErrRoomNotExisted
	}

	err := l.appCtx.RoomService().DestroyRoom(req.RoomId)
	if err != nil {
		return err
	}

	if len(room.Participants) > 0 {
		members := make([]int64, 0)
		for _, p := range room.Participants {
			members = append(members, p.UId)
		}
		signal := dto.MakeEndCallSignal(
			room.Id, "", req.UId, time.Now().UnixMilli(),
		)
		pushMessage := &msgDto.PushMessageReq{
			UIds:        members,
			Type:        l.appCtx.SignalType(),
			Body:        signal.JsonString(),
			OfflinePush: true,
		}
		_, _ = l.appCtx.MsgApi().PushMessage(pushMessage, claims)
	}
	return nil
}
