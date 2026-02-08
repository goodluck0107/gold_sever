package product_deliver

import (
	"encoding/json"
	"github.com/open-source/game/chess.git/pkg/static"
	"testing"
)

func TestDeliverProduct(t *testing.T) {
	var msg static.Msg_DeliverProduct

	var data = "{\"type\":27,\"product_type\":4, \"num\":30,   \"uid\":1000001,  \"extra\":\"{\\\"expend_num\\\":30,\\\"expend_type\\\":7,\\\"got\\\":[{\\\"wt\\\":3,\\\"wn\\\":38888}]}\"}"

	err := json.Unmarshal([]byte(data), &msg)

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%+v\n", msg)
}
