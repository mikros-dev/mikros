// Package http provides a comprehensive set of HTTP request parameter binding utilities
// for Go applications. It enables automatic extraction and type conversion of request
// data (query parameters, headers, path parameters, and JSON body) into Go structs
// using struct field tags and reflection.
//
// # Overview
//
// This package offers several binding functions that can extract data from different
// parts of an HTTP request.
//
// All binding functions support rich type conversion including basic types, pointers,
// slices, time.Time, time.Duration, and types implementing encoding.TextUnmarshaler.
//
// # Basic Usage
//
// The simplest way to bind request data is using the specialized binding functions:
//
//	type UserRequest struct {
//		ID     int    `json:"id"`
//		Name   string `json:"name"`
//		Active bool   `json:"active"`
//	}
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//		var req UserRequest
//		if err := BindQuery(r, &req); err != nil {
//			http.Error(w, err.Error(), http.StatusBadRequest)
//			return
//		}
//		// Use req.ID, req.Name, req.Active...
//	}
//
// # Multi-Source Binding
//
// The Bind function supports extracting data from different request sources based on
// struct field tags. Use the `http` tag to specify the data source:
//
//	type RequestParams struct {
//		UserID   string    `json:"user_id" http:"loc=path"`        // From URL path
//		Filter   string    `json:"filter" http:"loc=query"`        // From query string
//		APIKey   string    `json:"api_key" http:"loc=header"`      // From headers
//		PageSize int       `json:"page_size" http:"loc=query"`     // From query string
//		Created  time.Time `json:"created" http:"loc=query,time_format=2006-01-02"`
//	}
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//		var params RequestParams
//		if err := Bind(r, &params); err != nil {
//			http.Error(w, err.Error(), http.StatusBadRequest)
//			return
//		}
//		// All fields populated from their respective sources
//	}
//
// # Field Name Resolution
//
// Field names are resolved in the following order:
//
//  1. `json` struct tag value (e.g., `json:"user_name"`)
//  2. Snake case conversion if FallbackSnakeCase is enabled (e.g., "UserName" -> "user_name")
//  3. Lowercase field name (e.g., "UserName" -> "username")
//
// Fields tagged with `json:"-"` are skipped during binding.
//
// # Slice and Multiple Value Handling
//
// Slices are populated from multiple parameter values or CSV-formatted single values:
//
//	type FilterRequest struct {
//		Tags []string `json:"tags"`  // Multiple ?tags=foo&tags=bar or single ?tags=foo,bar
//		IDs  []int    `json:"ids"`   // Multiple ?ids=1&ids=2 or single ?ids=1,2,3
//	}
//
// CSV parsing is controlled by BindOptions.
//
// # TextUnmarshaler Support
//
// Types implementing encoding.TextUnmarshaler can be bound directly:
//
//	type Status int
//
//	func (s *Status) UnmarshalText(data []byte) error {
//		switch string(data) {
//		case "active":
//			*s = 1
//		case "inactive":
//			*s = 0
//		default:
//			return fmt.Errorf("invalid status: %s", data)
//		}
//		return nil
//	}
//
//	type UserRequest struct {
//		Status Status `json:"status"`
//	}
package http
