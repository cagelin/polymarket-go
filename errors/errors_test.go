package errors_test

import (
	"context"
	"errors"
	"testing"

	sdkerrors "github.com/cagelin/polymarket-go/errors"
)

func TestWrappedCausesRemainDiscoverable(t *testing.T) {
	for _, failure := range []error{
		&sdkerrors.TransportError{Cause: context.Canceled},
		&sdkerrors.UnexpectedResponseError{Message: "bad response", Cause: context.Canceled},
		&sdkerrors.SigningError{Cause: context.Canceled},
	} {
		if !errors.Is(failure, context.Canceled) || failure.Error() == "" {
			t.Errorf("wrapped %T = %v, want nonempty error retaining context.Canceled", failure, failure)
		}
	}
	rejection := &sdkerrors.RequestRejectedError{Message: "rejected", Status: 400}
	for _, failure := range []error{
		&sdkerrors.RFQRequestRejectedError{Rejection: rejection},
		&sdkerrors.OrderHeartbeatMismatchError{HeartbeatID: "id", Rejection: rejection},
	} {
		var discovered *sdkerrors.RequestRejectedError
		if !errors.As(failure, &discovered) || !errors.Is(discovered, rejection) || failure.Error() != "rejected" {
			t.Errorf("wrapped %T = %v, want original RequestRejectedError", failure, failure)
		}
	}
}
