package adapter

type AdapterInterface interface {
	Brief() string
	Check() error
	Parse() error
	Process() error
	String() string
}
