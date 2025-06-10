// Copyright 2025 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package srv

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	XFrameOptionsDENY                               = "DENY"
	XFrameOptionsSAMEORIGIN                         = "SAMEORIGIN"
	XPermittedCrossDomainPoliciesNONE               = "none"
	XPermittedCrossDomainPoliciesMASTER_ONLY        = "master-only"
	XPermittedCrossDomainPoliciesBY_CONTENT_TYPE    = "by-content-type"
	XPermittedCrossDomainPoliciesBY_FTP_FILENAME    = "by-ftp-filename"
	XPermittedCrossDomainPoliciesALL                = "all"
	XPermittedCrossDomainPoliciesNONE_THIS_RESPONSE = "none-this-response"

	TransferEncodingChunked  = "chunked"
	TransferEncodingCompress = "compress"
	TransferEncodingDeflate  = "deflate"
	TransferEncodingGzip     = "gzip"
)

type BodyFn func(w io.Writer) error

// Response represents an HTTP response that can be customized with status codes, headers, and body content.
// It provides a fluent interface for building responses with various common HTTP status codes and payloads.
type Response struct {
	StatusCode int
	headers    http.Header
	cookies    []*http.Cookie
	bodyFn     BodyFn
	jsonBody   any
	rawBody    []byte
	afterWrite []func()
}

// Respond creates a new Response with default status code 200 OK and empty headers.
func Respond() *Response {
	return &Response{
		StatusCode: http.StatusOK,
		headers:    http.Header{},
		cookies:    make([]*http.Cookie, 0),
		afterWrite: make([]func(), 0),
	}
}

// Status sets the HTTP status code for the response.
func (r *Response) Status(status int) *Response {
	r.StatusCode = status
	return r
}

// Created sets the HTTP status code to 201 Created and optionally sets the response body.
func (r *Response) Created(body ...any) *Response {
	return r.statusWithBody(http.StatusCreated, body...)
}

// NoContent sets the HTTP status code to 204 No Content.
func (r *Response) NoContent() *Response {
	r.StatusCode = http.StatusNoContent
	return r
}

// MovedPermanently sets the HTTP status code to 301 Moved Permanently and sets the Location header.
func (r *Response) MovedPermanently(location string) *Response {
	r.StatusCode = http.StatusMovedPermanently
	r.headers.Set("Location", location)
	return r
}

// Found sets the HTTP status code to 302 Found and sets the Location header.
func (r *Response) Found(location string) *Response {
	r.StatusCode = http.StatusFound
	r.headers.Set("Location", location)
	return r
}

// NotModified sets the HTTP status code to 304 Not Modified.
func (r *Response) NotModified() *Response {
	r.StatusCode = http.StatusNotModified
	return r
}

// BadRequest sets the HTTP status code to 400 Bad Request and optionally sets the response body.
func (r *Response) BadRequest(body ...any) *Response {
	return r.statusWithBody(http.StatusBadRequest, body...)
}

// Unauthorized sets the HTTP status code to 401 Unauthorized and optionally sets the response body.
func (r *Response) Unauthorized(body ...any) *Response {
	return r.statusWithBody(http.StatusUnauthorized, body...)
}

// Forbidden sets the HTTP status code to 403 Forbidden and optionally sets the response body.
func (r *Response) Forbidden(body ...any) *Response {
	return r.statusWithBody(http.StatusForbidden, body...)
}

// NotFound sets the HTTP status code to 404 Not Found and optionally sets the response body.
func (r *Response) NotFound(body ...any) *Response {
	return r.statusWithBody(http.StatusNotFound, body...)
}

// MethodNotAllowed sets the HTTP status code to 405 Method Not Allowed and optionally sets the response body.
func (r *Response) MethodNotAllowed(body ...any) *Response {
	return r.statusWithBody(http.StatusMethodNotAllowed, body...)
}

// NotAcceptable sets the HTTP status code to 406 Not Acceptable and optionally sets the response body.
func (r *Response) NotAcceptable(body ...any) *Response {
	return r.statusWithBody(http.StatusNotAcceptable, body...)
}

// ProxyAuthRequired sets the HTTP status code to 407 Proxy Authentication Required.
func (r *Response) ProxyAuthRequired(body ...any) *Response {
	return r.statusWithBody(http.StatusProxyAuthRequired, body...)
}

// Conflict sets the HTTP status code to 409 Conflict and optionally sets the response body.
func (r *Response) Conflict(body ...any) *Response {
	return r.statusWithBody(http.StatusConflict, body...)
}

// PreconditionFailed sets the HTTP status code to 412 Precondition Failed.
func (r *Response) PreconditionFailed() *Response {
	r.StatusCode = http.StatusPreconditionFailed
	return r
}

func (r *Response) InternalServerError(body ...any) *Response {
	return r.statusWithBody(http.StatusInternalServerError, body...)
}

func (r *Response) statusWithBody(status int, body ...any) *Response {
	r.StatusCode = status
	if len(body) > 0 {
		return r.Json(body[0])
	}
	return r
}

// Error sets the HTTP status code to 500 Internal Server Error and sets the response body to an ErrorDto.
// If err is nil, the error message will be empty. Otherwise, the error message will be set to err.Error().
func (r *Response) Error(err error) *Response {
	r.StatusCode = http.StatusInternalServerError
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	return r.Json(ErrorDto{
		Code:    "InternalServerError",
		Message: msg,
	})
}

// Header sets a header in the response.
func (r *Response) Header(key, value string) *Response {
	r.headers.Set(key, value)
	return r
}

// WwwAuthenticate sets the "WWW-Authenticate" header in the response.
func (r *Response) WwwHauthenticate(challenge string) *Response {
	r.headers.Set("WWW-Authenticate", challenge)
	return r
}

// ProxyAuthenticate sets the "Proxy-Authenticate" header in the response.
func (r *Response) ProxyAuthenticate(challenge string) *Response {
	r.headers.Set("Proxy-Authenticate", challenge)
	return r
}

// Age sets the "Age" header in the response.
func (r *Response) Age(deltaSeconds int) *Response {
	if deltaSeconds < 0 {
		panic("deltaSeconds must be greater than or equal to 0")
	}
	r.headers.Set("Age", strconv.Itoa(deltaSeconds))
	return r
}

// CacheControl sets the "Cache-Control" header in the response.
func (r *Response) CacheControl(directive string) *Response {
	r.headers.Set("Cache-Control", directive)
	return r
}

// ClearSiteData sets the "Clear-Site-Data" header in the response.
func (r *Response) ClearSiteData(directive string) *Response {
	r.headers.Set("Clear-Site-Data", directive)
	return r
}

// Expires sets the "Expires" header in the response.
// The time will be automatically converted to UTC and formatted according to RFC 7231.
func (r *Response) Expires(t time.Time) *Response {
	r.headers.Set("Expires", t.UTC().Format(http.TimeFormat))
	return r
}

// NoVarySearch sets the "No-Vary-Search" header in the response.
func (r *Response) NoVarySearch(rules string) *Response {
	r.headers.Set("No-Vary-Search", rules)
	return r
}

// LastModified sets the "Last-Modified" header in the response.
// The time will be automatically converted to UTC and formatted according to RFC 7231.
func (r *Response) LastModified(t time.Time) *Response {
	r.headers.Set("Last-Modified", t.UTC().Format(http.TimeFormat))
	return r
}

// ETag sets the "ETag" header in the response. The etag value will be automatically wrapped in quotes.
func (r *Response) ETag(etag string) *Response {
	r.headers.Set("ETag", `"`+etag+`"`)
	return r
}

// Vary sets the "Vary" header in the response.
func (r *Response) Vary(headers ...string) *Response {
	r.headers.Set("Vary", strings.Join(headers, ", "))
	return r
}

// Connection sets the "Connection" header in the response.
func (r *Response) Connection(value string) *Response {
	r.headers.Set("Connection", value)
	return r
}

// KeepAlive sets the "Keep-Alive" header in the response.
func (r *Response) KeepAlive(timeout int, max int) *Response {
	r.headers.Set("Keep-Alive", fmt.Sprintf("timeout=%d, max=%d", timeout, max))
	return r
}

// Accept sets the "Accept" header in the response.
func (r *Response) Accept(value string) *Response {
	r.headers.Set("Accept", value)
	return r
}

// AcceptEncoding sets the "Accept-Encoding" header in the response.
func (r *Response) AcceptEncoding(value string) *Response {
	r.headers.Set("Accept-Encoding", value)
	return r
}

// AcceptPatch sets the "Accept-Patch" header in the response.
func (r *Response) AcceptPatch(value string) *Response {
	r.headers.Set("Accept-Patch", value)
	return r
}

// AcceptPost sets the "Accept-Post" header in the response.
func (r *Response) AcceptPost(value string) *Response {
	r.headers.Set("Accept-Post", value)
	return r
}

// Cookie adds a Set-Cookie header to the ResponseWriter's headers.
// The provided cookie must have a valid Name. Invalid cookies may be silently dropped.
func (r *Response) Cookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) *Response {
	if path == "" {
		path = "/"
	}
	return r.CookieRaw(&http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

// CookieRaw adds a Set-Cookie header to the ResponseWriter's headers.
// The provided cookie must have a valid Name. Invalid cookies may be silently dropped.
func (r *Response) CookieRaw(cookie *http.Cookie) *Response {
	r.cookies = append(r.cookies, cookie)
	return r
}

// AccessControlAllowCredentials sets the "Access-Control-Allow-Credentials" header in the response.
func (r *Response) AccessControlAllowCredentials() *Response {
	r.headers.Set("Access-Control-Allow-Credentials", "true")
	return r
}

// AccessControlAllowHeaders sets the "Access-Control-Allow-Headers" header in the response.
func (r *Response) AccessControlAllowHeaders(headers ...string) *Response {
	r.headers.Set("Access-Control-Allow-Headers", strings.Join(headers, ", "))
	return r
}

// AccessControlAllowMethods sets the "Access-Control-Allow-Methods" header in the response.
func (r *Response) AccessControlAllowMethods(methods ...string) *Response {
	r.headers.Set("Access-Control-Allow-Methods", strings.Join(methods, ", "))
	return r
}

// AccessControlAllowOrigin sets the "Access-Control-Allow-Origin" header in the response.
func (r *Response) AccessControlAllowOrigin(origin string) *Response {
	r.headers.Set("Access-Control-Allow-Origin", origin)
	return r
}

// AccessControlExposeHeaders sets the "Access-Control-Expose-Headers" header in the response.
func (r *Response) AccessControlExposeHeaders(headers ...string) *Response {
	r.headers.Set("Access-Control-Expose-Headers", strings.Join(headers, ", "))
	return r
}

// AccessControlMaxAge sets the "Access-Control-Max-Age" header in the response.
func (r *Response) AccessControlMaxAge(maxAge int) *Response {
	if maxAge < 0 {
		panic("maxAge must be greater than or equal to 0")
	}
	r.headers.Set("Access-Control-Max-Age", strconv.Itoa(maxAge))
	return r
}

// TimingAllowOrigin sets the "Timing-Allow-Origin" header in the response.
func (r *Response) TimingAllowOrigin(origin string) *Response {
	r.headers.Set("Timing-Allow-Origin", origin)
	return r
}

// ContentDisposition sets the "Content-Disposition" header in the response.
func (r *Response) ContentDisposition(disposition string) *Response {
	r.headers.Set("Content-Disposition", disposition)
	return r
}

// ContentLength sets the "Content-Length" header in the response.
func (r *Response) ContentLength(length int64) *Response {
	r.headers.Set("Content-Length", strconv.FormatInt(length, 10))
	return r
}

// ContentType sets the "Content-Type" header in the response.
func (r *Response) ContentType(contentType string) *Response {
	r.headers.Set("Content-Type", contentType)
	return r
}

// ContentEncoding sets the "Content-Encoding" header in the response.
func (r *Response) ContentEncoding(encoding string) *Response {
	r.headers.Set("Content-Encoding", encoding)
	return r
}

// ContentLanguage sets the "Content-Language" header in the response.
func (r *Response) ContentLanguage(language string) *Response {
	r.headers.Set("Content-Language", language)
	return r
}

// ContentLocation sets the "Content-Location" header in the response.
func (r *Response) ContentLocation(location string) *Response {
	r.headers.Set("Content-Location", location)
	return r
}

// PreferenceApplied sets the "Preference-Applied" header in the response.
func (r *Response) PreferenceApplied(preference string) *Response {
	r.headers.Set("Preference-Applied", preference)
	return r
}

// Via sets the "Via" header in the response.
// The value will be added to the existing "Via" header.
func (r *Response) Via(via string) *Response {
	r.headers.Add("Via", via)
	return r
}

// AcceptRanges sets the "Accept-Ranges" header in the response.
func (r *Response) AcceptRanges() *Response {
	r.headers.Add("Accept-Ranges", "bytes")
	return r
}

// ContentRange sets the "Content-Range" header in the response.
func (r *Response) ContentRange(value string) *Response {
	r.headers.Set("Content-Range", value)
	return r
}

// Location sets the "Location" header in the response.
func (r *Response) Location(location string) *Response {
	r.headers.Set("Location", location)
	return r
}

// Refresh sets the "Refresh" header in the response.
func (r *Response) Refresh(timeSeconds int, url string) *Response {
	if url == "" {
		r.headers.Set("Refresh", strconv.Itoa(timeSeconds))
	} else {
		r.headers.Set("Refresh", fmt.Sprintf("%d;url=%s", timeSeconds, url))
	}
	return r
}

// ReferrerPolicy sets the "Referrer-Policy" header in the response.
func (r *Response) ReferrerPolicy(policy string) *Response {
	r.headers.Set("Referrer-Policy", policy)
	return r
}

// Allow sets the "Allow" header in the response.
func (r *Response) Allow(methods ...string) *Response {
	r.headers.Set("Allow", strings.Join(methods, ", "))
	return r
}

// Server sets the "Server" header in the response.
func (r *Response) Server(server string) *Response {
	r.headers.Set("Server", server)
	return r
}

// CrossOriginEmbedderPolicy sets the "Cross-Origin-Embedder-Policy" header in the response.
func (r *Response) CrossOriginEmbedderPolicy(policy string) *Response {
	r.headers.Set("Cross-Origin-Embedder-Policy", policy)
	return r
}

// CrossOriginOpenerPolicy sets the "Cross-Origin-Opener-Policy" header in the response.
func (r *Response) CrossOriginOpenerPolicy(policy string) *Response {
	r.headers.Set("Cross-Origin-Opener-Policy", policy)
	return r
}

// CrossOriginResourcePolicy sets the "Cross-Origin-Resource-Policy" header in the response.
func (r *Response) CrossOriginResourcePolicy(policy string) *Response {
	r.headers.Set("Cross-Origin-Resource-Policy", policy)
	return r
}

// ContentSecurityPolicy sets the "Content-Security-Policy" header in the response.
func (r *Response) ContentSecurityPolicy(directive string) *Response {
	r.headers.Set("Content-Security-Policy", directive)
	return r
}

// ContentSecurityPolicyReportOnly sets the "Content-Security-Policy-Report-Only" header in the response.
func (r *Response) ContentSecurityPolicyReportOnly(directive string) *Response {
	r.headers.Set("Content-Security-Policy-Report-Only", directive)
	return r
}

// StrictTransportSecurity sets the "Strict-Transport-Security" header in the response.
func (r *Response) StrictTransportSecurity(value string) *Response {
	r.headers.Set("Strict-Transport-Security", value)
	return r
}

// XContentTypeOptions sets the "X-Content-Type-Options" header in the response.
func (r *Response) XContentTypeOptions() *Response {
	r.headers.Set("X-Content-Type-Options", "nosniff")
	return r
}

// XFrameOptions sets the "X-Frame-Options" header in the response.
func (r *Response) XFrameOptions(directive string) *Response {
	r.headers.Set("X-Frame-Options", directive)
	return r
}

// XPermittedCrossDomainPolicies sets the "X-Permitted-Cross-Domain-Policies" header in the response.
func (r *Response) XPermittedCrossDomainPolicies(directive string) *Response {
	r.headers.Set("X-Permitted-Cross-Domain-Policies", directive)
	return r
}

// XPoweredBy sets the "X-Powered-By" header in the response.
func (r *Response) XPoweredBy(application string) *Response {
	r.headers.Set("X-Powered-By", application)
	return r
}

// ReportingEndpoints sets the "Reporting-Endpoints" header in the response.
func (r *Response) ReportingEndpoints(endpoints ...string) *Response {
	r.headers.Set("Reporting-Endpoints", strings.Join(endpoints, ", "))
	return r
}

// TransferEncoding sets the "Transfer-Encoding" header in the response.
func (r *Response) TransferEncoding(encodings ...string) *Response {
	r.headers.Set("Transfer-Encoding", strings.Join(encodings, ", "))
	return r
}

// Trailer sets the "Trailer" header in the response.
func (r *Response) Trailer(headerNames string) *Response {
	r.headers.Set("Trailer", headerNames)
	return r
}

// Date sets the "Date" header in the response.
func (r *Response) Date(t time.Time) *Response {
	r.headers.Set("Date", t.UTC().Format(http.TimeFormat))
	return r
}

// Link sets the "Link" header in the response.
func (r *Response) Link(link string) *Response {
	r.headers.Set("Link", link)
	return r
}

// RetryAfterSeconds sets the "Retry-After" header in the response.
func (r *Response) RetryAfterSeconds(seconds int) *Response {
	r.headers.Set("Retry-After", strconv.Itoa(seconds))
	return r
}

// RetryAfterDate sets the "Retry-After" header in the response.
func (r *Response) RetryAfterDate(t time.Time) *Response {
	r.headers.Set("Retry-After", t.UTC().Format(http.TimeFormat))
	return r
}

// ServerTiming sets the "Server-Timing" header in the response.
func (r *Response) ServerTiming(timing string) *Response {
	r.headers.Set("Server-Timing", timing)
	return r
}

// ServiceWorkerAllowed sets the "Service-Worker-Allowed" header in the response.
func (r *Response) ServiceWorkerAllowed(scope string) *Response {
	r.headers.Set("Service-Worker-Allowed", scope)
	return r
}

// SourceMap sets the "SourceMap" header in the response.
func (r *Response) SourceMap(url string) *Response {
	r.headers.Set("SourceMap", url)
	return r
}

// HxLocation sets the HX-Location header.
func (r *Response) HxLocation(location string) *Response {
	r.headers.Set("HX-Location", location)
	return r
}

// HxPushUrl sets the HX-Push-Url header.
func (r *Response) HxPushUrl(url string) *Response {
	r.headers.Set("HX-Push-Url", url)
	return r
}

// HXRedirect sets the HX-Redirect header.
func (r *Response) HxRedirect(location string) *Response {
	r.headers.Set("HX-Redirect", location)
	return r
}

// HxRefresh sets the HX-Refresh header to true.
func (r *Response) HxRefresh() *Response {
	r.headers.Set("HX-Refresh", "true")
	return r
}

// HxReplaceUrl sets the HX-Replace-Url header.
func (r *Response) HxReplaceUrl(url string) *Response {
	r.headers.Set("HX-Replace-Url", url)
	return r
}

// HxReswap sets the HX-Reswap header.
func (r *Response) HxReswap(value string) *Response {
	r.headers.Set("HX-Reswap", value)
	return r
}

// HxRetarget sets the HX-Retarget header.
func (r *Response) HxRetarget(selector string) *Response {
	r.headers.Set("HX-Retarget", selector)
	return r
}

// HxReselect sets the HX-Reselect header.
func (r *Response) HxReselect(selector string) *Response {
	r.headers.Set("HX-Reselect", selector)
	return r
}

// HxTrigger sets the HX-Trigger header.
func (r *Response) HxTrigger(event string) *Response {
	r.headers.Set("HX-Trigger", event)
	return r
}

// HxTriggerAfterSettle sets the HX-Trigger-After-Settle header.
func (r *Response) HxTriggerAfterSettle(event string) *Response {
	r.headers.Set("HX-Trigger-After-Settle", event)
	return r
}

// HxTriggerAfterSwap sets the HX-Trigger-After-Swap header.
func (r *Response) HxTriggerAfterSwap(event string) *Response {
	r.headers.Set("HX-Trigger-After-Swap", event)
	return r
}

// Json sets the response body to a JSON-encoded representation of the provided data.
// The Content-Type header is automatically set to "application/json;charset=UTF-8".
func (r *Response) Json(data any) *Response {
	r.jsonBody = data
	r.ContentType("application/json;charset=UTF-8")
	return r
}

// Html sets the response body to an HTML string.
// The Content-Type header is automatically set to "text/html;charset=UTF-8".
func (r *Response) Html(html string) *Response {
	r.rawBody = []byte(html)
	r.ContentType("text/html;charset=UTF-8")
	return r
}

// Text sets the response body to a plain text string.
// The Content-Type header is automatically set to "text/plain;charset=UTF-8".
func (r *Response) Text(text string) *Response {
	r.rawBody = []byte(text)
	r.ContentType("text/plain;charset=UTF-8")
	return r
}

// Body sets the response body to the provided data and sets the Content-Type header.
func (r *Response) Body(contentType string, data []byte) *Response {
	r.rawBody = data
	r.headers.Set("Content-Type", contentType)
	return r
}

func (r *Response) BodyFn(contentType string, bodyFn BodyFn) *Response {
	r.bodyFn = bodyFn
	r.headers.Set("Content-Type", contentType)
	return r
}

// Write writes the response to the http.ResponseWriter.
// It sets the headers and writes the body to the writer.
func (r *Response) Write(w http.ResponseWriter) error {
	defer func() {
		for _, fn := range r.afterWrite {
			fn()
		}
	}()

	for k, vals := range r.headers {
		for _, val := range vals {
			w.Header().Add(k, val)
		}
	}
	for _, cookie := range r.cookies {
		http.SetCookie(w, cookie)
	}

	body := r.rawBody
	if r.jsonBody != nil {
		b, err := json.Marshal(r.jsonBody)
		if err != nil {
			return err
		}
		body = b
	}
	w.WriteHeader(r.StatusCode)
	if r.bodyFn != nil {
		return r.bodyFn(w)
	}
	if _, err := w.Write(body); err != nil {
		return err
	}

	return nil
}

// AfterWrite adds a function to be called after the response is written.
func (r *Response) AfterWrite(fn func()) *Response {
	r.afterWrite = append(r.afterWrite, fn)
	return r
}
