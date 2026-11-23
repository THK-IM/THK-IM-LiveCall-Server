package room

import (
	"testing"

	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
)

func TestMediaEventsDoNotChangeMembership(t *testing.T) {
	m := memberState{Participant: dto.Participant{UId: 17, JoinTime: 123}, IssuedKey: "one"}
	if m.streamChanged("unknown", true) || !m.streamChanged("one", true) || m.streamChanged("one", true) {
		t.Fatal("start must match the issued session and be idempotent")
	}
	m.IssuedKey = "two"
	if !m.streamChanged("two", true) || m.streamChanged("one", false) || m.StreamKey != "two" {
		t.Fatal("old stop removed the replacement publisher")
	}
	if !m.streamChanged("two", false) || m.streamChanged("two", true) || m.streamChanged("two", false) {
		t.Fatal("stop must retire the key and ignore late starts")
	}
	if m.JoinTime != 123 || m.LeaveTime != 0 || m.Role != dto.Audience || m.StreamKey != "" {
		t.Fatalf("stopping changed room membership: %+v", m)
	}
	m.IssuedKey = "three"
	if !m.streamChanged("three", false) || m.streamChanged("three", true) {
		t.Fatal("stop before ready must cancel a pending publication")
	}
}
