/*
** description("").
** copyright('open-im,www.open-im.io').
** author("fg,Gordon@tuoyun.net").
** time(2021/9/8 14:54).
 */
package ws_local_server

import (
	"encoding/json"
	"open_im_sdk/pkg/constant"
	"reflect"
	"runtime"
)

type Req struct {
	ReqFuncName string `json:"reqFuncName" `
	OperationID string `json:"operationID"`
	Data        string `json:"data"`
	UserID      string `json:"userID"`
}

func (ws *WServer) DoLogin(m Req, conn *UserConn) {
	UserRouteRwLock.RLock()
	defer UserRouteRwLock.RUnlock()
	urm, ok := UserRouteMap[m.UserID]
	if !ok {
		wrapSdkLog("","login", "user first login: ", m)
		refR := GenUserRouterNoLock(m.UserID)
		params := []reflect.Value{reflect.ValueOf(m.Data), reflect.ValueOf(m.OperationID)}
		vf, ok := (*refR.refName)[m.ReqFuncName]
		if ok {
			vf.Call(params)
		} else {
			wrapSdkLog("","login", "no func name: ", m.ReqFuncName, m)
			SendOneConnMessage(EventData{m.ReqFuncName, StatusBadParameter, StatusText(StatusBadParameter), "", m.OperationID}, conn)
		}

	} else {
		if urm.wsRouter.getMyLoginStatus() == constant.LoginSuccess {
			//send ok
			SendOneConnMessage(EventData{"Login", 0, "ok", "", m.OperationID}, conn)
		} else {
			wrapSdkLog("", "login", "login status pending, try after 5 second ", urm.wsRouter.getMyLoginStatus(), m.UserID)
			SendOneConnMessage(EventData{"Login", StatusLoginPending, StatusText(StatusLoginPending), "", m.OperationID}, conn)
		}
	}
}

func (ws *WServer) msgParse(conn *UserConn, jsonMsg []byte) {
	m := Req{}
	if err := json.Unmarshal(jsonMsg, &m); err != nil {
		SendOneConnMessage(EventData{"error", 100, "Unmarshal failed", "", ""}, conn)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			SendOneConnMessage(EventData{m.ReqFuncName, StatusBadParameter, StatusText(StatusBadParameter), "", m.OperationID}, conn)
			wrapSdkLog("","msgParse", "bad request, panic is ", r)
			buf := make([]byte, 1<<16)
			runtime.Stack(buf, true)
			wrapSdkLog("","msgParse", "call", string(buf))
		}
	}()

	wrapSdkLog("","msgParse","recv request from web: ", "reqFuncName ", m.ReqFuncName, "data ", m.Data, "recv jsonMsg: ", string(jsonMsg))

	if m.ReqFuncName == "Login" {
		//	rwLock.Lock()
		ws.DoLogin(m, conn)
		wrapSdkLog("","msgParse", m)
		//	rwLock.Unlock()
		return
	}

	UserRouteRwLock.RLock()
	defer UserRouteRwLock.RUnlock()
	//	rwLock.RLock()
	//	defer rwLock.RUnlock()
	urm, ok := UserRouteMap[m.UserID]
	if !ok {
		wrapSdkLog("","msgParse", "user not login failed, must login first: ", m.UserID)
		SendOneConnMessage(EventData{"Login", StatusNoLogin, StatusText(StatusNoLogin), "", m.OperationID}, conn)
		return
	}
	parms := []reflect.Value{reflect.ValueOf(m.Data), reflect.ValueOf(m.OperationID)}
	vf, ok := (*urm.refName)[m.ReqFuncName]
	if ok {
		vf.Call(parms)
	} else {
		wrapSdkLog("","msgParse", "no func ", m.ReqFuncName)
		SendOneConnMessage(EventData{m.ReqFuncName, StatusBadParameter, StatusText(StatusBadParameter), "", m.OperationID}, conn)
	}

}
