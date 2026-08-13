package types

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/runtime/protoiface"

	proto "github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/CosmosContracts/juno/v31/x/stream/types/encoding"
)

// RouterInvoker routes gRPC queries through BaseApp's GRPCQueryRouter.
type RouterInvoker struct {
	app   *baseapp.BaseApp
	codec codec.BinaryCodec
}

var (
	errBaseAppNil              = errors.New("baseapp cannot be nil")
	errCodecNil                = errors.New("binary codec cannot be nil")
	errRequestNil              = errors.New("request message cannot be nil")
	errResponseNil             = errors.New("response message cannot be nil")
	errRequestNotProto         = errors.New("request type does not implement gogoproto.Message")
	errNoHybridHandler         = errors.New("no hybrid handler registered")
	errNoHandlerSucceeded      = errors.New("no handler succeeded")
	errResponseNotProtoMessage = errors.New("response does not implement proto.Message")
)

// NewRouterInvoker creates a new RouterInvoker bound to the given BaseApp and codec.
func NewRouterInvoker(app *baseapp.BaseApp, binaryCodec codec.BinaryCodec) (*RouterInvoker, error) {
	if app == nil {
		return nil, errBaseAppNil
	}
	if binaryCodec == nil {
		return nil, errCodecNil
	}
	return &RouterInvoker{
		app:   app,
		codec: binaryCodec,
	}, nil
}

// Invoke dispatches the request via the router using the latest committed height
// and fills the provided response message.
func (ri *RouterInvoker) Invoke(req protoiface.MessageV1, resp protoiface.MessageV1) error {
	if req == nil {
		return errRequestNil
	}
	if resp == nil {
		return errResponseNil
	}

	router := ri.app.GRPCQueryRouter()
	reqMsg, ok := req.(proto.Message)
	if !ok {
		return fmt.Errorf("%w: %T", errRequestNotProto, req)
	}

	requestName := proto.MessageName(reqMsg)
	handlers := router.HybridHandlerByRequestName(requestName)
	if len(handlers) == 0 {
		return fmt.Errorf("%w: %s", errNoHybridHandler, requestName)
	}

	var height int64
	if ri.app != nil {
		height = ri.app.LastCommitID().Version
	}
	sdkCtx, err := ri.app.CreateQueryContext(height, false)
	if err != nil {
		return fmt.Errorf("create query context: %w", err)
	}

	var lastErr error
	for _, handler := range handlers {
		if handler == nil {
			continue
		}
		if err := handler(sdkCtx, req, resp); err != nil {
			lastErr = err
			continue
		}
		return nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("%w: %s", errNoHandlerSucceeded, requestName)
	}
	return lastErr
}

// BuildRequest constructs a protobuf request message using the provided method
// descriptor and field map.
func (*RouterInvoker) BuildRequest(desc *encoding.MethodDescriptor, fields map[string]string) (protoiface.MessageV1, error) {
	return encoding.BuildRequestMessage(desc, fields)
}

// BuildResponse allocates an empty response message for the provided method descriptor.
func (*RouterInvoker) BuildResponse(desc *encoding.MethodDescriptor) (protoiface.MessageV1, error) {
	return encoding.NewMessageByName(desc.ResponseType)
}

// Execute builds request/response messages and invokes the query using the
// latest committed height.
func (ri *RouterInvoker) Execute(desc *encoding.MethodDescriptor, fields map[string]string) (proto.Message, error) {
	reqMsg, err := ri.BuildRequest(desc, fields)
	if err != nil {
		return nil, err
	}
	respMsg, err := ri.BuildResponse(desc)
	if err != nil {
		return nil, err
	}
	if err := ri.Invoke(reqMsg, respMsg); err != nil {
		return nil, err
	}

	respProto, ok := respMsg.(proto.Message)
	if !ok {
		return nil, errResponseNotProtoMessage
	}
	return respProto, nil
}
