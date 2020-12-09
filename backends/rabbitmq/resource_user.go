package rabbitmq

import (
	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
	"github.com/teseraio/ensemble/operator"
)

// User is a user in Rabbitmq
type User struct {
	operator.BaseResource `schema:",squash"`

	// Username is the name of the user
	Username string

	// Password is the password of the user
	Password string
}

// GetName implements the Resource interface
func (u *User) GetName() string {
	return "User"
}

// Delete implements the User interface
func (u *User) Delete(req interface{}) error {
	client := req.(*rabbithole.Client)

	if _, err := client.DeleteUser(u.ID); err != nil {
		return err
	}
	return nil
}

// Reconcile implements the Resource interface
func (u *User) Reconcile(req interface{}) error {
	client := req.(*rabbithole.Client)

	settings := rabbithole.UserSettings{
		Name:     u.Username,
		Password: u.Password,
	}
	if _, err := client.PutUser(u.Username, settings); err != nil {
		return err
	}
	u.SetID(u.Username)
	return nil
}
