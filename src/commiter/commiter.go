package commiter

type Commiter interface {
	Commit(msg interface{}) error
	GetMsg(id uint64) (msg interface{}, err error)
	Consume() (msg interface{}, err error)
	Total() uint64
	CurId() uint64
	MaxId() uint64
}
