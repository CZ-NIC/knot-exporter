package libknot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetVersion tests the GetVersion function
func TestGetVersion(t *testing.T) {
	// GetVersion returns a string
	version := GetVersion()
	assert.NotEmpty(t, version)
}

// TestCtlError tests the CtlError implementation
func TestCtlError(t *testing.T) {
	// Test CtlError without data
	err := &CtlError{message: "test error"}
	assert.Equal(t, "test error", err.Error())

	// Test CtlError with data
	errWithData := &CtlError{
		message: "test error with data",
		data: &CtlData{
			Section: "test-section",
			Item:    "test-item",
			Zone:    "example.com",
		},
	}
	assert.Contains(t, errWithData.Error(), "test error with data")
	assert.Contains(t, errWithData.Error(), "test-section")
	assert.Contains(t, errWithData.Error(), "example.com")

	// Test derived error types
	connectErr := &CtlErrorConnect{CtlError{message: "connection error"}}
	assert.Equal(t, "connection error", connectErr.Error())

	sendErr := &CtlErrorSend{CtlError{message: "send error"}}
	assert.Equal(t, "send error", sendErr.Error())

	receiveErr := &CtlErrorReceive{CtlError{message: "receive error"}}
	assert.Equal(t, "receive error", receiveErr.Error())

	remoteErr := &CtlErrorRemote{CtlError{message: "remote error"}}
	assert.Equal(t, "remote error", remoteErr.Error())
}

// TestCtlType tests the CtlType definitions
func TestCtlType(t *testing.T) {
	assert.Equal(t, CtlType(0), CtlTypeEnd)
	assert.Equal(t, CtlType(1), CtlTypeData)
	assert.Equal(t, CtlType(2), CtlTypeExtra)
	assert.Equal(t, CtlType(3), CtlTypeBlock)
}

// TestCtlData tests the CtlData structure
func TestCtlData(t *testing.T) {
	// Create a CtlData object
	data := CtlData{
		Section: "test-section",
		ID:      "test-id",
		Item:    "test-item",
		Zone:    "example.com",
		Data:    "test-data",
	}

	// Verify the data
	assert.Equal(t, "test-section", data.Section)
	assert.Equal(t, "test-id", data.ID)
	assert.Equal(t, "test-item", data.Item)
	assert.Equal(t, "example.com", data.Zone)
	assert.Equal(t, "test-data", data.Data)
}

// TestCtlNew tests the New function
func TestCtlNew(t *testing.T) {
	// Create a new Ctl object
	ctl := New()

	// Should not be nil if libknot is available
	if ctl != nil {
		defer ctl.Close()
		assert.NotNil(t, ctl)
		assert.NotNil(t, ctl.ctl)
	}
	// If nil, libknot allocation failed (which is okay in test environment)
}

// TestCtlSetTimeout tests the SetTimeout function
func TestCtlSetTimeout(t *testing.T) {
	ctl := New()
	if ctl == nil {
		t.Skip("libknot not available")
	}
	defer ctl.Close()

	// Test valid timeout
	assert.NotPanics(t, func() {
		ctl.SetTimeout(1000)
	})

	// Test zero timeout
	assert.NotPanics(t, func() {
		ctl.SetTimeout(0)
	})

	// Test negative timeout (should be handled safely)
	assert.NotPanics(t, func() {
		ctl.SetTimeout(-1)
	})

	// Test very large timeout (should use max value)
	assert.NotPanics(t, func() {
		ctl.SetTimeout(999999999999)
	})
}

// TestCtlConnectInvalidPath tests Connect with invalid path
func TestCtlConnectInvalidPath(t *testing.T) {
	ctl := New()
	if ctl == nil {
		t.Skip("libknot not available")
	}
	defer ctl.Close()

	// Try to connect to non-existent socket
	err := ctl.Connect("/nonexistent/socket.sock")

	// Should return an error
	assert.Error(t, err)
	assert.IsType(t, &CtlErrorConnect{}, err)
}

// TestCtlSendCommandBeforeConnect tests SendCommand before connecting
func TestCtlSendCommandBeforeConnect(t *testing.T) {
	ctl := New()
	if ctl == nil {
		t.Skip("libknot not available")
	}
	defer ctl.Close()

	// Try to send command without connecting
	err := ctl.SendCommand("status")

	// Should return an error
	assert.Error(t, err)
	assert.IsType(t, &CtlErrorSend{}, err)
}

// TestCtlSendCommandWithTypeBeforeConnect tests SendCommandWithType before connecting
func TestCtlSendCommandWithTypeBeforeConnect(t *testing.T) {
	ctl := New()
	if ctl == nil {
		t.Skip("libknot not available")
	}
	defer ctl.Close()

	// Try to send command with type without connecting
	err := ctl.SendCommandWithType("zone-read", "SOA")

	// Should return an error
	assert.Error(t, err)
	assert.IsType(t, &CtlErrorSend{}, err)
}

// TestCtlReceiveResponseBeforeConnect tests ReceiveResponse before connecting
func TestCtlReceiveResponseBeforeConnect(t *testing.T) {
	ctl := New()
	if ctl == nil {
		t.Skip("libknot not available")
	}
	defer ctl.Close()

	// Try to receive response without connecting
	_, _, err := ctl.ReceiveResponse()

	// Should return an error
	assert.Error(t, err)
	assert.IsType(t, &CtlErrorReceive{}, err)
}

// TestCtlClose tests the Close function
func TestCtlClose(t *testing.T) {
	ctl := New()
	if ctl == nil {
		t.Skip("libknot not available")
	}

	// Close should not panic
	assert.NotPanics(t, func() {
		ctl.Close()
	})

	// Closing again should not panic
	assert.NotPanics(t, func() {
		ctl.Close()
	})
}

// TestCtlCloseNil tests Close on nil Ctl
func TestCtlCloseNil(t *testing.T) {
	ctl := &Ctl{ctl: nil}

	// Close on nil ctl should not panic
	assert.NotPanics(t, func() {
		ctl.Close()
	})
}
