package blockchain

type TxOutput struct {
	Value     int
	PublicKey string
}

func (out *TxOutput) CanBeUnlocked(address string) bool {
	return out.PublicKey == address
}

type TxInput struct {
	ID  []byte
	Out int
	Sig string
}

func (in *TxInput) CanUnlock(address string) bool {
	return in.Sig == address
}
