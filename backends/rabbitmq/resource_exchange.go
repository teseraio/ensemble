package rabbitmq

import (
	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/schema"
)

func exchange() *operator.Resource2 {
	return &operator.Resource2{
		Name: "Exchange",
		Schema: schema.Schema2{
			Spec: &schema.Record{
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
