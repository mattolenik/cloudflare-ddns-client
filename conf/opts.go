package conf

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type StringOption struct {
	Name        string
	Description string
	Default     string
	flags       *pflag.FlagSet
}

func (o *StringOption) Get() string {
	return viper.GetString(o.Name)
}

func (o *StringOption) Bind(flags *pflag.FlagSet) *StringOption {
	o.flags = flags
	flags.String(o.Name, o.Default, o.Description)
	return o
}

func (o *StringOption) BindVar(flags *pflag.FlagSet, v *string) *StringOption {
	o.flags = flags
	flags.StringVar(v, o.Name, o.Default, o.Description)
	return o
}

func (o *StringOption) Viper() {
	if o.flags == nil {
		panic("Must call Bind or BindVar before Viper")
	}
	viper.BindPFlag(o.Name, o.flags.Lookup(o.Name))
	viper.SetDefault(o.Name, o.Default)
}

type StringOptionP struct {
	Name        string
	ShortName   string
	Description string
	Default     string
	flags       *pflag.FlagSet
}

func (o *StringOptionP) Get() string {
	return viper.GetString(o.Name)
}

func (o *StringOptionP) Bind(flags *pflag.FlagSet) *StringOptionP {
	o.flags = flags
	flags.String(o.Name, o.Default, o.Description)
	return o
}

func (o *StringOptionP) BindVar(flags *pflag.FlagSet, v *string) *StringOptionP {
	o.flags = flags
	flags.StringVar(v, o.Name, o.Default, o.Description)
	return o
}

func (o *StringOptionP) Viper() {
	if o.flags == nil {
		panic("Must call Bind or BindVar before Viper")
	}
	viper.BindPFlag(o.Name, o.flags.Lookup(o.Name))
	viper.SetDefault(o.Name, o.Default)
}

type BoolOption struct {
	Name        string
	Description string
	Default     bool
	flags       *pflag.FlagSet
}

func (o *BoolOption) Get() bool {
	return viper.GetBool(o.Name)
}

func (o *BoolOption) Bind(flags *pflag.FlagSet) *BoolOption {
	o.flags = flags
	flags.Bool(o.Name, o.Default, o.Description)
	return o
}

func (o *BoolOption) BindVar(flags *pflag.FlagSet, v *bool) *BoolOption {
	o.flags = flags
	flags.BoolVar(v, o.Name, o.Default, o.Description)
	return o
}

func (o *BoolOption) Viper() {
	if o.flags == nil {
		panic("Must call Bind or BindVar before Viper")
	}
	viper.BindPFlag(o.Name, o.flags.Lookup(o.Name))
	viper.SetDefault(o.Name, o.Default)
}

type BoolOptionP struct {
	Name        string
	ShortName   string
	Description string
	Default     bool
	flags       *pflag.FlagSet
}

func (o *BoolOptionP) Get() bool {
	return viper.GetBool(o.Name)
}

func (o *BoolOptionP) Bind(flags *pflag.FlagSet) *BoolOptionP {
	o.flags = flags
	flags.BoolP(o.Name, o.ShortName, o.Default, o.Description)
	return o
}

func (o *BoolOptionP) BindVar(flags *pflag.FlagSet, v *bool) *BoolOptionP {
	o.flags = flags
	flags.BoolVarP(v, o.Name, o.ShortName, o.Default, o.Description)
	return o
}

func (o *BoolOptionP) Viper() {
	if o.flags == nil {
		panic("Must call Bind or BindVar before Viper")
	}
	viper.BindPFlag(o.Name, o.flags.Lookup(o.Name))
	viper.SetDefault(o.Name, o.Default)
}
