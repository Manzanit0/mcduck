// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: receipts.v1/receipts.proto

package receiptsv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ListReceiptsSince int32

const (
	ListReceiptsSince_LIST_RECEIPTS_SINCE_UNSPECIFIED    ListReceiptsSince = 0
	ListReceiptsSince_LIST_RECEIPTS_SINCE_CURRENT_MONTH  ListReceiptsSince = 1
	ListReceiptsSince_LIST_RECEIPTS_SINCE_PREVIOUS_MONTH ListReceiptsSince = 2
	ListReceiptsSince_LIST_RECEIPTS_SINCE_ALL_TIME       ListReceiptsSince = 3
)

// Enum value maps for ListReceiptsSince.
var (
	ListReceiptsSince_name = map[int32]string{
		0: "LIST_RECEIPTS_SINCE_UNSPECIFIED",
		1: "LIST_RECEIPTS_SINCE_CURRENT_MONTH",
		2: "LIST_RECEIPTS_SINCE_PREVIOUS_MONTH",
		3: "LIST_RECEIPTS_SINCE_ALL_TIME",
	}
	ListReceiptsSince_value = map[string]int32{
		"LIST_RECEIPTS_SINCE_UNSPECIFIED":    0,
		"LIST_RECEIPTS_SINCE_CURRENT_MONTH":  1,
		"LIST_RECEIPTS_SINCE_PREVIOUS_MONTH": 2,
		"LIST_RECEIPTS_SINCE_ALL_TIME":       3,
	}
)

func (x ListReceiptsSince) Enum() *ListReceiptsSince {
	p := new(ListReceiptsSince)
	*p = x
	return p
}

func (x ListReceiptsSince) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ListReceiptsSince) Descriptor() protoreflect.EnumDescriptor {
	return file_receipts_v1_receipts_proto_enumTypes[0].Descriptor()
}

func (ListReceiptsSince) Type() protoreflect.EnumType {
	return &file_receipts_v1_receipts_proto_enumTypes[0]
}

func (x ListReceiptsSince) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ListReceiptsSince.Descriptor instead.
func (ListReceiptsSince) EnumDescriptor() ([]byte, []int) {
	return file_receipts_v1_receipts_proto_rawDescGZIP(), []int{0}
}

type ReceiptStatus int32

const (
	ReceiptStatus_RECEIPT_STATUS_UNSPECIFIED    ReceiptStatus = 0
	ReceiptStatus_RECEIPT_STATUS_PENDING_REVIEW ReceiptStatus = 1
	ReceiptStatus_RECEIPT_STATUS_REVIEWED       ReceiptStatus = 2
)

// Enum value maps for ReceiptStatus.
var (
	ReceiptStatus_name = map[int32]string{
		0: "RECEIPT_STATUS_UNSPECIFIED",
		1: "RECEIPT_STATUS_PENDING_REVIEW",
		2: "RECEIPT_STATUS_REVIEWED",
	}
	ReceiptStatus_value = map[string]int32{
		"RECEIPT_STATUS_UNSPECIFIED":    0,
		"RECEIPT_STATUS_PENDING_REVIEW": 1,
		"RECEIPT_STATUS_REVIEWED":       2,
	}
)

func (x ReceiptStatus) Enum() *ReceiptStatus {
	p := new(ReceiptStatus)
	*p = x
	return p
}

func (x ReceiptStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ReceiptStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_receipts_v1_receipts_proto_enumTypes[1].Descriptor()
}

func (ReceiptStatus) Type() protoreflect.EnumType {
	return &file_receipts_v1_receipts_proto_enumTypes[1]
}

func (x ReceiptStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ReceiptStatus.Descriptor instead.
func (ReceiptStatus) EnumDescriptor() ([]byte, []int) {
	return file_receipts_v1_receipts_proto_rawDescGZIP(), []int{1}
}

type CreateReceiptRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ReceiptFiles [][]byte `protobuf:"bytes,1,rep,name=receipt_files,json=receiptFiles,proto3" json:"receipt_files,omitempty"`
}

func (x *CreateReceiptRequest) Reset() {
	*x = CreateReceiptRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_receipts_v1_receipts_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateReceiptRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateReceiptRequest) ProtoMessage() {}

func (x *CreateReceiptRequest) ProtoReflect() protoreflect.Message {
	mi := &file_receipts_v1_receipts_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateReceiptRequest.ProtoReflect.Descriptor instead.
func (*CreateReceiptRequest) Descriptor() ([]byte, []int) {
	return file_receipts_v1_receipts_proto_rawDescGZIP(), []int{0}
}

func (x *CreateReceiptRequest) GetReceiptFiles() [][]byte {
	if x != nil {
		return x.ReceiptFiles
	}
	return nil
}

type CreateReceiptResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ReceiptIds []uint64 `protobuf:"varint,1,rep,packed,name=receipt_ids,json=receiptIds,proto3" json:"receipt_ids,omitempty"`
}

func (x *CreateReceiptResponse) Reset() {
	*x = CreateReceiptResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_receipts_v1_receipts_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateReceiptResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateReceiptResponse) ProtoMessage() {}

func (x *CreateReceiptResponse) ProtoReflect() protoreflect.Message {
	mi := &file_receipts_v1_receipts_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateReceiptResponse.ProtoReflect.Descriptor instead.
func (*CreateReceiptResponse) Descriptor() ([]byte, []int) {
	return file_receipts_v1_receipts_proto_rawDescGZIP(), []int{1}
}

func (x *CreateReceiptResponse) GetReceiptIds() []uint64 {
	if x != nil {
		return x.ReceiptIds
	}
	return nil
}

type UpdateReceiptRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id            uint64                 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Vendor        *string                `protobuf:"bytes,2,opt,name=vendor,proto3,oneof" json:"vendor,omitempty"`
	PendingReview *bool                  `protobuf:"varint,3,opt,name=pending_review,json=pendingReview,proto3,oneof" json:"pending_review,omitempty"`
	Date          *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=date,proto3,oneof" json:"date,omitempty"`
}

func (x *UpdateReceiptRequest) Reset() {
	*x = UpdateReceiptRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_receipts_v1_receipts_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateReceiptRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateReceiptRequest) ProtoMessage() {}

func (x *UpdateReceiptRequest) ProtoReflect() protoreflect.Message {
	mi := &file_receipts_v1_receipts_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateReceiptRequest.ProtoReflect.Descriptor instead.
func (*UpdateReceiptRequest) Descriptor() ([]byte, []int) {
	return file_receipts_v1_receipts_proto_rawDescGZIP(), []int{2}
}

func (x *UpdateReceiptRequest) GetId() uint64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *UpdateReceiptRequest) GetVendor() string {
	if x != nil && x.Vendor != nil {
		return *x.Vendor
	}
	return ""
}

func (x *UpdateReceiptRequest) GetPendingReview() bool {
	if x != nil && x.PendingReview != nil {
		return *x.PendingReview
	}
	return false
}

func (x *UpdateReceiptRequest) GetDate() *timestamppb.Timestamp {
	if x != nil {
		return x.Date
	}
	return nil
}

type UpdateReceiptResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *UpdateReceiptResponse) Reset() {
	*x = UpdateReceiptResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_receipts_v1_receipts_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateReceiptResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateReceiptResponse) ProtoMessage() {}

func (x *UpdateReceiptResponse) ProtoReflect() protoreflect.Message {
	mi := &file_receipts_v1_receipts_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateReceiptResponse.ProtoReflect.Descriptor instead.
func (*UpdateReceiptResponse) Descriptor() ([]byte, []int) {
	return file_receipts_v1_receipts_proto_rawDescGZIP(), []int{3}
}

type DeleteReceiptRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id uint64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *DeleteReceiptRequest) Reset() {
	*x = DeleteReceiptRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_receipts_v1_receipts_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteReceiptRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteReceiptRequest) ProtoMessage() {}

func (x *DeleteReceiptRequest) ProtoReflect() protoreflect.Message {
	mi := &file_receipts_v1_receipts_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteReceiptRequest.ProtoReflect.Descriptor instead.
func (*DeleteReceiptRequest) Descriptor() ([]byte, []int) {
	return file_receipts_v1_receipts_proto_rawDescGZIP(), []int{4}
}

func (x *DeleteReceiptRequest) GetId() uint64 {
	if x != nil {
		return x.Id
	}
	return 0
}

type DeleteReceiptResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DeleteReceiptResponse) Reset() {
	*x = DeleteReceiptResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_receipts_v1_receipts_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteReceiptResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteReceiptResponse) ProtoMessage() {}

func (x *DeleteReceiptResponse) ProtoReflect() protoreflect.Message {
	mi := &file_receipts_v1_receipts_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteReceiptResponse.ProtoReflect.Descriptor instead.
func (*DeleteReceiptResponse) Descriptor() ([]byte, []int) {
	return file_receipts_v1_receipts_proto_rawDescGZIP(), []int{5}
}

type ListReceiptsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Since  *ListReceiptsSince `protobuf:"varint,1,opt,name=since,proto3,enum=receipts.v1.ListReceiptsSince,oneof" json:"since,omitempty"`
	Status *ReceiptStatus     `protobuf:"varint,2,opt,name=status,proto3,enum=receipts.v1.ReceiptStatus,oneof" json:"status,omitempty"`
}

func (x *ListReceiptsRequest) Reset() {
	*x = ListReceiptsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_receipts_v1_receipts_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListReceiptsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListReceiptsRequest) ProtoMessage() {}

func (x *ListReceiptsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_receipts_v1_receipts_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListReceiptsRequest.ProtoReflect.Descriptor instead.
func (*ListReceiptsRequest) Descriptor() ([]byte, []int) {
	return file_receipts_v1_receipts_proto_rawDescGZIP(), []int{6}
}

func (x *ListReceiptsRequest) GetSince() ListReceiptsSince {
	if x != nil && x.Since != nil {
		return *x.Since
	}
	return ListReceiptsSince_LIST_RECEIPTS_SINCE_UNSPECIFIED
}

func (x *ListReceiptsRequest) GetStatus() ReceiptStatus {
	if x != nil && x.Status != nil {
		return *x.Status
	}
	return ReceiptStatus_RECEIPT_STATUS_UNSPECIFIED
}

type Receipt struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       uint64                 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Status   ReceiptStatus          `protobuf:"varint,2,opt,name=status,proto3,enum=receipts.v1.ReceiptStatus" json:"status,omitempty"`
	Vendor   string                 `protobuf:"bytes,3,opt,name=vendor,proto3" json:"vendor,omitempty"`
	Date     *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=date,proto3" json:"date,omitempty"`
	Expenses []*Expense             `protobuf:"bytes,5,rep,name=expenses,proto3" json:"expenses,omitempty"`
}

func (x *Receipt) Reset() {
	*x = Receipt{}
	if protoimpl.UnsafeEnabled {
		mi := &file_receipts_v1_receipts_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Receipt) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Receipt) ProtoMessage() {}

func (x *Receipt) ProtoReflect() protoreflect.Message {
	mi := &file_receipts_v1_receipts_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Receipt.ProtoReflect.Descriptor instead.
func (*Receipt) Descriptor() ([]byte, []int) {
	return file_receipts_v1_receipts_proto_rawDescGZIP(), []int{7}
}

func (x *Receipt) GetId() uint64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Receipt) GetStatus() ReceiptStatus {
	if x != nil {
		return x.Status
	}
	return ReceiptStatus_RECEIPT_STATUS_UNSPECIFIED
}

func (x *Receipt) GetVendor() string {
	if x != nil {
		return x.Vendor
	}
	return ""
}

func (x *Receipt) GetDate() *timestamppb.Timestamp {
	if x != nil {
		return x.Date
	}
	return nil
}

func (x *Receipt) GetExpenses() []*Expense {
	if x != nil {
		return x.Expenses
	}
	return nil
}

type Expense struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          uint64                 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Date        *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=date,proto3" json:"date,omitempty"`
	Category    string                 `protobuf:"bytes,3,opt,name=category,proto3" json:"category,omitempty"`
	Subcategory string                 `protobuf:"bytes,4,opt,name=subcategory,proto3" json:"subcategory,omitempty"`
	Description string                 `protobuf:"bytes,5,opt,name=description,proto3" json:"description,omitempty"`
	Amount      uint64                 `protobuf:"varint,6,opt,name=amount,proto3" json:"amount,omitempty"`
}

func (x *Expense) Reset() {
	*x = Expense{}
	if protoimpl.UnsafeEnabled {
		mi := &file_receipts_v1_receipts_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Expense) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Expense) ProtoMessage() {}

func (x *Expense) ProtoReflect() protoreflect.Message {
	mi := &file_receipts_v1_receipts_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Expense.ProtoReflect.Descriptor instead.
func (*Expense) Descriptor() ([]byte, []int) {
	return file_receipts_v1_receipts_proto_rawDescGZIP(), []int{8}
}

func (x *Expense) GetId() uint64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Expense) GetDate() *timestamppb.Timestamp {
	if x != nil {
		return x.Date
	}
	return nil
}

func (x *Expense) GetCategory() string {
	if x != nil {
		return x.Category
	}
	return ""
}

func (x *Expense) GetSubcategory() string {
	if x != nil {
		return x.Subcategory
	}
	return ""
}

func (x *Expense) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Expense) GetAmount() uint64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

type ListReceiptsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Receipts []*Receipt `protobuf:"bytes,1,rep,name=receipts,proto3" json:"receipts,omitempty"`
}

func (x *ListReceiptsResponse) Reset() {
	*x = ListReceiptsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_receipts_v1_receipts_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListReceiptsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListReceiptsResponse) ProtoMessage() {}

func (x *ListReceiptsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_receipts_v1_receipts_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListReceiptsResponse.ProtoReflect.Descriptor instead.
func (*ListReceiptsResponse) Descriptor() ([]byte, []int) {
	return file_receipts_v1_receipts_proto_rawDescGZIP(), []int{9}
}

func (x *ListReceiptsResponse) GetReceipts() []*Receipt {
	if x != nil {
		return x.Receipts
	}
	return nil
}

var File_receipts_v1_receipts_proto protoreflect.FileDescriptor

var file_receipts_v1_receipts_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x2f, 0x72, 0x65,
	0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x72, 0x65,
	0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x3b, 0x0a, 0x14, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x23, 0x0a, 0x0d, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x5f, 0x66, 0x69,
	0x6c, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x0c, 0x72, 0x65, 0x63, 0x65, 0x69,
	0x70, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x22, 0x38, 0x0a, 0x15, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x1f, 0x0a, 0x0b, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x5f, 0x69, 0x64, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x04, 0x52, 0x0a, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x49, 0x64,
	0x73, 0x22, 0xcb, 0x01, 0x0a, 0x14, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x63, 0x65,
	0x69, 0x70, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x02, 0x69, 0x64, 0x12, 0x1b, 0x0a, 0x06, 0x76, 0x65,
	0x6e, 0x64, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x06, 0x76, 0x65,
	0x6e, 0x64, 0x6f, 0x72, 0x88, 0x01, 0x01, 0x12, 0x2a, 0x0a, 0x0e, 0x70, 0x65, 0x6e, 0x64, 0x69,
	0x6e, 0x67, 0x5f, 0x72, 0x65, 0x76, 0x69, 0x65, 0x77, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x48,
	0x01, 0x52, 0x0d, 0x70, 0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77,
	0x88, 0x01, 0x01, 0x12, 0x33, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x48, 0x02, 0x52,
	0x04, 0x64, 0x61, 0x74, 0x65, 0x88, 0x01, 0x01, 0x42, 0x09, 0x0a, 0x07, 0x5f, 0x76, 0x65, 0x6e,
	0x64, 0x6f, 0x72, 0x42, 0x11, 0x0a, 0x0f, 0x5f, 0x70, 0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x5f,
	0x72, 0x65, 0x76, 0x69, 0x65, 0x77, 0x42, 0x07, 0x0a, 0x05, 0x5f, 0x64, 0x61, 0x74, 0x65, 0x22,
	0x17, 0x0a, 0x15, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x26, 0x0a, 0x14, 0x44, 0x65, 0x6c, 0x65,
	0x74, 0x65, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x02, 0x69, 0x64,
	0x22, 0x17, 0x0a, 0x15, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70,
	0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x9e, 0x01, 0x0a, 0x13, 0x4c, 0x69,
	0x73, 0x74, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x39, 0x0a, 0x05, 0x73, 0x69, 0x6e, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x1e, 0x2e, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x4c,
	0x69, 0x73, 0x74, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x53, 0x69, 0x6e, 0x63, 0x65,
	0x48, 0x00, 0x52, 0x05, 0x73, 0x69, 0x6e, 0x63, 0x65, 0x88, 0x01, 0x01, 0x12, 0x37, 0x0a, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x72,
	0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x63, 0x65, 0x69,
	0x70, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x48, 0x01, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x88, 0x01, 0x01, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x73, 0x69, 0x6e, 0x63, 0x65, 0x42,
	0x09, 0x0a, 0x07, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0xc7, 0x01, 0x0a, 0x07, 0x52,
	0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x02, 0x69, 0x64, 0x12, 0x32, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74,
	0x73, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x65,
	0x6e, 0x64, 0x6f, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x76, 0x65, 0x6e, 0x64,
	0x6f, 0x72, 0x12, 0x2e, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x04, 0x64, 0x61,
	0x74, 0x65, 0x12, 0x30, 0x0a, 0x08, 0x65, 0x78, 0x70, 0x65, 0x6e, 0x73, 0x65, 0x73, 0x18, 0x05,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x2e,
	0x76, 0x31, 0x2e, 0x45, 0x78, 0x70, 0x65, 0x6e, 0x73, 0x65, 0x52, 0x08, 0x65, 0x78, 0x70, 0x65,
	0x6e, 0x73, 0x65, 0x73, 0x22, 0xc1, 0x01, 0x0a, 0x07, 0x45, 0x78, 0x70, 0x65, 0x6e, 0x73, 0x65,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x02, 0x69, 0x64,
	0x12, 0x2e, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x04, 0x64, 0x61, 0x74, 0x65,
	0x12, 0x1a, 0x0a, 0x08, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x12, 0x20, 0x0a, 0x0b,
	0x73, 0x75, 0x62, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0b, 0x73, 0x75, 0x62, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x12, 0x20,
	0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x16, 0x0a, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x48, 0x0a, 0x14, 0x4c, 0x69, 0x73, 0x74,
	0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x30, 0x0a, 0x08, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x14, 0x2e, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x2e, 0x76, 0x31,
	0x2e, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x52, 0x08, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70,
	0x74, 0x73, 0x2a, 0xa9, 0x01, 0x0a, 0x11, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x63, 0x65, 0x69,
	0x70, 0x74, 0x73, 0x53, 0x69, 0x6e, 0x63, 0x65, 0x12, 0x23, 0x0a, 0x1f, 0x4c, 0x49, 0x53, 0x54,
	0x5f, 0x52, 0x45, 0x43, 0x45, 0x49, 0x50, 0x54, 0x53, 0x5f, 0x53, 0x49, 0x4e, 0x43, 0x45, 0x5f,
	0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x25, 0x0a,
	0x21, 0x4c, 0x49, 0x53, 0x54, 0x5f, 0x52, 0x45, 0x43, 0x45, 0x49, 0x50, 0x54, 0x53, 0x5f, 0x53,
	0x49, 0x4e, 0x43, 0x45, 0x5f, 0x43, 0x55, 0x52, 0x52, 0x45, 0x4e, 0x54, 0x5f, 0x4d, 0x4f, 0x4e,
	0x54, 0x48, 0x10, 0x01, 0x12, 0x26, 0x0a, 0x22, 0x4c, 0x49, 0x53, 0x54, 0x5f, 0x52, 0x45, 0x43,
	0x45, 0x49, 0x50, 0x54, 0x53, 0x5f, 0x53, 0x49, 0x4e, 0x43, 0x45, 0x5f, 0x50, 0x52, 0x45, 0x56,
	0x49, 0x4f, 0x55, 0x53, 0x5f, 0x4d, 0x4f, 0x4e, 0x54, 0x48, 0x10, 0x02, 0x12, 0x20, 0x0a, 0x1c,
	0x4c, 0x49, 0x53, 0x54, 0x5f, 0x52, 0x45, 0x43, 0x45, 0x49, 0x50, 0x54, 0x53, 0x5f, 0x53, 0x49,
	0x4e, 0x43, 0x45, 0x5f, 0x41, 0x4c, 0x4c, 0x5f, 0x54, 0x49, 0x4d, 0x45, 0x10, 0x03, 0x2a, 0x6f,
	0x0a, 0x0d, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12,
	0x1e, 0x0a, 0x1a, 0x52, 0x45, 0x43, 0x45, 0x49, 0x50, 0x54, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55,
	0x53, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12,
	0x21, 0x0a, 0x1d, 0x52, 0x45, 0x43, 0x45, 0x49, 0x50, 0x54, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55,
	0x53, 0x5f, 0x50, 0x45, 0x4e, 0x44, 0x49, 0x4e, 0x47, 0x5f, 0x52, 0x45, 0x56, 0x49, 0x45, 0x57,
	0x10, 0x01, 0x12, 0x1b, 0x0a, 0x17, 0x52, 0x45, 0x43, 0x45, 0x49, 0x50, 0x54, 0x5f, 0x53, 0x54,
	0x41, 0x54, 0x55, 0x53, 0x5f, 0x52, 0x45, 0x56, 0x49, 0x45, 0x57, 0x45, 0x44, 0x10, 0x02, 0x32,
	0xf6, 0x02, 0x0a, 0x0f, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x58, 0x0a, 0x0d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x63,
	0x65, 0x69, 0x70, 0x74, 0x12, 0x21, 0x2e, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x2e,
	0x76, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x22, 0x2e, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70,
	0x74, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x63, 0x65,
	0x69, 0x70, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x58, 0x0a,
	0x0d, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x12, 0x21,
	0x2e, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x22, 0x2e, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x2e,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x58, 0x0a, 0x0d, 0x44, 0x65, 0x6c, 0x65, 0x74,
	0x65, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x12, 0x21, 0x2e, 0x72, 0x65, 0x63, 0x65, 0x69,
	0x70, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x63,
	0x65, 0x69, 0x70, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x22, 0x2e, 0x72, 0x65,
	0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65,
	0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x12, 0x55, 0x0a, 0x0c, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74,
	0x73, 0x12, 0x20, 0x2e, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x2e,
	0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x21, 0x2e, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x2e, 0x76,
	0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0xa5, 0x01, 0x0a, 0x0f, 0x63, 0x6f, 0x6d,
	0x2e, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x42, 0x0d, 0x52, 0x65,
	0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x36, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x61, 0x6e, 0x7a, 0x61, 0x6e,
	0x69, 0x74, 0x30, 0x2f, 0x6d, 0x63, 0x64, 0x75, 0x63, 0x6b, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x72,
	0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x3b, 0x72, 0x65, 0x63, 0x65, 0x69,
	0x70, 0x74, 0x73, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x52, 0x58, 0x58, 0xaa, 0x02, 0x0b, 0x52, 0x65,
	0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x0b, 0x52, 0x65, 0x63, 0x65,
	0x69, 0x70, 0x74, 0x73, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x17, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70,
	0x74, 0x73, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0xea, 0x02, 0x0c, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x3a, 0x3a, 0x56, 0x31,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_receipts_v1_receipts_proto_rawDescOnce sync.Once
	file_receipts_v1_receipts_proto_rawDescData = file_receipts_v1_receipts_proto_rawDesc
)

func file_receipts_v1_receipts_proto_rawDescGZIP() []byte {
	file_receipts_v1_receipts_proto_rawDescOnce.Do(func() {
		file_receipts_v1_receipts_proto_rawDescData = protoimpl.X.CompressGZIP(file_receipts_v1_receipts_proto_rawDescData)
	})
	return file_receipts_v1_receipts_proto_rawDescData
}

var file_receipts_v1_receipts_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_receipts_v1_receipts_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_receipts_v1_receipts_proto_goTypes = []any{
	(ListReceiptsSince)(0),        // 0: receipts.v1.ListReceiptsSince
	(ReceiptStatus)(0),            // 1: receipts.v1.ReceiptStatus
	(*CreateReceiptRequest)(nil),  // 2: receipts.v1.CreateReceiptRequest
	(*CreateReceiptResponse)(nil), // 3: receipts.v1.CreateReceiptResponse
	(*UpdateReceiptRequest)(nil),  // 4: receipts.v1.UpdateReceiptRequest
	(*UpdateReceiptResponse)(nil), // 5: receipts.v1.UpdateReceiptResponse
	(*DeleteReceiptRequest)(nil),  // 6: receipts.v1.DeleteReceiptRequest
	(*DeleteReceiptResponse)(nil), // 7: receipts.v1.DeleteReceiptResponse
	(*ListReceiptsRequest)(nil),   // 8: receipts.v1.ListReceiptsRequest
	(*Receipt)(nil),               // 9: receipts.v1.Receipt
	(*Expense)(nil),               // 10: receipts.v1.Expense
	(*ListReceiptsResponse)(nil),  // 11: receipts.v1.ListReceiptsResponse
	(*timestamppb.Timestamp)(nil), // 12: google.protobuf.Timestamp
}
var file_receipts_v1_receipts_proto_depIdxs = []int32{
	12, // 0: receipts.v1.UpdateReceiptRequest.date:type_name -> google.protobuf.Timestamp
	0,  // 1: receipts.v1.ListReceiptsRequest.since:type_name -> receipts.v1.ListReceiptsSince
	1,  // 2: receipts.v1.ListReceiptsRequest.status:type_name -> receipts.v1.ReceiptStatus
	1,  // 3: receipts.v1.Receipt.status:type_name -> receipts.v1.ReceiptStatus
	12, // 4: receipts.v1.Receipt.date:type_name -> google.protobuf.Timestamp
	10, // 5: receipts.v1.Receipt.expenses:type_name -> receipts.v1.Expense
	12, // 6: receipts.v1.Expense.date:type_name -> google.protobuf.Timestamp
	9,  // 7: receipts.v1.ListReceiptsResponse.receipts:type_name -> receipts.v1.Receipt
	2,  // 8: receipts.v1.ReceiptsService.CreateReceipt:input_type -> receipts.v1.CreateReceiptRequest
	4,  // 9: receipts.v1.ReceiptsService.UpdateReceipt:input_type -> receipts.v1.UpdateReceiptRequest
	6,  // 10: receipts.v1.ReceiptsService.DeleteReceipt:input_type -> receipts.v1.DeleteReceiptRequest
	8,  // 11: receipts.v1.ReceiptsService.ListReceipts:input_type -> receipts.v1.ListReceiptsRequest
	3,  // 12: receipts.v1.ReceiptsService.CreateReceipt:output_type -> receipts.v1.CreateReceiptResponse
	5,  // 13: receipts.v1.ReceiptsService.UpdateReceipt:output_type -> receipts.v1.UpdateReceiptResponse
	7,  // 14: receipts.v1.ReceiptsService.DeleteReceipt:output_type -> receipts.v1.DeleteReceiptResponse
	11, // 15: receipts.v1.ReceiptsService.ListReceipts:output_type -> receipts.v1.ListReceiptsResponse
	12, // [12:16] is the sub-list for method output_type
	8,  // [8:12] is the sub-list for method input_type
	8,  // [8:8] is the sub-list for extension type_name
	8,  // [8:8] is the sub-list for extension extendee
	0,  // [0:8] is the sub-list for field type_name
}

func init() { file_receipts_v1_receipts_proto_init() }
func file_receipts_v1_receipts_proto_init() {
	if File_receipts_v1_receipts_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_receipts_v1_receipts_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*CreateReceiptRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_receipts_v1_receipts_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*CreateReceiptResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_receipts_v1_receipts_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*UpdateReceiptRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_receipts_v1_receipts_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*UpdateReceiptResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_receipts_v1_receipts_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*DeleteReceiptRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_receipts_v1_receipts_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*DeleteReceiptResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_receipts_v1_receipts_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*ListReceiptsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_receipts_v1_receipts_proto_msgTypes[7].Exporter = func(v any, i int) any {
			switch v := v.(*Receipt); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_receipts_v1_receipts_proto_msgTypes[8].Exporter = func(v any, i int) any {
			switch v := v.(*Expense); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_receipts_v1_receipts_proto_msgTypes[9].Exporter = func(v any, i int) any {
			switch v := v.(*ListReceiptsResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_receipts_v1_receipts_proto_msgTypes[2].OneofWrappers = []any{}
	file_receipts_v1_receipts_proto_msgTypes[6].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_receipts_v1_receipts_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_receipts_v1_receipts_proto_goTypes,
		DependencyIndexes: file_receipts_v1_receipts_proto_depIdxs,
		EnumInfos:         file_receipts_v1_receipts_proto_enumTypes,
		MessageInfos:      file_receipts_v1_receipts_proto_msgTypes,
	}.Build()
	File_receipts_v1_receipts_proto = out.File
	file_receipts_v1_receipts_proto_rawDesc = nil
	file_receipts_v1_receipts_proto_goTypes = nil
	file_receipts_v1_receipts_proto_depIdxs = nil
}
