package worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"

	"smtp-worker/domain"
	"smtp-worker/infrastructure/smtp"

	"github.com/nats-io/nats.go"
)

// Mocks
type MockLogger struct {
	messages []string
	mu       sync.Mutex
}

func (m *MockLogger) Printf(format string, v ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, fmt.Sprintf(format, v...))
}

func (m *MockLogger) LastMessage() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.messages) == 0 {
		return ""
	}
	return m.messages[len(m.messages)-1]
}

type MockSMTPClient struct {
	SendFunc   func(email *domain.Email) (*smtp.DeliveryResult, error)
	sendCalled bool
}

func (m *MockSMTPClient) Send(email *domain.Email) (*smtp.DeliveryResult, error) {
	m.sendCalled = true
	if m.SendFunc != nil {
		return m.SendFunc(email)
	}
	return &smtp.DeliveryResult{Success: true}, nil
}

type MockNatsMsg struct {
	data         []byte
	ackCallCount int
}

func (m *MockNatsMsg) Ack(opts ...nats.AckOpt) error {
	m.ackCallCount++
	return nil
}

func (m *MockNatsMsg) Data() []byte {
	return m.data
}

// --- Tests ---

func TestProcessMessage_Success(t *testing.T) {
	mockLogger := &MockLogger{}
	mockSMTP := &MockSMTPClient{}

	email := domain.Email{To: "test@example.com", Subject: "Test"}
	emailJSON, _ := json.Marshal(email)
	mockMsg := &MockNatsMsg{data: emailJSON}

	testWorker := &Worker{
		smtpClient: mockSMTP,
		logger:     mockLogger,
	}

	testWorker.processMessage(mockMsg)

	if !mockSMTP.sendCalled {
		t.Error("smtpClient.Send wurde nicht aufgerufen, sollte es aber")
	}

	if mockMsg.ackCallCount != 1 {
		t.Errorf("msg.Ack wurde %d Mal aufgerufen, erwartet wurde 1 Mal", mockMsg.ackCallCount)
	}
}

func TestProcessMessage_SMTPFailure(t *testing.T) {
	mockLogger := &MockLogger{}
	mockSMTP := &MockSMTPClient{
		SendFunc: func(email *domain.Email) (*smtp.DeliveryResult, error) {
			return nil, errors.New("connection refused")
		},
	}

	email := domain.Email{To: "test@example.com", Subject: "Test"}
	emailJSON, _ := json.Marshal(email)
	mockMsg := &MockNatsMsg{data: emailJSON}

	testWorker := &Worker{
		smtpClient: mockSMTP,
		logger:     mockLogger,
	}

	testWorker.processMessage(mockMsg)

	if !mockSMTP.sendCalled {
		t.Error("smtpClient.Send wurde nicht aufgerufen")
	}

	if mockMsg.ackCallCount > 0 {
		t.Errorf("msg.Ack wurde aufgerufen, sollte es aber nicht bei einem SMTP-Fehler")
	}
}

func TestProcessMessage_InvalidJSON(t *testing.T) {
	mockLogger := &MockLogger{}
	mockSMTP := &MockSMTPClient{}

	invalidData := []byte(`{"to": "test@example.com", "subject": "Test"`)
	mockMsg := &MockNatsMsg{data: invalidData}

	testWorker := &Worker{
		smtpClient: mockSMTP,
		logger:     mockLogger,
	}

	testWorker.processMessage(mockMsg)

	if mockSMTP.sendCalled {
		t.Error("smtpClient.Send wurde aufgerufen, sollte es aber nicht bei ungültigem JSON")
	}

	if mockMsg.ackCallCount != 1 {
		t.Errorf("msg.Ack wurde nicht aufgerufen, sollte aber, um eine ungültige Nachricht zu verwerfen")
	}
}
