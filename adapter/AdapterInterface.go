package adapter

type AdapterInterface interface {
	CastItemList(interface{}) error
	Process() error
}
