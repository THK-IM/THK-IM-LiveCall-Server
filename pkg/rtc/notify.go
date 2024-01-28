package rtc

import "encoding/json"

type Notify struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func NewStreamNotify(msg string) *string {
	notify := &Notify{
		Type:    "NewStream",
		Message: msg,
	}
	if js, err := json.Marshal(notify); err != nil {
		return nil
	} else {
		jsStr := string(js)
		return &jsStr
	}
}

func RemoveStreamNotify(msg string) *string {
	notify := &Notify{
		Type:    "RemoveStream",
		Message: msg,
	}
	if js, err := json.Marshal(notify); err != nil {
		return nil
	} else {
		jsStr := string(js)
		return &jsStr
	}
}

func DataChanelMsg(msg string) *string {
	notify := &Notify{
		Type:    "DataChannelMsg",
		Message: msg,
	}
	if js, err := json.Marshal(notify); err != nil {
		return nil
	} else {
		jsStr := string(js)
		return &jsStr
	}
}
