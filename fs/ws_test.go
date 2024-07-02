package fs_test

import (
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/stretchr/testify/assert"
)

func TestWSCloseTypeFromInt(t *testing.T) {
	tests := []struct {
		input int
		want  fs.WSCloseType
	}{
		{999, fs.WSCloseTypeInvalid},
		{1000, fs.WSCloseNormalClosure},
		{1001, fs.WSCloseGoingAway},
		{1002, fs.WSCloseProtocolError},
		{1003, fs.WSCloseUnsupportedData},
		{1004, fs.WSCloseSkipReserved},
		{1005, fs.WSCloseNoStatusReceived},
		{1006, fs.WSCloseAbnormalClosure},
		{1007, fs.WSCloseInvalidFramePayloadData},
		{1008, fs.WSClosePolicyViolation},
		{1009, fs.WSCloseMessageTooBig},
		{1010, fs.WSCloseMandatoryExtension},
		{1011, fs.WSCloseInternalServerErr},
		{1012, fs.WSCloseServiceRestart},
		{1013, fs.WSCloseTryAgainLater},
		{1014, fs.WSCloseSkipBadGateway},
		{1015, fs.WSCloseTLSHandshake},
		{1016, fs.WSCloseTypeInvalid},
	}

	for _, test := range tests {
		got := fs.WSCloseTypeFromInt(test.input)
		assert.Equal(t, test.want, got)
	}
}
func TestWSCloseTypeInt(t *testing.T) {
	tests := []struct {
		input fs.WSCloseType
		want  int
	}{
		{fs.WSCloseTypeInvalid, 999},
		{fs.WSCloseNormalClosure, 1000},
		{fs.WSCloseGoingAway, 1001},
		{fs.WSCloseProtocolError, 1002},
		{fs.WSCloseUnsupportedData, 1003},
		{fs.WSCloseSkipReserved, 1004},
		{fs.WSCloseNoStatusReceived, 1005},
		{fs.WSCloseAbnormalClosure, 1006},
		{fs.WSCloseInvalidFramePayloadData, 1007},
		{fs.WSClosePolicyViolation, 1008},
		{fs.WSCloseMessageTooBig, 1009},
		{fs.WSCloseMandatoryExtension, 1010},
		{fs.WSCloseInternalServerErr, 1011},
		{fs.WSCloseServiceRestart, 1012},
		{fs.WSCloseTryAgainLater, 1013},
		{fs.WSCloseSkipBadGateway, 1014},
		{fs.WSCloseTLSHandshake, 1015},
	}

	for _, test := range tests {
		got := test.input.Int()
		assert.Equal(t, test.want, got)
	}

	assert.Equal(t, fs.WSCloseType(1016).Int(), fs.WSCloseTypeInvalid.Int())
}

func TestWSCloseTypeValid(t *testing.T) {
	tests := []struct {
		input fs.WSCloseType
		want  bool
	}{
		{fs.WSCloseTypeInvalid, false},
		{fs.WSCloseNormalClosure, true},
		{fs.WSCloseGoingAway, true},
		{fs.WSCloseProtocolError, true},
		{fs.WSCloseUnsupportedData, true},
		{fs.WSCloseSkipReserved, true},
		{fs.WSCloseNoStatusReceived, true},
		{fs.WSCloseAbnormalClosure, true},
		{fs.WSCloseInvalidFramePayloadData, true},
		{fs.WSClosePolicyViolation, true},
		{fs.WSCloseMessageTooBig, true},
		{fs.WSCloseMandatoryExtension, true},
		{fs.WSCloseInternalServerErr, true},
		{fs.WSCloseServiceRestart, true},
		{fs.WSCloseTryAgainLater, true},
		{fs.WSCloseSkipBadGateway, true},
		{fs.WSCloseTLSHandshake, true},
	}

	for _, test := range tests {
		got := test.input.Valid()
		assert.Equal(t, test.want, got)
	}
}

func TestWSMessageTypeFromInt(t *testing.T) {
	tests := []struct {
		input int
		want  fs.WSMessageType
	}{
		{0, fs.WSMessageInvalid},
		{1, fs.WSMessageText},
		{2, fs.WSMessageBinary},
		{8, fs.WSMessageClose},
		{9, fs.WSMessagePing},
		{10, fs.WSMessagePong},
		{11, fs.WSMessageInvalid},
	}
	for _, test := range tests {
		got := fs.WSMessageTypeFromInt(test.input)
		assert.Equal(t, test.want, got)
	}
}

func TestWSMessageTypeInt(t *testing.T) {
	tests := []struct {
		input fs.WSMessageType
		want  int
	}{
		{fs.WSMessageInvalid, 0},
		{fs.WSMessageText, 1},
		{fs.WSMessageBinary, 2},
		{fs.WSMessageClose, 8},
		{fs.WSMessagePing, 9},
		{fs.WSMessagePong, 10},
	}
	for _, test := range tests {
		got := test.input.Int()
		assert.Equal(t, test.want, got)
	}

	assert.Equal(t, fs.WSMessageType(999).Int(), fs.WSMessageInvalid.Int())
}

func TestWSMessageTypeValid(t *testing.T) {
	tests := []struct {
		input fs.WSMessageType
		want  bool
	}{
		{fs.WSMessageInvalid, false},
		{fs.WSMessageText, true},
		{fs.WSMessageBinary, true},
		{fs.WSMessageClose, true},
		{fs.WSMessagePing, true},
		{fs.WSMessagePong, true},
	}
	for _, test := range tests {
		got := test.input.Valid()
		assert.Equal(t, test.want, got)
	}
}
