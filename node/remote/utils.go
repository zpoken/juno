package remote

import (
	"context"
	"regexp"
	"strconv"

	// grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	HTTPProtocols = regexp.MustCompile("https?://")
)

// GetHeightRequestContext adds the height to the context for querying the state at a given height
func GetHeightRequestContext(context context.Context, height int64) context.Context {
	return metadata.AppendToOutgoingContext(
		context,
		// grpctypes.GRPCBlockHeightHeader,
		strconv.FormatInt(height, 10),
	)
}
