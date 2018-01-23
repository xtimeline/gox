package httpx

// https://github.com/sony/gobreaker
type HttpBreaker interface {
	Allow() (done func(success bool), err error)
}
