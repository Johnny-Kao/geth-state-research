package catalyst

import (
	"math"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Verify the protocol-required empty result when the entire requested range
// is beyond the latest known block, including uint64 addition overflow.
func TestResearchG01PayloadBodiesOverflowRange(t *testing.T) {
	node, ethservice, _ := setupBodies(t)
	defer node.Close()
	api := newConsensusAPIWithoutHeartbeat(ethservice)
	for _, tc := range []struct {
		start, count uint64
	}{
		{math.MaxUint64, 2},
		{math.MaxUint64 - 1, 3},
		{math.MaxUint64 - 1023, 1024},
	} {
		v1, err := api.GetPayloadBodiesByRangeV1(hexutil.Uint64(tc.start), hexutil.Uint64(tc.count))
		if err != nil || len(v1) != 0 {
			t.Errorf("V1 start=%d count=%d: got len=%d err=%v, want empty result", tc.start, tc.count, len(v1), err)
		}
		v2, err := api.GetPayloadBodiesByRangeV2(hexutil.Uint64(tc.start), hexutil.Uint64(tc.count))
		if err != nil || len(v2) != 0 {
			t.Errorf("V2 start=%d count=%d: got len=%d err=%v, want empty result", tc.start, tc.count, len(v2), err)
		}
	}
}
