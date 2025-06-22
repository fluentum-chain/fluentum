package types

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/gogo/protobuf/proto"
)

const (
	// ModuleName defines the module name
	ModuleName = "fluentum"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_fluentum"
)

var (
	_ sdk.Msg = &MsgCreateFluentum{}
	_ sdk.Msg = &MsgUpdateFluentum{}
	_ sdk.Msg = &MsgDeleteFluentum{}

	// ModuleCdc defines the module codec
	ModuleCdc = codec.NewProtoCodec(types.NewInterfaceRegistry())
)

// MsgCreateFluentum defines the CreateFluentum message
type MsgCreateFluentum struct {
	Creator string `protobuf:"bytes,1,opt,name=creator,proto3" json:"creator,omitempty"`
	Index   string `protobuf:"bytes,2,opt,name=index,proto3" json:"index,omitempty"`
	Title   string `protobuf:"bytes,3,opt,name=title,proto3" json:"title,omitempty"`
	Body    string `protobuf:"bytes,4,opt,name=body,proto3" json:"body,omitempty"`
}

// ProtoMessage implements proto.Message
func (msg *MsgCreateFluentum) ProtoMessage() {}

// Reset implements proto.Message
func (msg *MsgCreateFluentum) Reset() {
	*msg = MsgCreateFluentum{}
}

// String implements proto.Message
func (msg *MsgCreateFluentum) String() string {
	return proto.CompactTextString(msg)
}

// NewMsgCreateFluentum creates a new MsgCreateFluentum instance
func NewMsgCreateFluentum(creator string, index string, title string, body string) *MsgCreateFluentum {
	return &MsgCreateFluentum{
		Creator: creator,
		Index:   index,
		Title:   title,
		Body:    body,
	}
}

// Route returns the message route
func (msg *MsgCreateFluentum) Route() string {
	return RouterKey
}

// Type returns the message type
func (msg *MsgCreateFluentum) Type() string {
	return "CreateFluentum"
}

// GetSigners returns the message signers
func (msg *MsgCreateFluentum) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message sign bytes
func (msg *MsgCreateFluentum) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validates the message
func (msg *MsgCreateFluentum) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return fmt.Errorf("invalid creator address (%s)", err)
	}
	return nil
}

// MsgUpdateFluentum defines the UpdateFluentum message
type MsgUpdateFluentum struct {
	Creator string `protobuf:"bytes,1,opt,name=creator,proto3" json:"creator,omitempty"`
	Index   string `protobuf:"bytes,2,opt,name=index,proto3" json:"index,omitempty"`
	Title   string `protobuf:"bytes,3,opt,name=title,proto3" json:"title,omitempty"`
	Body    string `protobuf:"bytes,4,opt,name=body,proto3" json:"body,omitempty"`
}

// ProtoMessage implements proto.Message
func (msg *MsgUpdateFluentum) ProtoMessage() {}

// Reset implements proto.Message
func (msg *MsgUpdateFluentum) Reset() {
	*msg = MsgUpdateFluentum{}
}

// String implements proto.Message
func (msg *MsgUpdateFluentum) String() string {
	return proto.CompactTextString(msg)
}

// NewMsgUpdateFluentum creates a new MsgUpdateFluentum instance
func NewMsgUpdateFluentum(creator string, index string, title string, body string) *MsgUpdateFluentum {
	return &MsgUpdateFluentum{
		Creator: creator,
		Index:   index,
		Title:   title,
		Body:    body,
	}
}

// Route returns the message route
func (msg *MsgUpdateFluentum) Route() string {
	return RouterKey
}

// Type returns the message type
func (msg *MsgUpdateFluentum) Type() string {
	return "UpdateFluentum"
}

// GetSigners returns the message signers
func (msg *MsgUpdateFluentum) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message sign bytes
func (msg *MsgUpdateFluentum) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validates the message
func (msg *MsgUpdateFluentum) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return fmt.Errorf("invalid creator address (%s)", err)
	}
	return nil
}

// MsgDeleteFluentum defines the DeleteFluentum message
type MsgDeleteFluentum struct {
	Creator string `protobuf:"bytes,1,opt,name=creator,proto3" json:"creator,omitempty"`
	Index   string `protobuf:"bytes,2,opt,name=index,proto3" json:"index,omitempty"`
}

// ProtoMessage implements proto.Message
func (msg *MsgDeleteFluentum) ProtoMessage() {}

// Reset implements proto.Message
func (msg *MsgDeleteFluentum) Reset() {
	*msg = MsgDeleteFluentum{}
}

// String implements proto.Message
func (msg *MsgDeleteFluentum) String() string {
	return proto.CompactTextString(msg)
}

// NewMsgDeleteFluentum creates a new MsgDeleteFluentum instance
func NewMsgDeleteFluentum(creator string, index string) *MsgDeleteFluentum {
	return &MsgDeleteFluentum{
		Creator: creator,
		Index:   index,
	}
}

// Route returns the message route
func (msg *MsgDeleteFluentum) Route() string {
	return RouterKey
}

// Type returns the message type
func (msg *MsgDeleteFluentum) Type() string {
	return "DeleteFluentum"
}

// GetSigners returns the message signers
func (msg *MsgDeleteFluentum) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message sign bytes
func (msg *MsgDeleteFluentum) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validates the message
func (msg *MsgDeleteFluentum) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return fmt.Errorf("invalid creator address (%s)", err)
	}
	return nil
}

// Fluentum defines the Fluentum struct
type Fluentum struct {
	Creator string `protobuf:"bytes,1,opt,name=creator,proto3" json:"creator,omitempty"`
	Index   string `protobuf:"bytes,2,opt,name=index,proto3" json:"index,omitempty"`
	Title   string `protobuf:"bytes,3,opt,name=title,proto3" json:"title,omitempty"`
	Body    string `protobuf:"bytes,4,opt,name=body,proto3" json:"body,omitempty"`
}

// ProtoMessage implements proto.Message
func (msg *Fluentum) ProtoMessage() {}

// Reset implements proto.Message
func (msg *Fluentum) Reset() {
	*msg = Fluentum{}
}

// String implements proto.Message
func (msg *Fluentum) String() string {
	return proto.CompactTextString(msg)
}

// NewFluentum creates a new Fluentum instance
func NewFluentum(creator string, index string, title string, body string) *Fluentum {
	return &Fluentum{
		Creator: creator,
		Index:   index,
		Title:   title,
		Body:    body,
	}
}

// Params defines the parameters for the module
type Params struct{}

// ParamKeyTable returns the parameter key table
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable()
}

// Stub query types to fix build errors
type QueryParamsRequest struct{}

func (m *QueryParamsRequest) Reset()         { *m = QueryParamsRequest{} }
func (m *QueryParamsRequest) String() string { return "QueryParamsRequest" }
func (m *QueryParamsRequest) ProtoMessage()  {}

type QueryParamsResponse struct{}

func (m *QueryParamsResponse) Reset()         { *m = QueryParamsResponse{} }
func (m *QueryParamsResponse) String() string { return "QueryParamsResponse" }
func (m *QueryParamsResponse) ProtoMessage()  {}

type QueryAllFluentumRequest struct {
	Pagination *PageRequest `protobuf:"bytes,1,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

func (m *QueryAllFluentumRequest) Reset()         { *m = QueryAllFluentumRequest{} }
func (m *QueryAllFluentumRequest) String() string { return "QueryAllFluentumRequest" }
func (m *QueryAllFluentumRequest) ProtoMessage()  {}

type QueryAllFluentumResponse struct {
	Fluentum   []*Fluentum   `protobuf:"bytes,1,rep,name=fluentum,proto3" json:"fluentum,omitempty"`
	Pagination *PageResponse `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

func (m *QueryAllFluentumResponse) Reset()         { *m = QueryAllFluentumResponse{} }
func (m *QueryAllFluentumResponse) String() string { return "QueryAllFluentumResponse" }
func (m *QueryAllFluentumResponse) ProtoMessage()  {}

type QueryGetFluentumRequest struct {
	Index string `protobuf:"bytes,1,opt,name=index,proto3" json:"index,omitempty"`
}

func (m *QueryGetFluentumRequest) Reset()         { *m = QueryGetFluentumRequest{} }
func (m *QueryGetFluentumRequest) String() string { return "QueryGetFluentumRequest" }
func (m *QueryGetFluentumRequest) ProtoMessage()  {}

type QueryGetFluentumResponse struct {
	Fluentum *Fluentum `protobuf:"bytes,1,opt,name=fluentum,proto3" json:"fluentum,omitempty"`
}

func (m *QueryGetFluentumResponse) Reset()         { *m = QueryGetFluentumResponse{} }
func (m *QueryGetFluentumResponse) String() string { return "QueryGetFluentumResponse" }
func (m *QueryGetFluentumResponse) ProtoMessage()  {}

// PageRequest and PageResponse for pagination
type PageRequest struct {
	Key        []byte `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Offset     uint64 `protobuf:"varint,2,opt,name=offset,proto3" json:"offset,omitempty"`
	Limit      uint64 `protobuf:"varint,3,opt,name=limit,proto3" json:"limit,omitempty"`
	CountTotal bool   `protobuf:"varint,4,opt,name=count_total,json=countTotal,proto3" json:"count_total,omitempty"`
	Reverse    bool   `protobuf:"varint,5,opt,name=reverse,proto3" json:"reverse,omitempty"`
}

func (m *PageRequest) Reset()         { *m = PageRequest{} }
func (m *PageRequest) String() string { return "PageRequest" }
func (m *PageRequest) ProtoMessage()  {}

type PageResponse struct {
	NextKey []byte `protobuf:"bytes,1,opt,name=next_key,json=nextKey,proto3" json:"next_key,omitempty"`
	Total   uint64 `protobuf:"varint,2,opt,name=total,proto3" json:"total,omitempty"`
}

func (m *PageResponse) Reset()         { *m = PageResponse{} }
func (m *PageResponse) String() string { return "PageResponse" }
func (m *PageResponse) ProtoMessage()  {}

// Stub query client interface
type QueryClient interface {
	Params(ctx context.Context, in *QueryParamsRequest, opts ...interface{}) (*QueryParamsResponse, error)
	FluentumAll(ctx context.Context, in *QueryAllFluentumRequest, opts ...interface{}) (*QueryAllFluentumResponse, error)
	Fluentum(ctx context.Context, in *QueryGetFluentumRequest, opts ...interface{}) (*QueryGetFluentumResponse, error)
}

// NewQueryClient creates a new query client
func NewQueryClient(cc interface{}) QueryClient {
	return &queryClient{cc: cc}
}

type queryClient struct {
	cc interface{}
}

func (c *queryClient) Params(ctx context.Context, in *QueryParamsRequest, opts ...interface{}) (*QueryParamsResponse, error) {
	return &QueryParamsResponse{}, nil
}

func (c *queryClient) FluentumAll(ctx context.Context, in *QueryAllFluentumRequest, opts ...interface{}) (*QueryAllFluentumResponse, error) {
	return &QueryAllFluentumResponse{
		Fluentum:   []*Fluentum{},
		Pagination: &PageResponse{},
	}, nil
}

func (c *queryClient) Fluentum(ctx context.Context, in *QueryGetFluentumRequest, opts ...interface{}) (*QueryGetFluentumResponse, error) {
	return &QueryGetFluentumResponse{
		Fluentum: &Fluentum{},
	}, nil
}
