package peer

import "github.com/pion/webrtc/v3"

// Payload wraps a message payload
type Payload struct {
	Type          string                     `json:"type"`
	ConnectionID  string                     `json:"connectionId"`
	Metadata      interface{}                `json:"metadata,omitempty"`
	Label         string                     `json:"label,omitempty"`
	Serialization string                     `json:"serialization,omitempty"`
	Reliable      bool                       `json:"reliable,omitempty"`
	Candidate     string                     `json:"candidate,omitempty"`
	SDP           *webrtc.SessionDescription `json:"sdp,omitempty"`
	Browser       string                     `json:"browser,omitempty"`
}

// Message the IMessage implementation
type Message struct {
	Type    string  `json:"type"`
	Payload Payload `json:"payload"`
	Src     string  `json:"src"`
	Dst     string  `json:"dst,omitempty"`
}

// GetPayload returns the message payload
func (m Message) GetPayload() Payload {
	return m.Payload
}

// GetSrc returns the message src
func (m Message) GetSrc() string {
	return m.Src
}

// GetDst returns the message dst
func (m Message) GetDst() string {
	return m.Dst
}

// GetType returns the message payload
func (m Message) GetType() string {
	return m.Type
}