package sync

import (
	"context"
	"testing"
	"time"

	"github.com/NethermindEth/juno/blockchain"
	"github.com/NethermindEth/juno/db/pebble"
	"github.com/NethermindEth/juno/testsource"
	"github.com/NethermindEth/juno/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncBlocks(t *testing.T) {
	gw, closeFn := testsource.NewTestGateway(utils.MAINNET)
	defer closeFn()
	testBlockchain := func(t *testing.T, bc *blockchain.Blockchain) bool {
		return assert.NoError(t, func() error {
			headBlock, err := bc.Head()
			require.NoError(t, err)

			height := int(headBlock.Number)
			for height >= 0 {
				b, err := gw.BlockByNumber(context.Background(), uint64(height))
				if err != nil {
					return err
				}

				block, err := bc.GetBlockByNumber(uint64(height))
				require.NoError(t, err)

				assert.Equal(t, b, block)
				height--
			}
			return nil
		}())
	}
	log := utils.NewNopZapLogger()
	t.Run("sync multiple blocks in an empty db", func(t *testing.T) {
		testDB := pebble.NewMemTest()
		bc := blockchain.New(testDB, utils.MAINNET)
		synchronizer := NewSynchronizer(bc, gw, log)
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(time.Second)
			cancel()
		}()
		require.NoError(t, synchronizer.Run(ctx))

		testBlockchain(t, bc)
	})
	t.Run("sync multiple blocks in a non-empty db", func(t *testing.T) {
		testDB := pebble.NewMemTest()
		bc := blockchain.New(testDB, utils.MAINNET)
		b0, err := gw.BlockByNumber(context.Background(), 0)
		require.NoError(t, err)
		s0, err := gw.StateUpdate(context.Background(), 0)
		require.NoError(t, err)
		require.NoError(t, bc.Store(b0, s0, nil))

		synchronizer := NewSynchronizer(bc, gw, log)
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(time.Second)
			cancel()
		}()
		require.NoError(t, synchronizer.Run(ctx))

		testBlockchain(t, bc)
	})
}
