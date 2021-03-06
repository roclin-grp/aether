// Services > Config Store
// This module handles saving and reading values from a config user file.

package configstore

import (
	"aether-core/services/fingerprinting"
	// "aether-core/services/randomhashgen"
	"aether-core/services/signaturing"
	"crypto/ecdsa"
	// "crypto/elliptic"
	"crypto/x509"
	// "encoding/hex"
	pb "aether-core/backend/metrics/proto"
	"encoding/json"
	"errors"
	"fmt"
	// "github.com/davecgh/go-spew/spew"
	// "github.com/fatih/color"
	"crypto/elliptic"
	"encoding/hex"
	cdir "github.com/shibukawa/configdir"
	"log"
	"runtime"
	"sync"
	"time"
)

// Config interface, so that we can actually have methods that take either frontend or backend config.

type Config interface {
	BlankCheck()
	SanityCheck()
}

/*
This package handles any data that gets saved to the user profile. This is important because everything that does not get saved into the database gets saved into this. Also important is this is where we allow multiple users to use the same database.
*/

// 0) UTILITY FUNCTIONS

func trace() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	result := fmt.Sprintf("%s,:%d %s", frame.File, frame.Line, frame.Function)
	return result
}

func invalidDataError(input interface{}) error {
	return errors.New(fmt.Sprintf("An invalid value for this setting was provided by the user / application (in Set) or by the storage backend (in Get). Value provided: %#v", input))
}

// Maximums

const (
	maxPOWBailoutSeconds            = 3600 // 1h
	maxTimeblockSizeMinutes         = 360  // 6h
	maxPastBlocksToCheck            = 28   // 360*28 = 7 days online before cache generation can start
	maxCacheGenerationIntervalHours = 168  // 7 days
	maxAbsolutePageSize             = 1000000
	maxPOWStrength                  = 63 // Our PoWs are 64 bytes long
	maxLocationSize                 = 2500
)

const (
	maxInt8   = 1<<7 - 1
	minInt8   = -1 << 7
	maxInt16  = 1<<15 - 1
	minInt16  = -1 << 15
	maxInt32  = 1<<31 - 1
	minInt32  = -1 << 31
	maxInt64  = 1<<63 - 1
	minInt64  = -1 << 63
	maxUint8  = 1<<8 - 1
	maxUint16 = 1<<16 - 1
	maxUint32 = 1<<32 - 1
	maxUint64 = 1<<64 - 1
)

/*
(ll. 116-138) Verily at the first Chaos came to be, but next wide-bosomed Earth, the ever-sure foundations of all the deathless ones who hold the peaks of snowy Olympus, and dim Tartarus in the depth of the wide-pathed Earth, and Eros, fairest among the deathless gods, who unnerves the limbs and overcomes the mind and wise counsels of all gods and all men within them. From Chaos came forth Erebus and black Night; but of Night were born Aether and Day, whom she conceived and bare from union in love with Erebus.
*/

const (
	night = 4386570
)

// 1) BACKEND

// Backend sub-entities

type EntityPageSizes struct {
	Boards            int
	BoardIndexes      int
	Threads           int
	ThreadIndexes     int
	Posts             int
	PostIndexes       int
	Votes             int
	VoteIndexes       int
	Addresses         int
	AddressIndexes    int
	Keys              int
	KeyIndexes        int
	Truststates       int
	TruststateIndexes int
}

type MinimumPoWStrengths struct {
	Board            int
	BoardUpdate      int
	Thread           int
	ThreadUpdate     int
	Post             int
	PostUpdate       int
	Vote             int
	VoteUpdate       int
	Key              int
	KeyUpdate        int
	Truststate       int
	TruststateUpdate int
}

/*

This is an exact copy of the api.Subprotocol. This is here because we cannot import api here — it creates a circular reference. I've tried splitting API in many ways to avoid this issue, but each of the solutions to do that cause a lot more problems elsewhere since structs defined in the API have methods that reference other libraries, and moving those methods out of the structs mean the code gets a lot messier, etc. In short, unlikely as this sounds, creating a shim here and translating on the fly is the cleanest solution.

https://play.golang.org/p/x8wk4d7oAar <- an example of casting a shim to its actual thing. This could be worth it for the address as well, but address is a multi level entity so it might be not a one shot cast.. or maybe it would. Let's see. Ah yeah it doesn't work.

*/

type SubprotocolShim struct {
	Name              string   `json:"name"`
	VersionMajor      uint8    `json:"version_major"`
	VersionMinor      uint16   `json:"version_minor"`
	SupportedEntities []string `json:"supported_entities"`
}

// Defaults

const (
	defaultNetworkHeadDays                         = 14
	defaultNetworkMemoryDays                       = 180
	defaultLocalMemoryDays                         = 180
	defaultPoWBailoutTimeSeconds                   = 30
	defaultTimeBlockSizeMinutes                    = 5
	defaultPastBlocksToCheck                       = 3
	defaultCacheGenerationIntervalHours            = 24
	defaultPOSTResponseExpiryMinutes               = 30
	defaultConnectionTimeout                       = 60 * time.Second
	defaultTCPConnectTimeout                       = 3 * time.Second
	defaultTLSHandshakeTimeout                     = 1 * time.Second
	defaultPingerPageSize                          = 100
	defaultOnlineAddressFinderPageSize             = 99
	defaultDispatchExclusionExpiryForLiveAddress   = 5 * time.Second // this is normally minute TODO
	defaultDispatchExclusionExpiryForStaticAddress = 72 * time.Hour
	defaultPowStrength                             = 20
	defaultExternalIp                              = "0.0.0.0" // Localhost, if this is still 0.0.0.0 at any point in the future we failed at finding this out.
	defaultExternalIpType                          = 4         // IPv4
	defaultExternalPort                            = 49999
	defaultDbEngine                                = "sqlite" // 'sqlite' or 'mysql'
	defaultDBIp                                    = "127.0.0.1"
	defaultDbPort                                  = 3306
	defaultDbUsername                              = "aether-app-db-access-user"
	defaultDbPassword                              = "exventoveritas"
)

// Default entity page sizes

const (
	defaultBoardsPageSize            = 2000  // 0.2x
	defaultBoardIndexesPageSize      = 8000  // 0.025x
	defaultThreadsPageSize           = 400   // 1x
	defaultThreadIndexesPageSize     = 16000 // 0.025x
	defaultPostsPageSize             = 400   // 1x
	defaultPostIndexesPageSize       = 12000 // 0.033x
	defaultVotesPageSize             = 2000  // 0.2x
	defaultVoteIndexesPageSize       = 4000  // 0.1x
	defaultAddressesPageSize         = 16000 // 0.025x
	defaultAddressIndexesPageSize    = 16000 // 0.025x - Address is its own index
	defaultKeysPageSize              = 2000  // 0.2x
	defaultKeyIndexesPageSize        = 20000 // 0.02x
	defaultTruststatesPageSize       = 6000  // 0.025x
	defaultTruststateIndexesPageSize = 15000 // 0.01x
	// Every regular page is about 500kb that way.
	// Every index page is about 1mb.
)

// Hardcoded version numbers specific to this build
const (
	clientVersionMajor   = 2
	clientVersionMinor   = 0
	clientVersionPatch   = 0
	clientName           = "Aether"
	protocolVersionMajor = 1
	protocolVersionMinor = 0
)

/*
Backend configuration.

## NetworkHeadDays
Days  of data that will be broadcast out in form of caches.

## NetworkMemoryDays
Days of data that will be provided to network upon request.

## LocalMemoryDays
Days of data to be kept before deletion.

## LastCacheGenerationTimestamp
The last time a new cache was generated locally.

## EntityPageSizes
How many entities will be put in a response page in POST responses and caches.

## MinimumPoWStrengths
The minimum number of zeros hashcash algorithm needs to have at the beginning of the PoW to accept it as valid.

## PoWBailoutTimeSeconds
How long does it take before a PoW timestamp is marked unattainable by the local computer. This is to make sure that the app doesn't keep attempting forever for an unattainably strong PoW it attempted to generate.

## TimeBlockSizeMinutes
Related to: SyncConfirmations. This library splits the recent past into blocks of time, and checks whether there was at least one successful sync in every block to determine heuristically whether this node is tracking the head, or not. This value determines the size of the time blocks.

## PastBlocksToCheck
Related to above. This value determines how many time blocks will be checked.

## CacheGenerationIntervalHours
How often does the node generate a new cache. By default, it generates a new cache every day.

## ClientVersionMajor
Major version of the client software (Aether). x.0.0

## ClientVersionMinor
Minor version of the client software (Aether). 0.x.0

## ClientVersionPatch
Patch version of the client software (Aether). 0.0.x

## ClientName
Name of the client that this node is part of. (Aether)

## ProtocolVersionMajor
Major version of the Mim protocol that content is served over.

## ProtocolVersionMinor
Minor version of the Mim protocol that content is served over.

## POSTResponseExpiryMinutes
When a remote node makes a request via a POST response, a post response is generated, saved as a temporary file, and the access instructions are sent to a remote node. Remote node has a certain amount of time from this point on to fetch this response, around 30 minutes. After 30 minutes, this response is deleted.

## ConnectionTimeout
How long the local node tries to attempt to connect to a remote node before deeming it unusable.

## TCPConnectTimeout
How long the local node tries to attempt to establish a TCP connection to a remote node before deeming it unusable.

## TLSHandshakeTimeout
How long the local node tries to attempt to complete a TLS handshake to a remote node before deeming it unusable.

## PingerPageSize
Pinger goes through all available addresses to find out whether they are online or not. This is done to keep a list of nodes that are usually online and in a connectable state. Pinger does this in form of pages (because there are occasionally more addresses available than there are sockets available in the local machine). This number determines how many nodes Pinger will attempt to connect at the same time.

## OnlineAddressFinderPageSize
This page size is slightly different than above. This one is for the local database call. Effectively, it looks at the most recent X addresses in the database to find ones that were active recently, and if that page does not yield enough online addresses, moves to the next page. This is to prevent querying a huge number of addresses.

## DispatchExclusionExpiryForLiveAddress
This is how long we wait until we reconnect to the same live address to look for updates.

## DispatchExclusionExpiryForStaticAddress
This is how long we wait until we reconnect to the same static address to look for updates.

## LoggingLevel
How deeply do we want to keep logs, or if any. 0 is no logs, 1 is medium, 2 is deep logs.

## ExternalIp
The external IP of this machine.

## ExternalIpType
The external IP type of this machine. 4: IPv4, 6: IPv6, 3: URL (in case of static)

## ExternalPort
The external port type of this machine.

## LastStaticAddressConnectionTimestamp
The last time we synced with a static node.

## LastLiveAddressConnectionTimestamp
The last time we synced with a live node.

## ServingSubprotocols
The subprotocols that this machine supports. In this case, c0 and dweb.

## NodeId
The node id of this machine. This is a randomly generated number. It does not have much significance beyond letting remote nodes keep their sync timestamps in check.

## UserDirectory
Where we save the backend , and if this node has a frontend, the frontend profile. This directory is given by the OS.

## CachesDirectory
Where we save the caches. This directory is given by the OS.

## Initialised
Whether the configuration file is properly initialised. If this is false, the initialisation did not complete.

## DbEngine
DbEngine allows the user to choose the database they want to use. SQLite is better for local installations where the app stays running on a desktop machine. It is simple and fast. MySQL is better when there are multiple users on the same backend, and it's a lot more robust against concurrent accesses. The preferred MySQL implementation is MariaDB, but original MySQL should also work.

Important: Do not forget that you have to create a DB called "aetherdb" in your preferred SQL engine with read/write access for the Username you give below.

(I thought of making this an iota and saving the numbers in this slot instead of string, but then that would make other parts of the code harder to read, because a DbEngine named 0 gives no information about what db engine it is, and you'd need to refer to this file to understand. I'd rather be infinitesimally less efficient and require less human RAM to read.)

## DbIP
This is the IP of the SQL server, if not SQLite3. By default, it's 127.0.0.1.

## DbPort
Port of the SQL server, if not SQLite3. By default, it's 3306 (MySQL default port)

## DbUsername
DbUsername is the username of the account that has read/write access to the "aetherdb" database, if not SQLite3. By default it's "aether-app-db-access-user".

## DbPassword
The password of the DB user, if not SQLite3. By default it's "exventoveritas". It's highly recommended that you change this.

## MetricsLevel
## MetricsToken

## BackendKeyPair
Backend key pair is the key for this specific backend by which it signs the pages it creates. This is a combination of both private and public keys.

## AllowUnsignedEntities
If this is set to true, the node accepts posts that are anonymous. (But still with PoW and Fingerprint). This is disabled by default.

*/

// Every time you add a new item here, please add getters, setters and to blankcheck method

// And before you think "hm, these would be better if they were private with lowercase letters.. that means you can't export them with JSON. Been there."

// Backend config base
type BackendConfig struct {
	NetworkHeadDays                         uint                // 14
	NetworkMemoryDays                       uint                // 180
	LocalMemoryDays                         uint                // 180
	LastCacheGenerationTimestamp            uint64              //
	EntityPageSizes                         EntityPageSizes     //
	MinimumPoWStrengths                     MinimumPoWStrengths //
	PoWBailoutTimeSeconds                   uint                // 30
	TimeBlockSizeMinutes                    uint                // 5
	PastBlocksToCheck                       uint                // 3
	CacheGenerationIntervalHours            uint                // 24
	ClientVersionMajor                      uint8               // 2 addr
	ClientVersionMinor                      uint16              // 0 addr
	ClientVersionPatch                      uint16              // 0 addr
	ClientName                              string              // Aether addr
	ProtocolVersionMajor                    uint8               // 1 (This refers to Mim, not subprotocols) addr
	ProtocolVersionMinor                    uint16              // 0 addr
	POSTResponseExpiryMinutes               uint                // 30
	ConnectionTimeout                       time.Duration
	TCPConnectTimeout                       time.Duration
	TLSHandshakeTimeout                     time.Duration
	PingerPageSize                          uint
	OnlineAddressFinderPageSize             uint
	DispatchExclusionExpiryForLiveAddress   time.Duration
	DispatchExclusionExpiryForStaticAddress time.Duration
	LoggingLevel                            uint
	ExternalIp                              string // addr
	ExternalIpType                          uint8
	ExternalPort                            uint16
	LastStaticAddressConnectionTimestamp    uint64
	LastLiveAddressConnectionTimestamp      uint64
	ServingSubprotocols                     []SubprotocolShim
	NodeId                                  string
	UserDirectory                           string
	CachesDirectory                         string
	Initialised                             bool // False by default, init to set true
	DbEngine                                string
	DbIp                                    string // Only applies to non-sqlite
	DbPort                                  uint16 // Only applies to non-sqlite
	DbUsername                              string // Only applies to non-sqlite
	DbPassword                              string // Only applies to non-sqlite
	MetricsLevel                            uint8  // 0: no metrics transmitted
	MetricsToken                            string // If metrics level is not zero, metrics token is the anonymous identifier for the metrics server. Resetting this to 0 makes this node behave like a new node as far as metrics go, but if you don't want metrics to be collected, you can set it through the application or set the metrics level to zero in the JSON settings file.
	BackendKeyPair                          []byte
	MarshaledBackendPublicKey               string
	AllowUnsignedEntities                   bool
}

// GETTERS AND SETTERS

/*
Q: Why do we even have these?

Because some of our types are not directly convertible to JSON, like the public / private key pairs.

Having this kind of set interface allows us to replace storage implementations later in the process without disrupting the rest of the app. The get / setter methods might look simple now, but they have no guarantee of being so in the future.

Q: Why the pain of uint as much as possible, then converting to ints?

Because we do not want users to provide negative values and make the application behave unpredictably. Any negative value should make the app not even start at all.
*/

// Getters
func (config *BackendConfig) GetLocalMemoryDays() int {
	if config.LocalMemoryDays < night &&
		config.LocalMemoryDays > 0 {
		return int(config.LocalMemoryDays)
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.LocalMemoryDays) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetNetworkMemoryDays() int {
	if config.NetworkMemoryDays < night &&
		config.NetworkMemoryDays > 0 {
		return int(config.NetworkMemoryDays)
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.NetworkMemoryDays) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetNetworkHeadDays() int {
	if config.NetworkHeadDays < night &&
		config.NetworkHeadDays > 0 {
		return int(config.NetworkHeadDays)
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.NetworkHeadDays) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetLastCacheGenerationTimestamp() int64 {
	if config.LastCacheGenerationTimestamp < maxInt64 { // can be zero
		return int64(config.LastCacheGenerationTimestamp)
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.LastCacheGenerationTimestamp) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetEntityPageSizes() EntityPageSizes {
	if config.EntityPageSizes.Boards < maxAbsolutePageSize &&
		config.EntityPageSizes.Boards > 0 &&
		config.EntityPageSizes.BoardIndexes < maxAbsolutePageSize &&
		config.EntityPageSizes.BoardIndexes > 0 &&
		config.EntityPageSizes.Threads < maxAbsolutePageSize &&
		config.EntityPageSizes.Threads > 0 &&
		config.EntityPageSizes.ThreadIndexes < maxAbsolutePageSize &&
		config.EntityPageSizes.ThreadIndexes > 0 &&
		config.EntityPageSizes.Posts < maxAbsolutePageSize &&
		config.EntityPageSizes.Posts > 0 &&
		config.EntityPageSizes.PostIndexes < maxAbsolutePageSize &&
		config.EntityPageSizes.PostIndexes > 0 &&
		config.EntityPageSizes.Keys < maxAbsolutePageSize &&
		config.EntityPageSizes.Keys > 0 &&
		config.EntityPageSizes.KeyIndexes < maxAbsolutePageSize &&
		config.EntityPageSizes.KeyIndexes > 0 &&
		config.EntityPageSizes.Votes < maxAbsolutePageSize &&
		config.EntityPageSizes.Votes > 0 &&
		config.EntityPageSizes.VoteIndexes < maxAbsolutePageSize &&
		config.EntityPageSizes.VoteIndexes > 0 &&
		config.EntityPageSizes.Truststates < maxAbsolutePageSize &&
		config.EntityPageSizes.Truststates > 0 &&
		config.EntityPageSizes.TruststateIndexes < maxAbsolutePageSize &&
		config.EntityPageSizes.TruststateIndexes > 0 {
		return config.EntityPageSizes
	} else {
		log.Fatal(fmt.Sprintf("%#v", invalidDataError(config.EntityPageSizes)) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return EntityPageSizes{}
}
func (config *BackendConfig) GetMinimumPoWStrengths() MinimumPoWStrengths {
	if config.MinimumPoWStrengths.Board < maxPOWStrength &&
		config.MinimumPoWStrengths.Board > 0 &&
		config.MinimumPoWStrengths.BoardUpdate < maxPOWStrength &&
		config.MinimumPoWStrengths.BoardUpdate > 0 &&
		config.MinimumPoWStrengths.Thread < maxPOWStrength &&
		config.MinimumPoWStrengths.Thread > 0 &&
		config.MinimumPoWStrengths.ThreadUpdate < maxPOWStrength &&
		config.MinimumPoWStrengths.ThreadUpdate > 0 &&
		config.MinimumPoWStrengths.Post < maxPOWStrength &&
		config.MinimumPoWStrengths.Post > 0 &&
		config.MinimumPoWStrengths.PostUpdate < maxPOWStrength &&
		config.MinimumPoWStrengths.PostUpdate > 0 &&
		config.MinimumPoWStrengths.Vote < maxPOWStrength &&
		config.MinimumPoWStrengths.Vote > 0 &&
		config.MinimumPoWStrengths.VoteUpdate < maxPOWStrength &&
		config.MinimumPoWStrengths.VoteUpdate > 0 &&
		config.MinimumPoWStrengths.Key < maxPOWStrength &&
		config.MinimumPoWStrengths.Key > 0 &&
		config.MinimumPoWStrengths.KeyUpdate < maxPOWStrength &&
		config.MinimumPoWStrengths.KeyUpdate > 0 &&
		config.MinimumPoWStrengths.Truststate < maxPOWStrength &&
		config.MinimumPoWStrengths.Truststate > 0 &&
		config.MinimumPoWStrengths.TruststateUpdate < maxPOWStrength &&
		config.MinimumPoWStrengths.TruststateUpdate > 0 {
		return config.MinimumPoWStrengths
	} else {
		log.Fatal(fmt.Sprintf("%#v", invalidDataError(config.MinimumPoWStrengths)) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return MinimumPoWStrengths{}
}
func (config *BackendConfig) GetPoWBailoutTimeSeconds() int {
	if config.PoWBailoutTimeSeconds < maxPOWBailoutSeconds &&
		config.PoWBailoutTimeSeconds > 0 {
		return int(config.PoWBailoutTimeSeconds)
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.PoWBailoutTimeSeconds) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetTimeBlockSizeMinutes() int {
	if config.TimeBlockSizeMinutes < maxTimeblockSizeMinutes &&
		config.TimeBlockSizeMinutes > 0 {
		return int(config.TimeBlockSizeMinutes)
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.TimeBlockSizeMinutes) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetPastBlocksToCheck() int {
	if config.PastBlocksToCheck < maxPastBlocksToCheck &&
		config.PastBlocksToCheck > 0 {
		return int(config.PastBlocksToCheck)
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.PastBlocksToCheck) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetCacheGenerationIntervalHours() int {
	if config.CacheGenerationIntervalHours < maxCacheGenerationIntervalHours &&
		config.CacheGenerationIntervalHours > 0 {
		return int(config.CacheGenerationIntervalHours)
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.CacheGenerationIntervalHours) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetClientVersionMajor() uint8 {
	if config.ClientVersionMajor < maxUint8 &&
		config.ClientVersionMajor > 0 {
		return config.ClientVersionMajor
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.ClientVersionMajor) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetClientVersionMinor() uint16 {
	if config.ClientVersionMinor < maxUint16 { // can be zero
		return config.ClientVersionMinor
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.ClientVersionMinor) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetClientVersionPatch() uint16 {
	if config.ClientVersionPatch < maxUint16 { // can be zero
		return config.ClientVersionPatch
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.ClientVersionPatch) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetClientName() string {
	if len(config.ClientName) < maxUint8 &&
		len(config.ClientName) > 0 {
		return config.ClientName
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.ClientName) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return ""
}
func (config *BackendConfig) GetProtocolVersionMajor() uint8 {
	if config.ProtocolVersionMajor < maxUint8 &&
		config.ProtocolVersionMajor > 0 {
		return config.ProtocolVersionMajor
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.ProtocolVersionMajor) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetProtocolVersionMinor() uint16 {
	if config.ProtocolVersionMinor < maxUint16 { // can be zero
		return config.ProtocolVersionMinor
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.ProtocolVersionMinor) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetPOSTResponseExpiryMinutes() int {
	if config.POSTResponseExpiryMinutes < maxInt32 &&
		config.POSTResponseExpiryMinutes > 0 {
		return int(config.POSTResponseExpiryMinutes)
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.POSTResponseExpiryMinutes) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetConnectionTimeout() time.Duration {
	if config.ConnectionTimeout >= 1*time.Second { // Any value under is probably an attack.
		return config.ConnectionTimeout
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.ConnectionTimeout) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return time.Duration(0)
}
func (config *BackendConfig) GetTCPConnectTimeout() time.Duration {
	if config.TCPConnectTimeout >= 1*time.Second { // Any value under is probably an attack.
		return config.TCPConnectTimeout
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.TCPConnectTimeout) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return time.Duration(0)
}
func (config *BackendConfig) GetTLSHandshakeTimeout() time.Duration {
	if config.TLSHandshakeTimeout >= 1*time.Second { // Any value under is probably an attack.
		return config.TLSHandshakeTimeout
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.TLSHandshakeTimeout) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return time.Duration(0)
}
func (config *BackendConfig) GetPingerPageSize() int {
	if config.PingerPageSize < maxInt32 &&
		config.PingerPageSize > 0 {
		return int(config.PingerPageSize)
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.PingerPageSize) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetOnlineAddressFinderPageSize() int {
	if config.OnlineAddressFinderPageSize < maxInt32 &&
		config.OnlineAddressFinderPageSize > 0 {
		return int(config.OnlineAddressFinderPageSize)
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.OnlineAddressFinderPageSize) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetDispatchExclusionExpiryForLiveAddress() time.Duration {
	if config.DispatchExclusionExpiryForLiveAddress >= 1*time.Microsecond { // Any value under is probably an attack. TODO THIS IS NORMALLY A MINUTE
		return config.DispatchExclusionExpiryForLiveAddress
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.DispatchExclusionExpiryForLiveAddress) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return time.Duration(0)
}
func (config *BackendConfig) GetDispatchExclusionExpiryForStaticAddress() time.Duration {
	if config.DispatchExclusionExpiryForStaticAddress >= 1*time.Minute { // Any value under is probably an attack.
		return config.DispatchExclusionExpiryForStaticAddress
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.DispatchExclusionExpiryForStaticAddress) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return time.Duration(0)
}
func (config *BackendConfig) GetLoggingLevel() int {
	if config.LoggingLevel < maxInt32 { // can be zero
		return int(config.LoggingLevel)
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.LoggingLevel) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetExternalIp() string {
	if len(config.ExternalIp) < maxLocationSize &&
		len(config.ExternalIp) > 0 {
		return config.ExternalIp
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.ExternalIp) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return ""
}
func (config *BackendConfig) GetLastStaticAddressConnectionTimestamp() int64 {
	if config.LastStaticAddressConnectionTimestamp < maxInt64 { // can be zero
		return int64(config.LastStaticAddressConnectionTimestamp)
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.LastStaticAddressConnectionTimestamp) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetLastLiveAddressConnectionTimestamp() int64 {
	if config.LastLiveAddressConnectionTimestamp < maxInt64 { // can be zero
		return int64(config.LastLiveAddressConnectionTimestamp)
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.LastLiveAddressConnectionTimestamp) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}

func (config *BackendConfig) GetInitialised() bool {
	return config.Initialised
}
func (config *BackendConfig) GetServingSubprotocols() []SubprotocolShim {
	for _, val := range config.ServingSubprotocols {
		if len(val.SupportedEntities) == 0 {
			log.Fatal(invalidDataError(fmt.Sprintf("%#v", val.SupportedEntities) + " Trace: " + trace()))
		}
	}
	return config.ServingSubprotocols
}
func (config *BackendConfig) GetExternalIpType() uint8 {
	if config.ExternalIpType == 6 || config.ExternalIpType == 4 || config.ExternalIpType == 3 { // 6: ipv6, 4: ipv4, 3: URL (useful in static nodes)
		return config.ExternalIpType
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.ExternalIpType) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetNodeId() string {
	if len(config.NodeId) == 64 {
		return config.NodeId
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.NodeId) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return ""
}
func (config *BackendConfig) GetExternalPort() uint16 {
	if config.ExternalPort < maxUint16 && config.ExternalPort > 0 {
		return config.ExternalPort
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.ExternalPort) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetUserDirectory() string {
	if len(config.UserDirectory) < maxUint16 &&
		len(config.UserDirectory) > 0 {
		return config.UserDirectory
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.UserDirectory) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return ""
}
func (config *BackendConfig) GetCachesDirectory() string {
	if len(config.CachesDirectory) < maxUint16 &&
		len(config.CachesDirectory) > 0 {
		return config.CachesDirectory
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.CachesDirectory) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return ""
}
func (config *BackendConfig) GetDbEngine() string {
	if config.DbEngine == "sqlite" || config.DbEngine == "mysql" {
		return config.DbEngine
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.DbEngine) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return ""
}
func (config *BackendConfig) GetDbIp() string {
	if len(config.DbIp) < maxLocationSize &&
		len(config.DbIp) > 0 {
		return config.DbIp
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.DbIp) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return ""
}
func (config *BackendConfig) GetDbPort() uint16 {
	if config.DbPort < maxUint16 && config.DbPort > 0 {
		return config.DbPort
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.DbPort) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}
func (config *BackendConfig) GetDbUsername() string {
	if len(config.DbUsername) < maxUint8 &&
		len(config.DbUsername) > 0 {
		return config.DbUsername
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.DbUsername) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return ""
}
func (config *BackendConfig) GetDbPassword() string {
	if len(config.DbPassword) < maxUint8 &&
		len(config.DbPassword) > 0 {
		return config.DbPassword
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.DbPassword) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return ""
}

func (config *BackendConfig) GetMetricsLevel() uint8 {
	if config.MetricsLevel == 0 || config.MetricsLevel == 1 { // 0: no metrics, 1: anonymous metrics
		return config.MetricsLevel
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.MetricsLevel) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}

func (config *BackendConfig) GetMetricsToken() string {
	if len(config.MetricsToken) < 65 &&
		len(config.MetricsToken) >= 0 {
		return config.MetricsToken
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.MetricsToken) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return ""
}

func (config *BackendConfig) GetBackendKeyPair() *ecdsa.PrivateKey {
	keyPair, err := x509.ParseECPrivateKey(config.BackendKeyPair)
	if err != nil {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v, Error: %#v ", config.BackendKeyPair, err) + " Trace: " + trace()))
	}
	return keyPair
}

func (config *BackendConfig) GetMarshaledBackendPublicKey() string {
	return config.MarshaledBackendPublicKey
}

func (config *BackendConfig) GetAllowUnsignedEntities() bool {
	return config.AllowUnsignedEntities
}

/*****************************************************************************/

// Setters

func (config *BackendConfig) SetLocalMemoryDays(val int) error {
	if val > 0 {
		config.LocalMemoryDays = uint(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetNetworkMemoryDays(val int) error {
	if val > 0 {
		config.NetworkMemoryDays = uint(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetNetworkHeadDays(val int) error {
	if val > 0 {
		config.NetworkHeadDays = uint(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetLastCacheGenerationTimestamp(val int64) error {
	if val > 0 {
		config.LastCacheGenerationTimestamp = uint64(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetEntityPageSizes(val EntityPageSizes) error {

	if val.Boards < maxAbsolutePageSize &&
		val.Boards > 0 &&
		val.BoardIndexes < maxAbsolutePageSize &&
		val.BoardIndexes > 0 &&
		val.Threads < maxAbsolutePageSize &&
		val.Threads > 0 &&
		val.ThreadIndexes < maxAbsolutePageSize &&
		val.ThreadIndexes > 0 &&
		val.Posts < maxAbsolutePageSize &&
		val.Posts > 0 &&
		val.PostIndexes < maxAbsolutePageSize &&
		val.PostIndexes > 0 &&
		val.Keys < maxAbsolutePageSize &&
		val.Keys > 0 &&
		val.KeyIndexes < maxAbsolutePageSize &&
		val.KeyIndexes > 0 &&
		val.Votes < maxAbsolutePageSize &&
		val.Votes > 0 &&
		val.VoteIndexes < maxAbsolutePageSize &&
		val.VoteIndexes > 0 &&
		val.Truststates < maxAbsolutePageSize &&
		val.Truststates > 0 &&
		val.TruststateIndexes < maxAbsolutePageSize &&
		val.TruststateIndexes > 0 {
		config.EntityPageSizes = val
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetMinimumPoWStrengths(powStr int) error {
	var mps MinimumPoWStrengths
	if powStr > 4 && powStr < maxPOWStrength {
		mps.Board = powStr
		mps.BoardUpdate = powStr
		mps.Thread = powStr
		mps.ThreadUpdate = powStr
		mps.Post = powStr
		mps.PostUpdate = powStr
		mps.Vote = powStr
		mps.VoteUpdate = powStr
		mps.Key = powStr
		mps.KeyUpdate = powStr
		mps.Truststate = powStr
		mps.TruststateUpdate = powStr
		config.MinimumPoWStrengths = mps
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		} else {
			return invalidDataError(fmt.Sprintf("%#v", powStr) + " Trace: " + trace())
		}
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetPoWBailoutTimeSeconds(val int) error {
	if val > 0 {
		config.PoWBailoutTimeSeconds = uint(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetTimeBlockSizeMinutes(val int) error {
	if val > 0 {
		config.TimeBlockSizeMinutes = uint(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetPastBlocksToCheck(val int) error {
	if val > 0 {
		config.PastBlocksToCheck = uint(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetCacheGenerationIntervalHours(val int) error {
	if val > 0 {
		config.CacheGenerationIntervalHours = uint(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetClientVersionMajor(val int) error {
	if val > 0 && val < maxUint8 {
		config.ClientVersionMajor = uint8(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetClientVersionMinor(val int) error {
	if val >= 0 && val < maxUint16 {
		config.ClientVersionMinor = uint16(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetClientVersionPatch(val int) error {
	if val >= 0 && val < maxUint16 {
		config.ClientVersionPatch = uint16(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetClientName(val string) error {
	if len(val) > 0 {
		config.ClientName = val
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetProtocolVersionMajor(val int) error {
	if val > 0 && val < maxUint8 {
		config.ProtocolVersionMajor = uint8(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetProtocolVersionMinor(val int) error {
	if val >= 0 && val < maxUint16 {
		config.ProtocolVersionMinor = uint16(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetPOSTResponseExpiryMinutes(val int) error {
	if val >= 0 {
		config.POSTResponseExpiryMinutes = uint(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetConnectionTimeout(val time.Duration) error {
	if val >= 1*time.Second { // Any value under is probably an attack.
		config.ConnectionTimeout = val
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetTCPConnectTimeout(val time.Duration) error {
	if val >= 1*time.Second { // Any value under is probably an attack.
		config.TCPConnectTimeout = val
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetTLSHandshakeTimeout(val time.Duration) error {
	if val >= 1*time.Second { // Any value under is probably an attack.
		config.TLSHandshakeTimeout = val
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetPingerPageSize(val int) error {
	if val > 0 {
		config.PingerPageSize = uint(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetOnlineAddressFinderPageSize(val int) error {
	if val > 0 {
		config.OnlineAddressFinderPageSize = uint(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetDispatchExclusionExpiryForLiveAddress(val time.Duration) error {
	if val >= 1*time.Microsecond { // TODO THIS IS NORMALLY A MINUTE Any value under is probably an attack.
		config.DispatchExclusionExpiryForLiveAddress = val
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetDispatchExclusionExpiryForStaticAddress(val time.Duration) error {
	if val >= 1*time.Minute { // Any value under is probably an attack.
		config.DispatchExclusionExpiryForStaticAddress = val
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetLoggingLevel(val int) error {
	if val >= 0 {
		config.LoggingLevel = uint(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetExternalIp(val string) error {
	if len(val) > 0 && len(val) < maxLocationSize {
		config.ExternalIp = val
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetLastStaticAddressConnectionTimestamp(val int64) error {
	if val > 0 {
		config.LastStaticAddressConnectionTimestamp = uint64(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetLastLiveAddressConnectionTimestamp(val int64) error {
	if val > 0 {
		config.LastLiveAddressConnectionTimestamp = uint64(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetInitialised(val bool) error {
	config.Initialised = true
	commitErr := config.Commit()
	if commitErr != nil {
		return commitErr
	}
	return nil
}
func (config *BackendConfig) SetServingSubprotocols(subprotocols []interface{}) error {
	var castSubprots []SubprotocolShim
	for _, val := range subprotocols {
		item, ok := val.(SubprotocolShim)
		if !ok {
			return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
		}
		castSubprots = append(castSubprots, item)
	}
	config.ServingSubprotocols = castSubprots
	commitErr := config.Commit()
	if commitErr != nil {
		return commitErr
	}
	return nil
}
func (config *BackendConfig) SetExternalIpType(val int) error {
	if val == 6 || val == 4 || val == 3 {
		config.ExternalIpType = uint8(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetNodeId(val string) error {
	if len(val) == 64 {
		config.NodeId = val
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetExternalPort(val int) error {
	if val > 0 && val < maxUint16 {
		config.ExternalPort = uint16(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetUserDirectory(val string) error {
	if len(val) > 0 && len(val) < maxUint16 {
		config.UserDirectory = val
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetCachesDirectory(val string) error {
	if len(val) > 0 && len(val) < maxUint16 {
		config.CachesDirectory = val
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetDbEngine(val string) error {
	if val == "mysql" || val == "sqlite" {
		config.DbEngine = val
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetDbIp(val string) error {
	if len(val) > 0 && len(val) < maxLocationSize {
		config.DbIp = val
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetDbPort(val int) error {
	if val > 0 && val < maxUint16 {
		config.DbPort = uint16(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetDbUsername(val string) error {
	if len(val) > 0 && len(val) < maxUint8 {
		config.DbUsername = val
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}
func (config *BackendConfig) SetDbPassword(val string) error {
	if len(val) > 0 && len(val) < maxUint8 {
		config.DbPassword = val
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}

func (config *BackendConfig) SetMetricsLevel(val int) error {
	if val == 0 || val == 1 {
		config.MetricsLevel = uint8(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}

func (config *BackendConfig) SetMetricsToken(val string) error {
	if len(val) >= 0 && len(val) < 65 {
		config.MetricsToken = val
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}

func (config *BackendConfig) SetBackendKeyPair(val *ecdsa.PrivateKey) error {
	derEncodedKeyPair, err := x509.MarshalECPrivateKey(val)
	if err != nil {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	config.BackendKeyPair = derEncodedKeyPair
	commitErr := config.Commit()
	if commitErr != nil {
		return commitErr
	}
	return nil
}

// The only way to set this is to set backend key pair first.
func (config *BackendConfig) SetMarshaledBackendPublicKey(val *ecdsa.PrivateKey) error {
	config.MarshaledBackendPublicKey = hex.EncodeToString(elliptic.Marshal(
		elliptic.P521(), val.X, val.Y))
	commitErr := config.Commit()
	if commitErr != nil {
		return commitErr
	}
	return nil
}

func (config *BackendConfig) SetAllowUnsignedEntities(val bool) error {
	config.AllowUnsignedEntities = val
	commitErr := config.Commit()
	if commitErr != nil {
		return commitErr
	}
	return nil
}

/*****************************************************************************/

// BlankCheck looks at all variables and if it finds they're at their zero value, sets the default value for it. This is a guard against a new item being added to the config store as a result of a version update, but it being zero value. If a zero'd value is found, we change it to its default before anything else happens. This also effectively runs at the first pass to set the defaults.

func (config *BackendConfig) BlankCheck() {
	if config.NetworkHeadDays == 0 {
		config.SetNetworkHeadDays(defaultNetworkHeadDays)
	}
	if config.NetworkMemoryDays == 0 {
		config.SetNetworkMemoryDays(defaultNetworkMemoryDays)
	}
	if config.LocalMemoryDays == 0 {
		config.SetLocalMemoryDays(defaultLocalMemoryDays)
	}
	// ::LastCacheGenerationTimestamp: can be zero, no need to blank check.
	if config.MinimumPoWStrengths.Board == 0 ||
		config.MinimumPoWStrengths.BoardUpdate == 0 ||
		config.MinimumPoWStrengths.Thread == 0 ||
		config.MinimumPoWStrengths.ThreadUpdate == 0 ||
		config.MinimumPoWStrengths.Post == 0 ||
		config.MinimumPoWStrengths.PostUpdate == 0 ||
		config.MinimumPoWStrengths.Vote == 0 ||
		config.MinimumPoWStrengths.VoteUpdate == 0 ||
		config.MinimumPoWStrengths.Key == 0 ||
		config.MinimumPoWStrengths.KeyUpdate == 0 ||
		config.MinimumPoWStrengths.Truststate == 0 ||
		config.MinimumPoWStrengths.TruststateUpdate == 0 {
		config.SetMinimumPoWStrengths(defaultPowStrength)
	}
	if config.EntityPageSizes.Boards == 0 ||
		config.EntityPageSizes.BoardIndexes == 0 ||
		config.EntityPageSizes.Threads == 0 ||
		config.EntityPageSizes.ThreadIndexes == 0 ||
		config.EntityPageSizes.Posts == 0 ||
		config.EntityPageSizes.PostIndexes == 0 ||
		config.EntityPageSizes.Votes == 0 ||
		config.EntityPageSizes.VoteIndexes == 0 ||
		config.EntityPageSizes.Keys == 0 ||
		config.EntityPageSizes.KeyIndexes == 0 ||
		config.EntityPageSizes.Truststates == 0 ||
		config.EntityPageSizes.TruststateIndexes == 0 ||
		config.EntityPageSizes.Addresses == 0 ||
		config.EntityPageSizes.AddressIndexes == 0 {
		config.setDefaultEntityPageSizes()
	}
	if config.PoWBailoutTimeSeconds == 0 {
		config.SetPoWBailoutTimeSeconds(defaultPoWBailoutTimeSeconds)
	}
	if config.TimeBlockSizeMinutes == 0 {
		config.SetTimeBlockSizeMinutes(defaultTimeBlockSizeMinutes)
	}
	if config.PastBlocksToCheck == 0 {
		config.SetPastBlocksToCheck(defaultPastBlocksToCheck)
	}
	if config.CacheGenerationIntervalHours == 0 {
		config.SetCacheGenerationIntervalHours(defaultCacheGenerationIntervalHours)
	}
	if config.ClientVersionMajor == 0 {
		config.SetClientVersionMajor(clientVersionMajor)
	}
	if config.ClientVersionMinor != clientVersionMinor {
		config.SetClientVersionMinor(clientVersionMinor)
	}
	if config.ClientVersionPatch != clientVersionPatch {
		config.SetClientVersionPatch(clientVersionPatch)
	}
	if config.ClientName == "" || config.ClientName != clientName {
		config.SetClientName(clientName)
	}
	if config.ProtocolVersionMajor == 0 || config.ProtocolVersionMajor != protocolVersionMajor {
		config.SetProtocolVersionMajor(protocolVersionMajor)
	}
	if config.ProtocolVersionMinor != protocolVersionMinor {
		config.SetProtocolVersionMinor(protocolVersionMinor)
	}
	if config.POSTResponseExpiryMinutes == 0 {
		config.SetPOSTResponseExpiryMinutes(defaultPOSTResponseExpiryMinutes)
	}
	if config.ConnectionTimeout == 0 {
		config.SetConnectionTimeout(defaultConnectionTimeout)
	}
	if config.TCPConnectTimeout == 0 {
		config.SetTCPConnectTimeout(defaultTCPConnectTimeout)
	}
	if config.TLSHandshakeTimeout == 0 {
		config.SetTLSHandshakeTimeout(defaultTLSHandshakeTimeout)
	}
	if config.PingerPageSize == 0 {
		config.SetPingerPageSize(defaultPingerPageSize)
	}
	if config.OnlineAddressFinderPageSize == 0 {
		config.SetOnlineAddressFinderPageSize(defaultOnlineAddressFinderPageSize)
	}
	if config.DispatchExclusionExpiryForLiveAddress == 0 {
		config.SetDispatchExclusionExpiryForLiveAddress(defaultDispatchExclusionExpiryForLiveAddress)
	}
	if config.DispatchExclusionExpiryForStaticAddress == 0 {
		config.SetDispatchExclusionExpiryForStaticAddress(defaultDispatchExclusionExpiryForStaticAddress)
	}
	// ::LoggingLevel: can be zero, no need to blank check.
	// if config.LoggingLevel == 0 {
	// 	config.SetLoggingLevel(2)
	// }
	if config.ExternalIp == "" {
		config.SetExternalIp(defaultExternalIp)
	}
	if config.ExternalIpType == 0 {
		config.SetExternalIpType(defaultExternalIpType)
	}
	if config.ExternalPort == 0 {
		config.SetExternalPort(defaultExternalPort)
	}
	// ::LastStaticAddressConnectionTimestamp: can be zero, no need to blank check.
	// ::LastLiveAddressConnectionTimestamp: can be zero, no need to blank check.
	var servingSubprotocolsNeedRegeneration bool
	if len(config.ServingSubprotocols) == 0 {
		servingSubprotocolsNeedRegeneration = true
	} else {
		for _, val := range config.ServingSubprotocols {
			if len(val.SupportedEntities) == 0 {
				servingSubprotocolsNeedRegeneration = true
			}
		}
	}
	if servingSubprotocolsNeedRegeneration {
		c0 := SubprotocolShim{Name: "c0", VersionMajor: 1, VersionMinor: 0, SupportedEntities: []string{"board", "thread", "post", "vote", "key", "truststate"}}
		dweb := SubprotocolShim{Name: "dweb", VersionMajor: 1, VersionMinor: 0, SupportedEntities: []string{"page"}}
		config.SetServingSubprotocols([]interface{}{c0, dweb})
	}
	if len(config.UserDirectory) == 0 {
		config.SetUserDirectory(cdir.New(Btc.OrgIdentifier, Btc.AppIdentifier).QueryFolders(cdir.Global)[0].Path)
	}
	if len(config.CachesDirectory) == 0 {
		config.SetCachesDirectory(cdir.New(Btc.OrgIdentifier, Btc.AppIdentifier).QueryCacheFolder().Path)
	}
	if !config.Initialised {
		config.SetInitialised(true)
	}
	if len(config.DbEngine) == 0 {
		config.SetDbEngine(defaultDbEngine)
	}
	if len(config.DbIp) == 0 {
		config.SetDbIp(defaultDBIp)
	}
	if config.DbPort == 0 {
		config.SetDbPort(defaultDbPort)
	}
	if len(config.DbUsername) == 0 {
		config.SetDbUsername(defaultDbUsername)
	}
	if len(config.DbPassword) == 0 {
		config.SetDbPassword(defaultDbPassword)
	}
	// ::MetricsLevel: can be zero, no need to blank check.
	// ::MetricsToken: can be zero, no need to blank check.
	if len(config.BackendKeyPair) == 0 {
		privKey, _ := signaturing.CreateKeyPair()
		config.SetBackendKeyPair(privKey)
	}
	// This needs to be after Backend key pair generation.
	if len(config.MarshaledBackendPublicKey) == 0 {
		config.SetMarshaledBackendPublicKey(config.GetBackendKeyPair())
	}
	// This needs to be after key pair generation, because it uses the key pair. Node Id is the Fingerprint of the public key of the backend.
	if config.NodeId == "" {
		nodeid := fingerprinting.Create(config.GetMarshaledBackendPublicKey())
		config.SetNodeId(nodeid)
	}
	// ::AllowUnsignedEntities: can be false, no need to blank check.
}

/*
Backend config sanity check.Everything you add to above, needs to also be added to the sanity check. This runs at the initialisation at the beginning of the program, and it checks that the values actually make sense. Sanity checks also run on gets and sets, but they don't normally run at startup. This function covers that base.
*/
func (config *BackendConfig) SanityCheck() {
	if !config.GetInitialised() {
		log.Fatal("Backend configuration is not initialised. Please initialise it before use.")
	} else {
		// If there is an error, the appropriate getter function will fail and crash the app.
		config.GetLocalMemoryDays()
		config.GetNetworkMemoryDays()
		config.GetNetworkHeadDays()
		config.GetLastCacheGenerationTimestamp()
		config.GetEntityPageSizes()
		config.GetMinimumPoWStrengths()
		config.GetPoWBailoutTimeSeconds()
		config.GetTimeBlockSizeMinutes()
		config.GetPastBlocksToCheck()
		config.GetCacheGenerationIntervalHours()
		config.GetClientVersionMajor()
		config.GetClientVersionMinor()
		config.GetClientVersionPatch()
		config.GetClientName()
		config.GetProtocolVersionMajor()
		config.GetProtocolVersionMinor()
		config.GetPOSTResponseExpiryMinutes()
		config.GetConnectionTimeout()
		config.GetTCPConnectTimeout()
		config.GetTLSHandshakeTimeout()
		config.GetPingerPageSize()
		config.GetOnlineAddressFinderPageSize()
		config.GetDispatchExclusionExpiryForLiveAddress()
		config.GetDispatchExclusionExpiryForStaticAddress()
		config.GetLoggingLevel()
		config.GetExternalIp()
		config.GetLastStaticAddressConnectionTimestamp()
		config.GetLastLiveAddressConnectionTimestamp()
		config.GetServingSubprotocols()
		config.GetDbEngine()
		config.GetDbIp()
		config.GetDbPort()
		config.GetDbPassword()
		config.GetMetricsLevel()
		config.GetMetricsToken()
		config.GetBackendKeyPair()
		// Below are location sensitive. Needs to happen after Backend Key Pair generation (above).
		config.GetMarshaledBackendPublicKey()
		config.GetNodeId()
	}
}

/*
Commit saves the file to memory. This is usually called after a Set operation.
*/
func (config *BackendConfig) Commit() error {
	if Btc.PermConfigReadOnly {
		return nil
	}
	Btc.ConfigMutex.Lock()
	defer Btc.ConfigMutex.Unlock()
	confAsByte, err3 := json.MarshalIndent(config, "", "    ")
	if err3 != nil {
		log.Fatal(fmt.Sprintf("JSON marshaler encountered an error while marshaling this config into JSON. Config: %#v, Error: %#v", config, err3))
	}
	configDirs := cdir.New(Btc.OrgIdentifier, Btc.AppIdentifier)
	folders := configDirs.QueryFolders(cdir.Global)
	err := folders[0].WriteFile("backend_config.json", confAsByte)
	if err != nil {
		return err
	}
	return nil
}

// Cycle commits the whole struct into memory, generating fields in JSON that were newly added.
func (config *BackendConfig) Cycle() error {
	err := config.Commit()
	if err != nil {
		return err
	}
	return nil
}

// The default base size is 1x (The thread size). At the base size, a page gets 100 entries.
func (config *BackendConfig) setDefaultEntityPageSizes() {
	var eps EntityPageSizes
	eps.Boards = defaultBoardsPageSize
	eps.BoardIndexes = defaultBoardIndexesPageSize
	eps.Threads = defaultThreadsPageSize
	eps.ThreadIndexes = defaultThreadIndexesPageSize
	eps.Posts = defaultPostsPageSize
	eps.PostIndexes = defaultPostIndexesPageSize
	eps.Votes = defaultVotesPageSize
	eps.VoteIndexes = defaultVoteIndexesPageSize
	eps.Addresses = defaultAddressesPageSize
	eps.AddressIndexes = defaultAddressIndexesPageSize
	eps.Keys = defaultKeysPageSize
	eps.KeyIndexes = defaultKeyIndexesPageSize
	eps.Truststates = defaultTruststatesPageSize
	eps.TruststateIndexes = defaultTruststateIndexesPageSize
	config.SetEntityPageSizes(eps)
}

// ===========================================

// 2) FRONTEND

// Frontend config base
type FrontendConfig struct {
	UserKeyPair  []byte
	Initialised  bool   // False by default, init to set true
	MetricsLevel uint8  // 0: no metrics transmitted
	MetricsToken string // If metrics level is not zero, metrics token is the anonymous identifier for the metrics server. Resetting this to 0 makes this node behave like a new node as far as metrics go, but if you don't want metrics to be collected, you can set it through the application or set the metrics level to zero in the JSON settings file.
}

// Getters and setters

// Getters

func (config *FrontendConfig) GetUserKeyPair() *ecdsa.PrivateKey {
	keyPair, err := x509.ParseECPrivateKey(config.UserKeyPair)
	if err != nil {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v, Error: %#v ", config.UserKeyPair, err) + " Trace: " + trace()))
	}
	return keyPair
}

func (config *FrontendConfig) GetInitialised() bool {
	return config.Initialised
}

func (config *FrontendConfig) GetMetricsLevel() uint8 {
	if config.MetricsLevel == 0 || config.MetricsLevel == 1 { // 0: no metrics, 1: anonymous metrics
		return config.MetricsLevel
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.MetricsLevel) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return 0
}

func (config *FrontendConfig) GetMetricsToken() string {
	if len(config.MetricsToken) < 65 &&
		len(config.MetricsToken) >= 0 {
		return config.MetricsToken
	} else {
		log.Fatal(invalidDataError(fmt.Sprintf("%#v", config.MetricsToken) + " Trace: " + trace()))
	}
	log.Fatal("This should never happen." + trace())
	return ""
}

/*****************************************************************************/

// Setters

func (config *FrontendConfig) SetUserKeyPair(val *ecdsa.PrivateKey) error {
	derEncodedKeyPair, err := x509.MarshalECPrivateKey(val)
	if err != nil {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	config.UserKeyPair = derEncodedKeyPair
	commitErr := config.Commit()
	if commitErr != nil {
		return commitErr
	}
	return nil
}

func (config *FrontendConfig) SetInitialised(val bool) error {
	config.Initialised = val
	commitErr := config.Commit()
	if commitErr != nil {
		return commitErr
	}
	return nil
}

func (config *FrontendConfig) SetMetricsLevel(val int) error {
	if val == 0 || val == 1 {
		config.MetricsLevel = uint8(val)
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}

func (config *FrontendConfig) SetMetricsToken(val string) error {
	if len(val) >= 0 && len(val) < 65 {
		config.MetricsToken = val
		commitErr := config.Commit()
		if commitErr != nil {
			return commitErr
		}
		return nil
	} else {
		return invalidDataError(fmt.Sprintf("%#v", val) + " Trace: " + trace())
	}
	log.Fatal("This should never happen." + trace())
	return nil
}

/*****************************************************************************/

// Frontend config methods

func (config *FrontendConfig) BlankCheck() {
	if len(config.UserKeyPair) == 0 {
		privKey, _ := signaturing.CreateKeyPair()
		config.SetUserKeyPair(privKey)
	}
	if !config.Initialised {
		config.SetInitialised(true)
	}
	// ::MetricsLevel: can be zero, no need to blank check.
	// ::MetricsToken: can be zero, no need to blank check.
}
func (config *FrontendConfig) SanityCheck() {
	if !config.GetInitialised() {
		log.Fatal("Frontend configuration is not initialised. Please initialise it before use.")
	} else {
		config.GetUserKeyPair()
		config.GetMetricsLevel()
		config.GetMetricsToken()
	}
}

/*
Commit saves the file to memory. This is usually called after a Set operation.
*/
func (config *FrontendConfig) Commit() error {
	if Ftc.PermConfigReadOnly {
		return nil
	}
	Ftc.ConfigMutex.Lock()
	defer Ftc.ConfigMutex.Unlock()
	confAsByte, err3 := json.MarshalIndent(config, "", "    ")
	if err3 != nil {
		log.Fatal(fmt.Sprintf("JSON marshaler encountered an error while marshaling this config into JSON. Config: %#v, Error: %#v", config, err3))
	}
	configDirs := cdir.New(Btc.OrgIdentifier, Btc.AppIdentifier)
	folders := configDirs.QueryFolders(cdir.Global)
	err := folders[0].WriteFile("frontend_config.json", confAsByte)
	if err != nil {
		return err
	}
	return nil
}

// Cycle commits the whole struct into memory, generating fields in JSON that were newly added.
func (config *FrontendConfig) Cycle() error {
	err := config.Commit()
	if err != nil {
		return err
	}
	return nil
}

/*****************************************************************************/

// 3) CONFIG METHODS

/*
EstablishBackendConfig establishes the connection with the config file, and makes it available as an object to the rest of the application.
*/
func EstablishBackendConfig() (*BackendConfig, error) {
	var config BackendConfig
	configDirs := cdir.New(Btc.OrgIdentifier, Btc.AppIdentifier)
	folder := configDirs.QueryFolderContainsFile("backend_config.json")
	if folder != nil {
		configJson, _ := folder.ReadFile("backend_config.json")
		err := json.Unmarshal(configJson, &config)
		if err != nil || fmt.Sprintf("%#v", string(configJson)) == "\"{}\"" {
			return &config, errors.New(fmt.Sprintf("Back-end configuration file is corrupted. Please fix the configuration file, or delete it. If deleted a new configuration will be generated with default values. Error: %#v, ConfigJson: %#v", err, string(configJson)))
		}
	}
	// Folder is nil - the configuration file in question does not exist. Ask to create.
	config.BlankCheck()
	config.SanityCheck()
	return &config, nil
}

/*
EstablishFrontendConfig establishes the connection with the config file, and makes it available as an object to the rest of the application.
*/
func EstablishFrontendConfig() (*FrontendConfig, error) {
	var config FrontendConfig
	configDirs := cdir.New(Btc.OrgIdentifier, Btc.AppIdentifier)
	folder := configDirs.QueryFolderContainsFile("frontend_config.json")
	if folder != nil {
		configJson, _ := folder.ReadFile("frontend_config.json")
		err := json.Unmarshal(configJson, &config)
		if err != nil || fmt.Sprintf("%#v", string(configJson)) == "\"{}\"" {
			return &config, errors.New(fmt.Sprintf("Front-end configuration file is corrupted. Please fix the configuration file, or delete it. If deleted a new configuration will be generated with default values. Error: %#v, ConfigJson: %#v", err, string(configJson)))
		}
	}
	// Folder is nil - the configuration file in question does not exist. Ask to create.
	config.BlankCheck()
	config.SanityCheck()
	return &config, nil
}

/*****************************************************************************/

// TRANSIENT CONFIG

// These are the items that are set in runtime, and do not change until the application closes. This is different from the application state in the way that they're set-once for the runtime.

// These do not have getters and setters.

var Btc BackendTransientConfig
var Ftc FrontendTransientConfig

// Backend

// Default entity versions for this version of the app. This is not user adjustable.

const (
	defaultBoardEntityVersion      = 1
	defaultThreadEntityVersion     = 1
	defaultPostEntityVersion       = 1
	defaultVoteEntityVersion       = 1
	defaultKeyEntityVersion        = 1
	defaultTruststateEntityVersion = 1
	defaultAddressEntityVersion    = 1
)

type EntityVersions struct {
	Board      int
	Thread     int
	Post       int
	Vote       int
	Key        int
	Truststate int
	Address    int
}

/*
#### NONCOMMITTED ITEMS

## PermConfigReadOnly
When enabled, this prevents anything from saved into the config. This value itself is NOT saved into the config, so when the application restarts, this value is reset to false. This is useful in the case that you provide flags to the executable, but you don't want the values in the flags to be permanently saved into the config file. Any flags being provided into the executable will set this to true, therefore any runs with flags will effectively treat the config as read-only.

## AppIdentifier
This is the name of the app as registered to the operating system. This is useful to have here, because what we can do is we can vary this number in the swarm testing (petridish) and each of these nodes will act like a network in a single local machine, each with their own databases and different config files.

## OrgIdentifier
Same as above, but it's probably best to keep it under the same org name just to keep the local machine clean.

## PrintToStdout
This is useful because the logging things the normal kind does not pass the output to the swarm test orchestrator. This flag being enabled routes the logs to stdout so that the orchestrator can show it.

## MetricsDebugMode
This being enabled temporarily makes this node send much more detailed metrics more frequently, so that network connectivity issues can be debugged. This is a transient config on purpose, so that this cannot be enabled permanently. If a frontend connects to a backend with debug mode enabled, it has to show a warning to its user that says this backend node has debugging enabled, and only connect if the user agrees. Mind that the backend doesn't have to be truthful about whether it has the debug mode on. Having this mode on does not immediately compromise the frontend's privacy / identity, but the longer the frontend stays on that backend and the more actions a user commits, the higher the likelihood.

## ExternalPortVerified
Whether the port that was in the config was actually checked to be free and clear. This is important because we'll check once before the server starts to run, and when it starts, that port will no longer be available, and will start to return 'not available'. That will make all subsequent checks fail and that will trigger the port to be moved to a port that is free - but not bound to any server, since the server is bound to the old port, and that in fact is the reason the checks return false.

## SwarmNodeId
This is the number that this specific node will route to the main swarm orchestrator when it's reporting logs. Make sure that the App identifier (Usually in the format of "Aether-N") matches this number N, or it can be confusing.

## ShutdownInitiated
This is set when the shutdown of the backend service is initiated. The processes that take a long time to return should be checking this value periodically, and if it is set, they should stop whatever they're doing and do a graceful shutdown.

## DispatcherExclusions
This is the temporary exclusions for the dispatcher. When you connect to a node, that node is placed in the exclusions list for a while, so that you don't repeatedly keep connecting back to that node again.

## StopStaticDispatcherCycle
This is the channel to send the message to when you want to stop the static dispatcher repeated task.

## StopAddressScannerCycle
This is the channel to send the message to when you want to stop the address scanner repeated task.

## StopUPNPCycle
This is the channel to send the message to when you want to stop the UPNP mapper repeated task.

## StopCacheGenerationCycle
This is the channel to send the message to when you want to stop the cache generator repeated task.

## AddressesScannerActive
This is the mutex that gets activated when the address scanner is active, so that it cannot be triggered twice at the same time.

## LiveDispatchRunning
This is the mutex that gets activated when the live dispatcher is active, so that it cannot be triggered twice at the same time.

## StaticDispatchRunning
This is the mutex that gets activated when the static dispatcher is active, so that it cannot be triggered twice at the same time.

## CurrentMetricsPage
This is the current metrics struct that we are building to send to the metrics server, if enabled.

## ConfigMutex
This is the mutex that prevents configuration from being written from multiple places.

## FingerprintCheckEnabled
Determines whether the entities coming over from the wire are fingerprint-checked for integrity.

## SignatureCheckEnabled
Determines whether the entities coming over from the wire are signature-checked for ownership.

## ProofOfWorkCheckEnabled
Determines whether the entities coming over from the wire are PoW-checked for anti-spam.

## PageSignatureCheckEnabled
Determines whether the pages (entity containers) coming over from the wire are signature-checked for integrity.
*/

type BackendTransientConfig struct {
	PermConfigReadOnly        bool
	AppIdentifier             string
	OrgIdentifier             string
	PrintToStdout             bool
	MetricsDebugMode          bool
	TooManyConnections        bool
	ExternalPortVerified      bool
	SwarmNodeId               int
	ShutdownInitiated         bool
	DispatcherExclusions      map[*interface{}]time.Time
	StopLiveDispatcherCycle   chan bool
	StopStaticDispatcherCycle chan bool
	StopAddressScannerCycle   chan bool
	StopUPNPCycle             chan bool
	StopCacheGenerationCycle  chan bool
	AddressesScannerActive    bool
	LiveDispatchRunning       bool
	StaticDispatchRunning     bool
	CurrentMetricsPage        pb.Metrics
	ConfigMutex               *sync.Mutex
	FingerprintCheckEnabled   bool
	SignatureCheckEnabled     bool
	ProofOfWorkCheckEnabled   bool
	PageSignatureCheckEnabled bool
	EntityVersions            EntityVersions
}

// Set transient backend config defaults. Only need to set defaults that are not the type default.

func (config *BackendTransientConfig) SetDefaults() {
	config.AppIdentifier = "Aether"
	config.OrgIdentifier = "Air Labs"
	config.ConfigMutex = &sync.Mutex{}
	config.FingerprintCheckEnabled = true
	config.SignatureCheckEnabled = true
	config.ProofOfWorkCheckEnabled = true
	config.PageSignatureCheckEnabled = true
	ev := EntityVersions{
		Board:      defaultBoardEntityVersion,
		Thread:     defaultThreadEntityVersion,
		Post:       defaultPostEntityVersion,
		Vote:       defaultVoteEntityVersion,
		Key:        defaultKeyEntityVersion,
		Truststate: defaultTruststateEntityVersion,
		Address:    defaultAddressEntityVersion,
	}
	config.EntityVersions = ev
}

// Frontend

type FrontendTransientConfig struct {
	PermConfigReadOnly bool
	MetricsDebugMode   bool
	ConfigMutex        *sync.Mutex
}

// Set transient frontend config defaults

func (config *FrontendTransientConfig) SetDefaults() {
	config.PermConfigReadOnly = false
	config.ConfigMutex = &sync.Mutex{}
}
