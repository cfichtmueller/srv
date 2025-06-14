// Copyright 2025 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package srv

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	SecFetchSiteCrossSite  = "cross-site"
	SecFetchSiteSameOrigin = "same-origin"
	SecFetchSiteSameSite   = "same-site"
	SecFetchSiteNone       = "none"
	SecFetchModeCors       = "cors"
	SecFetchModeNavigate   = "navigate"
	SecFetchModeNoCors     = "no-cors"
	SecFetchModeSameOrigin = "same-origin"
	SecFetchModeWebsocket  = "websocket"

	SecFetchDestAudio         = "audio"
	SecFetchDestAudioWorklet  = "audioworklet"
	SecFetchDestDocument      = "document"
	SecFetchDestEmbed         = "embed"
	SecFetchDestEmpty         = "empty"
	SecFetchDestFencedframe   = "fencedframe"
	SecFetchDestFont          = "font"
	SecFetchDestFrame         = "frame"
	SecFetchDestIframe        = "iframe"
	SecFetchDestImage         = "image"
	SecFetchDestManifest      = "manifest"
	SecFetchDestObject        = "object"
	SecFetchDestPaintworklet  = "paintworklet"
	SecFetchDestReport        = "report"
	SecFetchDestScript        = "script"
	SecFetchDestServiceworker = "serviceworker"
	SecFetchDestSharedworker  = "sharedworker"
	SecFetchDestStyle         = "style"
	SecFetchDestTrack         = "track"
	SecFetchDestVideo         = "video"
	SecFetchDestWebidentity   = "webidentity"
	SecFetchDestWorker        = "worker"
	SecFetchDestXslt          = "xslt"

	SecPurposePrefetch = "prefetch"
)

var (
	ErrNoBody = errors.New("no requestbody")
)

type contextConfig struct {
	maxMultipartMemory int64
	ipResolver         *IPResolver
}

// Context represents the context of an HTTP request.
type Context struct {
	conf        *contextConfig
	w           http.ResponseWriter
	r           *http.Request
	queryParsed bool
	query       url.Values
	formCache   url.Values
	values      map[string]any
	ipResolved  bool
	ipAddresses []string
}

// NewContext creates a new Context with the given http.ResponseWriter and http.Request.
func NewContext(w http.ResponseWriter, r *http.Request, conf *contextConfig) *Context {
	return &Context{
		w:      w,
		r:      r,
		values: make(map[string]any),
		conf:   conf,
	}
}

// Request returns the http.Request associated with the Context.
func (c *Context) Request() *http.Request {
	return c.r
}

// ClientIP returns the client IP address from the request. When proxies are trusted,
// the address is resolved from proxy headers like X-Forwarded-For. Otherwise, the
// direct remote address is used.
func (c *Context) ClientIP() string {
	if !c.ipResolved {
		c.ipAddresses = c.conf.ipResolver.Resolve(c.r)
		c.ipResolved = true
	}
	return c.ipAddresses[0]
}

// RemoteIP returns the remote IP address from the request.
func (c *Context) RemoteIP() string {
	if !c.ipResolved {
		c.ipAddresses = c.conf.ipResolver.Resolve(c.r)
		c.ipResolved = true
	}
	return c.ipAddresses[len(c.ipAddresses)-1]
}

// PathValue returns the value of the specified path parameter from the request.
func (c *Context) PathValue(name string) string {
	return c.r.PathValue(name)
}

// HasQuery checks if the request has a query parameter with the given key.
func (c *Context) HasQuery(key string) bool {
	if !c.queryParsed {
		c.query = c.r.URL.Query()
	}
	return c.query.Has(key)
}

// Query returns the value of the specified query parameter from the request.
func (c *Context) Query(key string) string {
	if !c.queryParsed {
		c.query = c.r.URL.Query()
	}
	return c.query.Get(key)
}

// IntQuery is a shortcut for IntQueryOrDefault(key, 0)
func (c *Context) IntQuery(key string) (int, *Response) {
	return c.IntQueryOrDefault(key, 0)
}

func (c *Context) IntQueryOrDefault(key string, defaultValue int) (int, *Response) {
	val := c.Query(key)
	if val == "" {
		return defaultValue, nil
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0, Respond().BadRequest(ErrorDto{
			Code:    "BadRequest",
			Message: "invalid value for '" + key + "'",
		})
	}
	return i, nil
}

func (c *Context) StringQuery(key string) (string, *Response) {
	return c.StringQueryOrDefault(key, "")
}

func (c *Context) StringQueryOrDefault(key string, defaultValue string) (string, *Response) {
	val := c.Query(key)
	if val == "" {
		return defaultValue, nil
	}
	s, err := url.QueryUnescape(val)
	if err != nil {
		return "", Respond().BadRequest(ErrorDto{
			Code:    "BadRequest",
			Message: "invalid value for '" + key + "'",
		})
	}
	return s, nil
}

// Header returns the value of the specified header from the request.
func (c *Context) Header(name string) string {
	return c.r.Header.Get(name)
}

// Authorization returns the value of the Authorization header.
func (c *Context) Authorization() string {
	return c.Header("Authorization")
}

// ProxyAuthorization returns the value of the Proxy-Authorization header.
func (c *Context) ProxyAuthorization() string {
	return c.Header("Proxy-Authorization")
}

// CacheControl returns the value of the Cache-Control header.
func (c *Context) CacheControl() string {
	return c.Header("Cache-Control")
}

// IfMatch returns the value of the If-Match header.
func (c *Context) IfMatch() string {
	return c.Header("If-Match")
}

// IfNoneMatch returns the value of the If-None-Match header.
func (c *Context) IfNoneMatch() string {
	return c.Header("If-None-Match")
}

// IfModifiedSince returns the value of the If-Modified-Since header.
func (c *Context) IfModifiedSince() (time.Time, bool, error) {
	ims := c.r.Header.Get("If-Modified-Since")
	if ims == "" {
		return time.Time{}, false, nil
	}
	t, err := http.ParseTime(ims)
	if err != nil {
		return time.Time{}, false, err
	}
	return t, true, nil
}

// IfUnmodifiedSince returns the value of the If-Unmodified-Since header.
func (c *Context) IfUnmodifiedSince() (time.Time, bool, error) {
	ium := c.r.Header.Get("If-Unmodified-Since")
	if ium == "" {
		return time.Time{}, false, nil
	}
	t, err := http.ParseTime(ium)
	if err != nil {
		return time.Time{}, false, err
	}
	return t, true, nil
}

// Connection returns the value of the Connection header.
func (c *Context) Connection() string {
	return c.Header("Connection")
}

// KeepAlive returns the value of the Keep-Alive header.
func (c *Context) KeepAlive() string {
	return c.Header("Keep-Alive")
}

// Accept returns the value of the Accept header.
func (c *Context) Accept() string {
	return c.Header("Accept")
}

// AcceptEncoding returns the value of the Accept-Encoding header.
func (c *Context) AcceptEncoding() string {
	return c.Header("Accept-Encoding")
}

// AcceptLanguage returns the value of the Accept-Language header.
func (c *Context) AcceptLanguage() string {
	return c.Header("Accept-Language")
}

// Expect returns the value of the Expect header.
func (c *Context) Expect() string {
	return c.Header("Expect")
}

// MaxForwards returns the value of the Max-Forwards header.
func (c *Context) MaxForwards() (int, bool, error) {
	raw := c.Header("Max-Forwards")
	if raw == "" {
		return 0, false, nil
	}
	i, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false, err
	}
	return i, true, nil
}

// Cookies returns the cookies from the request.
func (c *Context) Cookies() []*http.Cookie {
	return c.r.Cookies()
}

// Cookie returns the named cookie provided in the request or
// ErrNoCookie if not found. And return the named cookie is unescaped.
// If multiple cookies match the given name, only one cookie will
// be returned.
func (c *Context) Cookie(name string) (string, error) {
	cookie, err := c.r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// AccessControlRequestHeaders returns the value of the Access-Control-Request-Headers header.
func (c *Context) AccessControlRequestHeaders() ([]string, bool) {
	h := c.Header("Access-Control-Request-Headers")
	if h == "" {
		return nil, false
	}
	return strings.Split(h, ", "), true
}

// AccessControlRequestMethod returns the value of the Access-Control-Request-Method header.
func (c *Context) AccessControlRequestMethod() string {
	return c.Header("Access-Control-Request-Method")
}

// Origin returns the value of the Origin header.
func (c *Context) Origin() string {
	return c.Header("Origin")
}

// ContentDisposition returns the value of the Content-Disposition header.
func (c *Context) ContentDisposition() string {
	return c.Header("Content-Disposition")
}

// ContentLength returns the value of the Content-Length header.
func (c *Context) ContentLength() (int64, bool) {
	raw := c.Header("Content-Length")
	if raw == "" {
		return 0, false
	}
	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false
	}
	return i, true
}

// ContentType returns the value of the Content-Type header.
func (c *Context) ContentType() string {
	return c.Header("Content-Type")
}

// ContentEncoding returns the value of the Content-Encoding header.
func (c *Context) ContentEncoding() string {
	return c.Header("Content-Encoding")
}

// ContentLanguage returns the value of the Content-Language header.
func (c *Context) ContentLanguage() string {
	return c.Header("Content-Language")
}

// ContentLocation returns the value of the Content-Location header.
func (c *Context) ContentLocation() string {
	return c.Header("Content-Location")
}

// Prefer returns the value of the Prefer header.
func (c *Context) Prefer() string {
	return c.Header("Prefer")
}

// Forwarded returns the value of the Forwarded header.
func (c *Context) Forwarded() string {
	return c.Header("Forwarded")
}

// Via returns the value of the Via header.
func (c *Context) Via() string {
	return c.Header("Via")
}

// Range returns the value of the Range header.
func (c *Context) Range() string {
	return c.Header("Range")
}

// IfRange returns the value of the If-Range header.
func (c *Context) IfRange() string {
	return c.Header("If-Range")
}

// From returns the value of the From header.
func (c *Context) From() string {
	return c.Header("From")
}

// Host returns the value of the Host header.
func (c *Context) Host() string {
	return c.Header("Host")
}

// Referer returns the value of the Referer header.
func (c *Context) Referer() string {
	return c.Header("Referer")
}

// UserAgent returns the value of the User-Agent header.
func (c *Context) UserAgent() string {
	return c.Header("User-Agent")
}

// Server returns the value of the Server header.
func (c *Context) Server() string {
	return c.Header("Server")
}

// UpgradeInsecureRequests returns the value of the Upgrade-Insecure-Requests header.
func (c *Context) UpgradeInsecureRequests() (bool, bool) {
	raw := c.Header("Upgrade-Insecure-Requests")
	if raw == "" {
		return false, false
	}
	return true, true
}

// SecFetchSite returns the value of the Sec-Fetch-Site header.
func (c *Context) SecFetchSite() string {
	return c.Header("Sec-Fetch-Site")
}

// SecFetchMode returns the value of the Sec-Fetch-Mode header.
func (c *Context) SecFetchMode() string {
	return c.Header("Sec-Fetch-Mode")
}

// SecFetchUser returns the value of the Sec-Fetch-User header.
func (c *Context) SecFetchUser() bool {
	raw := c.Header("Sec-Fetch-User")
	return raw == "?1"
}

// SecFetchDest returns the value of the Sec-Fetch-Dest header.
func (c *Context) SecFetchDest() string {
	return c.Header("Sec-Fetch-Dest")
}

// SecPurpose returns the value of the Sec-Purpose header.
func (c *Context) SecPurpose() string {
	return c.Header("Sec-Purpose")
}

// ServiceWorkerNavigationPreload returns the value of the Service-Worker-Navigation-Preload header.
func (c *Context) ServiceWorkerNavigationPreload() string {
	return c.Header("Service-Worker-Navigation-Preload")
}

// TransferEncoding returns the value of the Transfer-Encoding header.
func (c *Context) TransferEncoding() string {
	return c.Header("Transfer-Encoding")
}

// TE returns the value of the TE header.
func (c *Context) TE() string {
	return c.Header("TE")
}

// Trailer returns the value of the Trailer header.
func (c *Context) Trailer() string {
	return c.Header("Trailer")
}

// Date returns the value of the Date header.
func (c *Context) Date() (time.Time, bool) {
	raw := c.Header("Date")
	if raw == "" {
		return time.Time{}, false
	}
	t, err := http.ParseTime(raw)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// Link returns the value of the Link header.
func (c *Context) Link() string {
	return c.Header("Link")
}

// ServiceWorker returns the value of the Service-Worker header.
func (c *Context) ServiceWorker() bool {
	return c.Header("Service-Worker") == "script"
}

// ConditionalIfMatch makes the request conditional. Returns a response when the precondition fails.
func (c *Context) ConditionalIfMatch(localEtag string) *Response {
	remoteEtag := c.r.Header.Get("If-Match")
	if remoteEtag == "" || "\""+localEtag+"\"" == remoteEtag {
		return nil
	}
	return Respond().PreconditionFailed()
}

// ConditionalIfNoneMatch makes the request conditional. Returns a response when the precondition fails.
func (c *Context) ConditionalIfNoneMatch(localEtag string) *Response {
	remoteEtag := c.r.Header.Get("If-None-Match")
	if remoteEtag == "" || "\""+localEtag+"\"" != remoteEtag {
		return nil
	}
	if c.r.Method == http.MethodGet || c.r.Method == http.MethodHead {
		return Respond().NotModified().ETag(localEtag)
	}
	return Respond().PreconditionFailed()
}

// ConditionalIfModifiedSince makes the request conditional. Returns a response when the precondition fails.
func (c *Context) ConditionalIfModifiedSince(lastModified ...time.Time) *Response {
	t, ok, err := c.IfModifiedSince()
	if err != nil {
		return Respond().BadRequest(ErrorDto{
			Code:    "BadRequest",
			Message: "invalid value for 'If-Modified-Since'",
		})
	}
	if !ok {
		return nil
	}
	lm := maxTime(lastModified).Truncate(time.Second)
	if lm.After(t) {
		return nil
	}
	return Respond().NotModified().LastModified(lm)
}

// BindJSON tries to bind a json payload. Returns a response if the binding was unsuccessful
func (c *Context) BindJSON(data any) *Response {
	b, err := io.ReadAll(c.r.Body)
	if err != nil {
		return respondInternalServerError(err)
	}
	if len(b) == 0 {
		return respondError(http.StatusBadRequest, "RequestBodyMissing", "request body is missing")
	}
	if err := json.Unmarshal(b, data); err != nil {
		return respondError(http.StatusBadRequest, "InvalidRequestBody", err.Error())
	}
	v, ok := data.(Validatable)
	if ok {
		if err := v.Validate(); err != nil {
			if v, ok := err.(*ValidationError); ok {
				return Respond().BadRequest(v)
			}
			return respondError(http.StatusBadRequest, "BadRequest", err.Error())
		}
	}
	return nil
}

// FormValues returns the values from a POST urlencoded form or multipart form
func (c *Context) FormValues() url.Values {
	if c.formCache == nil {
		c.parseForm()
	}
	return c.formCache
}

func (c *Context) parseForm() {
	c.formCache = make(url.Values)
	if err := c.r.ParseMultipartForm(c.conf.maxMultipartMemory); err != nil {
		if !errors.Is(err, http.ErrNotMultipart) {
			slog.Error("unable to parse multipart form", "error", err)
		}
	}
	c.formCache = c.r.PostForm
}

// HxBoosted returns true if the request is an HX-Boosted request.
func (c *Context) HxBoosted() bool {
	return c.Header("HX-Boosted") == "true"
}

// HxCurrentUrl returns the current URL from the HX-Current-URL header.
func (c *Context) HxCurrentUrl() string {
	return c.Header("HX-Current-URL")
}

// HxHistoryRestoreRequest returns true if this request is for history restoration.
func (c *Context) HxHistoryRestoreRequest() bool {
	return c.Header("HX-History-Restore-Request") == "true"
}

// HxPrompt returns the user response to an hx-prompt.
func (c *Context) HxPrompt() string {
	return c.Header("HX-Prompt")
}

// HxRequest returns true if the request is an HX request.
func (c *Context) HxRequest() bool {
	return c.Header("HX-Request") == "true"
}

// HxTarget returns the target element of the request.
func (c *Context) HxTarget() string {
	return c.Header("HX-Target")
}

// HxTriggerName returns the name of the triggered element if it exists.
func (c *Context) HxTriggerName() string {
	return c.Header("HX-Trigger-Name")
}

// HxTrigger returns the ID of the triggered element if it exists.
func (c *Context) HxTrigger() string {
	return c.Header("HX-Trigger")
}

// GetRawData reads the request body and returns the raw data.
// Returns ErrNoBody if the request body is nil.
func (c *Context) GetRawData() ([]byte, error) {
	if c.r.Body == nil {
		return nil, ErrNoBody
	}
	return io.ReadAll(c.r.Body)
}

func (c *Context) Set(key string, value any) {
	c.values[key] = value
}

func (c *Context) Get(key string) (any, bool) {
	v, ok := c.values[key]
	return v, ok
}

func (c *Context) MustGet(key string) any {
	v, ok := c.values[key]
	if !ok {
		panic("didn't find key '" + key + "' in context")
	}
	return v
}

func (c *Context) Deadline() (time.Time, bool) {
	return c.r.Context().Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.r.Context().Done()
}

func (c *Context) Err() error {
	return c.r.Context().Err()
}

func (c *Context) Value(key any) any {
	return c.r.Context().Value(key)
}

func respondInternalServerError(err error) *Response {
	return respondError(http.StatusInternalServerError, "InternalServerError", err.Error())
}

func respondError(statusCode int, code, message string) *Response {
	return Respond().Status(statusCode).Json(ErrorDto{
		Code:    code,
		Message: message,
	})
}
