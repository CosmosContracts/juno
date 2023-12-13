package keeper

import (
	"context"
	b64 "encoding/base64"
	"fmt"

	"github.com/CosmosContracts/juno/v19/x/youtube/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.MsgServer = &msgServer{}

// msgServer is a wrapper of Keeper.
type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the x/youtube MsgServer interface.
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{
		Keeper: k,
	}
}

// Upload implements types.MsgServer.
func (ms *msgServer) Upload(ctx context.Context, msg *types.MsgUploadContentBlob) (*types.MsgUploadContentBlobResponse, error) {
	// type MsgUploadContentBlob struct {
	// 	// When saving data, we prefix it with the sender so that every sender can
	// 	// upload at the same time with unique video ids
	// 	// NOTE: This requires only 1 upload at a time.
	// 	Sender  string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	// 	IdKey   uint64 `protobuf:"varint,2,opt,name=id_key,json=idKey,proto3" json:"id_key,omitempty" yaml:"id_key"`
	// 	Content []byte `protobuf:"bytes,3,opt,name=content,proto3" json:"content,omitempty" yaml:"content"`
	// }

	// save this to a stroe
	return nil, ms.Keeper.SetUploadContent(sdk.UnwrapSDKContext(ctx), msg)
}

// UploadMetadata implements types.MsgServer.
func (ms *msgServer) UploadMetadata(ctx context.Context, msg *types.MsgUploadMetadata) (*types.MsgUploadMetadataResponse, error) {
	return nil, ms.Keeper.SetUploadMetadata(sdk.UnwrapSDKContext(ctx), msg)
}

// helpers

func (k Keeper) SetUploadContent(ctx sdk.Context, msg *types.MsgUploadContentBlob) error {
	store := ctx.KVStore(k.storeKey)
	// bz := k.cdc.MustMarshal(msg)

	// TODO: add the title of the checksum here to allow for per title uploads for ids

	key := GetKey(msg.Sender, msg.IdKey)
	store.Set(key, msg.Content)

	return nil
}

func (k Keeper) SetUploadMetadata(ctx sdk.Context, msg *types.MsgUploadMetadata) error {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(msg)

	// this assumes the user already knows the title of the video
	// indexing wise it would be better to use a sha256 checksum, but still need to store titles somewhere
	// per unique content uploader
	key := GetMetadataKey(msg.Sender, msg.Title)
	store.Set(key, bz)

	return nil
}

// GetParams returns the current x/youtube module parameters.
func (k Keeper) GetContent(ctx sdk.Context, creator sdk.Address, id uint64) []byte {
	store := ctx.KVStore(k.storeKey)

	key := GetKey(creator.String(), id)

	return store.Get(key)
}

func (k Keeper) GetMetadata(ctx sdk.Context, creator sdk.Address, title string) *types.MsgUploadMetadata {
	store := ctx.KVStore(k.storeKey)

	key := GetMetadataKey(creator.String(), title)

	bz := store.Get(key)
	var metadata types.MsgUploadMetadata
	k.cdc.MustUnmarshal(bz, &metadata)

	return &metadata
}

func GetKey(sender string, id uint64) []byte {
	return []byte(sender + fmt.Sprintf("%d", id))
}

// GetMetadataKey may be better suited to not depend on the title?
// at a minimum I can use this as a unique videoId based off the sha256 checksum
func GetMetadataKey(sender string, title string) []byte {
	base64Title := b64.StdEncoding.EncodeToString([]byte(title))
	return []byte(sender + "|" + base64Title)
}
