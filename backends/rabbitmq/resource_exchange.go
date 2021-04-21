package rabbitmq

import (
	"fmt"

	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/schema"
)

func exchange() *operator.Resource2 {
	return &operator.Resource2{
		Name: "Exchange",
		Schema: &schema.Record{
			Fields: map[string]*schema.Field{
				"name": {
					Type: schema.TypeString,
				},
				"vhost": {
					Type: schema.TypeString,
				},
				"settings": {
					Type: &schema.Record{
						Fields: map[string]*schema.Field{
							"type": {
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
		DeleteFn: func(req *operator.CallbackRequest) error {
			client := req.Client.(*rabbithole.Client)

			vhost := req.Get("vhost").(string)
			exchange := req.Get("name").(string)

			if _, err := client.DeleteExchange(vhost, exchange); err != nil {
				return err
			}
			return nil
		},
		ApplyFn: func(req *operator.CallbackRequest) error {
			client := req.Client.(*rabbithole.Client)

			vhost := req.Get("vhost").(string)
			exchange := req.Get("name").(string)

			settings := rabbithole.ExchangeSettings{
				Type: req.Get("settings.type").(string),
			}
			if _, err := client.DeclareExchange(vhost, exchange, settings); err != nil {
				return err
			}
			return nil
		},
	}
}

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

// Init implements the Resource interface
func (e *Exchange) Init(spec map[string]interface{}) error {
	if !contains([]string{"direct", "topic", "headers", "fanout"}, e.Settings.Type) {
		return fmt.Errorf("exchange type %s is invalid", e.Settings.Type)
	}
	return nil
}

func contains(j []string, i string) bool {
	for _, o := range j {
		if o == i {
			return true
		}
	}
	return false
}
