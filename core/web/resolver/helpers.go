package resolver

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/smartcontractkit/chainlink/core/assets"
	"github.com/smartcontractkit/chainlink/core/bridges"
	"github.com/smartcontractkit/chainlink/core/store/models"
)

// These methods should be moved into the service

// ValidateBridgeTypeNotExist checks that a bridge has not already been created
func ValidateBridgeTypeNotExist(bt *bridges.BridgeTypeRequest, orm bridges.ORM) error {
	fe := models.NewJSONAPIErrors()
	_, err := orm.FindBridge(bt.Name)
	if err == nil {
		fe.Add(fmt.Sprintf("Bridge Type %v already exists", bt.Name))
	}
	if err != nil && err != sql.ErrNoRows {
		fe.Add(fmt.Sprintf("Error determining if bridge type %v already exists", bt.Name))
	}
	return fe.CoerceEmptyToNil()
}

// ValidateBridgeType checks that the bridge type doesn't have a duplicate
// or invalid name or invalid url
func ValidateBridgeType(bt *bridges.BridgeTypeRequest, orm bridges.ORM) error {
	fe := models.NewJSONAPIErrors()
	if len(bt.Name.String()) < 1 {
		fe.Add("No name specified")
	}
	if _, err := bridges.NewTaskType(bt.Name.String()); err != nil {
		fe.Merge(err)
	}
	u := bt.URL.String()
	if len(strings.TrimSpace(u)) == 0 {
		fe.Add("URL must be present")
	}
	if bt.MinimumContractPayment != nil &&
		bt.MinimumContractPayment.Cmp(assets.NewLinkFromJuels(0)) < 0 {
		fe.Add("MinimumContractPayment must be positive")
	}
	return fe.CoerceEmptyToNil()
}
