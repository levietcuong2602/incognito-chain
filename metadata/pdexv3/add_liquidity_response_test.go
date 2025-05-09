package pdexv3

import (
	"testing"

	"github.com/levietcuong2602/incognito-chain/common"
	metadataCommon "github.com/levietcuong2602/incognito-chain/metadata/common"
)

func TestAddLiquidityResponse_ValidateSanityData(t *testing.T) {
	type fields struct {
		MetadataBase metadataCommon.MetadataBase
		status       string
		txReqID      string
	}
	type args struct {
		chainRetriever      metadataCommon.ChainRetriever
		shardViewRetriever  metadataCommon.ShardViewRetriever
		beaconViewRetriever metadataCommon.BeaconViewRetriever
		beaconHeight        uint64
		tx                  metadataCommon.Transaction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		want1   bool
		wantErr bool
	}{
		{
			name: "Status is null",
			fields: fields{
				status: "",
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "txReqID is invalid",
			fields: fields{
				status: common.PDEContributionRefundChainStatus,
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "txReqID is empty",
			fields: fields{
				status:  common.PDEContributionRefundChainStatus,
				txReqID: common.Hash{}.String(),
			},
			args:    args{},
			want:    false,
			want1:   false,
			wantErr: true,
		},
		{
			name: "Valid Input",
			fields: fields{
				status:  common.PDEContributionRefundChainStatus,
				txReqID: common.PRVIDStr,
			},
			args:    args{},
			want:    true,
			want1:   true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &AddLiquidityResponse{
				MetadataBase: tt.fields.MetadataBase,
				status:       tt.fields.status,
				txReqID:      tt.fields.txReqID,
			}
			got, got1, err := response.ValidateSanityData(tt.args.chainRetriever, tt.args.shardViewRetriever, tt.args.beaconViewRetriever, tt.args.beaconHeight, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddLiquidityResponse.ValidateSanityData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AddLiquidityResponse.ValidateSanityData() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("AddLiquidityResponse.ValidateSanityData() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
