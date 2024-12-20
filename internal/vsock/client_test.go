// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0
package vsock

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ultravioletrs/cocos/manager"
	"google.golang.org/protobuf/proto"
)

// MockConn implements net.Conn for testing purposes.
type MockConn struct {
	ReadData    []byte
	WrittenData []byte
	ReadErr     error
	WriteErr    error
	closed      bool
	mu          sync.Mutex
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, io.EOF
	}
	if len(m.ReadData) == 0 {
		return 0, io.EOF // Ensure we handle this case more predictably
	}
	if m.ReadErr != nil {
		return 0, m.ReadErr
	}
	n = copy(b, m.ReadData)
	m.ReadData = m.ReadData[n:]
	return n, nil
}

func (m *MockConn) Write(b []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, errors.New("connection closed")
	}
	if m.WriteErr != nil {
		return 0, m.WriteErr
	}
	m.WrittenData = append(m.WrittenData, b...)
	return len(b), nil
}

func (m *MockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

// Implement other net.Conn methods with empty implementations.
func (m *MockConn) LocalAddr() net.Addr                { return nil }
func (m *MockConn) RemoteAddr() net.Addr               { return nil }
func (m *MockConn) SetDeadline(t time.Time) error      { return nil }
func (m *MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *MockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestAckReader_Read(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{"Valid message", []byte("Hello, World!"), false},
		{"Empty message", []byte{}, false},
		{"Message at max size", make([]byte, maxMessageSize), false},
		{"Message exceeds max size", make([]byte, maxMessageSize+1), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &MockConn{}
			ar := NewAckReader(mockConn)

			// Prepare mock data
			messageID := uint32(1)
			messageLen := uint32(len(tt.data))
			mockData := make([]byte, 8+len(tt.data))
			binary.LittleEndian.PutUint32(mockData[:4], messageID)
			binary.LittleEndian.PutUint32(mockData[4:8], messageLen)
			copy(mockData[8:], tt.data)
			mockConn.ReadData = mockData

			data, err := ar.Read()

			if (err != nil) != tt.wantErr {
				t.Errorf("AckReader.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !bytes.Equal(data, tt.data) {
					t.Errorf("AckReader.Read() got = %v, want %v", data, tt.data)
				}

				// Check if ACK was sent
				if len(mockConn.WrittenData) != 4 {
					t.Errorf("AckReader.Read() did not send ACK")
				} else {
					ackID := binary.LittleEndian.Uint32(mockConn.WrittenData)
					if ackID != messageID {
						t.Errorf("AckReader.Read() sent wrong ACK ID, got %d, want %d", ackID, messageID)
					}
				}
			}
		})
	}
}

func TestAckReader_ReadProto(t *testing.T) {
	tests := []struct {
		name    string
		msg     *manager.ClientStreamMessage
		wantErr bool
	}{
		{"Valid proto message", &manager.ClientStreamMessage{}, false},
		{"Empty proto message", &manager.ClientStreamMessage{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &MockConn{}
			ar := NewAckReader(mockConn)

			// Prepare mock data
			protoData, _ := proto.Marshal(tt.msg)
			messageID := uint32(1)
			messageLen := uint32(len(protoData))
			mockData := make([]byte, 8+len(protoData))
			binary.LittleEndian.PutUint32(mockData[:4], messageID)
			binary.LittleEndian.PutUint32(mockData[4:8], messageLen)
			copy(mockData[8:], protoData)
			mockConn.ReadData = mockData

			receivedMsg := &manager.ClientStreamMessage{}
			err := ar.ReadProto(receivedMsg)

			if (err != nil) != tt.wantErr {
				t.Errorf("AckReader.ReadProto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if receivedMsg.Message != tt.msg.Message {
					t.Errorf("AckReader.ReadProto() got = %v, want %v", receivedMsg, tt.msg)
				}

				// Check if ACK was sent
				if len(mockConn.WrittenData) != 4 {
					t.Errorf("AckReader.ReadProto() did not send ACK")
				} else {
					ackID := binary.LittleEndian.Uint32(mockConn.WrittenData)
					if ackID != messageID {
						t.Errorf("AckReader.ReadProto() sent wrong ACK ID, got %d, want %d", ackID, messageID)
					}
				}
			}
		})
	}
}

func TestNewAckWriter(t *testing.T) {
	mockConn := &MockConn{}
	writer := NewAckWriter(mockConn)

	if _, ok := writer.(io.Writer); !ok {
		t.Errorf("NewAckWriter() did not return an io.Writer")
	}
}

func TestNewAckReader(t *testing.T) {
	mockConn := &MockConn{}
	reader := NewAckReader(mockConn)

	assert.NotNil(t, reader)
}

func TestAckWriter_Close(t *testing.T) {
	mockConn := &MockConn{}
	aw := NewAckWriter(mockConn)

	err := aw.Close()
	if err != nil {
		t.Errorf("AckWriter.Close() error = %v, wantErr %v", err, nil)
	}

	if !mockConn.closed {
		t.Errorf("AckWriter.Close() did not close the connection")
	}
}

func TestAckWriter_Write(t *testing.T) {
	tests := []struct {
		name          string
		input         []byte
		expectErr     bool
		expectedError string
	}{
		{
			name:          "Message exceeds max size",
			input:         make([]byte, maxMessageSize+1),
			expectErr:     true,
			expectedError: "message size exceeds maximum allowed size",
		},
		{
			name:      "Write succeeds",
			input:     []byte("Hello, world!"),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &MockConn{
				mu: sync.Mutex{},
			}

			writer := NewAckWriter(mockConn)
			defer writer.Close()

			if tt.expectErr {
				writer.(*AckWriter).ctx.Done()
			}

			time.Sleep(100 * time.Millisecond)

			n, err := writer.Write(tt.input)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
				assert.Zero(t, n)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.input), n)
			}
		})
	}
}

func TestAckWriter_CleanupOldMessages(t *testing.T) {
	mockConn := &MockConn{}
	writer := NewAckWriter(mockConn).(*AckWriter)
	defer writer.Close()

	for i := uint32(1); i <= maxConcurrent+10; i++ {
		msg := &Message{
			ID:      i,
			Content: []byte("test"),
			Status:  StatusAcknowledged,
		}
		writer.messageStore.Store(i, msg)
	}

	writer.cleanupOldMessages(maxConcurrent + 11)

	var count int
	writer.messageStore.Range(func(key, value interface{}) bool {
		count++
		return true
	})

	assert.LessOrEqual(t, count, maxConcurrent)
}

func TestAckReader_LargeMessage(t *testing.T) {
	mockConn := &MockConn{}
	reader := NewAckReader(mockConn)

	largeMessage := make([]byte, maxMessageSize-1)
	for i := range largeMessage {
		largeMessage[i] = byte(i % 256)
	}

	messageID := uint32(1)
	messageLen := uint32(len(largeMessage))
	mockData := make([]byte, 8+len(largeMessage))
	binary.LittleEndian.PutUint32(mockData[:4], messageID)
	binary.LittleEndian.PutUint32(mockData[4:8], messageLen)
	copy(mockData[8:], largeMessage)
	mockConn.ReadData = mockData

	data, err := reader.Read()
	assert.NoError(t, err)
	assert.Equal(t, largeMessage, data)

	assert.Equal(t, 4, len(mockConn.WrittenData))
	ackID := binary.LittleEndian.Uint32(mockConn.WrittenData)
	assert.Equal(t, messageID, ackID)
}

func TestAckWriter_FailedSends(t *testing.T) {
	mockConn := &MockConn{
		WriteErr: errors.New("write error"),
	}
	writer := NewAckWriter(mockConn).(*AckWriter)
	defer writer.Close()

	// Add some messages to the channel
	for i := 0; i < 5; i++ {
		msg := &Message{
			ID:      uint32(i + 1),
			Content: []byte(fmt.Sprintf("Message %d", i+1)),
			Status:  StatusPending,
		}
		writer.pendingMessages <- msg
	}

	// Wait for the messages to be sent
	time.Sleep(100 * time.Millisecond)

	// Check that the messages were marked as failed
	writer.messageStore.Range(func(key, value interface{}) bool {
		msg := value.(*Message)
		assert.Equal(t, StatusFailed, msg.Status)
		return true
	})
}
