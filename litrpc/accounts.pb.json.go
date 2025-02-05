// Code generated by falafel 0.9.1. DO NOT EDIT.
// source: lit-accounts.proto

package litrpc

import (
	"context"

	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

func RegisterAccountsJSONCallbacks(registry map[string]func(ctx context.Context,
	conn *grpc.ClientConn, reqJSON string, callback func(string, error))) {

	marshaler := &gateway.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		},
	}

	registry["litrpc.Accounts.CreateAccount"] = func(ctx context.Context,
		conn *grpc.ClientConn, reqJSON string, callback func(string, error)) {

		req := &CreateAccountRequest{}
		err := marshaler.Unmarshal([]byte(reqJSON), req)
		if err != nil {
			callback("", err)
			return
		}

		client := NewAccountsClient(conn)
		resp, err := client.CreateAccount(ctx, req)
		if err != nil {
			callback("", err)
			return
		}

		respBytes, err := marshaler.Marshal(resp)
		if err != nil {
			callback("", err)
			return
		}
		callback(string(respBytes), nil)
	}

	registry["litrpc.Accounts.UpdateAccount"] = func(ctx context.Context,
		conn *grpc.ClientConn, reqJSON string, callback func(string, error)) {

		req := &UpdateAccountRequest{}
		err := marshaler.Unmarshal([]byte(reqJSON), req)
		if err != nil {
			callback("", err)
			return
		}

		client := NewAccountsClient(conn)
		resp, err := client.UpdateAccount(ctx, req)
		if err != nil {
			callback("", err)
			return
		}

		respBytes, err := marshaler.Marshal(resp)
		if err != nil {
			callback("", err)
			return
		}
		callback(string(respBytes), nil)
	}

	registry["litrpc.Accounts.UpdateBalance"] = func(ctx context.Context,
		conn *grpc.ClientConn, reqJSON string, callback func(string, error)) {

		req := &UpdateAccountBalanceRequest{}
		err := marshaler.Unmarshal([]byte(reqJSON), req)
		if err != nil {
			callback("", err)
			return
		}

		client := NewAccountsClient(conn)
		resp, err := client.UpdateBalance(ctx, req)
		if err != nil {
			callback("", err)
			return
		}

		respBytes, err := marshaler.Marshal(resp)
		if err != nil {
			callback("", err)
			return
		}
		callback(string(respBytes), nil)
	}

	registry["litrpc.Accounts.ListAccounts"] = func(ctx context.Context,
		conn *grpc.ClientConn, reqJSON string, callback func(string, error)) {

		req := &ListAccountsRequest{}
		err := marshaler.Unmarshal([]byte(reqJSON), req)
		if err != nil {
			callback("", err)
			return
		}

		client := NewAccountsClient(conn)
		resp, err := client.ListAccounts(ctx, req)
		if err != nil {
			callback("", err)
			return
		}

		respBytes, err := marshaler.Marshal(resp)
		if err != nil {
			callback("", err)
			return
		}
		callback(string(respBytes), nil)
	}

	registry["litrpc.Accounts.AccountInfo"] = func(ctx context.Context,
		conn *grpc.ClientConn, reqJSON string, callback func(string, error)) {

		req := &AccountInfoRequest{}
		err := marshaler.Unmarshal([]byte(reqJSON), req)
		if err != nil {
			callback("", err)
			return
		}

		client := NewAccountsClient(conn)
		resp, err := client.AccountInfo(ctx, req)
		if err != nil {
			callback("", err)
			return
		}

		respBytes, err := marshaler.Marshal(resp)
		if err != nil {
			callback("", err)
			return
		}
		callback(string(respBytes), nil)
	}

	registry["litrpc.Accounts.RemoveAccount"] = func(ctx context.Context,
		conn *grpc.ClientConn, reqJSON string, callback func(string, error)) {

		req := &RemoveAccountRequest{}
		err := marshaler.Unmarshal([]byte(reqJSON), req)
		if err != nil {
			callback("", err)
			return
		}

		client := NewAccountsClient(conn)
		resp, err := client.RemoveAccount(ctx, req)
		if err != nil {
			callback("", err)
			return
		}

		respBytes, err := marshaler.Marshal(resp)
		if err != nil {
			callback("", err)
			return
		}
		callback(string(respBytes), nil)
	}
}
