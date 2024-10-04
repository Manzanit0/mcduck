// @generated by protoc-gen-connect-es v1.5.0 with parameter "target=ts,import_extension=.ts"
// @generated from file auth.v1/auth.proto (package auth.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { ConnectTelegramRequest, ConnectTelegramResponse, LoginRequest, LoginResponse, RegisterRequest, RegisterResponse } from "./auth_pb.ts";
import { MethodKind } from "@bufbuild/protobuf";

/**
 * @generated from service auth.v1.AuthService
 */
export const AuthService = {
  typeName: "auth.v1.AuthService",
  methods: {
    /**
     * @generated from rpc auth.v1.AuthService.Register
     */
    register: {
      name: "Register",
      I: RegisterRequest,
      O: RegisterResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc auth.v1.AuthService.Login
     */
    login: {
      name: "Login",
      I: LoginRequest,
      O: LoginResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc auth.v1.AuthService.ConnectTelegram
     */
    connectTelegram: {
      name: "ConnectTelegram",
      I: ConnectTelegramRequest,
      O: ConnectTelegramResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;

