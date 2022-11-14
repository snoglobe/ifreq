package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"strconv"
)

type RequestType byte

const (
	RetrieveInfo RequestType = iota
	RequestStream
	RequestPeers
	GoingOffAir
	GoingOnAir
)

type Request struct {
	reqType RequestType
	payload any
}

type Response struct {
	payload any
	success bool
}

func NewRequest(reqType RequestType, payload any) *Request {
	return &Request{reqType, payload}
}

func (r *Request) Type() RequestType {
	return r.reqType
}

func (r *Request) Payload() any {
	return r.payload
}

func (r *Request) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":    r.reqType,
		"payload": r.payload,
	})
}

func (r *Request) UnmarshalJSON(data []byte) error {
	var req map[string]any
	if err := json.Unmarshal(data, &req); err != nil {
		return err
	}

	r.reqType = RequestType(req["type"].(float64))
	r.payload = req["payload"]

	return nil
}

func NewResponse(payload any, success bool) *Response {
	return &Response{payload, success}
}

func (r *Response) Payload() any {
	return r.payload
}

func (r *Response) Success() bool {
	return r.success
}

func (r *Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"payload": r.payload,
		"success": r.success,
	})
}

func (r *Response) UnmarshalJSON(data []byte) error {
	var resp map[string]any
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	r.payload = resp["payload"]
	r.success = resp["success"].(bool)

	return nil
}

func SendRequest(req *Request, ip net.IP, port uint16) *Response {
	log.Printf("sending request of type %d with payload %v to %s:%d", req.Type(), req.payload, ip, port)
	reqData, _ := json.Marshal(req)

	conn, err := net.Dial("tcp", ip.String()+":"+strconv.Itoa(int(port)))
	if err != nil {
		return NewResponse(err.Error(), false)
	}
	defer conn.Close()

	_, err = conn.Write(reqData)
	if err != nil {
		return NewResponse(err.Error(), false)
	}

	respData := bytes.Buffer{}
	_, err = respData.ReadFrom(conn)
	if err != nil {
		return NewResponse(err.Error(), false)
	}

	println(respData.String())

	resp := &Response{}
	err = json.Unmarshal(respData.Bytes(), resp)
	if err != nil {
		return NewResponse(err.Error(), false)
	}

	return resp
}
