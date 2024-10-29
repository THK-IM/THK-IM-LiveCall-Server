package room

const (
	DestroyRoomEventKey = "DestroyRoomEvent"
)

type DestroyRoomEvent struct {
	RoomId string `json:"room_id"`
}
