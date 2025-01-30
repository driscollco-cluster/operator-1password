package interfaces

import (
	"github.com/driscollco-core/log"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../mocks/mock-log.go -package=mocks . Log
type Log interface {
	log.Log
}
