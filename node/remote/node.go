package remote

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/rs/zerolog/log"
	cryptov1alpha1 "go.buf.build/grpc/go/penumbra-zone/penumbra/penumbra/core/crypto/v1alpha1"
	dexv1alpha1 "go.buf.build/grpc/go/penumbra-zone/penumbra/penumbra/core/dex/v1alpha1"
	stakev1alpha1 "go.buf.build/grpc/go/penumbra-zone/penumbra/penumbra/core/stake/v1alpha1"
	"io"
	"net/http"
	"strings"
	"time"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"google.golang.org/grpc"

	constypes "github.com/tendermint/tendermint/consensus/types"
	tmjson "github.com/tendermint/tendermint/libs/json"

	"github.com/zpoken/juno/v3/node"

	httpclient "github.com/tendermint/tendermint/rpc/client/http"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	jsonrpcclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	penumbra "go.buf.build/grpc/go/penumbra-zone/penumbra/penumbra/client/v1alpha1"
)

var (
	_ node.Node = &Node{}
)

// Node implements a wrapper around both a Tendermint RPCConfig client and a
// chain SDK REST client that allows for essential data queries.
type Node struct {
	ctx                  context.Context
	codec                codec.Codec
	client               *httpclient.HTTP
	grpcConnection       *grpc.ClientConn
	obliviousQueryClient penumbra.ObliviousQueryServiceClient
	specificQueryClient  penumbra.SpecificQueryServiceClient
}

// NewNode allows to build a new Node instance
func NewNode(cfg *Details, codec codec.Codec) (*Node, error) {
	httpClient, err := jsonrpcclient.DefaultHTTPClient(cfg.RPC.Address)
	if err != nil {
		return nil, err
	}

	// Tweak the transport
	httpTransport, ok := (httpClient.Transport).(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("invalid HTTP Transport: %T", httpTransport)
	}
	httpTransport.MaxConnsPerHost = cfg.RPC.MaxConnections

	rpcClient, err := httpclient.NewWithClient(cfg.RPC.Address, "/websocket", httpClient)
	if err != nil {
		return nil, err
	}

	err = rpcClient.Start()
	if err != nil {
		return nil, err
	}

	grpcConnection, err := CreateGrpcConnection(cfg.GRPC)
	if err != nil {
		return nil, err
	}

	return &Node{
		ctx:                  context.Background(),
		codec:                codec,
		client:               rpcClient,
		grpcConnection:       grpcConnection,
		obliviousQueryClient: penumbra.NewObliviousQueryServiceClient(grpcConnection),
		specificQueryClient:  penumbra.NewSpecificQueryServiceClient(grpcConnection),
	}, nil
}

// Genesis implements node.Node
func (cp *Node) Genesis() (*tmctypes.ResultGenesis, error) {
	res, err := cp.client.Genesis(cp.ctx)
	if err != nil && strings.Contains(err.Error(), "use the genesis_chunked API instead") {
		return cp.getGenesisChunked()
	}
	return res, err
}

// getGenesisChunked gets the genesis data using the chinked API instead
func (cp *Node) getGenesisChunked() (*tmctypes.ResultGenesis, error) {
	bz, err := cp.getGenesisChunksStartingFrom(0)
	if err != nil {
		return nil, err
	}

	var genDoc *tmtypes.GenesisDoc
	err = tmjson.Unmarshal(bz, &genDoc)
	if err != nil {
		return nil, err
	}

	return &tmctypes.ResultGenesis{Genesis: genDoc}, nil
}

// getGenesisChunksStartingFrom returns all the genesis chunks data starting from the chunk with the given id
func (cp *Node) getGenesisChunksStartingFrom(id uint) ([]byte, error) {
	res, err := cp.client.GenesisChunked(cp.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error while getting genesis chunk %d out of %d", id, res.TotalChunks)
	}

	bz, err := base64.StdEncoding.DecodeString(res.Data)
	if err != nil {
		return nil, fmt.Errorf("error while decoding genesis chunk %d out of %d", id, res.TotalChunks)
	}

	if id == uint(res.TotalChunks-1) {
		return bz, nil
	}

	nextChunk, err := cp.getGenesisChunksStartingFrom(id + 1)
	if err != nil {
		return nil, err
	}

	return append(bz, nextChunk...), nil
}

// ConsensusState implements node.Node
func (cp *Node) ConsensusState() (*constypes.RoundStateSimple, error) {
	state, err := cp.client.ConsensusState(context.Background())
	if err != nil {
		return nil, err
	}

	var data constypes.RoundStateSimple
	err = tmjson.Unmarshal(state.RoundState, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// LatestHeight implements node.Node
func (cp *Node) LatestHeight() (int64, error) {
	status, err := cp.client.Status(cp.ctx)
	if err != nil {
		return -1, err
	}

	height := status.SyncInfo.LatestBlockHeight
	return height, nil
}

// ChainID implements node.Node
func (cp *Node) ChainID() (string, error) {
	status, err := cp.client.Status(cp.ctx)
	if err != nil {
		return "", err
	}

	chainID := status.NodeInfo.Network
	return chainID, err
}

// Validators implements node.Node
func (cp *Node) Validators(height int64) (*tmctypes.ResultValidators, error) {
	vals := &tmctypes.ResultValidators{
		BlockHeight: height,
	}

	page := 1
	perPage := 100 // maximum 100 entries per page
	stop := false
	for !stop {
		result, err := cp.client.Validators(cp.ctx, &height, &page, &perPage)
		if err != nil {
			return nil, err
		}
		vals.Validators = append(vals.Validators, result.Validators...)
		vals.Count += result.Count
		vals.Total = result.Total
		page++
		stop = vals.Count == vals.Total
	}

	return vals, nil
}

// Block implements node.Node
func (cp *Node) Block(height int64) (*tmctypes.ResultBlock, error) {
	return cp.client.Block(cp.ctx, &height)
}

// BlockResults implements node.Node
func (cp *Node) BlockResults(height int64) (*tmctypes.ResultBlockResults, error) {
	return cp.client.BlockResults(cp.ctx, &height)
}

// Txs implements node.Node
func (cp *Node) Txs(block *tmctypes.ResultBlock) ([]*tmctypes.ResultTx, error) {

	txs := make(map[string]*tmctypes.ResultTx)

	log.Printf("block.Block.Txs %v", len(block.Block.Txs))

	chainID, _ := cp.ChainID()
	blockRangeStream, _ := cp.obliviousQueryClient.CompactBlockRange(context.Background(), &penumbra.CompactBlockRangeRequest{
		ChainId:     chainID,
		StartHeight: uint64(block.Block.Height),
		EndHeight:   uint64(block.Block.Height),
		KeepAlive:   false,
	})

	for {
		in, err := blockRangeStream.Recv()
		if err == io.EOF {
			break
		}
		if len(in.CompactBlock.NotePayloads) == 0 {
			break
		}
		for _, note := range in.CompactBlock.NotePayloads {

			hexTx := hex.EncodeToString(note.Source.Inner)

			resultTx, err := cp.client.Tx(cp.ctx, note.Source.Inner, false)
			if err != nil {
				break
			}
			log.Printf(resultTx.TxResult.String())

			txs[hexTx] = resultTx

		}

	}

	v := make([]*tmctypes.ResultTx, 0, len(txs))

	for _, value := range txs {
		log.Printf(value.TxResult.String())
		v = append(v, value)
	}
	return v, nil
}

func (cp *Node) ValidatorsInfo(height int64) ([]*stakev1alpha1.ValidatorInfo, error) {

	chainID, _ := cp.ChainID()

	validators := make([]*stakev1alpha1.ValidatorInfo, 0)

	validatorInfo, err := cp.obliviousQueryClient.ValidatorInfo(context.Background(), &penumbra.ValidatorInfoRequest{
		ChainId:      chainID,
		ShowInactive: true,
	})

	for {
		in, err := validatorInfo.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error().Str("module", "staking").Err(err)
		}
		if in == nil {
			continue
		}
		validators = append(validators, in.ValidatorInfo)

	}

	return validators, err
}

func getTradingPairs() []*dexv1alpha1.TradingPair {

	pairs := make([]*dexv1alpha1.TradingPair, 0)

	upenumbra, _ := base64.StdEncoding.DecodeString("KeqcLzNx9qSH5+lcJHBB9KNW+YPrBk5dKzvPMiypahA=")
	ugm, _ := base64.StdEncoding.DecodeString("HW2Eq3UZVSBttoUwUi/MUtE7rr2UU7/UH500byp7OAc=")
	ugn, _ := base64.StdEncoding.DecodeString("nwPDkQq3OvLnBwGTD+nmv1Ifb2GEmFCgNHrU++9BsRE=")

	tradingPair1 := &dexv1alpha1.TradingPair{
		Asset_1: &cryptov1alpha1.AssetId{
			Inner: ugm,
		},
		Asset_2: &cryptov1alpha1.AssetId{
			Inner: ugn,
		},
	}

	pairs = append(pairs, tradingPair1)

	tradingPair2 := &dexv1alpha1.TradingPair{
		Asset_1: &cryptov1alpha1.AssetId{
			Inner: ugm,
		},
		Asset_2: &cryptov1alpha1.AssetId{
			Inner: upenumbra,
		},
	}

	pairs = append(pairs, tradingPair2)

	tradingPair3 := &dexv1alpha1.TradingPair{
		Asset_1: &cryptov1alpha1.AssetId{
			Inner: upenumbra,
		},
		Asset_2: &cryptov1alpha1.AssetId{
			Inner: ugn,
		},
	}

	pairs = append(pairs, tradingPair3)

	return pairs
}

func (cp *Node) SwapOutputData(height int64) ([]*dexv1alpha1.BatchSwapOutputData, error) {

	pairs := getTradingPairs()

	swapData := make([]*dexv1alpha1.BatchSwapOutputData, 0, len(pairs))

	for _, pair := range pairs {
		data, err := cp.specificQueryClient.BatchSwapOutputData(context.Background(), &penumbra.BatchSwapOutputDataRequest{
			Height:      uint64(height),
			TradingPair: pair,
		})
		if err != nil {
			swapData = append(swapData, nil)

		} else {

			swapData = append(swapData, data.Data)
			log.Printf("SwapOutputData append ", data)
		}

	}

	return swapData, nil

}

func (cp *Node) CPMMReserves(height int64) ([]*dexv1alpha1.Reserves, error) {

	pairs := getTradingPairs()

	reserves := make([]*dexv1alpha1.Reserves, 0, len(pairs))

	for _, pair := range pairs {
		data, err := cp.specificQueryClient.StubCPMMReserves(context.Background(), &penumbra.StubCPMMReservesRequest{
			TradingPair: pair,
		})
		if err != nil {
			reserves = append(reserves, nil)

		} else {

			reserves = append(reserves, data.Reserves)
			log.Printf("CPMMReserves append ", data)
		}

	}

	return reserves, nil
}

// TxSearch implements node.Node
func (cp *Node) TxSearch(query string, page *int, perPage *int, orderBy string) (*tmctypes.ResultTxSearch, error) {
	return cp.client.TxSearch(cp.ctx, query, false, page, perPage, orderBy)
}

// SubscribeEvents implements node.Node
func (cp *Node) SubscribeEvents(subscriber, query string) (<-chan tmctypes.ResultEvent, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	eventCh, err := cp.client.Subscribe(ctx, subscriber, query)
	return eventCh, cancel, err
}

// SubscribeNewBlocks implements node.Node
func (cp *Node) SubscribeNewBlocks(subscriber string) (<-chan tmctypes.ResultEvent, context.CancelFunc, error) {
	return cp.SubscribeEvents(subscriber, "tm.event = 'NewBlock'")
}

// Stop implements node.Node
func (cp *Node) Stop() {
	err := cp.client.Stop()
	if err != nil {
		panic(fmt.Errorf("error while stopping proxy: %s", err))
	}

	err = cp.grpcConnection.Close()
	if err != nil {
		panic(fmt.Errorf("error while closing gRPC connection: %s", err))
	}
}
