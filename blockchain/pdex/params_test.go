package pdex

import (
	"testing"

	"github.com/incognitochain/incognito-chain/common"
)

func TestParams_IsZeroValue(t *testing.T) {
	type fields struct {
		DefaultFeeRateBPS               uint
		FeeRateBPS                      map[string]uint
		PRVDiscountPercent              uint
		TradingProtocolFeePercent       uint
		TradingStakingPoolRewardPercent uint
		PDEXRewardPoolPairsShare        map[string]uint
		StakingPoolsShare               map[string]uint
		StakingRewardTokens             []common.Hash
		MintNftRequireAmount            uint64
		MaxOrdersPerNft                 uint
		OrderMiningRewardRatioBPS       map[string]uint
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "is zero value",
			fields: fields{
				FeeRateBPS:                make(map[string]uint),
				StakingPoolsShare:         make(map[string]uint),
				PDEXRewardPoolPairsShare:  make(map[string]uint),
				StakingRewardTokens:       []common.Hash{},
				OrderMiningRewardRatioBPS: make(map[string]uint),
			},
			want: true,
		},
		{
			name: "not zero value",
			fields: fields{
				DefaultFeeRateBPS: 30,
				FeeRateBPS: map[string]uint{
					"abc": 12,
				},
				PRVDiscountPercent:              25,
				TradingProtocolFeePercent:       0,
				TradingStakingPoolRewardPercent: 10,
				PDEXRewardPoolPairsShare:        map[string]uint{},
				StakingPoolsShare: map[string]uint{
					common.PRVIDStr: 10,
				},
				StakingRewardTokens:  []common.Hash{},
				MintNftRequireAmount: 1000000000,
				MaxOrdersPerNft:      10,
				OrderMiningRewardRatioBPS: map[string]uint{
					"abs": 100,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &Params{
				DefaultFeeRateBPS:               tt.fields.DefaultFeeRateBPS,
				FeeRateBPS:                      tt.fields.FeeRateBPS,
				PRVDiscountPercent:              tt.fields.PRVDiscountPercent,
				TradingProtocolFeePercent:       tt.fields.TradingProtocolFeePercent,
				TradingStakingPoolRewardPercent: tt.fields.TradingStakingPoolRewardPercent,
				PDEXRewardPoolPairsShare:        tt.fields.PDEXRewardPoolPairsShare,
				StakingPoolsShare:               tt.fields.StakingPoolsShare,
				StakingRewardTokens:             tt.fields.StakingRewardTokens,
				MintNftRequireAmount:            tt.fields.MintNftRequireAmount,
				MaxOrdersPerNft:                 tt.fields.MaxOrdersPerNft,
				OrderMiningRewardRatioBPS:       tt.fields.OrderMiningRewardRatioBPS,
			}
			if got := params.IsZeroValue(); got != tt.want {
				t.Errorf("Params.IsZeroValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
