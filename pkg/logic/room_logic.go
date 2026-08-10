package logic

import (
	"time"

	"github.com/sirupsen/logrus"
	baseDto "github.com/thk-im/thk-im-base-server/dto"
	baseErrorx "github.com/thk-im/thk-im-base-server/errorx"
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/errorx"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/room"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/signal"
)

type RoomLogic struct {
	appCtx        *app.Context
	roomService   room.Service
	signalService signal.Service
}

func NewRoomLogic(appCtx *app.Context) *RoomLogic {
	return &RoomLogic{
		appCtx:        appCtx,
		roomService:   room.NewCloudflareSFURoomService(appCtx),
		signalService: signal.NewSignalService(appCtx),
	}
}

func (l RoomLogic) CreateRoom(req *dto.RoomCreateReq, claims baseDto.ThkClaims) (*dto.RoomJoinResp, error) {
	return l.roomService.CreateRoom(req, claims)
}

func (l RoomLogic) QueryRoom(id string, claims baseDto.ThkClaims) (*dto.Room, error) {
	return l.roomService.FindRoomById(id, claims)
}

func (l RoomLogic) JoinRoom(req *dto.RoomJoinReq, claims baseDto.ThkClaims) (*dto.RoomJoinResp, error) {
	roomVo, err := l.roomService.FindRoomById(req.RoomId, claims)
	if err != nil {
		return nil, err
	}
	if roomVo == nil {
		return nil, errorx.ErrRoomNotExisted
	}

	resp, errRequestJoin := l.roomService.RequestJoinRoom(req, claims)
	if errRequestJoin != nil {
		return nil, errRequestJoin
	}

	{
		members := make([]int64, 0)
		for _, p := range roomVo.Participants {
			if p.UId != req.UId {
				members = append(members, p.UId)
			}
		}

		s := dto.MakeAcceptRequestSignal(
			roomVo.Id, "", req.UId, time.Now().UnixMilli(),
		)
		_ = l.signalService.PushSignal(s, members, claims)
	}

	return resp, nil
}

func (l RoomLogic) CallRoomMembers(req *dto.RoomCallReq, claims baseDto.ThkClaims) error {
	roomVo, err := l.roomService.FindRoomById(req.RoomId, claims)
	if err != nil {
		return err
	}
	if roomVo == nil {
		return errorx.ErrRoomNotExisted
	}

	if len(req.Members) == 0 {
		return baseErrorx.ErrParamsError
	}

	memberMap := make(map[int64]*dto.Participant)
	for _, member := range roomVo.Participants {
		memberMap[member.UId] = member
	}
	for _, member := range req.Members {
		if memberMap[member] == nil {
			_ = l.roomService.AddRoomMember(roomVo.Id, member, claims)
		}
	}

	s := dto.MakeBeingRequestedSignal(
		roomVo.Id, req.Members, roomVo.Mode, req.Msg, req.UId, roomVo.CreateTime,
		time.Now().UnixMilli()+req.Duration*1000,
	)
	errPush := l.signalService.PushSignal(s, req.Members, claims)
	return errPush
}

func (l RoomLogic) CancelCallRoomMembers(req *dto.CancelCallingReq, claims baseDto.ThkClaims) error {
	roomVo, err := l.roomService.FindRoomById(req.RoomId, claims)
	if err != nil {
		return err
	}
	if roomVo == nil {
		return errorx.ErrRoomNotExisted
	}

	if len(req.Members) > 0 {
		s := dto.MakeCancelRequestingSignal(
			roomVo.Id, req.Msg, roomVo.CreateTime, time.Now().UnixMilli(),
		)
		errPush := l.signalService.PushSignal(s, req.Members, claims)
		if errPush != nil {
			return errPush
		}
	}

	shouldEnd := false
	joinedMembers := make([]int64, 0)
	for _, p := range roomVo.Participants {
		if p.JoinTime > 0 {
			joinedMembers = append(joinedMembers, p.UId)
		}
	}
	if len(joinedMembers) > 1 {
		shouldEnd = false
	} else {
		shouldEnd = true
	}

	if shouldEnd {
		return l.roomService.DestroyRoom(roomVo.Id, claims)
	}

	return nil
}

func (l RoomLogic) InviteJoinRoom(req *dto.InviteJoinRoomReq, claims baseDto.ThkClaims) error {
	roomVo, err := l.roomService.FindRoomById(req.RoomId, claims)
	if err != nil {
		return err
	}
	if roomVo == nil {
		return errorx.ErrRoomNotExisted
	}

	memberMap := make(map[int64]*dto.Participant)
	for _, member := range roomVo.Participants {
		memberMap[member.UId] = member
	}
	for _, member := range req.InviteUIds {
		if memberMap[member] == nil {
			_ = l.roomService.AddRoomMember(roomVo.Id, member, claims)
		}
	}

	s := dto.MakeBeingRequestedSignal(
		roomVo.Id, req.InviteUIds, roomVo.Mode, req.Msg, req.UId, roomVo.CreateTime,
		time.Now().UnixMilli()+req.Duration*1000,
	)
	errPush := l.signalService.PushSignal(s, req.InviteUIds, claims)
	return errPush
}

func (l RoomLogic) RefuseJoinRoom(req *dto.RefuseJoinRoomReq, claims baseDto.ThkClaims) error {
	roomVo, err := l.roomService.FindRoomById(req.RoomId, claims)
	if err != nil {
		return err
	}
	if roomVo == nil {
		return errorx.ErrRoomNotExisted
	}
	members := make([]int64, 0)
	for _, p := range roomVo.Participants {
		if p.UId != req.UId {
			members = append(members, p.UId)
		}
	}

	errRefuse := l.roomService.RefuseJoinRoom(req.RoomId, req.UId, req.IsBusy, claims)
	if errRefuse != nil {
		return errRefuse
	}

	s := dto.MakeRejectRequestSignal(
		roomVo.Id, req.Msg, req.UId, time.Now().UnixMilli(),
	)

	errPush := l.signalService.PushSignal(s, members, claims)
	return errPush
}

func (l RoomLogic) RoomMemberLeave(req *dto.RoomMemberLeaveReq, claims baseDto.ThkClaims) error {
	roomVo, err := l.roomService.FindRoomById(req.RoomId, claims)
	if err != nil {
		return err
	}
	if roomVo == nil {
		return nil
	}
	members := make([]int64, 0)
	for _, p := range roomVo.Participants {
		members = append(members, p.UId)
	}
	if len(members) > 0 {
		s := dto.MakeHangupSignal(
			roomVo.Id, req.Msg, req.UId, time.Now().UnixMilli(),
		)
		errPush := l.signalService.PushSignal(s, members, claims)
		return errPush
	}
	return nil
}

func (l RoomLogic) KickoffRoomMember(req *dto.KickoffMemberReq, claims baseDto.ThkClaims) error {
	roomVo, err := l.roomService.FindRoomById(req.RoomId, claims)
	if err != nil {
		return err
	}
	if roomVo == nil {
		return errorx.ErrRoomNotExisted
	}
	hasPermission := roomVo.OwnerId == req.UId
	if !hasPermission {
		return errorx.ErrNoPermission
	}
	members := make([]int64, 0)
	for _, p := range roomVo.Participants {
		members = append(members, p.UId)
	}
	s := dto.MakeKickMemberSignal(
		roomVo.Id, req.Msg, req.UId, time.Now().UnixMilli(), req.KickoffUIds,
	)
	errPush := l.signalService.PushSignal(s, members, claims)
	return errPush
}

func (l RoomLogic) DeleteRoom(req *dto.RoomDelReq, claims baseDto.ThkClaims) error {
	roomVo, errRoom := l.roomService.FindRoomById(req.RoomId, claims)
	if errRoom != nil {
		return errRoom
	}
	if roomVo == nil {
		l.appCtx.Logger().WithFields(logrus.Fields(claims)).Trace("deleteRoom roomVo is nil ", req)
		return nil
	}
	if len(roomVo.Participants) > 0 {
		members := make([]int64, 0)
		for _, p := range roomVo.Participants {
			members = append(members, p.UId)
		}
		s := dto.MakeEndCallSignal(
			roomVo.Id, "", req.UId, time.Now().UnixMilli(),
		)
		errPush := l.signalService.PushSignal(s, members, claims)
		if errPush != nil {
			return errPush
		}
	}
	err := l.roomService.DestroyRoom(req.RoomId, claims)
	if err != nil {
		return err
	}

	return nil
}

func (l RoomLogic) OnUserJoinEvent(event *dto.RoomUserJoinEvent, claims baseDto.ThkClaims) error {
	return l.roomService.OnUserJoinEvent(event, claims)
}

func (l RoomLogic) OnUserLeaveEvent(event *dto.RoomUserLevelEvent, claims baseDto.ThkClaims) error {
	return l.roomService.OnUserLeaveEvent(event, claims)
}

func (l RoomLogic) OnUserPushEvent(event *dto.RoomUserPushStreamEvent, claims baseDto.ThkClaims) error {
	err := l.roomService.OnUserPushEvent(event, claims)
	if err != nil {
		l.appCtx.Logger().Error("OnUserPushEvent OnUserPushEvent", event, err, claims)
	}
	return nil
}
