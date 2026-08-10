package signal

import (
	"encoding/json"
	"time"

	baseDto "github.com/thk-im/thk-im-base-server/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	msgDto "github.com/thk-im/thk-im-msgapi-server/pkg/dto"
)

type Service struct {
	appCtx *app.Context
}

const CallMsgType = 14

func NewSignalService(appCtx *app.Context) Service {
	return Service{appCtx: appCtx}
}

func (s Service) PushSignal(signal *dto.LiveCallSignal, toUIds []int64, claims baseDto.ThkClaims) error {
	if s.appCtx.MsgApi() == nil {
		return nil
	}
	pushMessage := &msgDto.PushMessageReq{
		UIds:        toUIds,
		Type:        400,
		Body:        signal.JsonString(),
		OfflinePush: true,
	}
	_, errPush := s.appCtx.MsgApi().PushMessage(pushMessage, claims)
	return errPush
}

func (s Service) SendLiveCallMsgByEnded(room *dto.Room, claims baseDto.ThkClaims) error {
	if s.appCtx.MsgApi() == nil {
		return nil
	}
	if room.SessionId == nil {
		return nil
	}

	roomJson, errJson := room.Json()
	if errJson != nil {
		s.appCtx.Logger().Error("SendLiveCallMsgByEnded", errJson)
	}
	s.appCtx.Logger().Trace("SendLiveCallMsgByEnded", roomJson)

	callMsg := dto.BuildCallMsg(room)
	msgBody, errBody := json.Marshal(callMsg)
	if errBody != nil {
		return errBody
	}
	req := &msgDto.SendMessageReq{
		CId:       s.appCtx.SnowflakeNode().Generate().Int64(),
		SId:       *room.SessionId,
		Type:      CallMsgType,
		CTime:     time.Now().UnixMilli(),
		Body:      string(msgBody),
		FUid:      room.OwnerId,
		RMsgId:    nil,
		AtUsers:   nil,
		Receivers: nil,
		ExtData:   nil,
	}
	_, errSend := s.appCtx.MsgApi().SendSessionMessage(req, claims)
	return errSend
}
