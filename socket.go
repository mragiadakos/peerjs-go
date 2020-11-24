package peer

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// SocketEvent carries an event from the socket
type SocketEvent struct {
	Type    string
	Message *Message
	Error   error
}

//NewSocket create a socket instance
func NewSocket(opts Options) Socket {
	s := Socket{
		Emitter: NewEmitter(),
		log:     createLogger("socket", opts.Debug),
	}
	s.opts = opts
	s.disconnected = true
	return s
}

//Socket abstract websocket exposing an event emitter like interface
type Socket struct {
	Emitter
	id           string
	opts         Options
	baseURL      string
	disconnected bool
	conn         *websocket.Conn
	log          *logrus.Entry
	mutex        sync.Mutex
}

func (s *Socket) buildBaseURL() string {
	proto := "ws"
	if s.opts.Secure {
		proto = "wss"
	}
	port := strconv.Itoa(s.opts.Port)
	return fmt.Sprintf(
		"%s://%s:%s%s/peerjs?key=%s",
		proto,
		s.opts.Host,
		port,
		s.opts.Path,
		s.opts.Key,
	)
}

//Start initiate the connection
func (s *Socket) Start(id string, token string) error {

	if !s.disconnected {
		return nil
	}

	if s.baseURL == "" {
		s.baseURL = s.buildBaseURL()
	}

	url := s.baseURL + fmt.Sprintf("&id=%s&token=%s", id, token)
	s.log.Debugf("Connecting to %s", url)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	s.conn = c

	// s.conn.SetCloseHandler(func(code int, text string) error {
	// 	s.log.Debug("Called close handler")
	// 	s.disconnected = true
	// 	s.Emit(SocketEventTypeDisconnected, SocketEvent{SocketEventTypeDisconnected, nil, nil})
	// 	return nil
	// })

	//  ws ping
	go func() {
		ticker := time.NewTicker(time.Millisecond * time.Duration(s.opts.PingInterval))
		defer func() {
			ticker.Stop()
			s.Close()
		}()
		for {
			select {
			case <-ticker.C:
				if s.conn == nil {
					return
				}
				s.mutex.Lock()
				if err := s.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					s.mutex.Unlock()
					return
				}
				s.mutex.Unlock()
				break
			}
		}
	}()

	// collect messages
	go func() {
		for {
			if s.conn == nil {
				return
			}

			msgType, raw, err := s.conn.ReadMessage()
			if err != nil {
				if ce, ok := err.(*websocket.CloseError); ok {
					switch ce.Code {
					case websocket.CloseNormalClosure,
						websocket.CloseGoingAway,
						websocket.CloseNoStatusReceived:
						return
					}
				}
				s.log.Warnf("WS read error: %s", err)
				continue
			}

			s.log.Infof("WS recv: %d %s", msgType, raw)

			if msgType == websocket.TextMessage {

				msg := Message{}
				err = json.Unmarshal(raw, &msg)
				if err != nil {
					s.log.Errorf("Failed to decode message=%s %s", string(raw), err)
				}

				s.Emit(SocketEventTypeMessage, SocketEvent{SocketEventTypeMessage, &msg, err})
			}

		}
	}()

	return nil
}

//Close close the websocket connection
func (s *Socket) Close() error {
	if s.disconnected {
		return nil
	}
	err := s.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
	if err != nil {
		s.log.Debugf("Failed to send close message: %s", err)
	}
	err = s.conn.Close()
	if err != nil {
		s.log.Warnf("WS close error: %s", err)
	}
	s.disconnected = true
	s.conn = nil
	return err
}

//Send send a message
func (s *Socket) Send(msg []byte) error {
	if s.conn == nil {
		return nil
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.conn.WriteMessage(websocket.TextMessage, msg)
}