package postgresql

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/tendermint/tendermint/crypto"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	dexv1alpha1 "go.buf.build/grpc/go/penumbra-zone/penumbra/penumbra/core/dex/v1alpha1"
	stakev1alpha1 "go.buf.build/grpc/go/penumbra-zone/penumbra/penumbra/core/stake/v1alpha1"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/zpoken/juno/v3/logging"

	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/lib/pq"

	"github.com/zpoken/juno/v3/database"
	"github.com/zpoken/juno/v3/types"
	"github.com/zpoken/juno/v3/types/config"
)

// Builder creates a database connection with the given database connection info
// from config. It returns a database connection handle or an error if the
// connection fails.
func Builder(ctx *database.Context) (database.Database, error) {
	sslMode := "disable"
	if ctx.Cfg.SSLMode != "" {
		sslMode = ctx.Cfg.SSLMode
	}

	schema := "public"
	if ctx.Cfg.Schema != "" {
		schema = ctx.Cfg.Schema
	}

	connStr := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s sslmode=%s search_path=%s",
		ctx.Cfg.Host, ctx.Cfg.Port, ctx.Cfg.Name, ctx.Cfg.User, sslMode, schema,
	)

	if ctx.Cfg.Password != "" {
		connStr += fmt.Sprintf(" password=%s", ctx.Cfg.Password)
	}

	postgresDb, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Set max open connections
	postgresDb.SetMaxOpenConns(ctx.Cfg.MaxOpenConnections)
	postgresDb.SetMaxIdleConns(ctx.Cfg.MaxIdleConnections)

	return &Database{
		SQL:            postgresDb,
		EncodingConfig: ctx.EncodingConfig,
		Logger:         ctx.Logger,
	}, nil
}

// type check to ensure interface is properly implemented
var _ database.Database = &Database{}

// Database defines a wrapper around a SQL database and implements functionality
// for data aggregation and exporting.
type Database struct {
	SQL            *sqlx.DB
	EncodingConfig *params.EncodingConfig
	Logger         logging.Logger
}

// createPartitionIfNotExists creates a new partition having the given partition id if not existing
func (db *Database) createPartitionIfNotExists(table string, partitionID int64) error {
	partitionTable := fmt.Sprintf("%s_%d", table, partitionID)

	stmt := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s PARTITION OF %s FOR VALUES IN (%d)",
		partitionTable,
		table,
		partitionID,
	)
	_, err := db.SQL.Exec(stmt)

	if err != nil {
		return err
	}

	return nil
}

// -------------------------------------------------------------------------------------------------------------------

// HasBlock implements database.Database
func (db *Database) HasBlock(height int64) (bool, error) {
	var res bool
	err := db.SQL.QueryRow(`SELECT EXISTS(SELECT 1 FROM block WHERE height = $1);`, height).Scan(&res)
	return res, err
}

// GetLastBlockHeight returns the last block height stored inside the database
func (db *Database) GetLastBlockHeight() (int64, error) {
	stmt := `SELECT height FROM block ORDER BY height DESC LIMIT 1;`

	var height int64
	err := db.SQL.QueryRow(stmt).Scan(&height)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			// If no rows stored in block table, return 0 as height
			return 0, nil
		}
		return 0, fmt.Errorf("error while getting last block height, error: %s", err)
	}

	return height, nil
}

// SaveBlock implements database.Database
func (db *Database) SaveBlock(block *types.Block) error {
	sqlStatement := `
INSERT INTO block (height, hash, num_txs, total_gas, proposer_address, timestamp)
VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING`

	proposerAddress := sql.NullString{Valid: len(block.ProposerAddress) != 0, String: block.ProposerAddress}
	_, err := db.SQL.Exec(sqlStatement,
		block.Height, block.Hash, block.TxNum, block.TotalGas, proposerAddress, block.Timestamp,
	)
	return err
}

// GetTotalBlocks implements database.Database
func (db *Database) GetTotalBlocks() int64 {
	var blockCount int64
	err := db.SQL.QueryRow(`SELECT count(*) FROM block;`).Scan(&blockCount)
	if err != nil {
		return 0
	}

	return blockCount
}

// SaveTx implements database.Database
func (db *Database) SaveTx(transaction *coretypes.ResultTx) error {

	sqlStatement := `
INSERT INTO transaction 
(hash, height, success, messages, memo, signatures, signer_infos, fee, gas_wanted, gas_used, raw_log, logs, partition_id) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) 
ON CONFLICT (hash, partition_id) DO UPDATE 
	SET height = excluded.height, 
		success = excluded.success, 
		messages = excluded.messages,
		memo = excluded.memo, 
		signatures = excluded.signatures, 
		signer_infos = excluded.signer_infos,
		fee = excluded.fee, 
		gas_wanted = excluded.gas_wanted, 
		gas_used = excluded.gas_used,
		raw_log = excluded.raw_log, 
		logs = excluded.logs`

	var msgs = make([]string, len(transaction.TxResult.Events))
	for index, msg := range transaction.TxResult.Events {
		bz, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		msgs[index] = string(bz)
	}
	msgsBz := fmt.Sprintf("[%s]", strings.Join(msgs, ","))

	sigInfoBz := "[]"

	logsBz, err := db.EncodingConfig.Amino.MarshalJSON(transaction.Tx)
	if err != nil {
		return err
	}

	var empty []string

	empty = append(empty, "sig")

	fee := "{}"

	atoi, err := strconv.Atoi(strconv.FormatInt(transaction.Height, 10))
	_, err = db.SQL.Exec(sqlStatement,
		transaction.Hash.String(), atoi, true,
		msgsBz, nil, pq.Array(empty),
		sigInfoBz, fee,
		transaction.TxResult.GasWanted,
		transaction.TxResult.GasUsed,
		transaction.Tx.String(),
		string(logsBz),
		1,
	)
	return err
}

// saveTxInsidePartition stores the given transaction inside the partition having the given id
func (db *Database) saveTxInsidePartition(tx *types.Tx, partitionID int64) error {
	sqlStatement := `
INSERT INTO transaction 
(hash, height, success, messages, memo, signatures, signer_infos, fee, gas_wanted, gas_used, raw_log, logs, partition_id) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) 
ON CONFLICT (hash, partition_id) DO UPDATE 
	SET height = excluded.height, 
		success = excluded.success, 
		messages = excluded.messages,
		memo = excluded.memo, 
		signatures = excluded.signatures, 
		signer_infos = excluded.signer_infos,
		fee = excluded.fee, 
		gas_wanted = excluded.gas_wanted, 
		gas_used = excluded.gas_used,
		raw_log = excluded.raw_log, 
		logs = excluded.logs`

	var sigs = make([]string, len(tx.Signatures))
	for index, sig := range tx.Signatures {
		sigs[index] = base64.StdEncoding.EncodeToString(sig)
	}

	var msgs = make([]string, len(tx.Body.Messages))
	for index, msg := range tx.Body.Messages {
		bz, err := db.EncodingConfig.Marshaler.MarshalJSON(msg)
		if err != nil {
			return err
		}
		msgs[index] = string(bz)
	}
	msgsBz := fmt.Sprintf("[%s]", strings.Join(msgs, ","))

	feeBz, err := db.EncodingConfig.Marshaler.MarshalJSON(tx.AuthInfo.Fee)
	if err != nil {
		return fmt.Errorf("failed to JSON encode tx fee: %s", err)
	}

	var sigInfos = make([]string, len(tx.AuthInfo.SignerInfos))
	for index, info := range tx.AuthInfo.SignerInfos {
		bz, err := db.EncodingConfig.Marshaler.MarshalJSON(info)
		if err != nil {
			return err
		}
		sigInfos[index] = string(bz)
	}
	sigInfoBz := fmt.Sprintf("[%s]", strings.Join(sigInfos, ","))

	logsBz, err := db.EncodingConfig.Amino.MarshalJSON(tx.Logs)
	if err != nil {
		return err
	}

	_, err = db.SQL.Exec(sqlStatement,
		tx.TxHash, tx.Height, tx.Successful(),
		msgsBz, tx.Body.Memo, pq.Array(sigs),
		sigInfoBz, string(feeBz),
		tx.GasWanted, tx.GasUsed, tx.RawLog, string(logsBz),
		partitionID,
	)
	return err
}

func (db *Database) SaveValidatorInfo(validator *stakev1alpha1.ValidatorInfo, height int64) error {

	bech32Prefix := sdk.GetConfig().GetBech32ConsensusPubPrefix()
	consPubKey, _ := bech32.ConvertAndEncode(bech32Prefix, validator.GetValidator().GetConsensusKey())

	validatorAddress := crypto.AddressHash(validator.GetValidator().GetConsensusKey())
	consAddr := sdk.ConsAddress(validatorAddress).String()

	indetityKey, _ := bech32.ConvertAndEncode("penumbravalid", validator.GetValidator().GetIdentityKey().GetIk())

	selfDelegationAccQuery := `
INSERT INTO account (address) VALUES ($1) ON CONFLICT DO NOTHING`

	_, _ = db.SQL.Exec(selfDelegationAccQuery, indetityKey)

	validatorQuery := `
INSERT INTO validator (consensus_address, consensus_pubkey) VALUES ($1, $2) ON CONFLICT DO NOTHING `

	_, _ = db.SQL.Exec(validatorQuery, consAddr, consPubKey)

	validatorInfoQuery := `
INSERT INTO validator_info (consensus_address, operator_address, self_delegate_address, max_change_rate, max_rate, height) 
	VALUES($1, $2, $3, $4, $5, $6)
	ON CONFLICT (consensus_address) DO UPDATE 
	SET consensus_address = excluded.consensus_address,
		operator_address = excluded.operator_address,
		max_change_rate = excluded.max_change_rate,
		max_rate = excluded.max_rate,
		height = excluded.height
WHERE validator_info.height <= excluded.height`

	_, err := db.SQL.Exec(validatorInfoQuery,
		ToNullString(consAddr),
		ToNullString(consAddr),
		ToNullString(indetityKey),
		validator.GetRateData().ValidatorRewardRate,
		validator.GetRateData().ValidatorRewardRate,
		height)

	// Insert the description
	stmt := `
INSERT INTO validator_description (
	validator_address, moniker, identity,  website, details, height
)
VALUES($1, $2, $3, $4, $5, $6)
ON CONFLICT (validator_address) DO UPDATE
    SET moniker = excluded.moniker, 
        identity = excluded.identity, 
        avatar_url = excluded.avatar_url,
        website = excluded.website, 
        security_contact = excluded.security_contact, 
        details = excluded.details,
        height = excluded.height
WHERE validator_description.height <= excluded.height`

	_, err = db.SQL.Exec(stmt,
		ToNullString(consAddr),
		ToNullString(validator.GetValidator().GetName()),
		ToNullString(validator.GetValidator().GetName()),
		ToNullString(validator.GetValidator().GetWebsite()),
		ToNullString(validator.GetValidator().GetDescription()),
		height,
	)
	if err != nil {
		//return fmt.Errorf("error while storing validator description: %s", err)
	}

	var commissiomBps = uint32(0)

	for _, stream := range validator.GetValidator().GetFundingStreams() {
		commissiomBps = commissiomBps + stream.RateBps
	}
	stmt = `
INSERT INTO validator_commission (validator_address, commission, height) 
VALUES ($1, $2, $3)
ON CONFLICT (validator_address) DO UPDATE 
    SET commission = excluded.commission, 
        min_self_delegation = excluded.min_self_delegation,
        height = excluded.height
WHERE validator_commission.height <= excluded.height`
	_, err = db.SQL.Exec(stmt, consAddr, float64(commissiomBps)/10000, height)
	if err != nil {
		//return fmt.Errorf("error while storing validator commission: %s", err)
	}

	stmt = `INSERT INTO validator_voting_power (validator_address, voting_power, height)VALUES ($1, $2, $3) ON CONFLICT (validator_address) DO UPDATE 
	SET voting_power = excluded.voting_power, 
		height = excluded.height
WHERE validator_voting_power.height <= excluded.height `

	_, err = db.SQL.Exec(stmt, consAddr, validator.GetStatus().VotingPower, height)

	statusStmt := `INSERT INTO validator_status (validator_address, status, jailed, tombstoned, height) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (validator_address) DO UPDATE 
	SET status = excluded.status,
	    jailed = excluded.jailed,
	    tombstoned = excluded.tombstoned,
	    height = excluded.height
WHERE validator_status.height <= excluded.height `

	if validator.GetStatus().GetState().State == 2 {
		_, err = db.SQL.Exec(statusStmt, consAddr, 3, validator.GetStatus().GetState().State == 3, validator.GetStatus().GetState().State == 4, height)
	} else {
		_, err = db.SQL.Exec(statusStmt, consAddr, validator.GetStatus().GetState().State, validator.GetStatus().GetState().State == 3, validator.GetStatus().GetState().State == 4, height)

	}

	if err != nil {

		return fmt.Errorf("error while storing validator infos: %s", err)
	}

	return nil
}

func ToNullString(value string) sql.NullString {
	value = strings.TrimSpace(value)
	return sql.NullString{
		Valid:  value != "",
		String: value,
	}
}

func (db *Database) SaveSwapOutputData(swapOutputData *dexv1alpha1.BatchSwapOutputData,
	height int64,
	tradingPairId int64) error {

	swapQuery := `
INSERT INTO swap_output_data (height, trading_pair_id, delta_1, delta_2, lambda_1, lambda_2, success) 
	VALUES($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE
		    SET
	 height = excluded.height,
		trading_pair_id = excluded.trading_pair_id,
		delta_1 = excluded.delta_1,
		delta_2 = excluded.delta_2,
		lambda_1 = excluded.lambda_1,
	    lambda_2 = excluded.lambda_2,
	    success = excluded.success`

	_, err := db.SQL.Exec(swapQuery,
		height,
		tradingPairId,
		swapOutputData.Delta_1,
		swapOutputData.Delta_2,
		swapOutputData.Lambda_1,
		swapOutputData.Lambda_2,
		swapOutputData.Success)

	return err
}

func (db *Database) SaveCPMMReserves(reserves *dexv1alpha1.Reserves, height int64, tradingPairId int64) error {

	r1 := int64(reserves.R1.Hi)

	r2 := int64(reserves.R2.Hi)

	db.Logger.Info("r1", "r1", r1)
	db.Logger.Info("r2", "r2", r1)

	swapQuery := `
INSERT INTO cpmm_reserve (height, trading_pair_id, r1, r2) 
	VALUES($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE
		    SET
	 height = excluded.height,
		trading_pair_id = excluded.trading_pair_id,
		r1 = excluded.r1,
		    r2 = excluded.r2`

	_, err := db.SQL.Exec(swapQuery,
		height,
		tradingPairId,
		r1,
		r2)

	return err
}

// HasValidator implements database.Database
func (db *Database) HasValidator(addr string) (bool, error) {
	var res bool
	stmt := `SELECT EXISTS(SELECT 1 FROM validator WHERE consensus_address = $1);`
	err := db.SQL.QueryRow(stmt, addr).Scan(&res)
	return res, err
}

// SaveValidators implements database.Database
func (db *Database) SaveValidators(validators []*types.Validator) error {
	if len(validators) == 0 {
		return nil
	}

	stmt := `INSERT INTO validator (consensus_address, consensus_pubkey) VALUES `

	var vparams []interface{}
	for i, val := range validators {
		vi := i * 2

		stmt += fmt.Sprintf("($%d, $%d),", vi+1, vi+2)
		vparams = append(vparams, val.ConsAddr, val.ConsPubKey)
	}

	stmt = stmt[:len(stmt)-1] // Remove trailing ,
	stmt += " ON CONFLICT DO NOTHING"
	_, err := db.SQL.Exec(stmt, vparams...)
	return err
}

// SaveCommitSignatures implements database.Database
func (db *Database) SaveCommitSignatures(signatures []*types.CommitSig) error {
	if len(signatures) == 0 {
		return nil
	}

	stmt := `INSERT INTO pre_commit (validator_address, height, timestamp, voting_power, proposer_priority) VALUES `

	var sparams []interface{}
	for i, sig := range signatures {
		si := i * 5

		stmt += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d),", si+1, si+2, si+3, si+4, si+5)
		sparams = append(sparams, sig.ValidatorAddress, sig.Height, sig.Timestamp, sig.VotingPower, sig.ProposerPriority)
	}

	stmt = stmt[:len(stmt)-1]
	stmt += " ON CONFLICT (validator_address, timestamp) DO NOTHING"
	_, err := db.SQL.Exec(stmt, sparams...)
	return err
}

// SaveMessage implements database.Database
func (db *Database) SaveMessage(msg *types.Message) error {
	var partitionID int64
	partitionSize := config.Cfg.Database.PartitionSize
	if partitionSize > 0 {
		partitionID = msg.Height / partitionSize
		err := db.createPartitionIfNotExists("message", partitionID)
		if err != nil {
			return err
		}
	}

	return db.saveMessageInsidePartition(msg, partitionID)
}

// saveMessageInsidePartition stores the given message inside the partition having the provided id
func (db *Database) saveMessageInsidePartition(msg *types.Message, partitionID int64) error {
	stmt := `
INSERT INTO message(transaction_hash, index, type, value, involved_accounts_addresses, height, partition_id) 
VALUES ($1, $2, $3, $4, $5, $6, $7) 
ON CONFLICT (transaction_hash, index, partition_id) DO UPDATE 
	SET height = excluded.height, 
		type = excluded.type,
		value = excluded.value,
		involved_accounts_addresses = excluded.involved_accounts_addresses`

	_, err := db.SQL.Exec(stmt, msg.TxHash, msg.Index, msg.Type, msg.Value, pq.Array(msg.Addresses), msg.Height, partitionID)
	return err
}

// Close implements database.Database
func (db *Database) Close() {
	err := db.SQL.Close()
	if err != nil {
		db.Logger.Error("error while closing connection", "err", err)
	}
}

// -------------------------------------------------------------------------------------------------------------------

// GetLastPruned implements database.PruningDb
func (db *Database) GetLastPruned() (int64, error) {
	var lastPrunedHeight int64
	err := db.SQL.QueryRow(`SELECT coalesce(MAX(last_pruned_height),0) FROM pruning LIMIT 1;`).Scan(&lastPrunedHeight)
	return lastPrunedHeight, err
}

// StoreLastPruned implements database.PruningDb
func (db *Database) StoreLastPruned(height int64) error {
	_, err := db.SQL.Exec(`DELETE FROM pruning`)
	if err != nil {
		return err
	}

	_, err = db.SQL.Exec(`INSERT INTO pruning (last_pruned_height) VALUES ($1)`, height)
	return err
}

// Prune implements database.PruningDb
func (db *Database) Prune(height int64) error {
	_, err := db.SQL.Exec(`DELETE FROM pre_commit WHERE height = $1`, height)
	if err != nil {
		return err
	}

	_, err = db.SQL.Exec(`
DELETE FROM message 
USING transaction 
WHERE message.transaction_hash = transaction.hash AND transaction.height = $1
`, height)
	return err
}
