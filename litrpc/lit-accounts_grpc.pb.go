// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package litrpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// AccountsClient is the client API for Accounts service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AccountsClient interface {
	// litcli: `accounts create`
	// CreateAccount adds an entry to the account database. This entry represents
	// an amount of satoshis (account balance) that can be spent using off-chain
	// transactions (e.g. paying invoices).
	//
	// Macaroons can be created to be locked to an account. This makes sure that
	// the bearer of the macaroon can only spend at most that amount of satoshis
	// through the daemon that has issued the macaroon.
	//
	// Accounts only assert a maximum amount spendable. Having a certain account
	// balance does not guarantee that the node has the channel liquidity to
	// actually spend that amount.
	CreateAccount(ctx context.Context, in *CreateAccountRequest, opts ...grpc.CallOption) (*CreateAccountResponse, error)
	// litcli: `accounts update`
	// UpdateAccount updates an existing account in the account database.
	UpdateAccount(ctx context.Context, in *UpdateAccountRequest, opts ...grpc.CallOption) (*Account, error)
	// litcli: `accounts updatebalance`
	// UpdateBalance adds or deducts an amount from an existing account in the
	// account database.
	UpdateBalance(ctx context.Context, in *UpdateAccountBalanceRequest, opts ...grpc.CallOption) (*Account, error)
	// litcli: `accounts list`
	// ListAccounts returns all accounts that are currently stored in the account
	// database.
	ListAccounts(ctx context.Context, in *ListAccountsRequest, opts ...grpc.CallOption) (*ListAccountsResponse, error)
	// litcli: `accounts info`
	// AccountInfo returns the account with the given ID or label.
	AccountInfo(ctx context.Context, in *AccountInfoRequest, opts ...grpc.CallOption) (*Account, error)
	// litcli: `accounts remove`
	// RemoveAccount removes the given account from the account database.
	RemoveAccount(ctx context.Context, in *RemoveAccountRequest, opts ...grpc.CallOption) (*RemoveAccountResponse, error)
}

type accountsClient struct {
	cc grpc.ClientConnInterface
}

func NewAccountsClient(cc grpc.ClientConnInterface) AccountsClient {
	return &accountsClient{cc}
}

func (c *accountsClient) CreateAccount(ctx context.Context, in *CreateAccountRequest, opts ...grpc.CallOption) (*CreateAccountResponse, error) {
	out := new(CreateAccountResponse)
	err := c.cc.Invoke(ctx, "/litrpc.Accounts/CreateAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountsClient) UpdateAccount(ctx context.Context, in *UpdateAccountRequest, opts ...grpc.CallOption) (*Account, error) {
	out := new(Account)
	err := c.cc.Invoke(ctx, "/litrpc.Accounts/UpdateAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountsClient) UpdateBalance(ctx context.Context, in *UpdateAccountBalanceRequest, opts ...grpc.CallOption) (*Account, error) {
	out := new(Account)
	err := c.cc.Invoke(ctx, "/litrpc.Accounts/UpdateBalance", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountsClient) ListAccounts(ctx context.Context, in *ListAccountsRequest, opts ...grpc.CallOption) (*ListAccountsResponse, error) {
	out := new(ListAccountsResponse)
	err := c.cc.Invoke(ctx, "/litrpc.Accounts/ListAccounts", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountsClient) AccountInfo(ctx context.Context, in *AccountInfoRequest, opts ...grpc.CallOption) (*Account, error) {
	out := new(Account)
	err := c.cc.Invoke(ctx, "/litrpc.Accounts/AccountInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountsClient) RemoveAccount(ctx context.Context, in *RemoveAccountRequest, opts ...grpc.CallOption) (*RemoveAccountResponse, error) {
	out := new(RemoveAccountResponse)
	err := c.cc.Invoke(ctx, "/litrpc.Accounts/RemoveAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AccountsServer is the server API for Accounts service.
// All implementations must embed UnimplementedAccountsServer
// for forward compatibility
type AccountsServer interface {
	// litcli: `accounts create`
	// CreateAccount adds an entry to the account database. This entry represents
	// an amount of satoshis (account balance) that can be spent using off-chain
	// transactions (e.g. paying invoices).
	//
	// Macaroons can be created to be locked to an account. This makes sure that
	// the bearer of the macaroon can only spend at most that amount of satoshis
	// through the daemon that has issued the macaroon.
	//
	// Accounts only assert a maximum amount spendable. Having a certain account
	// balance does not guarantee that the node has the channel liquidity to
	// actually spend that amount.
	CreateAccount(context.Context, *CreateAccountRequest) (*CreateAccountResponse, error)
	// litcli: `accounts update`
	// UpdateAccount updates an existing account in the account database.
	UpdateAccount(context.Context, *UpdateAccountRequest) (*Account, error)
	// litcli: `accounts updatebalance`
	// UpdateBalance adds or deducts an amount from an existing account in the
	// account database.
	UpdateBalance(context.Context, *UpdateAccountBalanceRequest) (*Account, error)
	// litcli: `accounts list`
	// ListAccounts returns all accounts that are currently stored in the account
	// database.
	ListAccounts(context.Context, *ListAccountsRequest) (*ListAccountsResponse, error)
	// litcli: `accounts info`
	// AccountInfo returns the account with the given ID or label.
	AccountInfo(context.Context, *AccountInfoRequest) (*Account, error)
	// litcli: `accounts remove`
	// RemoveAccount removes the given account from the account database.
	RemoveAccount(context.Context, *RemoveAccountRequest) (*RemoveAccountResponse, error)
	mustEmbedUnimplementedAccountsServer()
}

// UnimplementedAccountsServer must be embedded to have forward compatible implementations.
type UnimplementedAccountsServer struct {
}

func (UnimplementedAccountsServer) CreateAccount(context.Context, *CreateAccountRequest) (*CreateAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAccount not implemented")
}
func (UnimplementedAccountsServer) UpdateAccount(context.Context, *UpdateAccountRequest) (*Account, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateAccount not implemented")
}
func (UnimplementedAccountsServer) UpdateBalance(context.Context, *UpdateAccountBalanceRequest) (*Account, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateBalance not implemented")
}
func (UnimplementedAccountsServer) ListAccounts(context.Context, *ListAccountsRequest) (*ListAccountsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAccounts not implemented")
}
func (UnimplementedAccountsServer) AccountInfo(context.Context, *AccountInfoRequest) (*Account, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AccountInfo not implemented")
}
func (UnimplementedAccountsServer) RemoveAccount(context.Context, *RemoveAccountRequest) (*RemoveAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveAccount not implemented")
}
func (UnimplementedAccountsServer) mustEmbedUnimplementedAccountsServer() {}

// UnsafeAccountsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AccountsServer will
// result in compilation errors.
type UnsafeAccountsServer interface {
	mustEmbedUnimplementedAccountsServer()
}

func RegisterAccountsServer(s grpc.ServiceRegistrar, srv AccountsServer) {
	s.RegisterService(&Accounts_ServiceDesc, srv)
}

func _Accounts_CreateAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountsServer).CreateAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/litrpc.Accounts/CreateAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountsServer).CreateAccount(ctx, req.(*CreateAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Accounts_UpdateAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountsServer).UpdateAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/litrpc.Accounts/UpdateAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountsServer).UpdateAccount(ctx, req.(*UpdateAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Accounts_UpdateBalance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateAccountBalanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountsServer).UpdateBalance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/litrpc.Accounts/UpdateBalance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountsServer).UpdateBalance(ctx, req.(*UpdateAccountBalanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Accounts_ListAccounts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListAccountsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountsServer).ListAccounts(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/litrpc.Accounts/ListAccounts",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountsServer).ListAccounts(ctx, req.(*ListAccountsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Accounts_AccountInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AccountInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountsServer).AccountInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/litrpc.Accounts/AccountInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountsServer).AccountInfo(ctx, req.(*AccountInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Accounts_RemoveAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountsServer).RemoveAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/litrpc.Accounts/RemoveAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountsServer).RemoveAccount(ctx, req.(*RemoveAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Accounts_ServiceDesc is the grpc.ServiceDesc for Accounts service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Accounts_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "litrpc.Accounts",
	HandlerType: (*AccountsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateAccount",
			Handler:    _Accounts_CreateAccount_Handler,
		},
		{
			MethodName: "UpdateAccount",
			Handler:    _Accounts_UpdateAccount_Handler,
		},
		{
			MethodName: "UpdateBalance",
			Handler:    _Accounts_UpdateBalance_Handler,
		},
		{
			MethodName: "ListAccounts",
			Handler:    _Accounts_ListAccounts_Handler,
		},
		{
			MethodName: "AccountInfo",
			Handler:    _Accounts_AccountInfo_Handler,
		},
		{
			MethodName: "RemoveAccount",
			Handler:    _Accounts_RemoveAccount_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "lit-accounts.proto",
}
