package core

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/beacon"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
)

func TestResearchParallelStateDBLogsMatchSequential(t *testing.T) {
	// PUSH1 0; PUSH1 0; LOG0; STOP
	logger := common.HexToAddress("0x1000")
	env := newBALTestEnv(types.GenesisAlloc{
		logger: {Balance: big.NewInt(0), Code: []byte{0x60, 0x00, 0x60, 0x00, 0xa0, 0x00}},
	})
	engine := beacon.New(ethash.NewFaker())
	_, blocks, _ := GenerateChainWithGenesis(env.gspec, engine, 1, func(_ int, g *BlockGen) {
		g.AddTx(env.tx(0, &logger, big.NewInt(0), 100_000, 0, nil))
	})
	block := blocks[0]
	if block.AccessList() == nil {
		t.Fatal("expected Amsterdam block access list")
	}

	bc, err := NewBlockChain(rawdb.NewMemoryDatabase(), env.gspec, engine, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer bc.Stop()

	parState, err := bc.State()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewStateProcessor(bc).Process(context.Background(), block, parState, nil, nil, vm.Config{}, nil, nil); err != nil {
		t.Fatalf("parallel process: %v", err)
	}

	seqState, err := bc.State()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewStateProcessor(bc).Process(context.Background(), block, seqState, nil, nil, vm.Config{DisableParallelExecution: true}, nil, nil); err != nil {
		t.Fatalf("sequential process: %v", err)
	}

	parLogs, seqLogs := parState.Logs(), seqState.Logs()
	if len(seqLogs) == 0 {
		t.Fatal("test setup emitted no logs")
	}
	if len(parLogs) != len(seqLogs) {
		t.Fatalf("StateDB logs diverged: parallel=%d sequential=%d", len(parLogs), len(seqLogs))
	}
}
