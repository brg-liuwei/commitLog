package channer

type Channer interface {
	GetChan() (ch chan interface{})
}
