package fsm

type ApplyRequest struct {
	Command func() interface{}
}

type ApplyResponse struct {
	Data  interface{}
	Error error
}
