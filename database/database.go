package database

import (
	"github.com/cosmos/cosmos-sdk/simapp/params"

	"github.com/forbole/juno/v3/logging"

	databaseconfig "github.com/forbole/juno/v3/database/config"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/juno/v3/types"
)

// Database represents an abstract database that can be used to save data inside it
type Database interface {
	// HasBlock tells whether or not the database has already stored the block having the given height.
	// An error is returned if the operation fails.
	HasBlock(height int64) (bool, error)

	// SaveBlock will be called when a new block is parsed, passing the block itself
	// and the transactions contained inside that block.
	// An error is returned if the operation fails.
	// NOTE. For each transaction inside txs, SaveTx will be called as well.
	SaveBlock(block *types.Block) error

	// SaveTx will be called to save each transaction contained inside a block.
	// An error is returned if the operation fails.
	SaveTx(tx types.TxResponseTest) error

	// HasValidator returns true if a given validator by consensus address exists.
	// An error is returned if the operation fails.
	HasValidator(address string) (bool, error)

	// SaveValidators stores a list of validators if they do not already exist.
	// An error is returned if the operation fails.
	SaveValidators(validators []*types.Validator) error

	// SaveCommitSignatures stores a  slice of validator commit signatures.
	// An error is returned if the operation fails.
	SaveCommitSignatures(signatures []*types.CommitSig) error

	// SaveMessage stores a single message.
	// An error is returned if the operation fails.
	SaveMessage(msg *types.Message) error

	// SaveSupply stores a total supply value.
	// An error is returned if the operation fails.
	SaveSupply(coins sdk.Coins, height int64) error

	// SaveInflation stores the inflation value.
	// An error is returned if the operation fails.
	SaveInflation(inflation string, height int64) error

	// SaveStakingPool stores the staking pool value.
	// An error is returned if the operation fails.
	SaveStakingPool(pool *types.StakingPool) error

	// GetLastBlockHeight returns the latest block height stored in database.
	// An error is returned if the operation fails.
	GetLastBlockHeight() (int64, error)

	// SaveIBCParams stores the ibc tx params value.
	// An error is returned if the operation fails.
	SaveIBCParams(params *types.IBCParams) error

	// SaveAccountBalance stores the account balance value.
	// An error is returned if the operation fails.
	SaveAccountBalance(balances types.AccountBalance) error

	// SaveToken stores the token details.
	// An error is returned if the operation fails.
	SaveToken(token types.Token) error

	// Close closes the connection to the database
	Close()
}

// PruningDb represents a database that supports pruning properly
type PruningDb interface {
	// Prune prunes the data for the given height, returning any error
	Prune(height int64) error

	// StoreLastPruned saves the last height at which the database was pruned
	StoreLastPruned(height int64) error

	// GetLastPruned returns the last height at which the database was pruned
	GetLastPruned() (int64, error)
}

// Context contains the data that might be used to build a Database instance
type Context struct {
	Cfg            databaseconfig.Config
	EncodingConfig *params.EncodingConfig
	Logger         logging.Logger
}

// NewContext allows to build a new Context instance
func NewContext(cfg databaseconfig.Config, encodingConfig *params.EncodingConfig, logger logging.Logger) *Context {
	return &Context{
		Cfg:            cfg,
		EncodingConfig: encodingConfig,
		Logger:         logger,
	}
}

// Builder represents a method that allows to build any database from a given codec and configuration
type Builder func(ctx *Context) (Database, error)
