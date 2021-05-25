package rabbitmq

import (
	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/schema"
)

func vhost() *operator.Resource2 {
	return &operator.Resource2{
		Name: "VHost",
		Schema: schema.Schema2{
			Spec: &schema.Record{
				Fields: map[string]*schema.Field{
					"name": {
						Type:     schema.TypeString,
						Required: true,
						ForceNew: true,
					},
				},
			},
		},
		DeleteFn: func(req *operator.CallbackRequest) error {
			client := req.Client.(*rabbithole.Client)

			if _, err := client.DeleteVhost(req.Get("name").(string)); err != nil {
				return err
			}
			return nil
		},
		ApplyFn: func(req *operator.CallbackRequest) error {
			client := req.Client.(*rabbithole.Client)

			if _, err := client.PutVhost(req.Get("name").(string), rabbithole.VhostSettings{}); err != nil {
				return err
			}
			return nil
		},
	}
}
