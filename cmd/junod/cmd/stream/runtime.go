package stream

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	proto "github.com/cosmos/gogoproto/proto"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/client"
	clientflags "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	streamtypes "github.com/CosmosContracts/juno/v31/x/stream/types"
	"github.com/CosmosContracts/juno/v31/x/stream/types/encoding"
)

func runDynamicStreamCommand(cmd *cobra.Command, ctx client.Context, descriptor *encoding.MethodDescriptor, params map[string]string) error {
	if descriptor == nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "method descriptor required")
	}
	if ctx.GRPCClient == nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "grpc client not configured")
	}

	queryClient := streamtypes.NewQueryClient(ctx.GRPCClient)
	req := &streamtypes.StreamDynamicRequest{
		Module: descriptor.Module,
		Method: descriptor.StreamName,
		Params: params,
	}

	stream, err := queryClient.Stream(cmd.Context(), req)
	if err != nil {
		return err
	}

	return consumeStream(ctx, func() (proto.Message, error) {
		resp, err := stream.Recv()
		if err != nil {
			return nil, err
		}
		if resp.Result == nil {
			return nil, errorsmod.Wrap(sdkerrors.ErrInvalidType, "empty stream response")
		}
		return decodeDynamicAny(ctx.Codec, descriptor.ResponseType, resp.Result)
	})
}

func consumeStream(clientCtx client.Context, next func() (proto.Message, error)) error {
	for {
		msg, err := next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			//nolint:exhaustive // We only treat canceled and deadline errors specially.
			switch status.Code(err) {
			case codes.Canceled, codes.DeadlineExceeded:
				return nil
			default:
				return err
			}
		}

		if err := clientCtx.PrintProto(msg); err != nil {
			return err
		}
	}
}

func buildStreamingClientContext(cmd *cobra.Command) (client.Context, error) {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return client.Context{}, err
	}

	if clientCtx.GRPCClient == nil {
		grpcAddr, _ := cmd.Flags().GetString(clientflags.FlagGRPC)
		if grpcAddr == "" && clientCtx.Viper != nil {
			if viaFlag := clientCtx.Viper.GetString(clientflags.FlagGRPC); viaFlag != "" {
				grpcAddr = viaFlag
			}
			if grpcAddr == "" {
				grpcAddr = clientCtx.Viper.GetString("grpc.address")
			}
		}
		if grpcAddr == "" {
			grpcAddr = "localhost:9090"
		}

		useInsecure := true
		insecureChanged := cmd.Flags().Changed(clientflags.FlagGRPCInsecure)
		if insecureChanged {
			useInsecure, _ = cmd.Flags().GetBool(clientflags.FlagGRPCInsecure)
		}

		conn, err := dialEndpoint(
			grpcAddr,
			useInsecure,
			useInsecure && !insecureChanged,
		)
		if err != nil {
			return client.Context{}, fmt.Errorf("failed to dial gRPC endpoint %s: %w", grpcAddr, err)
		}

		clientCtx = clientCtx.WithGRPCClient(conn)
	}

	return clientCtx, nil
}

func dialEndpoint(target string, useInsecure bool, allowFallback bool) (*grpc.ClientConn, error) {
	if useInsecure {
		conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil || !allowFallback {
			return conn, err
		}
	}

	return grpc.NewClient(
		target,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})),
	)
}

func decodeDynamicAny(
	marshaler codec.Codec,
	typeName string,
	anyMsg *codectypes.Any,
) (proto.Message, error) {
	if anyMsg == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidType, "stream result is nil")
	}
	if marshaler == nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidType, "codec not configured for decoding %s", anyMsg.TypeUrl)
	}
	if typeName == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidType, "response type not specified")
	}

	msg, err := encoding.NewMessageByName(typeName)
	if err != nil {
		return nil, err
	}

	protoMsg, ok := msg.(proto.Message)
	if !ok {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidType, "type %s does not implement proto.Message", typeName)
	}

	if err := marshaler.Unmarshal(anyMsg.Value, protoMsg); err != nil {
		return nil, errorsmod.Wrapf(err, "decode %s", anyMsg.TypeUrl)
	}
	return protoMsg, nil
}
