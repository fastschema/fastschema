package fs

type WSClient interface {
	ID() string
	Write(message []byte, messageTypes ...WSMessageType) error
	Read() (messageType WSMessageType, message []byte, err error)
	Close(msgs ...string) error
	IsCloseNormal(err error) bool
}

type WSCloseType int

const (
	WSCloseTypeInvalid WSCloseType = iota + 999
	WSCloseNormalClosure
	WSCloseGoingAway
	WSCloseProtocolError
	WSCloseUnsupportedData
	WSCloseSkipReserved
	WSCloseNoStatusReceived
	WSCloseAbnormalClosure
	WSCloseInvalidFramePayloadData
	WSClosePolicyViolation
	WSCloseMessageTooBig
	WSCloseMandatoryExtension
	WSCloseInternalServerErr
	WSCloseServiceRestart
	WSCloseTryAgainLater
	WSCloseSkipBadGateway
	WSCloseTLSHandshake
	endWSCloseTypes
)

var (
	wsCloseTypeToInts = [...]int{
		WSCloseTypeInvalid:             999,
		WSCloseNormalClosure:           1000,
		WSCloseGoingAway:               1001,
		WSCloseProtocolError:           1002,
		WSCloseUnsupportedData:         1003,
		WSCloseSkipReserved:            1004,
		WSCloseNoStatusReceived:        1005,
		WSCloseAbnormalClosure:         1006,
		WSCloseInvalidFramePayloadData: 1007,
		WSClosePolicyViolation:         1008,
		WSCloseMessageTooBig:           1009,
		WSCloseMandatoryExtension:      1010,
		WSCloseInternalServerErr:       1011,
		WSCloseServiceRestart:          1012,
		WSCloseTryAgainLater:           1013,
		WSCloseSkipBadGateway:          1014,
		WSCloseTLSHandshake:            1015,
	}

	intToWSCloseTypes = map[int]WSCloseType{
		999:  WSCloseTypeInvalid,
		1000: WSCloseNormalClosure,
		1001: WSCloseGoingAway,
		1002: WSCloseProtocolError,
		1003: WSCloseUnsupportedData,
		1004: WSCloseSkipReserved,
		1005: WSCloseNoStatusReceived,
		1006: WSCloseAbnormalClosure,
		1007: WSCloseInvalidFramePayloadData,
		1008: WSClosePolicyViolation,
		1009: WSCloseMessageTooBig,
		1010: WSCloseMandatoryExtension,
		1011: WSCloseInternalServerErr,
		1012: WSCloseServiceRestart,
		1013: WSCloseTryAgainLater,
		1014: WSCloseSkipBadGateway,
		1015: WSCloseTLSHandshake,
	}
)

func WSCloseTypeFromInt(i int) WSCloseType {
	if t, ok := intToWSCloseTypes[i]; ok {
		return t
	}
	return WSCloseTypeInvalid
}

// Int returns the int value of the type
func (t WSCloseType) Int() int {
	if t < endWSCloseTypes {
		return wsCloseTypeToInts[t]
	}
	return wsCloseTypeToInts[WSCloseTypeInvalid]
}

// Valid reports if the given type if known type.
func (t WSCloseType) Valid() bool {
	return t > WSCloseTypeInvalid && t < endWSCloseTypes
}

type WSMessageType int

const (
	WSMessageInvalid WSMessageType = 0
	WSMessageText    WSMessageType = 1
	WSMessageBinary  WSMessageType = 2
	WSMessageClose   WSMessageType = 8
	WSMessagePing    WSMessageType = 9
	WSMessagePong    WSMessageType = 10
	endWSMessageTypes
)

var (
	wsMessageTypeToInts = map[WSMessageType]int{
		WSMessageInvalid: 0,
		WSMessageText:    1,
		WSMessageBinary:  2,
		WSMessageClose:   8,
		WSMessagePing:    9,
		WSMessagePong:    10,
	}

	intToWSMessageTypes = map[int]WSMessageType{
		0:  WSMessageInvalid,
		1:  WSMessageText,
		2:  WSMessageBinary,
		8:  WSMessageClose,
		9:  WSMessagePing,
		10: WSMessagePong,
	}
)

func WSMessageTypeFromInt(i int) WSMessageType {
	if t, ok := intToWSMessageTypes[i]; ok {
		return t
	}
	return WSMessageInvalid
}

// Int returns the int value of the type
func (t WSMessageType) Int() int {
	if t.Valid() {
		return wsMessageTypeToInts[t]
	}
	return wsMessageTypeToInts[WSMessageInvalid]
}

// Valid reports if the given type if known type.
func (t WSMessageType) Valid() bool {
	intVal, ok := wsMessageTypeToInts[t]
	return ok && intVal > 0
}
