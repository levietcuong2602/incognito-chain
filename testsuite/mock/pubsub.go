package mock

import "github.com/levietcuong2602/incognito-chain/pubsub"

type Pubsub struct{}

func (ps *Pubsub) PublishMessage(message *pubsub.Message) {
	return
}
