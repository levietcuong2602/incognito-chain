package pdexv3

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
)

func TestAcceptWithdrawLiquidity_FromStringSlice(t *testing.T) {
	type fields struct {
		poolPairID  string
		nftID       common.Hash
		tokenID     common.Hash
		tokenAmount uint64
		shareAmount uint64
		otaReceiver string
		txReqID     common.Hash
		shardID     byte
	}
	type args struct {
		source []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AcceptWithdrawLiquidity{
				poolPairID:  tt.fields.poolPairID,
				nftID:       tt.fields.nftID,
				tokenID:     tt.fields.tokenID,
				tokenAmount: tt.fields.tokenAmount,
				shareAmount: tt.fields.shareAmount,
				otaReceiver: tt.fields.otaReceiver,
				txReqID:     tt.fields.txReqID,
				shardID:     tt.fields.shardID,
			}
			if err := a.FromStringSlice(tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("AcceptWithdrawLiquidity.FromStringSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAcceptWithdrawLiquidity_StringSlice(t *testing.T) {
	type fields struct {
		poolPairID  string
		nftID       common.Hash
		tokenID     common.Hash
		tokenAmount uint64
		shareAmount uint64
		otaReceiver string
		txReqID     common.Hash
		shardID     byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AcceptWithdrawLiquidity{
				poolPairID:  tt.fields.poolPairID,
				nftID:       tt.fields.nftID,
				tokenID:     tt.fields.tokenID,
				tokenAmount: tt.fields.tokenAmount,
				shareAmount: tt.fields.shareAmount,
				otaReceiver: tt.fields.otaReceiver,
				txReqID:     tt.fields.txReqID,
				shardID:     tt.fields.shardID,
			}
			got, err := a.StringSlice()
			if (err != nil) != tt.wantErr {
				t.Errorf("AcceptWithdrawLiquidity.StringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AcceptWithdrawLiquidity.StringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
