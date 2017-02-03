package adapter

type AdapterInterface interface {
	Check() error
	GetBrief() string
	Parse() error
	Process() error
	String() string
}
