package dto

import "time"

type (
	PostMeetingReq struct {
		Title           string `json:"title"`
		PreferredRegion string `json:"preferred_region"`
	}

	PostMeetingResp struct {
		Id              string    `json:"id"`
		Title           string    `json:"title"`
		PreferredRegion string    `json:"preferred_region"`
		CreatedAt       time.Time `json:"created_at"`
	}

	AddParticipantReq struct {
		PresetName          string `json:"preset_name"`
		CustomParticipantID string `json:"custom_participant_id"`
	}

	AddParticipantResp struct {
		ID                  string    `json:"id"`
		Name                string    `json:"name"`
		Picture             string    `json:"picture"`
		CustomParticipantID string    `json:"custom_participant_id"`
		PresetName          string    `json:"preset_name"`
		CreatedAt           time.Time `json:"created_at"`
		UpdatedAt           time.Time `json:"updated_at"`
		Token               string    `json:"token"`
	}
)
