package types

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/juno/v3/types"
)

// BlockRow represents a single block row stored inside the database
type BlockRow struct {
	Height          int64          `db:"height"`
	Hash            string         `db:"hash"`
	TxNum           int64          `db:"num_txs"`
	TotalGas        int64          `db:"total_gas"`
	ProposerAddress sql.NullString `db:"proposer_address"`
	PreCommitsNum   int64          `db:"pre_commits"`
	Timestamp       time.Time      `db:"timestamp"`
}

// DbCoin represents the information stored inside the database about a single coin
type DbCoin struct {
	Denom  string
	Amount string
}

// NewDbCoin builds a DbCoin starting from an SDK Coin
func NewDbCoin(coin sdk.Coin) DbCoin {
	return DbCoin{
		Denom:  coin.Denom,
		Amount: coin.Amount.String(),
	}
}

// Value implements driver.Value
func (coin *DbCoin) Value() (driver.Value, error) {
	return fmt.Sprintf("(%s,%s)", coin.Denom, coin.Amount), nil
}

// _________________________________________________________

// DbCoins represents an array of coins
type DbCoins []*DbCoin

// NewDbCoins build a new DbCoins object starting from an array of coins
func NewDbCoins(coins sdk.Coins) DbCoins {
	dbCoins := make([]*DbCoin, 0)
	for _, coin := range coins {
		dbCoins = append(dbCoins, &DbCoin{Amount: coin.Amount.String(), Denom: coin.Denom})
	}
	return dbCoins
}

// _________________________________________________________

// NewDBSignatures returns signatures in string array
func NewDBSignatures(signaturesList []types.BDSignatures) []string {
	var signatures []string
	for _, index := range signaturesList {
		signatures = append(signatures, index.Signature)
	}
	return signatures
}

// _________________________________________________________

type GenesisRow struct {
	OneRowID      bool      `db:"one_row_id"`
	ChainID       string    `db:"chain_id"`
	Time          time.Time `db:"time"`
	InitialHeight int64     `db:"initial_height"`
}

func NewGenesisRow(chainID string, time time.Time, initialHeight int64) GenesisRow {
	return GenesisRow{
		OneRowID:      true,
		ChainID:       chainID,
		Time:          time,
		InitialHeight: initialHeight,
	}
}

// _________________________________________________________

func ToNullString(value string) sql.NullString {
	value = strings.TrimSpace(value)
	return sql.NullString{
		Valid:  value != "",
		String: value,
	}
}
