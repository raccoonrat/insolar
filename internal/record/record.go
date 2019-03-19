package record

import (
	"github.com/insolar/insolar"
)

// Standalone ==========================================================================================================

type Request interface {
	Reason() insolar.Reference

	Object() insolar.ID
	Payload() []byte
}

type Result interface {
	Reason() insolar.Reference

	Request() insolar.ID
	Payload() []byte
}

type Code interface {
	Reason() insolar.Reference

	CodeBlob() insolar.ID
}

// Object ==============================================================================================================

type Activate interface {
	Reason() insolar.Reference

	MemoryBlob() insolar.ID
	Image() insolar.Reference
	Parent() insolar.Reference
}

type Amend interface {
	Reason() insolar.Reference

	MemoryBlob() []byte
}

type Deactivate interface {
	Reason() insolar.Reference
}

// Utility =============================================================================================================\

type Blob interface {
	Reason() insolar.Reference

	Data() []byte
}

type Child interface {
	Reason() insolar.Reference

	Reference() insolar.Reference
}
