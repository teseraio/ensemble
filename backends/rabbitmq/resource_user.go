package rabbitmq

import (
	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/schema"
)

func user() *operator.Resource2 {
	return &operator.Resource2{
		Name: "User",
		Schema: schema.Schema2{
			Spec: &schema.Record{
				Fields: map[string]*schema.Field{
					"username": {
						Type:     schema.TypeString,
						ForceNew: true,
						Required: true,
					},
					"password": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
		},
		DeleteFn: func(req *operator.CallbackRequest) error {
			client := req.Client.(*rabbithole.Client)

			if _, err := client.DeleteUser(req.Get("username").(string)); err != nil {
				return err
			}
			return nil
		},
		ApplyFn: func(req *operator.CallbackRequest) error {
			client := req.Client.(*rabbithole.Client)

			username := req.Get("username").(string)
			settings := rabbithole.UserSettings{
				Name:     username,
				Password: req.Get("password").(string),
			}
			if _, err := client.PutUser(username, settings); err != nil {
				return err
			}

			return nil
		},
	}
}
