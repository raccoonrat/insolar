package contract

import (
	"github.com/insolar/insolar"
)

type Ledger interface {
	GetObject(GetObjectArgs) (ObjectResult, error)
}

type GetObjectArgs interface {
	Object() insolar.Reference
	IsValid() bool
}

type ObjectResult interface {
	Reference() insolar.Reference
	Memory() []byte
	StateID() insolar.ID
}

type RefIterator interface {
	Next() (insolar.Reference, error)
	HasNext() bool
}

// StateID returns reference to object state record.
StateID() *RecordID

// Memory fetches object memory from storage.
Memory() []byte

// IsPrototype determines if the object is a prototype.
IsPrototype() bool

// Code returns code reference.
Code() (*RecordRef, error)

// Prototype returns prototype reference.
Prototype() (*RecordRef, error)

// Children returns object's children references.
Children(pulse *PulseNumber) (RefIterator, error)

// ChildPointer returns the latest child for this object.
ChildPointer() *RecordID

// Parent returns object's parent.
Parent() *RecordRef
