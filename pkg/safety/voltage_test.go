package safety

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVoltageCheckCodingPass(t *testing.T) {
	result := CheckVoltage(12.8, OpCoding)
	assert.Equal(t, VoltageOK, result.Status)
}

func TestVoltageCheckCodingWarning(t *testing.T) {
	result := CheckVoltage(12.3, OpCoding)
	assert.Equal(t, VoltageWarning, result.Status)
	assert.Contains(t, result.Message, "12.3")
}

func TestVoltageCheckCodingBlocked(t *testing.T) {
	result := CheckVoltage(11.8, OpCoding)
	assert.Equal(t, VoltageBlocked, result.Status)
}

func TestVoltageCheckMultiCodingBlocked(t *testing.T) {
	result := CheckVoltage(12.3, OpMultiCoding)
	assert.Equal(t, VoltageBlocked, result.Status)
}

func TestVoltageCheckMultiCodingPass(t *testing.T) {
	result := CheckVoltage(12.8, OpMultiCoding)
	assert.Equal(t, VoltageOK, result.Status)
}

func TestVoltageCheckFlashBlocked(t *testing.T) {
	result := CheckVoltage(12.8, OpFlash)
	assert.Equal(t, VoltageBlocked, result.Status)
}

func TestVoltageCheckFlashPass(t *testing.T) {
	result := CheckVoltage(13.5, OpFlash)
	assert.Equal(t, VoltageOK, result.Status)
}

func TestVoltageCheckFlashAtThreshold(t *testing.T) {
	result := CheckVoltage(13.0, OpFlash)
	assert.Equal(t, VoltageOK, result.Status)
}

func TestVoltageCheckCodingExactlyAtBlock(t *testing.T) {
	result := CheckVoltage(12.0, OpCoding)
	assert.Equal(t, VoltageOK, result.Status)
}

func TestVoltageCheckCodingJustBelowBlock(t *testing.T) {
	result := CheckVoltage(11.99, OpCoding)
	assert.Equal(t, VoltageBlocked, result.Status)
}
