package table

// ---------------------------------------------------------------------------

type Table interface {
	Insert(docs ...interface{}) error
	FindOne(ret interface{}, sel interface{}) error
	FindAll(ret interface{}, sel interface{}) error
	RemoveAll(sel interface{}) error
	CloseSession() error
	CopySession() Table
}

type Creator interface {
	WithUniques(uniques ...interface{}) Creator
	WithIndexes(indexes ...interface{}) Creator
	New() (Table, error)
}

// ---------------------------------------------------------------------------

