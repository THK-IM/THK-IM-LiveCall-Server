package dto

import (
	"bufio"
	"strings"
)

// =========================
// 基础结构
// =========================

type SessionDescription struct {
	SDP  string `json:"sdp,omitempty"`
	Type string `json:"type,omitempty"` // offer / answer
}

// =========================
// Track
// =========================

type TrackObject struct {
	Location                 string           `json:"location,omitempty"` // local / remote
	Mid                      string           `json:"mid,omitempty"`
	SessionID                string           `json:"sessionId,omitempty"`
	TrackName                string           `json:"trackName,omitempty"`
	BidirectionalMediaStream bool             `json:"bidirectionalMediaStream,omitempty"`
	Kind                     string           `json:"kind,omitempty"` // audio / video
	Simulcast                *SimulcastConfig `json:"simulcast,omitempty"`
}

type SimulcastConfig struct {
	PreferredRid     string `json:"preferredRid,omitempty"`
	PriorityOrdering string `json:"priorityOrdering,omitempty"` // none / asciibetical

	RidNotAvailable string `json:"ridNotAvailable,omitempty"`
}

// =========================
// Close Track
// =========================

type CloseTrackObject struct {
	Mid string `json:"mid,omitempty"`
}

// =========================
// 通用 Tracks
// =========================

type TracksRequest struct {
	SessionDescription *SessionDescription `json:"sessionDescription,omitempty"`
	Tracks             []TrackObject       `json:"tracks,omitempty"`
	AutoDiscover       bool                `json:"autoDiscover,omitempty"`
}

type TracksResponse struct {
	ErrorCode                      string              `json:"errorCode,omitempty"`
	ErrorDescription               string              `json:"errorDescription,omitempty"`
	RequiresImmediateRenegotiation bool                `json:"requiresImmediateRenegotiation,omitempty"`
	SessionDescription             *SessionDescription `json:"sessionDescription,omitempty"`
	Tracks                         []TrackResponse     `json:"tracks,omitempty"`
}

type TrackResponse struct {
	TrackObject
	ErrorCode        string `json:"errorCode,omitempty"`
	ErrorDescription string `json:"errorDescription,omitempty"`
}

// =========================
// Session
// =========================

type NewSessionRequest struct {
	SessionDescription *SessionDescription `json:"sessionDescription,omitempty"`
}

type NewSessionResponse struct {
	ErrorCode          string              `json:"errorCode,omitempty"`
	ErrorDescription   string              `json:"errorDescription,omitempty"`
	SessionDescription *SessionDescription `json:"sessionDescription,omitempty"`
	SessionID          string              `json:"sessionId"`
}

// =========================
// Close Tracks
// =========================

type CloseTracksRequest struct {
	SessionDescription *SessionDescription `json:"sessionDescription,omitempty"`
	Tracks             []CloseTrackObject  `json:"tracks,omitempty"`
	Force              bool                `json:"force,omitempty"`
}

type CloseTracksResponse struct {
	ErrorCode                      string               `json:"errorCode,omitempty"`
	ErrorDescription               string               `json:"errorDescription,omitempty"`
	SessionDescription             *SessionDescription  `json:"sessionDescription,omitempty"`
	Tracks                         []CloseTrackResponse `json:"tracks,omitempty"`
	RequiresImmediateRenegotiation bool                 `json:"requiresImmediateRenegotiation,omitempty"`
}

type CloseTrackResponse struct {
	CloseTrackObject
	ErrorCode        string `json:"errorCode,omitempty"`
	ErrorDescription string `json:"errorDescription,omitempty"`
}

// =========================
// Session State
// =========================

type GetSessionStateResponse struct {
	ErrorCode        string              `json:"errorCode,omitempty"`
	ErrorDescription string              `json:"errorDescription,omitempty"`
	Tracks           []SessionTrackState `json:"tracks,omitempty"`
	DataChannels     []interface{}       `json:"dataChannels"`
}

type SessionTrackState struct {
	TrackObject
	Status string `json:"status,omitempty"` // active / inactive / waiting
}

// =========================
// Renegotiate
// =========================

type RenegotiateRequest struct {
	SessionDescription *SessionDescription `json:"sessionDescription,omitempty"`
}

type RenegotiateResponse struct {
	ErrorCode          string              `json:"errorCode,omitempty"`
	ErrorDescription   string              `json:"errorDescription,omitempty"`
	SessionDescription *SessionDescription `json:"sessionDescription,omitempty"`
}

// =========================
// Change Tracks（Map 版本）
// =========================

type ChangeTracksRequest struct {
	Tracks             map[string]TrackObject `json:"tracks,omitempty"`
	SessionDescription *SessionDescription    `json:"sessionDescription,omitempty"`
}

// =========================
// Update Tracks
// =========================

type UpdateTracksRequest struct {
	Tracks             []TrackObject       `json:"tracks,omitempty"`
	SessionDescription *SessionDescription `json:"sessionDescription,omitempty"`
}

type UpdateTracksResponse struct {
	ErrorCode                      string          `json:"errorCode,omitempty"`
	ErrorDescription               string          `json:"errorDescription,omitempty"`
	RequiresImmediateRenegotiation bool            `json:"requiresImmediateRenegotiation,omitempty"`
	Tracks                         []TrackResponse `json:"tracks,omitempty"`
}

// =========================
// Adapter
// =========================

type AdapterObject struct {
	Location    string `json:"location,omitempty"` // local / remote
	SessionID   string `json:"sessionId,omitempty"`
	TrackName   string `json:"trackName,omitempty"`
	Endpoint    string `json:"endpoint,omitempty"`
	OutputCodec string `json:"outputCodec,omitempty"` // pcm / jpeg
	InputCodec  string `json:"inputCodec,omitempty"`  // pcm
	Mode        string `json:"mode,omitempty"`        // stream / buffer
	AdapterID   string `json:"adapterId,omitempty"`
}

type AdapterStatObject struct {
	AdapterID      string  `json:"adapterId,omitempty"`
	BytesProcessed float64 `json:"bytesProcessed,omitempty"`
}

type NewAdapterRequest struct {
	Tracks []AdapterObject `json:"tracks,omitempty"`
}

type NewAdapterResponse struct {
	ErrorCode        string                 `json:"errorCode,omitempty"`
	ErrorDescription string                 `json:"errorDescription,omitempty"`
	Tracks           []AdapterResponseTrack `json:"tracks,omitempty"`
}

type AdapterResponseTrack struct {
	AdapterObject
	ErrorCode        string `json:"errorCode,omitempty"`
	ErrorDescription string `json:"errorDescription,omitempty"`
}

type CloseAdapterRequest struct {
	Tracks []struct {
		AdapterID string `json:"adapterId,omitempty"`
	} `json:"tracks,omitempty"`
}

type CloseAdapterResponse struct {
	ErrorCode        string                     `json:"errorCode,omitempty"`
	ErrorDescription string                     `json:"errorDescription,omitempty"`
	Tracks           []CloseAdapterResponseItem `json:"tracks,omitempty"`
}

type CloseAdapterResponseItem struct {
	AdapterStatObject
	ErrorCode        string `json:"errorCode,omitempty"`
	ErrorDescription string `json:"errorDescription,omitempty"`
}

// =========================
// Data Channel
// =========================

type DataChannelObject struct {
	Location        string `json:"location,omitempty"`
	SessionID       string `json:"sessionId,omitempty"`
	DataChannelName string `json:"dataChannelName,omitempty"`
	ID              int    `json:"id,omitempty"`
}

type DataChannelsRequest struct {
	DataChannels []DataChannelObject `json:"dataChannels,omitempty"`
}

type DataChannelsResponse struct {
	ErrorCode        string                    `json:"errorCode,omitempty"`
	ErrorDescription string                    `json:"errorDescription,omitempty"`
	DataChannels     []DataChannelResponseItem `json:"dataChannels,omitempty"`
}

type DataChannelResponseItem struct {
	DataChannelObject
	ErrorCode        string `json:"errorCode,omitempty"`
	ErrorDescription string `json:"errorDescription,omitempty"`
}

// =========================
// DataChannel Transport
// =========================

type EstablishDataChannelsTransportRequest struct {
	SessionDescription *SessionDescription `json:"sessionDescription,omitempty"`
	DataChannel        *DataChannelObject  `json:"dataChannel,omitempty"`
}

type EstablishDataChannelsTransportResponse struct {
	ErrorCode                      string              `json:"errorCode,omitempty"`
	ErrorDescription               string              `json:"errorDescription,omitempty"`
	SessionDescription             *SessionDescription `json:"sessionDescription,omitempty"`
	RequiresImmediateRenegotiation bool                `json:"requiresImmediateRenegotiation,omitempty"`
	DataChannel                    *DataChannelObject  `json:"dataChannel,omitempty"`
}

// =========================
// Close DataChannel
// =========================

type CloseDataChannelsRequest struct {
	DataChannels []DataChannelObject `json:"dataChannels,omitempty"`
}

type CloseDataChannelsResponse struct {
	ErrorCode        string                    `json:"errorCode,omitempty"`
	ErrorDescription string                    `json:"errorDescription,omitempty"`
	DataChannels     []DataChannelResponseItem `json:"dataChannels,omitempty"`
}

func ParseTracksFromSDP(sdp string, sessionId string) []TrackObject {
	var tracks []TrackObject

	scanner := bufio.NewScanner(strings.NewReader(sdp))

	var currentKind string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 找 m=
		if strings.HasPrefix(line, "m=") {
			if strings.HasPrefix(line, "m=audio") {
				currentKind = "audio"
			} else if strings.HasPrefix(line, "m=video") {
				currentKind = "video"
			} else {
				currentKind = ""
			}
		}

		// 找 mid
		if strings.HasPrefix(line, "a=mid:") && currentKind != "" {
			mid := strings.TrimPrefix(line, "a=mid:")

			trackName := ""
			if currentKind == "audio" {
				trackName = "mic"
			} else if currentKind == "video" {
				trackName = "camera"
			}

			tracks = append(tracks, TrackObject{
				Location:  "local",
				Mid:       mid,
				SessionID: sessionId,
				Kind:      currentKind,
				TrackName: trackName,
			})
		}
	}

	return tracks
}
