package errmgr

import (
	"github.com/olekukonko/errors"
	"testing"
)

func TestStaticErrors(t *testing.T) {
	tests := []struct {
		err      *errors.Error
		name     string
		expected string
		code     int
		retry    bool
		timeout  bool
	}{
		{ErrInvalidArg, "ErrInvalidArg", "invalid argument", CodeBadRequest, false, false},
		{ErrNotFound, "ErrNotFound", "not found", CodeNotFound, false, false},
		{ErrPermission, "ErrPermission", "permission denied", CodeForbidden, false, false},
		{ErrTimeout, "ErrTimeout", "operation timed out", 0, false, true},
		{ErrUnknown, "ErrUnknown", "unknown error", CodeInternalError, false, false},
		{ErrDBConnRetryable, "ErrDBConnRetryable", "database connection failed", 0, true, false},
		{ErrNetworkRetryable, "ErrNetworkRetryable", "network failure", 0, true, false},
		{ErrNetworkTimedOut, "ErrNetworkTimedOut", "network timeout", 0, true, true},
		{ErrServiceRetryable, "ErrServiceRetryable", "service unavailable", CodeServiceUnavailable, true, false},
		{ErrRateLimitRetryable, "ErrRateLimitRetryable", "rate limit exceeded", CodeTooManyRequests, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("Expected message %q, got %q", tt.expected, tt.err.Error())
			}
			if tt.err.Code() != tt.code {
				t.Errorf("Expected code %d, got %d", tt.code, tt.err.Code())
			}
			ctx := tt.err.Context()
			if tt.retry && (ctx == nil || !ctx["[error] retry"].(bool)) {
				t.Errorf("Expected retryable error, got context %v", ctx)
			}
			if tt.timeout && (ctx == nil || !ctx["[error] timeout"].(bool)) {
				t.Errorf("Expected timeout error, got context %v", ctx)
			}
		})
	}
}

func TestTemplatedErrors(t *testing.T) {
	tests := []struct {
		errFunc  func(...interface{}) *errors.Error
		name     string
		args     []interface{}
		expected string
		code     int
		category errors.ErrorCategory
	}{
		{ErrAuthFailed, "ErrAuthFailed", []interface{}{"user", "pass"}, "authentication failed for user: pass", CodeUnauthorized, ""},
		{ErrDBConnection, "ErrDBConnection", []interface{}{"mysql"}, "database connection failed: mysql", 0, CategoryDatabase},
		{ErrNetworkTimeout, "ErrNetworkTimeout", []interface{}{"host"}, "network timeout: host", 0, CategoryNetwork},
		{ErrFileNotFound, "ErrFileNotFound", []interface{}{"file.txt"}, "file (file.txt) not found", CodeNotFound, ""},
		{ErrValidationFailed, "ErrValidationFailed", []interface{}{"email"}, "validation failed: email", CodeBadRequest, ""},
		{ErrRateLimitExceeded, "ErrRateLimitExceeded", []interface{}{"user123"}, "rate limit exceeded: user123", CodeTooManyRequests, ""},
		{ErrUserNotFound, "ErrUserNotFound", []interface{}{"user123", "not in db"}, "user user123 not found: not in db", CodeNotFound, ""},
		{ErrMethodNotAllowed, "ErrMethodNotAllowed", []interface{}{"POST"}, "method POST not allowed", CodeMethodNotAllowed, ""},
		{ErrUnprocessable, "ErrUnprocessable", []interface{}{"data"}, "unprocessable entity: data", CodeUnprocessable, ""},
		{ErrBusinessRule, "ErrBusinessRule", []interface{}{"rule1"}, "business rule violation: rule1", 0, CategoryBusiness},
		{ErrIORead, "ErrIORead", []interface{}{"disk"}, "I/O read error: disk", 0, CategoryIO},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errFunc(tt.args...)
			if err.Error() != tt.expected {
				t.Errorf("Expected message %q, got %q", tt.expected, err.Error())
			}
			if err.Code() != tt.code {
				t.Errorf("Expected code %d, got %d", tt.code, err.Code())
			}
			if tt.category != "" {
				if cat := errors.Category(err); cat != string(tt.category) {
					t.Errorf("Expected category %q, got %q", tt.category, cat)
				}
			}
			err.Free()
		})
	}
}

func TestCategorizedTemplates(t *testing.T) {
	tests := []struct {
		errFunc  func(...interface{}) *errors.Error
		name     string
		args     []interface{}
		expected string
		category errors.ErrorCategory
		code     int
	}{
		{AuthFailed, "AuthFailed", []interface{}{"user", "reason"}, "authentication failed for user: reason", CategoryAuth, 0},
		{BusinessError, "BusinessError", []interface{}{"rule"}, "business error: rule", CategoryBusiness, 0},
		{DBError, "DBError", []interface{}{"query"}, "database error: query", CategoryDatabase, 0},
		{IOError, "IOError", []interface{}{"disk"}, "I/O error: disk", CategoryIO, 0},
		{NetworkError, "NetworkError", []interface{}{"host"}, "network failure: host", CategoryNetwork, 0},
		{SystemError, "SystemError", []interface{}{"crash"}, "system error: crash", CategorySystem, 0},
		{UserError, "UserError", []interface{}{"input"}, "user error: input", CategoryUser, 0},
		{ValidationError, "ValidationError", []interface{}{"format"}, "validation error: format", CategoryValidation, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errFunc(tt.args...)
			if err.Error() != tt.expected {
				t.Errorf("Expected message %q, got %q", tt.expected, err.Error())
			}
			if err.Code() != tt.code {
				t.Errorf("Expected code %d, got %d", tt.code, err.Code())
			}
			if cat := errors.Category(err); cat != string(tt.category) {
				t.Errorf("Expected category %q, got %q", tt.category, cat)
			}
			err.Free()
		})
	}
}
