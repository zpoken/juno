package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/types"
)

type TxResponseTest struct {
	Fee        BDCoin         `protobuf:"bytes,1,opt,name=fee,proto3" json:"fee,omitempty"`
	Memo       string         `protobuf:"bytes,2,opt,name=memo,proto3" json:"memo,omitempty"`
	Msg        []BDMsg        `protobuf:"bytes,3,opt,name=msg,proto3" json:"msg,omitempty"`
	Signatures []BDSignatures `protobuf:"bytes,4,opt,name=signatures,proto3" json:"signatures,omitempty"`
}

type FeeCoin struct {
	Denom  string `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty"`
	Amount string `protobuf:"bytes,2,opt,name=amount,proto3,customtype=Dec" json:"amount"`
}

type BDCoin struct {
	Amount []FeeCoin `protobuf:"bytes,1,opt,name=amount,proto3" json:"amount,omitempty"`
	Gas    string    `protobuf:"bytes,2,opt,name=gas,proto3,customtype=Dec" json:"gas"`
}

type BDSignatures struct {
	PubKey    BDPubKey `protobuf:"bytes,1,opt,name=pub_key,proto3" json:"pub_key,omitempty"`
	Signature string   `protobuf:"bytes,2,opt,name=signature,proto3,customtype=Dec" json:"signature"`
}

type BDPubKey struct {
	Type  string `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value,proto3,customtype=Dec" json:"value"`
}

type BDMsg struct {
	Type  string     `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Value BDMsgValue `protobuf:"bytes,2,opt,name=value,proto3,customtype=Dec" json:"value"`
}

type BDMsgValue struct {
	Amount           FeeCoin `protobuf:"bytes,1,opt,name=amount,proto3" json:"amount,omitempty"`
	DelegatorAddress string  `protobuf:"bytes,2,opt,name=delegator_address,proto3,customtype=Dec" json:"delegator_address"`
	ValidatorAddress string  `protobuf:"bytes,3,opt,name=validator_address,proto3,customtype=Dec" json:"validator_address"`
}

func NewTxResponseTest(
	fee BDCoin, memo string, msg []BDMsg, sig []BDSignatures,
) TxResponseTest {
	return TxResponseTest{
		Fee:        fee,
		Memo:       memo,
		Msg:        msg,
		Signatures: sig,
	}
}

type TxResponse struct {
	// The block height
	Height string `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
	// The transaction hash.
	Hash string `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
	// Namespace for the Code
	Codespace string `protobuf:"bytes,3,opt,name=codespace,proto3" json:"codespace,omitempty"`
	// Response code.
	Code uint32 `protobuf:"varint,4,opt,name=code,proto3" json:"code,omitempty"`
	// Result bytes, if any.
	Data string `protobuf:"bytes,5,opt,name=data,proto3" json:"data,omitempty"`
	// The output of the application's logger (raw string). May be
	// non-deterministic.
	RawLog string `protobuf:"bytes,6,opt,name=raw_log,json=rawLog,proto3" json:"raw_log,omitempty"`
	// The output of the application's logger (typed). May be non-deterministic.
	Logs sdk.ABCIMessageLogs `protobuf:"bytes,7,rep,name=logs,proto3,castrepeated=ABCIMessageLogs" json:"logs"`
	// Additional information. May be non-deterministic.
	Info string `protobuf:"bytes,8,opt,name=info,proto3" json:"info,omitempty"`
	// Amount of gas requested for transaction.
	GasWanted int64 `protobuf:"varint,9,opt,name=gas_wanted,json=gasWanted,proto3" json:"gas_wanted,omitempty"`
	// Amount of gas consumed by transaction.
	GasUsed int64 `protobuf:"varint,10,opt,name=gas_used,json=gasUsed,proto3" json:"gas_used,omitempty"`
	// The request transaction bytes.
	Tx *types.Any `protobuf:"bytes,11,opt,name=tx,proto3" json:"tx,omitempty"`
	// Time of the previous block. For heights > 1, it's the weighted median of
	// the timestamps of the valid votes in the block.LastCommit. For height == 1,
	// it's genesis time.
	Timestamp string `protobuf:"bytes,12,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	JsonRPC   string `protobuf:"bytes,13,opt,name=jsonrpc,proto3" json:"jsonrpc,omitempty"`
}

type InflationResponse struct {
	Inflation string `json:"inflation" yaml:"inflation"`
}

// StakingPool contains the data of the staking pool at the given height
type StakingPool struct {
	BondedTokens    sdk.Int
	NotBondedTokens sdk.Int
	Height          int64
}

// NewStakingPool allows to build a new StakingPool instance
func NewStakingPool(bondedTokens sdk.Int, notBondedTokens sdk.Int, height int64) *StakingPool {
	return &StakingPool{
		BondedTokens:    bondedTokens,
		NotBondedTokens: notBondedTokens,
		Height:          height,
	}
}
