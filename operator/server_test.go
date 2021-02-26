package operator

/*
type nullResource struct {
	BaseResource
}

func (n *nullResource) GetName() string {
	return ""
}

func (n *nullResource) Delete(conn interface{}) error {
	return nil
}

func (n *nullResource) GetID() string {
	return ""
}

func (n *nullResource) Get(conn interface{}) error {
	return nil
}

func (n *nullResource) Reconcile(conn interface{}) error {
	return nil
}

func (n *nullResource) Init(spec map[string]interface{}) error {
	return nil
}

func TestForceNew(t *testing.T) {
	var resource struct {
		*nullResource
		A string `schema:"a,force-new"`
	}
	cases := []struct {
		res    Resource
		new    string
		old    string
		expect bool
	}{
		{
			res:    resource,
			new:    `{"a": "b"}`,
			old:    `{"a": "c"}`,
			expect: true,
		},
		{
			res:    resource,
			new:    `{"a": "b"}`,
			old:    `{"a": "b"}`,
			expect: false,
		},
	}

	for _, c := range cases {
		found, err := isForceNew(c.res, &proto.ResourceSpec{Params: c.old}, &proto.ResourceSpec{Params: c.new})
		if err != nil {
			t.Fatal(err)
		}
		if c.expect != found {
			t.Fatal("err")
		}
	}
}
*/
