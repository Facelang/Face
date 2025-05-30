// Package errmgr provides common error definitions and categories for use across applications.
// These predefined errors are designed for consistency in error handling and can be used
// directly as immutable instances or copied for customization using Copy().
package errmgr

import (
	"github.com/olekukonko/errors"
)

// Common error categories used for organizing errors across different domains.
const (
	CategoryAuth       errors.ErrorCategory = "auth"       // Authentication-related errors (e.g., login failures)
	CategoryBusiness   errors.ErrorCategory = "business"   // Business logic errors (e.g., rule violations)
	CategoryDatabase   errors.ErrorCategory = "database"   // Database-related errors (e.g., connection issues)
	CategoryIO         errors.ErrorCategory = "io"         // Input/Output-related errors (e.g., file operations)
	CategoryNetwork    errors.ErrorCategory = "network"    // Network-related errors (e.g., timeouts, unreachable hosts)
	CategorySystem     errors.ErrorCategory = "system"     // System-level errors (e.g., resource exhaustion)
	CategoryUser       errors.ErrorCategory = "user"       // User-related errors (e.g., invalid input, permissions)
	CategoryValidation errors.ErrorCategory = "validation" // Validation-related errors (e.g., invalid input formats)
)

// Common HTTP status codes used for error responses, aligned with REST API conventions.
const (
	CodeBadRequest         = 400 // HTTP 400 Bad Request (client error, invalid input)
	CodeUnauthorized       = 401 // HTTP 401 Unauthorized (authentication required)
	CodeForbidden          = 403 // HTTP 403 Forbidden (access denied)
	CodeNotFound           = 404 // HTTP 404 Not Found (resource not found)
	CodeMethodNotAllowed   = 405 // HTTP 405 Method Not Allowed (unsupported method)
	CodeConflict           = 409 // HTTP 409 Conflict (resource conflict)
	CodeUnprocessable      = 422 // HTTP 422 Unprocessable Entity (semantic errors in request)
	CodeTooManyRequests    = 429 // HTTP 429 Too Many Requests (rate limiting)
	CodeInternalError      = 500 // HTTP 500 Internal Server Error (server failure)
	CodeNotImplemented     = 501 // HTTP 501 Not Implemented (feature not supported)
	CodeServiceUnavailable = 503 // HTTP 503 Service Unavailable (temporary unavailability)
)

// Generic Predefined Errors (Static)
// These are immutable instances suitable for direct use or copying with Copy().
// Errors requiring specific properties like WithRetryable() or WithTimeout() are defined here.
var (
	ErrInvalidArg         = errors.New("invalid argument").WithCode(CodeBadRequest)
	ErrNotFound           = errors.New("not found").WithCode(CodeNotFound)
	ErrPermission         = errors.New("permission denied").WithCode(CodeForbidden)
	ErrTimeout            = errors.New("operation timed out").WithTimeout()
	ErrUnknown            = errors.New("unknown error").WithCode(CodeInternalError)
	ErrDBConnRetryable    = errors.New("database connection failed").WithCategory(CategoryDatabase).WithRetryable()
	ErrNetworkRetryable   = errors.New("network failure").WithCategory(CategoryNetwork).WithRetryable()
	ErrNetworkTimedOut    = errors.New("network timeout").WithCategory(CategoryNetwork).WithTimeout().WithRetryable()
	ErrServiceRetryable   = errors.New("service unavailable").WithCode(CodeServiceUnavailable).WithRetryable()
	ErrRateLimitRetryable = errors.New("rate limit exceeded").WithCode(CodeTooManyRequests).WithRetryable()
)

// Authentication Errors (Templated)
// Use these by providing arguments, e.g., ErrAuthFailed("user@example.com", "invalid password").
var (
	ErrAuthFailed   = Coded("ErrAuthFailed", "authentication failed for %s: %s", CodeUnauthorized)
	ErrInvalidToken = Coded("ErrInvalidToken", "invalid authentication token: %s", CodeUnauthorized)
	ErrMissingCreds = Coded("ErrMissingCreds", "missing credentials: %s", CodeBadRequest)
	ErrTokenExpired = Coded("ErrTokenExpired", "authentication token expired: %s", CodeUnauthorized)
)

// Business Logic Errors (Templated)
// Example: ErrInsufficientFunds("account123", "balance too low").
var (
	ErrBusinessRule      = Categorized(CategoryBusiness, "ErrBusinessRule", "business rule violation: %s")
	ErrInsufficientFunds = Categorized(CategoryBusiness, "ErrInsufficientFunds", "insufficient funds: %s")
)

// Database Errors (Templated)
// Example: ErrDBConnection("mysql", "host unreachable").
var (
	ErrDBConnection = Categorized(CategoryDatabase, "ErrDBConnection", "database connection failed: %s")
	ErrDBConstraint = Coded("ErrDBConstraint", "database constraint violation: %s", CodeConflict)
	ErrDBQuery      = Categorized(CategoryDatabase, "ErrDBQuery", "database query failed: %s")
	ErrDBTimeout    = Categorized(CategoryDatabase, "ErrDBTimeout", "database operation timed out: %s")
)

// IO Errors (Templated)
// Example: ErrFileNotFound("/path/to/file").
var (
	ErrFileNotFound = Coded("ErrFileNotFound", "file (%s) not found", CodeNotFound)
	ErrIORead       = Categorized(CategoryIO, "ErrIORead", "I/O read error: %s")
	ErrIOWrite      = Categorized(CategoryIO, "ErrIOWrite", "I/O write error: %s")
)

// Network Errors (Templated)
// Example: ErrNetworkTimeout("http://example.com", "no response").
var (
	ErrNetworkConnRefused = Categorized(CategoryNetwork, "ErrNetworkConnRefused", "connection refused: %s")
	ErrNetworkTimeout     = Categorized(CategoryNetwork, "ErrNetworkTimeout", "network timeout: %s")
	ErrNetworkUnreachable = Categorized(CategoryNetwork, "ErrNetworkUnreachable", "network unreachable: %s")
)

// System Errors (Templated)
// Example: ErrResourceExhausted("memory", "out of memory").
var (
	ErrConfigInvalid     = Coded("ErrConfigInvalid", "invalid configuration: %s", CodeInternalError)
	ErrResourceExhausted = Coded("ErrResourceExhausted", "resource exhausted: %s", CodeServiceUnavailable)
	ErrSystemFailure     = Coded("ErrSystemFailure", "system failure: %s", CodeInternalError)
	ErrSystemUnhealthy   = Coded("ErrSystemUnhealthy", "system unhealthy: %s", CodeServiceUnavailable)
)

// User Errors (Templated)
// Example: ErrUserNotFound("user123", "not in database").
var (
	ErrUserLocked     = Coded("ErrUserLocked", "user %s is locked: %s", CodeForbidden)
	ErrUserNotFound   = Coded("ErrUserNotFound", "user %s not found: %s", CodeNotFound)
	ErrUserPermission = Coded("ErrUserPermission", "user %s lacks permission: %s", CodeForbidden)
	ErrUserSuspended  = Coded("ErrUserSuspended", "user %s is suspended: %s", CodeForbidden)
)

// Validation Errors (Templated)
// Example: ErrValidationFailed("email", "invalid email format").
var (
	ErrInvalidFormat    = Coded("ErrInvalidFormat", "invalid format: %s", CodeBadRequest)
	ErrValidationFailed = Coded("ErrValidationFailed", "validation failed: %s", CodeBadRequest)
)

// Additional REST API Errors (Templated)
// Example: ErrMethodNotAllowed("POST", "only GET allowed").
var (
	ErrConflict           = Coded("ErrConflict", "conflict occurred: %s", CodeConflict)
	ErrMethodNotAllowed   = Coded("ErrMethodNotAllowed", "method %s not allowed", CodeMethodNotAllowed)
	ErrNotImplemented     = Coded("ErrNotImplemented", "%s not implemented", CodeNotImplemented)
	ErrRateLimitExceeded  = Coded("ErrRateLimitExceeded", "rate limit exceeded: %s", CodeTooManyRequests)
	ErrServiceUnavailable = Coded("ErrServiceUnavailable", "service (%s) unavailable", CodeServiceUnavailable)
	ErrUnprocessable      = Coded("ErrUnprocessable", "unprocessable entity: %s", CodeUnprocessable)
)

// Additional Domain-Specific Errors (Templated)
// Example: ErrSerialization("json", "invalid data").
var (
	ErrDeserialization      = Define("ErrDeserialization", "deserialization error: %s")
	ErrExternalService      = Define("ErrExternalService", "external service (%s) error")
	ErrSerialization        = Define("ErrSerialization", "serialization error: %s")
	ErrUnsupportedOperation = Coded("ErrUnsupportedOperation", "unsupported operation %s", CodeNotImplemented)
)

// Predefined Templates with Categories (Templated)
// These are convenience wrappers with categories applied; use like AuthFailed("user", "reason").
var (
	AuthFailed      = Categorized(CategoryAuth, "AuthFailed", "authentication failed for %s: %s")
	BusinessError   = Categorized(CategoryBusiness, "BusinessError", "business error: %s")
	DBError         = Categorized(CategoryDatabase, "DBError", "database error: %s")
	IOError         = Categorized(CategoryIO, "IOError", "I/O error: %s")
	NetworkError    = Categorized(CategoryNetwork, "NetworkError", "network failure: %s")
	SystemError     = Categorized(CategorySystem, "SystemError", "system error: %s")
	UserError       = Categorized(CategoryUser, "UserError", "user error: %s")
	ValidationError = Categorized(CategoryValidation, "ValidationError", "validation error: %s")
)
