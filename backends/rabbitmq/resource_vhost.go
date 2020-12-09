package rabbitmq

import (
	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
	"github.com/teseraio/ensemble/operator"
)

// VHost is a Rabbitmq vhost
type VHost struct {
	operator.BaseResource `schema:",squash"`

	// Name is the name of the vhost
	Name string
}

// GetName implements the Resource intrface
func (v *VHost) GetName() string {
	return "VHost"
}

// Delete implements the Resource interface
func (v *VHost) Delete(req interface{}) error {
	client := req.(*rabbithole.Client)

	if _, err := client.DeleteVhost(v.Name); err != nil {
		return err
	}
	return nil
}

// Reconcile implements the Resource interface
func (v *VHost) Reconcile(req interface{}) error {
	client := req.(*rabbithole.Client)

	if _, err := client.PutVhost(v.Name, rabbithole.VhostSettings{}); err != nil {
		return err
	}
	return nil
}
