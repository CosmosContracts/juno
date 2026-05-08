package types

import (
	"encoding/json"
	"fmt"

	collcodec "cosmossdk.io/collections/codec"
)

// paramsValueCodec implements collcodec.ValueCodec[Params] using JSON.
// Params is a small, mostly-immutable governance config — JSON is plenty.
// (For larger / hot-path types we'd back this with proto + codec.CollValue.)
type paramsValueCodec struct{}

func ParamsValueCodec() collcodec.ValueCodec[Params] {
	return paramsValueCodec{}
}

func (paramsValueCodec) Encode(p Params) ([]byte, error)     { return json.Marshal(p) }
func (paramsValueCodec) EncodeJSON(p Params) ([]byte, error) { return json.Marshal(p) }
func (paramsValueCodec) Stringify(p Params) string           { return fmt.Sprintf("%+v", p) }
func (paramsValueCodec) ValueType() string                   { return "Params" }

func (paramsValueCodec) Decode(b []byte) (Params, error) {
	var p Params
	if err := json.Unmarshal(b, &p); err != nil {
		return Params{}, err
	}
	return p, nil
}

func (paramsValueCodec) DecodeJSON(b []byte) (Params, error) { return paramsValueCodec{}.Decode(b) }
