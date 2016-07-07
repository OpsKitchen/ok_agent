package adapter

type AdapterInterface interface {
	CheckItem() error
	ParseItem() error
	Process() error
}
