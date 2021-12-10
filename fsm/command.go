package fsm

type (
	Func         func() interface{}
	ApplyRequest struct {
		Command Func
	}
)

type ApplyResponse struct {
	Data  interface{}
	Error error
}
