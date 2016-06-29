package adapter

import (
	"encoding/json"
	"github.com/OpsKitchen/ok_agent/model/api/returndata"
)

type Augeas struct {
	itemList []returndata.Augeas
}

func (augeasAdapter *Augeas) CastItemList(list interface{}) error {
	var dataByte []byte
	dataByte, _ = json.Marshal(list)
	return json.Unmarshal(dataByte, &augeasAdapter.itemList)
}

func (augeasAdapter *Augeas) Process() error {
	var err error
	var item returndata.Augeas
	for _, item = range augeasAdapter.itemList {
		err = augeasAdapter.processAugeas(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (augeasAdapter *Augeas) processAugeas(item returndata.Augeas) error {
	return nil
}
