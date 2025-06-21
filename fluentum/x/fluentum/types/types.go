package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
)

// MsgCreateFluentum defines the CreateFluentum message
type MsgCreateFluentum struct {
	Creator string `protobuf:"bytes,1,opt,name=creator,proto3" json:"creator,omitempty"`
	Index   string `protobuf:"bytes,2,opt,name=index,proto3" json:"index,omitempty"`
	Title   string `protobuf:"bytes,3,opt,name=title,proto3" json:"title,omitempty"`
	Body    string `protobuf:"bytes,4,opt,name=body,proto3" json:"body,omitempty"`
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
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
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
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

// MsgDeleteFluentum defines the DeleteFluentum message
type MsgDeleteFluentum struct {
	Creator string `protobuf:"bytes,1,opt,name=creator,proto3" json:"creator,omitempty"`
	Index   string `protobuf:"bytes,2,opt,name=index,proto3" json:"index,omitempty"`
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
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
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

// NewFluentum creates a new Fluentum instance
func NewFluentum(creator string, index string, title string, body string) *Fluentum {
	return &Fluentum{
		Creator: creator,
		Index:   index,
		Title:   title,
		Body:    body,
	}
}
