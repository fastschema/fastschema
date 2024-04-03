package testutils

func CreateLogger(silences ...bool) *MockLogger {
	silences = append(silences, false)
	mockLogger := &MockLogger{
		Silence: silences[0],
	}

	return mockLogger
}
