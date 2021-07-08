package flagset

import (
	"flag"
	"fmt"
	"strings"
)

type Flagset struct {
	flags []*FlagVar
	set   *flag.FlagSet
}

func NewFlagSet(name string) *Flagset {
	return &Flagset{
		flags: []*FlagVar{},
		set:   flag.NewFlagSet(name, flag.ContinueOnError),
	}
}

type FlagVar struct {
	Name  string
	Usage string
}

func (f *Flagset) addFlag(fl *FlagVar) {
	f.flags = append(f.flags, fl)
}

func (f *Flagset) Help() string {
	str := "Options:\n\n"
	items := []string{}
	for _, item := range f.flags {
		items = append(items, fmt.Sprintf("  -%s\n    %s", item.Name, item.Usage))
	}
	return str + strings.Join(items, "\n\n")
}

func (f *Flagset) Parse(args []string) error {
	return f.set.Parse(args)
}

func (f *Flagset) Args() []string {
	return f.set.Args()
}

func (f *Flagset) BoolVar() {

}

type BoolFlag struct {
	Name    string
	Usage   string
	Default bool
	Value   *bool
}

func (f *Flagset) BoolFlag(b *BoolFlag) {
	f.addFlag(&FlagVar{
		Name:  b.Name,
		Usage: b.Usage,
	})
	f.set.BoolVar(b.Value, b.Name, b.Default, "")
}

type StringFlag struct {
	Name    string
	Usage   string
	Default string
	Value   *string
}

func (f *Flagset) StringFlag(b *StringFlag) {
	f.addFlag(&FlagVar{
		Name:  b.Name,
		Usage: b.Usage,
	})
	f.set.StringVar(b.Value, b.Name, b.Default, "")
}
