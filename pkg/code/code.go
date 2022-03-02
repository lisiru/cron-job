package code

import (
	"github.com/marmotedu/errors"
	"github.com/novalagung/gubrak"
	"net/http"
)

type ErrCode struct {
	C int
	HTTP int
	Ext string
	Ref string
}

func (e ErrCode) HTTPStatus() int {
	if e.HTTP ==0 {
		return http.StatusInternalServerError
	}
	return e.HTTP
}

func (e ErrCode) String() string {
	return e.Ext
}

func (e ErrCode) Reference() string {
	return e.Ref
}

func (e ErrCode) Code() int {
	return e.C
}



var _ errors.Coder = &ErrCode{}

func register(code int,httpStatus int,message string)  {
	found,_:=gubrak.Includes([]int{200,400,401,403,404,500},httpStatus)
	if !found {
		panic("http code not in `200,400,401,403,404,500`")
	}
	coder:=&ErrCode{
		C: code,
		HTTP: httpStatus,
		Ext: message,
	}
	errors.MustRegister(coder)
}
