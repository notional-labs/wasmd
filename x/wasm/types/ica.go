package types

import (
	"fmt"
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	connectiontypes "github.com/cosmos/ibc-go/v3/modules/core/03-connection/types"
)

const (
	// ControllerPortFormat is the expected port identifier format to which controller chains must conform
	// See (TODO: Link to spec when updated)
	ControllerPortFormat = "<wasm>.<controller-conn-seq>.<host-conn-seq>.<owner>"
)

// GeneratePortID generates an interchain accounts controller port identifier for the provided owner
// in the following format:
//
// 'ics-27-<connectionSequence>-<counterpartyConnectionSequence>-<owner-address>'
// https://github.com/seantking/ibc/tree/sean/ics-27-updates/spec/app/ics-027-interchain-accounts#registering--controlling-flows
// TODO: update link to spec
func GenerateICAPortID(owner, connectionID, counterpartyConnectionID string) (string, error) {
	if strings.TrimSpace(owner) == "" {
		return "", sdkerrors.Wrap(icatypes.ErrInvalidAccountAddress, "owner address cannot be empty")
	}

	connectionSeq, err := connectiontypes.ParseConnectionSequence(connectionID)
	if err != nil {
		return "", sdkerrors.Wrap(err, "invalid connection identifier")
	}

	counterpartyConnectionSeq, err := connectiontypes.ParseConnectionSequence(counterpartyConnectionID)
	if err != nil {
		return "", sdkerrors.Wrap(err, "invalid counterparty connection identifier")
	}

	return fmt.Sprint(
		"wasm", icatypes.Delimiter,
		connectionSeq, icatypes.Delimiter,
		counterpartyConnectionSeq, icatypes.Delimiter,
		owner,
	), nil
}
