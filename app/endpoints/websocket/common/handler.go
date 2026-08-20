package common

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	jsonpb "github.com/cosmos/gogoproto/jsonpb"
	proto "github.com/cosmos/gogoproto/proto"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CosmosContracts/juno/v31/x/stream/types"
	"github.com/CosmosContracts/juno/v31/x/stream/types/encoding"
)

const (
	// WebSocket configuration values used across handlers
	WriteWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Handler contains the shared state required to service WebSocket connections
type Handler struct {
	Logger         log.Logger
	Registry       *types.SubscriptionRegistry
	AppContextDone <-chan struct{}
	Upgrader       *websocket.Upgrader
}

// ConnectionParams contains parameters for establishing a WebSocket connection
type ConnectionParams struct {
	Writer   http.ResponseWriter
	Request  *http.Request
	Resolved *types.ResolvedStream
	Fetch    func(context.Context) (proto.Message, error)
}

// ServeConnection handles the standard WebSocket connection lifecycle
func (h *Handler) ServeConnection(params ConnectionParams) (retErr error) {
	if params.Resolved == nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "stream resolution missing")
	}
	if params.Fetch == nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "fetch function not provided")
	}
	if err := h.validateRequest(params); err != nil {
		return err
	}

	conn, err := h.upgradeConnection(params)
	if err != nil {
		return err
	}

	session := newConnectionSession(h, conn)
	defer session.finalize()
	defer func() {
		if r := recover(); r != nil {
			session.handlePanic(r)
		}
	}()

	registry := h.Registry
	if registry == nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "subscription registry not configured")
	}

	baseCtx, cancel := h.connectionContext()
	defer cancel()

	ctx, cancelClient := context.WithCancel(baseCtx)
	defer cancelClient()

	if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return err
	}
	conn.SetReadLimit(maxMessageSize)
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	go h.readPump(ctx, conn, cancelClient)

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	var writeMu sync.Mutex
	marshaler := jsonpb.Marshaler{OrigName: true, EmitDefaults: true}

	send := func(msg proto.Message) error {
		if msg == nil {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty response message")
		}

		payload, err := marshaler.MarshalToString(msg)
		if err != nil {
			return errorsmod.Wrap(err, "marshal response")
		}

		writeMu.Lock()
		defer writeMu.Unlock()

		if err := conn.SetWriteDeadline(time.Now().Add(WriteWait)); err != nil {
			return err
		}
		err = conn.WriteMessage(websocket.TextMessage, []byte(payload))

		return err
	}

	go h.pingLoop(ctx, cancelClient, conn, &writeMu, ticker)

	meta := types.ConnectionMetadata{
		RemoteAddr:   params.Request.RemoteAddr,
		ForwardedFor: params.Request.Header.Get("X-Forwarded-For"),
	}

	logger := h.Logger
	var onEvent func(encoding.StreamEvent)
	if logger != nil {
		subscriptionKey := types.SanitizeLog(params.Resolved.Key.String())
		onEvent = func(event encoding.StreamEvent) {
			logger.Debug("websocket stream event received",
				"module", event.Module,
				"method", event.Method,
				"params", types.SanitizeLogMap(event.Params),
				"subscription", subscriptionKey,
			)
		}
	}

	err = types.RunStream(ctx, types.RunStreamParams{
		Registry:       registry,
		Key:            params.Resolved.Key,
		ConnectionKind: types.ConnectionKindWebsocket,
		ConnectionMeta: meta,
		Fetch: func(runCtx context.Context) (proto.Message, error) {
			if err := runCtx.Err(); err != nil {
				return nil, err
			}
			return params.Fetch(runCtx)
		},
		Send:    send,
		OnEvent: onEvent,
		Logger:  logger,
		OnConnect: func(connectionID string) error {
			if !h.allowConnection(conn) {
				return types.ErrCircuitBreakerOpen
			}
			session.setConnectionID(connectionID)
			return nil
		},
	})

	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil
	}
	if errors.Is(err, types.ErrCircuitBreakerOpen) {
		return nil
	}
	if errors.Is(err, types.ErrSubscriptionLimit) {
		session.setCloseError(closeError{code: websocket.ClosePolicyViolation, msg: "subscription limit reached"})
		return err
	}
	if errors.Is(err, types.ErrConnectionLimit) {
		session.setCloseError(closeError{code: websocket.ClosePolicyViolation, msg: "connection limit exceeded"})
		return err
	}

	session.setCloseError(closeError{code: websocket.CloseInternalServerErr, msg: "internal server error"})
	return err
}

func (h *Handler) allowConnection(conn *websocket.Conn) bool {
	errorMsg := map[string]string{"error": "service temporarily unavailable"}
	if err := conn.SetWriteDeadline(time.Now().Add(WriteWait)); err != nil {
		h.Logger.Error("failed to set write deadline", "error", err)
	}
	if err := conn.WriteJSON(errorMsg); err != nil {
		h.Logger.Error("failed to write error message", "error", err)
	}
	if err := conn.Close(); err != nil {
		h.Logger.Error("failed to close connection", "error", err)
	}
	return false
}

// GetUpgrader returns a websocket upgrader with configurable CORS
func GetUpgrader(allowedOrigins []string) *websocket.Upgrader {
	allowAll := false
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		if origin == "*" {
			allowAll = true
			break
		}
		if origin != "" {
			allowed[origin] = struct{}{}
		}
	}

	return &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			if allowAll {
				return true
			}

			origin := r.Header.Get("Origin")
			if origin == "" {
				return true
			}

			if len(allowedOrigins) == 0 {
				return origin == "http://"+r.Host || origin == "https://"+r.Host
			}

			_, ok := allowed[origin]
			return ok
		},
	}
}

func (*Handler) readPump(ctx context.Context, conn *websocket.Conn, cancel context.CancelFunc) {
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			if cancel != nil {
				cancel()
			}
			return
		}
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func (*Handler) pingLoop(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn, writeMu *sync.Mutex, ticker *time.Ticker) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			writeMu.Lock()
			if err := conn.SetWriteDeadline(time.Now().Add(WriteWait)); err != nil {
				writeMu.Unlock()
				if cancel != nil {
					cancel()
				}
				return
			}
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				writeMu.Unlock()
				if cancel != nil {
					cancel()
				}
				return
			}
			writeMu.Unlock()
		}
	}
}

// CloseWithMessage closes the WebSocket connection with a specific close code and message
func CloseWithMessage(conn *websocket.Conn, code int, message string) error {
	if len(message) > 123 {
		message = message[:123]
	}
	deadline := time.Now().Add(WriteWait)
	if err := conn.SetWriteDeadline(deadline); err != nil {
		return err
	}
	closeMsg := websocket.FormatCloseMessage(code, message)
	err := conn.WriteMessage(websocket.CloseMessage, closeMsg)
	if err != nil {
		_ = conn.WriteControl(websocket.CloseMessage, closeMsg, deadline)
	}
	return err
}

func (h *Handler) validateRequest(params ConnectionParams) error {
	if h.Registry != nil && !h.Registry.CanAcceptConnection(types.ConnectionKindWebsocket) {
		http.Error(params.Writer, "connection limit exceeded", http.StatusServiceUnavailable)
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "connection limit exceeded")
	}

	return nil
}

func (h *Handler) upgradeConnection(params ConnectionParams) (*websocket.Conn, error) {
	conn, err := h.Upgrader.Upgrade(params.Writer, params.Request, nil)
	if err != nil {
		h.Logger.Error("websocket upgrade failed", "error", err)
		return nil, err
	}
	return conn, nil
}

func (h *Handler) connectionContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	if h.AppContextDone != nil {
		appDone := h.AppContextDone
		go func() {
			select {
			case <-appDone:
				cancel()
			case <-ctx.Done():
			}
		}()
	}
	return ctx, cancel
}

type closeError struct {
	code int
	msg  string
}

func (e closeError) Error() string {
	return e.msg
}

type connectionSession struct {
	handler      *Handler
	conn         *websocket.Conn
	connectionID string
	closeErr     error
}

func newConnectionSession(h *Handler, conn *websocket.Conn) *connectionSession {
	return &connectionSession{handler: h, conn: conn}
}

func (s *connectionSession) setConnectionID(id string) {
	s.connectionID = id
}

func (s *connectionSession) setCloseError(err error) {
	if err == nil {
		return
	}
	s.closeErr = err
}

func (s *connectionSession) handlePanic(recovered any) {
	s.handler.Logger.Error("panic in WebSocket handler", "panic", recovered)
	closeMsg := fmt.Sprintf("Internal error: %v", recovered)
	if len(closeMsg) > 123 {
		closeMsg = closeMsg[:123]
	}
	s.closeErr = closeError{code: websocket.CloseInternalServerErr, msg: closeMsg}
}

func (s *connectionSession) finalize() {
	if err := s.closeErr; err != nil {
		var ce closeError
		if errors.As(err, &ce) {
			s.handler.Logger.Info("sending close message", "code", ce.code, "msg", ce.msg)
			if err := CloseWithMessage(s.conn, ce.code, ce.msg); err != nil {
				s.handler.Logger.Error("failed to send close message", "error", err)
			}
		}
	} else {
		if err := CloseWithMessage(s.conn, websocket.CloseGoingAway, "server shutting down"); err != nil {
			s.handler.Logger.Error("failed to send close message", "error", err)
		}
	}

	if err := s.conn.Close(); err != nil && s.connectionID != "" {
		s.handler.Logger.Error("failed to close websocket connection", "error", err, "connection_id", s.connectionID)
	}
}
