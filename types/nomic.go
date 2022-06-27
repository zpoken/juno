package types

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
