package pingpong

import (
	"testapp/internal/service"
)

// myRepo lists the MyRepo methods this package calls. goro created this file once and
// never changes it, so the interface is yours to edit.
type myRepo interface {
	service.Transactor
	// TODO: add the MyRepo methods you need
}
