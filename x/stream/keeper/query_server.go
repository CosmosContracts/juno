package keeper

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	proto "github.com/cosmos/gogoproto/proto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/CosmosContracts/juno/v31/x/stream/types"
	"github.com/CosmosContracts/juno/v31/x/stream/types/encoding"
)

var _ types.QueryServer = queryServer{}

func NewQueryServer(keeper *Keeper) types.QueryServer {
	return queryServer{k: keeper}
}

type queryServer struct {
	k *Keeper
}

func (q queryServer) Stream(req *types.StreamDynamicRequest, srv types.Query_StreamServer) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	module := strings.ToLower(strings.TrimSpace(req.Module))
	method := strings.ToLower(strings.TrimSpace(req.Method))
	if module == "" || method == "" {
		return status.Error(codes.InvalidArgument, "module and method are required")
	}

	methodRegistry := q.k.MethodRegistry()
	if methodRegistry == nil {
		return status.Error(codes.Unavailable, "method registry not initialised")
	}

	resolved, err := types.ResolveStream(methodRegistry, module, method, req.Params)
	if err != nil {
		return classifyResolveError(err)
	}

	ctx, cancel := q.attachAppLifecycle(srv.Context())
	defer cancel()

	send := func(msg proto.Message) error {
		if msg == nil {
			return status.Error(codes.Internal, "empty response message")
		}
		anyResult, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			return status.Errorf(codes.Internal, "pack stream response: %v", err)
		}
		return srv.Send(&types.StreamDynamicResponse{Result: anyResult})
	}

	invoker := q.k.Invoker()

	fetch := func(fetchCtx context.Context) (proto.Message, error) {
		if err := fetchCtx.Err(); err != nil {
			return nil, err
		}
		result, err := invoker.Execute(resolved.Descriptor, resolved.Fields)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return result, nil
	}

	var onEvent func(encoding.StreamEvent)
	logger := q.k.Logger()
	if logger != nil {
		subscriptionKey := resolved.Key.String()
		onEvent = func(event encoding.StreamEvent) {
			logger.Debug("stream event received",
				"module", event.Module,
				"method", event.Method,
				"params", event.Params,
				"subscription", subscriptionKey,
			)
		}
	}

	registry := q.k.Registry()

	meta := types.ConnectionMetadata{}
	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		meta.RemoteAddr = p.Addr.String()
	}

	err = types.RunStream(ctx, types.RunStreamParams{
		Registry:       registry,
		Key:            resolved.Key,
		ConnectionKind: types.ConnectionKindGRPC,
		ConnectionMeta: meta,
		Fetch:          fetch,
		Send:           send,
		OnEvent:        onEvent,
		Logger:         logger,
	})
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, types.ErrConnectionLimit):
		return status.Error(codes.ResourceExhausted, "gRPC connection limit exceeded")
	case errors.Is(err, types.ErrSubscriptionLimit):
		return status.Error(codes.ResourceExhausted, "subscription limit reached")
	}

	if s, ok := status.FromError(err); ok && s.Code() != codes.Unknown {
		return err
	}
	return status.Error(codes.Internal, err.Error())
}

func (q queryServer) attachAppLifecycle(ctx context.Context) (context.Context, context.CancelFunc) {
	appCtx := q.k.GetAppContext()
	if appCtx == nil {
		return ctx, func() {}
	}

	cancelCtx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-appCtx.Done():
			q.k.Logger().Debug("app context cancelled, closing stream")
			cancel()
		case <-ctx.Done():
		}
	}()

	return cancelCtx, cancel
}

func classifyResolveError(err error) error {
	switch {
	case errors.Is(err, types.ErrMissingModuleOrMethod):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, types.ErrStreamUnavailable):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.InvalidArgument, err.Error())
	}
}
