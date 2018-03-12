// Code generated by protoc-gen-go. DO NOT EDIT.
// source: proto/metrics.proto

/*
Package metrics is a generated protocol buffer package.

It is generated from these files:
	proto/metrics.proto

It has these top-level messages:
	MetricsDeliveryResponse
	Machine
	Client
	Protocol
	Entity
	Connection
	NodeEntity
	Metrics
*/
package metrics

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Entity_EntityType int32

const (
	Entity_UNKNOWN    Entity_EntityType = 0
	Entity_BOARD      Entity_EntityType = 1
	Entity_THREAD     Entity_EntityType = 2
	Entity_POST       Entity_EntityType = 3
	Entity_VOTE       Entity_EntityType = 4
	Entity_KEY        Entity_EntityType = 5
	Entity_TRUSTSTATE Entity_EntityType = 6
	Entity_ADDRESS    Entity_EntityType = 7
)

var Entity_EntityType_name = map[int32]string{
	0: "UNKNOWN",
	1: "BOARD",
	2: "THREAD",
	3: "POST",
	4: "VOTE",
	5: "KEY",
	6: "TRUSTSTATE",
	7: "ADDRESS",
}
var Entity_EntityType_value = map[string]int32{
	"UNKNOWN":    0,
	"BOARD":      1,
	"THREAD":     2,
	"POST":       3,
	"VOTE":       4,
	"KEY":        5,
	"TRUSTSTATE": 6,
	"ADDRESS":    7,
}

func (x Entity_EntityType) String() string {
	return proto.EnumName(Entity_EntityType_name, int32(x))
}
func (Entity_EntityType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{4, 0} }

type Connection_Direction int32

const (
	Connection_INBOUND  Connection_Direction = 0
	Connection_OUTBOUND Connection_Direction = 1
)

var Connection_Direction_name = map[int32]string{
	0: "INBOUND",
	1: "OUTBOUND",
}
var Connection_Direction_value = map[string]int32{
	"INBOUND":  0,
	"OUTBOUND": 1,
}

func (x Connection_Direction) String() string {
	return proto.EnumName(Connection_Direction_name, int32(x))
}
func (Connection_Direction) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{5, 0} }

// Empty responses
type MetricsDeliveryResponse struct {
}

func (m *MetricsDeliveryResponse) Reset()                    { *m = MetricsDeliveryResponse{} }
func (m *MetricsDeliveryResponse) String() string            { return proto.CompactTextString(m) }
func (*MetricsDeliveryResponse) ProtoMessage()               {}
func (*MetricsDeliveryResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// Machines refer to specific backends or frontends that are parts of the network. (Frontends connect to their backends only, but they can still send frontend metrics). The list of these machines are public and available on the network, so this message does not contain any information that is not available publicly. Frontends are not connected to backends and they send their metrics completely separately with a different identifier, so that no backend can be matched to the frontend(s) it serves.
type Machine struct {
	NodeId       string                `protobuf:"bytes,1,opt,name=nodeId" json:"nodeId,omitempty"`
	MetricsToken *Machine_MetricsToken `protobuf:"bytes,2,opt,name=metricsToken" json:"metricsToken,omitempty"`
	Protocols    []*Protocol           `protobuf:"bytes,3,rep,name=protocols" json:"protocols,omitempty"`
	Client       *Client               `protobuf:"bytes,4,opt,name=client" json:"client,omitempty"`
	Address      string                `protobuf:"bytes,5,opt,name=address" json:"address,omitempty"`
	Port         int32                 `protobuf:"varint,6,opt,name=port" json:"port,omitempty"`
}

func (m *Machine) Reset()                    { *m = Machine{} }
func (m *Machine) String() string            { return proto.CompactTextString(m) }
func (*Machine) ProtoMessage()               {}
func (*Machine) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Machine) GetNodeId() string {
	if m != nil {
		return m.NodeId
	}
	return ""
}

func (m *Machine) GetMetricsToken() *Machine_MetricsToken {
	if m != nil {
		return m.MetricsToken
	}
	return nil
}

func (m *Machine) GetProtocols() []*Protocol {
	if m != nil {
		return m.Protocols
	}
	return nil
}

func (m *Machine) GetClient() *Client {
	if m != nil {
		return m.Client
	}
	return nil
}

func (m *Machine) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *Machine) GetPort() int32 {
	if m != nil {
		return m.Port
	}
	return 0
}

type Machine_MetricsToken struct {
	Token string `protobuf:"bytes,1,opt,name=token" json:"token,omitempty"`
}

func (m *Machine_MetricsToken) Reset()                    { *m = Machine_MetricsToken{} }
func (m *Machine_MetricsToken) String() string            { return proto.CompactTextString(m) }
func (*Machine_MetricsToken) ProtoMessage()               {}
func (*Machine_MetricsToken) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1, 0} }

func (m *Machine_MetricsToken) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

type Client struct {
	Name         string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	VersionMajor int32  `protobuf:"varint,2,opt,name=versionMajor" json:"versionMajor,omitempty"`
	VersionMinor int32  `protobuf:"varint,3,opt,name=versionMinor" json:"versionMinor,omitempty"`
	VersionPatch int32  `protobuf:"varint,4,opt,name=versionPatch" json:"versionPatch,omitempty"`
}

func (m *Client) Reset()                    { *m = Client{} }
func (m *Client) String() string            { return proto.CompactTextString(m) }
func (*Client) ProtoMessage()               {}
func (*Client) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Client) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Client) GetVersionMajor() int32 {
	if m != nil {
		return m.VersionMajor
	}
	return 0
}

func (m *Client) GetVersionMinor() int32 {
	if m != nil {
		return m.VersionMinor
	}
	return 0
}

func (m *Client) GetVersionPatch() int32 {
	if m != nil {
		return m.VersionPatch
	}
	return 0
}

type Protocol struct {
	Name              string   `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	VersionMajor      int32    `protobuf:"varint,2,opt,name=versionMajor" json:"versionMajor,omitempty"`
	VersionMinor      int32    `protobuf:"varint,3,opt,name=versionMinor" json:"versionMinor,omitempty"`
	SupportedEntities []string `protobuf:"bytes,4,rep,name=supportedEntities" json:"supportedEntities,omitempty"`
}

func (m *Protocol) Reset()                    { *m = Protocol{} }
func (m *Protocol) String() string            { return proto.CompactTextString(m) }
func (*Protocol) ProtoMessage()               {}
func (*Protocol) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Protocol) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Protocol) GetVersionMajor() int32 {
	if m != nil {
		return m.VersionMajor
	}
	return 0
}

func (m *Protocol) GetVersionMinor() int32 {
	if m != nil {
		return m.VersionMinor
	}
	return 0
}

func (m *Protocol) GetSupportedEntities() []string {
	if m != nil {
		return m.SupportedEntities
	}
	return nil
}

type Entity struct {
	EntityType         Entity_EntityType `protobuf:"varint,1,opt,name=entityType,enum=metrics.Entity_EntityType" json:"entityType,omitempty"`
	Fingerprint        string            `protobuf:"bytes,2,opt,name=fingerprint" json:"fingerprint,omitempty"`
	AddressLocation    string            `protobuf:"bytes,3,opt,name=addressLocation" json:"addressLocation,omitempty"`
	AddressSublocation string            `protobuf:"bytes,4,opt,name=addressSublocation" json:"addressSublocation,omitempty"`
	AddressPort        int32             `protobuf:"varint,5,opt,name=addressPort" json:"addressPort,omitempty"`
	LastUpdate         int64             `protobuf:"varint,6,opt,name=lastUpdate" json:"lastUpdate,omitempty"`
}

func (m *Entity) Reset()                    { *m = Entity{} }
func (m *Entity) String() string            { return proto.CompactTextString(m) }
func (*Entity) ProtoMessage()               {}
func (*Entity) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *Entity) GetEntityType() Entity_EntityType {
	if m != nil {
		return m.EntityType
	}
	return Entity_UNKNOWN
}

func (m *Entity) GetFingerprint() string {
	if m != nil {
		return m.Fingerprint
	}
	return ""
}

func (m *Entity) GetAddressLocation() string {
	if m != nil {
		return m.AddressLocation
	}
	return ""
}

func (m *Entity) GetAddressSublocation() string {
	if m != nil {
		return m.AddressSublocation
	}
	return ""
}

func (m *Entity) GetAddressPort() int32 {
	if m != nil {
		return m.AddressPort
	}
	return 0
}

func (m *Entity) GetLastUpdate() int64 {
	if m != nil {
		return m.LastUpdate
	}
	return 0
}

type Connection struct {
	Timestamp  int64                `protobuf:"varint,1,opt,name=timestamp" json:"timestamp,omitempty"`
	Successful bool                 `protobuf:"varint,2,opt,name=successful" json:"successful,omitempty"`
	Direction  Connection_Direction `protobuf:"varint,3,opt,name=direction,enum=metrics.Connection_Direction" json:"direction,omitempty"`
	Address    string               `protobuf:"bytes,4,opt,name=address" json:"address,omitempty"`
	Port       int32                `protobuf:"varint,5,opt,name=port" json:"port,omitempty"`
}

func (m *Connection) Reset()                    { *m = Connection{} }
func (m *Connection) String() string            { return proto.CompactTextString(m) }
func (*Connection) ProtoMessage()               {}
func (*Connection) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *Connection) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *Connection) GetSuccessful() bool {
	if m != nil {
		return m.Successful
	}
	return false
}

func (m *Connection) GetDirection() Connection_Direction {
	if m != nil {
		return m.Direction
	}
	return Connection_INBOUND
}

func (m *Connection) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *Connection) GetPort() int32 {
	if m != nil {
		return m.Port
	}
	return 0
}

type NodeEntity struct {
	Fingerprint            string `protobuf:"bytes,1,opt,name=fingerprint" json:"fingerprint,omitempty"`
	BoardsLastCheckin      int64  `protobuf:"varint,2,opt,name=boardsLastCheckin" json:"boardsLastCheckin,omitempty"`
	ThreadsLastCheckin     int64  `protobuf:"varint,3,opt,name=threadsLastCheckin" json:"threadsLastCheckin,omitempty"`
	PostsLastCheckin       int64  `protobuf:"varint,4,opt,name=postsLastCheckin" json:"postsLastCheckin,omitempty"`
	VotesLastCheckin       int64  `protobuf:"varint,5,opt,name=votesLastCheckin" json:"votesLastCheckin,omitempty"`
	KeysLastCheckin        int64  `protobuf:"varint,6,opt,name=keysLastCheckin" json:"keysLastCheckin,omitempty"`
	TruststatesLastCheckin int64  `protobuf:"varint,7,opt,name=truststatesLastCheckin" json:"truststatesLastCheckin,omitempty"`
	AddressesLastCheckin   int64  `protobuf:"varint,8,opt,name=addressesLastCheckin" json:"addressesLastCheckin,omitempty"`
}

func (m *NodeEntity) Reset()                    { *m = NodeEntity{} }
func (m *NodeEntity) String() string            { return proto.CompactTextString(m) }
func (*NodeEntity) ProtoMessage()               {}
func (*NodeEntity) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *NodeEntity) GetFingerprint() string {
	if m != nil {
		return m.Fingerprint
	}
	return ""
}

func (m *NodeEntity) GetBoardsLastCheckin() int64 {
	if m != nil {
		return m.BoardsLastCheckin
	}
	return 0
}

func (m *NodeEntity) GetThreadsLastCheckin() int64 {
	if m != nil {
		return m.ThreadsLastCheckin
	}
	return 0
}

func (m *NodeEntity) GetPostsLastCheckin() int64 {
	if m != nil {
		return m.PostsLastCheckin
	}
	return 0
}

func (m *NodeEntity) GetVotesLastCheckin() int64 {
	if m != nil {
		return m.VotesLastCheckin
	}
	return 0
}

func (m *NodeEntity) GetKeysLastCheckin() int64 {
	if m != nil {
		return m.KeysLastCheckin
	}
	return 0
}

func (m *NodeEntity) GetTruststatesLastCheckin() int64 {
	if m != nil {
		return m.TruststatesLastCheckin
	}
	return 0
}

func (m *NodeEntity) GetAddressesLastCheckin() int64 {
	if m != nil {
		return m.AddressesLastCheckin
	}
	return 0
}

type Metrics struct {
	Machine     *Machine             `protobuf:"bytes,1,opt,name=machine" json:"machine,omitempty"`
	Persistence *Metrics_Persistence `protobuf:"bytes,2,opt,name=persistence" json:"persistence,omitempty"`
	Network     *Metrics_Network     `protobuf:"bytes,3,opt,name=network" json:"network,omitempty"`
	Node        *Metrics_Node        `protobuf:"bytes,4,opt,name=node" json:"node,omitempty"`
	Validation  *Metrics_Validation  `protobuf:"bytes,5,opt,name=validation" json:"validation,omitempty"`
	Frontend    *Metrics_Frontend    `protobuf:"bytes,6,opt,name=frontend" json:"frontend,omitempty"`
}

func (m *Metrics) Reset()                    { *m = Metrics{} }
func (m *Metrics) String() string            { return proto.CompactTextString(m) }
func (*Metrics) ProtoMessage()               {}
func (*Metrics) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *Metrics) GetMachine() *Machine {
	if m != nil {
		return m.Machine
	}
	return nil
}

func (m *Metrics) GetPersistence() *Metrics_Persistence {
	if m != nil {
		return m.Persistence
	}
	return nil
}

func (m *Metrics) GetNetwork() *Metrics_Network {
	if m != nil {
		return m.Network
	}
	return nil
}

func (m *Metrics) GetNode() *Metrics_Node {
	if m != nil {
		return m.Node
	}
	return nil
}

func (m *Metrics) GetValidation() *Metrics_Validation {
	if m != nil {
		return m.Validation
	}
	return nil
}

func (m *Metrics) GetFrontend() *Metrics_Frontend {
	if m != nil {
		return m.Frontend
	}
	return nil
}

// Metrics related to the database state and caching.
type Metrics_Persistence struct {
	CurrentDatabaseSize                int64         `protobuf:"varint,1,opt,name=currentDatabaseSize" json:"currentDatabaseSize,omitempty"`
	CurrentCachesSize                  int64         `protobuf:"varint,2,opt,name=currentCachesSize" json:"currentCachesSize,omitempty"`
	ArrivedEntitiesSinceLastMetricsDbg []*Entity     `protobuf:"bytes,3,rep,name=arrivedEntitiesSinceLastMetricsDbg" json:"arrivedEntitiesSinceLastMetricsDbg,omitempty"`
	Orphans                            []*Entity     `protobuf:"bytes,4,rep,name=orphans" json:"orphans,omitempty"`
	NodeInsertionsSinceLastMetricsDbg  []*NodeEntity `protobuf:"bytes,5,rep,name=nodeInsertionsSinceLastMetricsDbg" json:"nodeInsertionsSinceLastMetricsDbg,omitempty"`
}

func (m *Metrics_Persistence) Reset()                    { *m = Metrics_Persistence{} }
func (m *Metrics_Persistence) String() string            { return proto.CompactTextString(m) }
func (*Metrics_Persistence) ProtoMessage()               {}
func (*Metrics_Persistence) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7, 0} }

func (m *Metrics_Persistence) GetCurrentDatabaseSize() int64 {
	if m != nil {
		return m.CurrentDatabaseSize
	}
	return 0
}

func (m *Metrics_Persistence) GetCurrentCachesSize() int64 {
	if m != nil {
		return m.CurrentCachesSize
	}
	return 0
}

func (m *Metrics_Persistence) GetArrivedEntitiesSinceLastMetricsDbg() []*Entity {
	if m != nil {
		return m.ArrivedEntitiesSinceLastMetricsDbg
	}
	return nil
}

func (m *Metrics_Persistence) GetOrphans() []*Entity {
	if m != nil {
		return m.Orphans
	}
	return nil
}

func (m *Metrics_Persistence) GetNodeInsertionsSinceLastMetricsDbg() []*NodeEntity {
	if m != nil {
		return m.NodeInsertionsSinceLastMetricsDbg
	}
	return nil
}

// Metrics related to the network state.
type Metrics_Network struct {
	InboundBandwidthConsumptionLast24H      int64         `protobuf:"varint,1,opt,name=inboundBandwidthConsumptionLast24h" json:"inboundBandwidthConsumptionLast24h,omitempty"`
	OutboundBandwidthConsumptionLast24H     int64         `protobuf:"varint,2,opt,name=outboundBandwidthConsumptionLast24h" json:"outboundBandwidthConsumptionLast24h,omitempty"`
	InboundBandwidthConsumptionLast7D       int64         `protobuf:"varint,3,opt,name=inboundBandwidthConsumptionLast7d" json:"inboundBandwidthConsumptionLast7d,omitempty"`
	OutboundBandwidthConsumptionLast7D      int64         `protobuf:"varint,4,opt,name=outboundBandwidthConsumptionLast7d" json:"outboundBandwidthConsumptionLast7d,omitempty"`
	InboundBandwidthConsumptionLast30D      int64         `protobuf:"varint,5,opt,name=inboundBandwidthConsumptionLast30d" json:"inboundBandwidthConsumptionLast30d,omitempty"`
	OutboundBandwidthConsumptionLast30D     int64         `protobuf:"varint,6,opt,name=outboundBandwidthConsumptionLast30d" json:"outboundBandwidthConsumptionLast30d,omitempty"`
	CurrentOnlineNodesCount                 int32         `protobuf:"varint,7,opt,name=currentOnlineNodesCount" json:"currentOnlineNodesCount,omitempty"`
	Last10InboundConnections                []*Connection `protobuf:"bytes,8,rep,name=last10InboundConnections" json:"last10InboundConnections,omitempty"`
	Last10OutboundConnections               []*Connection `protobuf:"bytes,9,rep,name=last10OutboundConnections" json:"last10OutboundConnections,omitempty"`
	Last10SuccessfulInboundConnections      []*Connection `protobuf:"bytes,10,rep,name=last10SuccessfulInboundConnections" json:"last10SuccessfulInboundConnections,omitempty"`
	Last10SuccessfulOutboundConnections     []*Connection `protobuf:"bytes,11,rep,name=last10SuccessfulOutboundConnections" json:"last10SuccessfulOutboundConnections,omitempty"`
	CurrentOnlineNodesDbg                   []*Machine    `protobuf:"bytes,12,rep,name=currentOnlineNodesDbg" json:"currentOnlineNodesDbg,omitempty"`
	Last100InboundConnectionsDbg            []*Connection `protobuf:"bytes,13,rep,name=last100InboundConnectionsDbg" json:"last100InboundConnectionsDbg,omitempty"`
	Last100OutboundConnectionsDbg           []*Connection `protobuf:"bytes,14,rep,name=last100OutboundConnectionsDbg" json:"last100OutboundConnectionsDbg,omitempty"`
	Last100SuccessfulInboundConnectionsDbg  []*Connection `protobuf:"bytes,15,rep,name=last100SuccessfulInboundConnectionsDbg" json:"last100SuccessfulInboundConnectionsDbg,omitempty"`
	Last100SuccessfulOutboundConnectionsDbg []*Connection `protobuf:"bytes,16,rep,name=last100SuccessfulOutboundConnectionsDbg" json:"last100SuccessfulOutboundConnectionsDbg,omitempty"`
}

func (m *Metrics_Network) Reset()                    { *m = Metrics_Network{} }
func (m *Metrics_Network) String() string            { return proto.CompactTextString(m) }
func (*Metrics_Network) ProtoMessage()               {}
func (*Metrics_Network) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7, 1} }

func (m *Metrics_Network) GetInboundBandwidthConsumptionLast24H() int64 {
	if m != nil {
		return m.InboundBandwidthConsumptionLast24H
	}
	return 0
}

func (m *Metrics_Network) GetOutboundBandwidthConsumptionLast24H() int64 {
	if m != nil {
		return m.OutboundBandwidthConsumptionLast24H
	}
	return 0
}

func (m *Metrics_Network) GetInboundBandwidthConsumptionLast7D() int64 {
	if m != nil {
		return m.InboundBandwidthConsumptionLast7D
	}
	return 0
}

func (m *Metrics_Network) GetOutboundBandwidthConsumptionLast7D() int64 {
	if m != nil {
		return m.OutboundBandwidthConsumptionLast7D
	}
	return 0
}

func (m *Metrics_Network) GetInboundBandwidthConsumptionLast30D() int64 {
	if m != nil {
		return m.InboundBandwidthConsumptionLast30D
	}
	return 0
}

func (m *Metrics_Network) GetOutboundBandwidthConsumptionLast30D() int64 {
	if m != nil {
		return m.OutboundBandwidthConsumptionLast30D
	}
	return 0
}

func (m *Metrics_Network) GetCurrentOnlineNodesCount() int32 {
	if m != nil {
		return m.CurrentOnlineNodesCount
	}
	return 0
}

func (m *Metrics_Network) GetLast10InboundConnections() []*Connection {
	if m != nil {
		return m.Last10InboundConnections
	}
	return nil
}

func (m *Metrics_Network) GetLast10OutboundConnections() []*Connection {
	if m != nil {
		return m.Last10OutboundConnections
	}
	return nil
}

func (m *Metrics_Network) GetLast10SuccessfulInboundConnections() []*Connection {
	if m != nil {
		return m.Last10SuccessfulInboundConnections
	}
	return nil
}

func (m *Metrics_Network) GetLast10SuccessfulOutboundConnections() []*Connection {
	if m != nil {
		return m.Last10SuccessfulOutboundConnections
	}
	return nil
}

func (m *Metrics_Network) GetCurrentOnlineNodesDbg() []*Machine {
	if m != nil {
		return m.CurrentOnlineNodesDbg
	}
	return nil
}

func (m *Metrics_Network) GetLast100InboundConnectionsDbg() []*Connection {
	if m != nil {
		return m.Last100InboundConnectionsDbg
	}
	return nil
}

func (m *Metrics_Network) GetLast100OutboundConnectionsDbg() []*Connection {
	if m != nil {
		return m.Last100OutboundConnectionsDbg
	}
	return nil
}

func (m *Metrics_Network) GetLast100SuccessfulInboundConnectionsDbg() []*Connection {
	if m != nil {
		return m.Last100SuccessfulInboundConnectionsDbg
	}
	return nil
}

func (m *Metrics_Network) GetLast100SuccessfulOutboundConnectionsDbg() []*Connection {
	if m != nil {
		return m.Last100SuccessfulOutboundConnectionsDbg
	}
	return nil
}

// Metrics related to local node health.
type Metrics_Node struct {
	SystemUptime        int64 `protobuf:"varint,1,opt,name=systemUptime" json:"systemUptime,omitempty"`
	NodeIsShuttingDown  bool  `protobuf:"varint,2,opt,name=nodeIsShuttingDown" json:"nodeIsShuttingDown,omitempty"`
	NodeIsStartingUp    bool  `protobuf:"varint,3,opt,name=nodeIsStartingUp" json:"nodeIsStartingUp,omitempty"`
	NodeIsTrackingHead  bool  `protobuf:"varint,4,opt,name=nodeIsTrackingHead" json:"nodeIsTrackingHead,omitempty"`
	NodeIsBehindARouter bool  `protobuf:"varint,5,opt,name=nodeIsBehindARouter" json:"nodeIsBehindARouter,omitempty"`
	LoadAvg1            int32 `protobuf:"varint,6,opt,name=loadAvg1" json:"loadAvg1,omitempty"`
	LoadAvg5            int32 `protobuf:"varint,7,opt,name=loadAvg5" json:"loadAvg5,omitempty"`
	LoadAvg15           int32 `protobuf:"varint,8,opt,name=loadAvg15" json:"loadAvg15,omitempty"`
}

func (m *Metrics_Node) Reset()                    { *m = Metrics_Node{} }
func (m *Metrics_Node) String() string            { return proto.CompactTextString(m) }
func (*Metrics_Node) ProtoMessage()               {}
func (*Metrics_Node) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7, 2} }

func (m *Metrics_Node) GetSystemUptime() int64 {
	if m != nil {
		return m.SystemUptime
	}
	return 0
}

func (m *Metrics_Node) GetNodeIsShuttingDown() bool {
	if m != nil {
		return m.NodeIsShuttingDown
	}
	return false
}

func (m *Metrics_Node) GetNodeIsStartingUp() bool {
	if m != nil {
		return m.NodeIsStartingUp
	}
	return false
}

func (m *Metrics_Node) GetNodeIsTrackingHead() bool {
	if m != nil {
		return m.NodeIsTrackingHead
	}
	return false
}

func (m *Metrics_Node) GetNodeIsBehindARouter() bool {
	if m != nil {
		return m.NodeIsBehindARouter
	}
	return false
}

func (m *Metrics_Node) GetLoadAvg1() int32 {
	if m != nil {
		return m.LoadAvg1
	}
	return 0
}

func (m *Metrics_Node) GetLoadAvg5() int32 {
	if m != nil {
		return m.LoadAvg5
	}
	return 0
}

func (m *Metrics_Node) GetLoadAvg15() int32 {
	if m != nil {
		return m.LoadAvg15
	}
	return 0
}

type Metrics_Validation struct {
	Failures []*Entity `protobuf:"bytes,1,rep,name=failures" json:"failures,omitempty"`
}

func (m *Metrics_Validation) Reset()                    { *m = Metrics_Validation{} }
func (m *Metrics_Validation) String() string            { return proto.CompactTextString(m) }
func (*Metrics_Validation) ProtoMessage()               {}
func (*Metrics_Validation) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7, 3} }

func (m *Metrics_Validation) GetFailures() []*Entity {
	if m != nil {
		return m.Failures
	}
	return nil
}

// Product metrics. TBD.
type Metrics_Frontend struct {
}

func (m *Metrics_Frontend) Reset()                    { *m = Metrics_Frontend{} }
func (m *Metrics_Frontend) String() string            { return proto.CompactTextString(m) }
func (*Metrics_Frontend) ProtoMessage()               {}
func (*Metrics_Frontend) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7, 4} }

func init() {
	proto.RegisterType((*MetricsDeliveryResponse)(nil), "metrics.MetricsDeliveryResponse")
	proto.RegisterType((*Machine)(nil), "metrics.Machine")
	proto.RegisterType((*Machine_MetricsToken)(nil), "metrics.Machine.MetricsToken")
	proto.RegisterType((*Client)(nil), "metrics.Client")
	proto.RegisterType((*Protocol)(nil), "metrics.Protocol")
	proto.RegisterType((*Entity)(nil), "metrics.Entity")
	proto.RegisterType((*Connection)(nil), "metrics.Connection")
	proto.RegisterType((*NodeEntity)(nil), "metrics.NodeEntity")
	proto.RegisterType((*Metrics)(nil), "metrics.Metrics")
	proto.RegisterType((*Metrics_Persistence)(nil), "metrics.Metrics.Persistence")
	proto.RegisterType((*Metrics_Network)(nil), "metrics.Metrics.Network")
	proto.RegisterType((*Metrics_Node)(nil), "metrics.Metrics.Node")
	proto.RegisterType((*Metrics_Validation)(nil), "metrics.Metrics.Validation")
	proto.RegisterType((*Metrics_Frontend)(nil), "metrics.Metrics.Frontend")
	proto.RegisterEnum("metrics.Entity_EntityType", Entity_EntityType_name, Entity_EntityType_value)
	proto.RegisterEnum("metrics.Connection_Direction", Connection_Direction_name, Connection_Direction_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for MetricsService service

type MetricsServiceClient interface {
	RequestMetricsToken(ctx context.Context, in *Machine, opts ...grpc.CallOption) (*Machine_MetricsToken, error)
	UploadMetrics(ctx context.Context, in *Metrics, opts ...grpc.CallOption) (*MetricsDeliveryResponse, error)
}

type metricsServiceClient struct {
	cc *grpc.ClientConn
}

func NewMetricsServiceClient(cc *grpc.ClientConn) MetricsServiceClient {
	return &metricsServiceClient{cc}
}

func (c *metricsServiceClient) RequestMetricsToken(ctx context.Context, in *Machine, opts ...grpc.CallOption) (*Machine_MetricsToken, error) {
	out := new(Machine_MetricsToken)
	err := grpc.Invoke(ctx, "/metrics.MetricsService/RequestMetricsToken", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricsServiceClient) UploadMetrics(ctx context.Context, in *Metrics, opts ...grpc.CallOption) (*MetricsDeliveryResponse, error) {
	out := new(MetricsDeliveryResponse)
	err := grpc.Invoke(ctx, "/metrics.MetricsService/UploadMetrics", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for MetricsService service

type MetricsServiceServer interface {
	RequestMetricsToken(context.Context, *Machine) (*Machine_MetricsToken, error)
	UploadMetrics(context.Context, *Metrics) (*MetricsDeliveryResponse, error)
}

func RegisterMetricsServiceServer(s *grpc.Server, srv MetricsServiceServer) {
	s.RegisterService(&_MetricsService_serviceDesc, srv)
}

func _MetricsService_RequestMetricsToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Machine)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServiceServer).RequestMetricsToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/metrics.MetricsService/RequestMetricsToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServiceServer).RequestMetricsToken(ctx, req.(*Machine))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricsService_UploadMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Metrics)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServiceServer).UploadMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/metrics.MetricsService/UploadMetrics",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServiceServer).UploadMetrics(ctx, req.(*Metrics))
	}
	return interceptor(ctx, in, info, handler)
}

var _MetricsService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "metrics.MetricsService",
	HandlerType: (*MetricsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RequestMetricsToken",
			Handler:    _MetricsService_RequestMetricsToken_Handler,
		},
		{
			MethodName: "UploadMetrics",
			Handler:    _MetricsService_UploadMetrics_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/metrics.proto",
}

func init() { proto.RegisterFile("proto/metrics.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 1425 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x58, 0xdd, 0x72, 0x13, 0xb7,
	0x17, 0x8f, 0xe3, 0xef, 0xe3, 0x90, 0x18, 0x85, 0x0f, 0xe3, 0x3f, 0xfc, 0x27, 0x2c, 0x1d, 0x08,
	0x94, 0x09, 0x60, 0x4a, 0x69, 0xcb, 0x4c, 0x67, 0x92, 0xd8, 0x0c, 0x0c, 0xc4, 0x4e, 0x65, 0x1b,
	0x86, 0xab, 0xce, 0x66, 0x57, 0x89, 0x97, 0xd8, 0xd2, 0x56, 0xd2, 0x9a, 0x49, 0x9f, 0xa0, 0xd7,
	0x9d, 0xde, 0x77, 0xfa, 0x22, 0x7d, 0x92, 0xf6, 0x11, 0xda, 0x07, 0xe8, 0x55, 0x47, 0x5a, 0xd9,
	0xbb, 0xf6, 0xae, 0x13, 0xdf, 0xf4, 0xca, 0xab, 0x73, 0x7e, 0xe7, 0x43, 0x47, 0x47, 0xe7, 0x1c,
	0x19, 0x36, 0x7d, 0xce, 0x24, 0x7b, 0x34, 0x22, 0x92, 0x7b, 0x8e, 0xd8, 0xd1, 0x2b, 0x54, 0x34,
	0x4b, 0xeb, 0x06, 0x5c, 0x3f, 0x08, 0x3f, 0x9b, 0x64, 0xe8, 0x8d, 0x09, 0x3f, 0xc3, 0x44, 0xf8,
	0x8c, 0x0a, 0x62, 0xfd, 0xbc, 0x0a, 0xc5, 0x03, 0xdb, 0x19, 0x78, 0x94, 0xa0, 0x6b, 0x50, 0xa0,
	0xcc, 0x25, 0xaf, 0xdd, 0x5a, 0x66, 0x2b, 0xb3, 0x5d, 0xc6, 0x66, 0x85, 0x76, 0x61, 0xcd, 0x68,
	0xea, 0xb1, 0x53, 0x42, 0x6b, 0xab, 0x5b, 0x99, 0xed, 0x4a, 0xe3, 0xd6, 0xce, 0xc4, 0x9a, 0x91,
	0xdf, 0x39, 0x88, 0x81, 0xf0, 0x8c, 0x08, 0x7a, 0x04, 0x65, 0xed, 0x93, 0xc3, 0x86, 0xa2, 0x96,
	0xdd, 0xca, 0x6e, 0x57, 0x1a, 0x97, 0xa7, 0xf2, 0x87, 0x86, 0x83, 0x23, 0x0c, 0xba, 0x07, 0x05,
	0x67, 0xe8, 0x11, 0x2a, 0x6b, 0x39, 0x6d, 0x6d, 0x63, 0x8a, 0xde, 0xd7, 0x64, 0x6c, 0xd8, 0xa8,
	0x06, 0x45, 0xdb, 0x75, 0x39, 0x11, 0xa2, 0x96, 0xd7, 0x5e, 0x4f, 0x96, 0x08, 0x41, 0xce, 0x67,
	0x5c, 0xd6, 0x0a, 0x5b, 0x99, 0xed, 0x3c, 0xd6, 0xdf, 0xf5, 0xcf, 0x60, 0x2d, 0xee, 0x25, 0xba,
	0x02, 0x79, 0xa9, 0xf7, 0x14, 0xee, 0x38, 0x5c, 0x58, 0x3f, 0x65, 0xa0, 0x10, 0x9a, 0x51, 0x4a,
	0xa8, 0x3d, 0x22, 0x86, 0xaf, 0xbf, 0x91, 0x05, 0x6b, 0x63, 0xc2, 0x85, 0xc7, 0xe8, 0x81, 0xfd,
	0x91, 0x71, 0x1d, 0x8f, 0x3c, 0x9e, 0xa1, 0xc5, 0x31, 0x1e, 0x65, 0xbc, 0x96, 0x9d, 0xc5, 0x28,
	0x5a, 0x0c, 0x73, 0x68, 0x4b, 0x67, 0xa0, 0x77, 0x1a, 0x61, 0x34, 0xcd, 0xfa, 0x25, 0x03, 0xa5,
	0x49, 0x7c, 0xfe, 0x53, 0x67, 0x1e, 0xc2, 0x65, 0x11, 0xf8, 0x2a, 0x48, 0xc4, 0x6d, 0x51, 0xe9,
	0x49, 0x8f, 0x88, 0x5a, 0x6e, 0x2b, 0xbb, 0x5d, 0xc6, 0x49, 0x86, 0xf5, 0xf7, 0x2a, 0x14, 0xf4,
	0xe2, 0x0c, 0x7d, 0x03, 0x40, 0xf4, 0x57, 0xef, 0xcc, 0x0f, 0x5d, 0x5b, 0x6f, 0xd4, 0xa7, 0xa7,
	0x15, 0x82, 0xcc, 0x8f, 0x42, 0xe0, 0x18, 0x1a, 0x6d, 0x41, 0xe5, 0xd8, 0xa3, 0x27, 0x84, 0xfb,
	0xdc, 0xa3, 0x52, 0xfb, 0x5e, 0xc6, 0x71, 0x12, 0xda, 0x86, 0x0d, 0x73, 0x9e, 0x6f, 0x99, 0x63,
	0x4b, 0x8f, 0x51, 0xed, 0x7d, 0x19, 0xcf, 0x93, 0xd1, 0x0e, 0x20, 0x43, 0xea, 0x06, 0x47, 0xc3,
	0x09, 0x38, 0xa7, 0xc1, 0x29, 0x1c, 0x65, 0xdb, 0x50, 0x0f, 0x55, 0x96, 0xe4, 0x75, 0x4c, 0xe2,
	0x24, 0xf4, 0x7f, 0x80, 0xa1, 0x2d, 0x64, 0xdf, 0x77, 0x6d, 0x49, 0x74, 0x1a, 0x65, 0x71, 0x8c,
	0x62, 0x7d, 0x04, 0x88, 0xf6, 0x85, 0x2a, 0x50, 0xec, 0xb7, 0xdf, 0xb4, 0x3b, 0xef, 0xdb, 0xd5,
	0x15, 0x54, 0x86, 0xfc, 0x5e, 0x67, 0x17, 0x37, 0xab, 0x19, 0x04, 0x50, 0xe8, 0xbd, 0xc2, 0xad,
	0xdd, 0x66, 0x75, 0x15, 0x95, 0x20, 0x77, 0xd8, 0xe9, 0xf6, 0xaa, 0x59, 0xf5, 0xf5, 0xae, 0xd3,
	0x6b, 0x55, 0x73, 0xa8, 0x08, 0xd9, 0x37, 0xad, 0x0f, 0xd5, 0x3c, 0x5a, 0x07, 0xe8, 0xe1, 0x7e,
	0xb7, 0xd7, 0xed, 0xed, 0xf6, 0x5a, 0xd5, 0x82, 0x52, 0xb8, 0xdb, 0x6c, 0xe2, 0x56, 0xb7, 0x5b,
	0x2d, 0x5a, 0x7f, 0x66, 0x00, 0xf6, 0x19, 0xa5, 0xc4, 0xd1, 0xce, 0xdf, 0x84, 0xb2, 0xf4, 0x46,
	0x44, 0x48, 0x7b, 0xe4, 0xeb, 0x98, 0x67, 0x71, 0x44, 0x50, 0x8e, 0x8b, 0xc0, 0x71, 0x88, 0x10,
	0xc7, 0xc1, 0x50, 0x47, 0xb5, 0x84, 0x63, 0x14, 0xf4, 0x02, 0xca, 0xae, 0xc7, 0x43, 0x55, 0x3a,
	0x9c, 0xeb, 0xb1, 0xdb, 0x1c, 0x59, 0xd9, 0x69, 0x4e, 0x40, 0x38, 0xc2, 0xc7, 0x2f, 0x5c, 0x2e,
	0xfd, 0xc2, 0xe5, 0xa3, 0x0b, 0x67, 0xdd, 0x85, 0xf2, 0x54, 0x8b, 0xda, 0xd1, 0xeb, 0xf6, 0x5e,
	0xa7, 0xdf, 0x6e, 0x56, 0x57, 0xd0, 0x1a, 0x94, 0x3a, 0xfd, 0x5e, 0xb8, 0xca, 0x58, 0xff, 0xac,
	0x02, 0xb4, 0x99, 0x4b, 0x4c, 0x52, 0xcd, 0x25, 0x46, 0x26, 0x99, 0x18, 0x0f, 0xe1, 0xf2, 0x11,
	0xb3, 0xb9, 0x2b, 0xde, 0xda, 0x42, 0xee, 0x0f, 0x88, 0x73, 0xea, 0x85, 0x95, 0x29, 0x8b, 0x93,
	0x0c, 0x95, 0x1c, 0x72, 0xc0, 0x89, 0x3d, 0x0b, 0xcf, 0x6a, 0x78, 0x0a, 0x07, 0x3d, 0x80, 0xaa,
	0xcf, 0x84, 0x9c, 0x41, 0xe7, 0x34, 0x3a, 0x41, 0x57, 0xd8, 0x31, 0x93, 0x64, 0x06, 0x9b, 0x0f,
	0xb1, 0xf3, 0x74, 0x95, 0xce, 0xa7, 0xe4, 0x6c, 0x06, 0x1a, 0xe6, 0xd5, 0x3c, 0x19, 0x7d, 0x09,
	0xd7, 0x24, 0x0f, 0x84, 0x14, 0xd2, 0x9e, 0xd3, 0x5d, 0xd4, 0x02, 0x0b, 0xb8, 0xa8, 0x01, 0x57,
	0xcc, 0x79, 0xcc, 0x4a, 0x95, 0xb4, 0x54, 0x2a, 0xcf, 0xfa, 0x75, 0x13, 0x8a, 0xa6, 0x2c, 0xa2,
	0x07, 0x50, 0x1c, 0x85, 0xf5, 0x5c, 0x47, 0xbd, 0xd2, 0xa8, 0xce, 0xd7, 0x79, 0x3c, 0x01, 0xa0,
	0x6f, 0xa1, 0xe2, 0xab, 0x1a, 0x22, 0x24, 0xa1, 0x0e, 0x31, 0x7d, 0xe1, 0x66, 0x84, 0x9f, 0xd4,
	0xf7, 0x08, 0x83, 0xe3, 0x02, 0xa8, 0x01, 0x45, 0x4a, 0xe4, 0x27, 0xc6, 0x4f, 0xf5, 0x51, 0x54,
	0x1a, 0xb5, 0x84, 0x6c, 0x3b, 0xe4, 0xe3, 0x09, 0x10, 0xdd, 0x87, 0x9c, 0x6a, 0x4b, 0xa6, 0x2d,
	0x5c, 0x4d, 0x0a, 0x30, 0x97, 0x60, 0x0d, 0x41, 0x2f, 0x00, 0xc6, 0xf6, 0xd0, 0x73, 0xc3, 0x4a,
	0x90, 0xd7, 0x02, 0xff, 0x4b, 0x08, 0xbc, 0x9b, 0x42, 0x70, 0x0c, 0x8e, 0x9e, 0x41, 0xe9, 0x98,
	0x33, 0x2a, 0x09, 0x75, 0xf5, 0x11, 0x55, 0x1a, 0x37, 0x12, 0xa2, 0x2f, 0x0d, 0x00, 0x4f, 0xa1,
	0xf5, 0x3f, 0x56, 0xa1, 0x12, 0xdb, 0x2f, 0x7a, 0x0c, 0x9b, 0x4e, 0xc0, 0x39, 0xa1, 0xb2, 0x69,
	0x4b, 0xfb, 0xc8, 0x16, 0xa4, 0xeb, 0xfd, 0x48, 0xcc, 0x95, 0x4d, 0x63, 0xa9, 0xc4, 0x36, 0xe4,
	0x7d, 0xdb, 0x19, 0x10, 0xa1, 0xf1, 0x26, 0xb1, 0x13, 0x0c, 0xf4, 0x3d, 0x58, 0x36, 0xe7, 0xde,
	0x38, 0xaa, 0xcd, 0x5d, 0x8f, 0x3a, 0x44, 0x9d, 0xee, 0xa4, 0xe5, 0x1f, 0x9d, 0x98, 0x8e, 0xbb,
	0x31, 0x57, 0x95, 0xf1, 0x12, 0xa2, 0xe8, 0x3e, 0x14, 0x19, 0xf7, 0x07, 0x36, 0x0d, 0xbb, 0x41,
	0x8a, 0x96, 0x09, 0x1f, 0xd9, 0x70, 0x5b, 0x4f, 0x0c, 0x54, 0x10, 0xae, 0x62, 0x98, 0xea, 0x4a,
	0x5e, 0x2b, 0xd9, 0x9c, 0x2a, 0x89, 0x2e, 0x3d, 0xbe, 0x58, 0xba, 0xfe, 0x17, 0x40, 0xd1, 0xa4,
	0x04, 0x6a, 0x83, 0xe5, 0xd1, 0x23, 0x16, 0x50, 0x77, 0xcf, 0xa6, 0xee, 0x27, 0xcf, 0x95, 0x83,
	0x7d, 0x46, 0x45, 0x30, 0xf2, 0x95, 0xb8, 0x12, 0x6c, 0x7c, 0x31, 0x30, 0x91, 0x5e, 0x02, 0x89,
	0x0e, 0xe1, 0x0e, 0x0b, 0xe4, 0x85, 0x0a, 0xc3, 0xa3, 0x58, 0x06, 0x8a, 0xde, 0xc2, 0xed, 0x0b,
	0xec, 0x3e, 0x77, 0x4d, 0x11, 0xba, 0x18, 0xa8, 0xf6, 0x7b, 0x91, 0xd1, 0xe7, 0xae, 0xa9, 0x52,
	0x4b, 0x20, 0x97, 0x88, 0xdf, 0xd3, 0xc7, 0xae, 0xa9, 0x64, 0x4b, 0x20, 0x97, 0x89, 0x9f, 0x52,
	0x58, 0x58, 0x2e, 0x7e, 0x4a, 0xe3, 0x57, 0x70, 0xdd, 0x64, 0x7c, 0x87, 0x0e, 0x3d, 0x4a, 0x54,
	0xae, 0x88, 0x7d, 0x16, 0x50, 0xa9, 0x8b, 0x60, 0x1e, 0x2f, 0x62, 0xa3, 0x0e, 0xd4, 0x54, 0xa3,
	0x7e, 0xf2, 0xf8, 0x75, 0xe8, 0x77, 0xd4, 0xd4, 0x44, 0xad, 0x34, 0x97, 0x81, 0x11, 0x0f, 0x2f,
	0x14, 0x42, 0xdf, 0xc1, 0x8d, 0x90, 0xd7, 0x31, 0x7e, 0xc7, 0x35, 0x96, 0x17, 0x6b, 0x5c, 0x2c,
	0x85, 0x1c, 0xb0, 0x42, 0x66, 0x77, 0xda, 0x99, 0x53, 0xbc, 0x85, 0xc5, 0xba, 0x97, 0x10, 0x47,
	0x04, 0xee, 0xcc, 0xa3, 0xd2, 0x76, 0x50, 0x59, 0x6c, 0x65, 0x19, 0x79, 0xf4, 0x12, 0xae, 0x26,
	0x8f, 0x42, 0x5d, 0xf7, 0x35, 0xad, 0x38, 0xd9, 0x43, 0xd2, 0xe1, 0xe8, 0x3d, 0xdc, 0x0c, 0xcd,
	0xa5, 0x9c, 0x81, 0x52, 0x77, 0x69, 0xb1, 0x9f, 0xe7, 0x0a, 0xa2, 0x0f, 0x70, 0xcb, 0xf0, 0x53,
	0xdc, 0x57, 0x9a, 0xd7, 0x17, 0x6b, 0x3e, 0x5f, 0x12, 0x9d, 0xc2, 0x5d, 0x03, 0x38, 0xef, 0x24,
	0x94, 0x8d, 0x8d, 0xc5, 0x36, 0x96, 0x54, 0x81, 0x46, 0x70, 0x2f, 0x81, 0x5c, 0xb0, 0xa3, 0xea,
	0x62, 0x6b, 0xcb, 0xea, 0xa8, 0xff, 0xbe, 0x0a, 0x39, 0x75, 0x38, 0xea, 0x09, 0x21, 0xce, 0x84,
	0x24, 0xa3, 0xbe, 0xaf, 0xe6, 0x4c, 0x53, 0x56, 0x67, 0x68, 0x6a, 0xc8, 0xd2, 0x15, 0x5c, 0x74,
	0x07, 0x81, 0x94, 0x1e, 0x3d, 0x69, 0xb2, 0x4f, 0xd4, 0x8c, 0x9f, 0x29, 0x1c, 0x35, 0x38, 0x19,
	0xaa, 0xb4, 0xb9, 0xa2, 0xf6, 0x7d, 0x5d, 0x0d, 0x4b, 0x38, 0x41, 0x8f, 0x74, 0xf7, 0xb8, 0xad,
	0x66, 0x96, 0x93, 0x57, 0xc4, 0x0e, 0x8b, 0xdd, 0x54, 0x77, 0x9c, 0xa3, 0xfa, 0x6e, 0x48, 0xdd,
	0x23, 0x03, 0x8f, 0xba, 0xbb, 0x98, 0x05, 0x92, 0x70, 0x5d, 0xcd, 0x4a, 0x38, 0x8d, 0x85, 0xea,
	0x50, 0x1a, 0x32, 0xdb, 0xdd, 0x1d, 0x9f, 0x3c, 0x31, 0x4f, 0xc6, 0xe9, 0x3a, 0xc6, 0x7b, 0x66,
	0x2a, 0xcf, 0x74, 0xad, 0x46, 0xf1, 0x09, 0xee, 0x99, 0x9e, 0xb2, 0xf2, 0x38, 0x22, 0xd4, 0xbf,
	0x06, 0x88, 0x06, 0x0c, 0xf4, 0x39, 0x94, 0x8e, 0x6d, 0x6f, 0x18, 0x70, 0x22, 0x6a, 0x99, 0xf4,
	0x6e, 0x3a, 0x05, 0xd4, 0x01, 0x4a, 0x93, 0x01, 0xa3, 0xf1, 0x5b, 0x06, 0xd6, 0x4d, 0x1b, 0xec,
	0x12, 0x3e, 0xf6, 0x1c, 0x82, 0x5e, 0xc1, 0x26, 0x26, 0x3f, 0x04, 0x64, 0xda, 0x1f, 0xc3, 0x17,
	0x6d, 0xe2, 0xaa, 0xd5, 0xcf, 0x7f, 0xa8, 0x5b, 0x2b, 0xa8, 0x05, 0x97, 0xfa, 0xbe, 0x72, 0x79,
	0x32, 0x03, 0x56, 0xe7, 0x27, 0x9d, 0xfa, 0xd6, 0x3c, 0x25, 0xf1, 0x47, 0xc2, 0xca, 0x51, 0x41,
	0xbf, 0xde, 0x9f, 0xfe, 0x1b, 0x00, 0x00, 0xff, 0xff, 0xb9, 0xf4, 0x21, 0x1b, 0x8c, 0x10, 0x00,
	0x00,
}
