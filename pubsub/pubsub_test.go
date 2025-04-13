package pubsub

import (
	"reflect"
	"sync"
	"testing"
)

func TestNewMessage(t *testing.T) {
	msg := NewMessage(TestTopic, 1)
	if msg.topic != TestTopic {
		t.Error("Wrong Topic")
	}
	value, ok := msg.Value.(int)
	if !ok {
		t.Error("Wrong value type")
	}
	if value != 1 {
		t.Error("Wrong value")
	}
}

func TestRegisterNewSubcriber(t *testing.T) {
	var pubsubManager = NewPubSubManager()
	id, event, err := pubsubManager.RegisterNewSubscriber(TestTopic)
	if err != nil {
		t.Errorf("Counter error %+v \n", err)
	}
	subMap, ok := pubsubManager.subscriberList[TestTopic]
	if !ok {
		t.Error("Can not get subcribe map by topic")
	}
	if subChan, ok := subMap[id]; !ok {
		t.Error("Can not get sub chan")
	} else {
		if !reflect.DeepEqual(event, subChan) {
			t.Error("Wrong Subchan")
		}
	}
}
func TestRegisterNewSubcribeWithUnregisteredTopic(t *testing.T) {
	var pubsubManager = NewPubSubManager()
	id, _, err := pubsubManager.RegisterNewSubscriber("ajsdkl;awjdkl")
	if id != 0 {
		t.Error("Wrong Event ID")
	}
	if pubsubErr, ok := err.(*PubSubError); !ok {
		t.Error("Wrong error type")
	} else {
		if pubsubErr.Code != ErrCodeMessage[UnregisteredTopicError].Code {
			t.Error("Wrong Error code")
		}
	}
}
func TestUnsubcribe(t *testing.T) {
	var pubsubManager = NewPubSubManager()
	id, _, _ := pubsubManager.RegisterNewSubscriber(TestTopic)
	pubsubManager.Unsubscribe(TestTopic, id)
	subMap, ok := pubsubManager.subscriberList[TestTopic]
	if !ok {
		t.Error("Can not get subcribe map by topic")
	}
	if _, ok := subMap[id]; ok {
		t.Error("Should have no sub chan")
	}
}
func TestPublishMessage(t *testing.T) {
	var pubsubManager = NewPubSubManager()
	pubsubManager.PublishMessage(NewMessage(TestTopic, "abc"))
	msgs, ok := pubsubManager.messageBroker[TestTopic]
	if !ok {
		t.Error("No Message found with this topic")
	}
	if len(msgs) != 1 {
		t.Errorf("Should have only 1 message %+v \n", len(pubsubManager.messageBroker[TestTopic]))
	}
	if msgs[0].topic != TestTopic {
		t.Error("Wrong Topic")
	}
	valueInterface := msgs[0].Value
	if value, ok := valueInterface.(string); !ok {
		t.Error("Wrong msg type")
	} else {
		if value != "abc" {
			t.Error("Wrong msg value")
		}
	}
}
func TestMessageBroken(t *testing.T) {
	var pubsubManager = NewPubSubManager()
	var wg sync.WaitGroup
	go pubsubManager.Start()
	id, event, err := pubsubManager.RegisterNewSubscriber(TestTopic)
	if err != nil {
		t.Error("Error when subcription")
	}
	wg.Add(1)
	go func(event chan *Message) {
		defer wg.Done()
		for msg := range event {
			topic := msg.topic
			if topic != TestTopic {
				t.Error("Wrong subcription topic")
			}
			if value, ok := msg.Value.(string); !ok {
				t.Error("Wrong value type")
			} else {
				if value != "abc" {
					t.Error("Unexpected value")
				}
			}
			close(event)
		}
	}(event)
	pubsubManager.PublishMessage(NewMessage(TestTopic, "abc"))
	wg.Wait()
	pubsubManager.Unsubscribe(TestTopic, id)
	return
}
func TestHasTopic(t *testing.T) {
	var pubsubManager = NewPubSubManager()
	if !pubsubManager.HasTopic(NewBeaconBlockTopic) {
		t.Error("Pubsub manager should have this topic")
	}
	if pubsubManager.HasTopic("lajsdlkjaskldj") {
		t.Error("Pubsub manager should not have this topic")
	}
}

func TestAddTopic(t *testing.T) {
	var pubsubManager = NewPubSubManager()
	if pubsubManager.HasTopic("lajsdlkjaskldj") {
		t.Error("Pubsub manager should not have this topic")
	}
	pubsubManager.AddTopic("lajsdlkjaskldj")
	if !pubsubManager.HasTopic("lajsdlkjaskldj") {
		t.Error("Pubsub manager should have this topic")
	}
}
