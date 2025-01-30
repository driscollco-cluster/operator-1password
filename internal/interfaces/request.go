package interfaces

import router "github.com/driscollco-core/http-router"

//go:generate go run go.uber.org/mock/mockgen -destination=../mocks/mock-request.go -package=mocks . Request
type Request interface {
	router.Request
}
