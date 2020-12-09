package rabbitmq

import (
	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
	"github.com/teseraio/ensemble/operator"
)

// Exchange is a Rabbitmq exchange
type Exchange struct {
	operator.BaseResource `schema:",squash"`

	// Name is the name of the exchange
	Name string

	// VHost is the name of the vhost
	VHost string

	// Settings are the settings of the exchange
	Settings *ExchangeSettings
}

// ExchangeSettings are the settings of the exchange
type ExchangeSettings struct {
	Type string
}

// GetName implements the Resource intrface
func (e *Exchange) GetName() string {
	return "Exchange"
}

// Delete implements the Resource interface
func (e *Exchange) Delete(req interface{}) error {
	client := req.(*rabbithole.Client)

	if _, err := client.DeleteExchange(e.VHost, e.Name); err != nil {
		return err
	}
	return nil
}

// Reconcile implements the Resource interface
func (e *Exchange) Reconcile(req interface{}) error {
	client := req.(*rabbithole.Client)

	settings := rabbithole.ExchangeSettings{
		Type: e.Settings.Type,
	}
	if _, err := client.DeclareExchange(e.VHost, e.Name, settings); err != nil {
		return err
	}
	return nil
}
