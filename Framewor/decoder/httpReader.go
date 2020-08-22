/*
 * NETCAP - Traffic Analysis Framework
 * Copyright (c) 2017-2020 Philipp Mieden <dreadl0ck [at] protonmail [dot] ch>
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package decoder

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/dreadl0ck/cryptoutils"

	"github.com/dreadl0ck/netcap/reassembly"
	"github.com/dreadl0ck/netcap/types"
	"github.com/dreadl0ck/netcap/utils"
)

const (
	headerContentType     = "Content-Type"
	headerContentEncoding = "Content-Encoding"

	methodPost = "POST"

	credentialsDecoderName = "Credentials"
)

// cmsHeadersList is the list of identifying headers for CMSs, frontend frameworks, ...
// nolint
var cmsHeadersList = []string{"powered", "X-CDN-Forward", "X-Pardot-Rsp", "Expires", "Weglot-Translated", "X-Powered-By", "X-Drupal-Cache", "X-AH-Environment", "X-Pardot-Route", "Derak-Umbrage", "MicrosoftSharePointTeamServices", "x-platform-processor", "X-Jenkins", "X-Wix-Request-Id", "X-Drectory-Script", "X-B3-Flags", "IBM-Web2-Location", "X-Generated-By", "azure-version", "DNNOutputCache", "x-shopify-stage", "content-disposition", "x-vercel-cache", "X-Powered-PublicCMS", "X-JIVE-USER-ID", "X-Foswikiuri", "cf-ray", "section-io-id", "X-Ghost-Cache-Status", "Access-Control-Allow-Headers", "x-fw-serve", "X-Swiftlet-Cache", "X-EC-Debug", "X-Varnish-Action", "azure-slotname", "X-Pardot-LB", "X-Mod-Pagespeed", "Cookie", "CMS-Version", "x-via-fastly", "X-SharePointHealthScore", "SPRequestGuid", "X-GitHub-Request-Id", "sw-context-token", "X-Akamai-Transformed", "X-Umbraco-Version", "X-Rocket-Nginx-Bypass", "x-fw-server", "x-platform-router", "X-Compressed-By", "Link", "Arastta", "Vary", "X-Protected-By", "WebDevSrc", "x-bubble-capacity-limit", "Fastly-Debug-Digest", "X-Akaunting", "server", "X-Fastcgi-Cache", "x-litespeed-cache", "x-platform-cluster", "X-Firefox-Spdy", "X-GoCache-CacheStatus", "X-B3-ParentSpanId", "X-Includable-Version", "host-header", "cf-cache-status", "x-bubble-capacity-used", "X-Confluence-Request-Time", "section-io-origin-status", "Server", "X-KoobooCMS-Version", "X-Wix-Server-Artifact-Id", "Servlet-engine", "X-Generator", "X-CF2", "X-Page-Speed", "X-CDN", "x-bubble-perf", "x-jive-chrome-wrapped", "X-Backdrop-Cache", "X-Jimdo-Instance", "X-Scholica-Version", "X-Shopery", "X-Supplied-By", "X-Jimdo-Wid", "X-Varnish-Age", "X-DataDome-CID", "X-Powered-By-Plesk", "x-shopid", "x-zendesk-user-id", "azure-sitename", "X-DataDome", "x-fw-static", "X-Foswikiaction", "X-StatusPage-Version", "X-AspNet-Version", "x-vercel-id", "x-pantheon-styx-hostname", "sw-version-id", "X-Dotclear-Static-Cache", "sw-language-id", "Link", "X-Arastta", "sw-invalidation-states", "X-Powered-CMS", "X-Advertising-By", "X-Hacker", "x-now-trace", "X-B3-TraceId", "X-NF-Request-ID", "Composed-By", "solodev_session", "X-Wix-Renderer-Server", "X-CF1", "x-oracle-dms-ecid", "X-Unbounce-PageId", "X-Elcodi", "OracleCommerceCloud-Version", "COMMERCE-SERVER-SOFTWARE", "x-platform-server", "X-WPE-Loopback-Upstream-Addr", "kbn-name", "X-Backside-Transport", "X-Amz-Cf-Id", "X-ATG-Version", "X-Flex-Lang", "X-Jive-Flow-Id", "X-Flow-Powered", "X-Fastly-Request-ID", "X-B3-Sampled", "SharePointHealthScore", "X-ServedBy", "kbn-version", "X-Varnish", "WP-Super-Cache", "Via", "X-Jive-Request-Id", "X-Rack-Cache", "X-dynaTrace-JS-Agent", "X-Streams-Distribution", "X-Lift-Version", "X-Spip-Cache", "section-io-origin-time-seconds", "X-B3-SpanId", "X-StatusPage-Skip-Logging", "X-epages-RequestId", "Itx-Generated-Timestamp", "X-XRDS-Location", "X-MCF-ID", "x-lw-cache", "x-fw-type", "X-JSL", "x-fw-hash", "x-sucuri-id", "X-Varnish-Hostname", "x-kinsta-cache", "X-Content-Encoded-By", "wpe-backend", "x-powered-by", "X-Varnish-Cache", "azure-regionname", "x-envoy-upstream-service-time", "Liferay-Portal", "Powered-By", "X-Pass-Why", "AR-PoweredBy", "X-Global-Transaction-ID", "x-sucuri-cache", "X-Tumblr-User"}

// cmsCookiesList is the list of identifying cookies for CMSs, frontend frameworks, ...
// nolint
var cmsCookiesList = []string{"MRHSHin", "PHPSESSI", "cookie_nam", "dps_site_i", "pyrocm", "_gphw_mod", "OSTSESSI", "i_like_gogit", "koken_referre", "_zendesk_shared_sessio", "cs_secure_sessio", "lm_onlin", "e107_t", "FOSWIKISTRIKEON", "memberstac", "Domai", "bigwareCsi", "vtex_sessio", "LastMRH_Sessio", "OJSSI", "LithiumVisito", "_kjb_sessio", "ekmpowersho", "swell-sessio", "CMSPreferredCultur", "kohanasessio", "iexexchanger_sessio", "NS_VE", "xf_csr", "MoodleSessio", "ASPSESSIO", "ipbWWLsession_i", "OpenGro", "CMSSESSI", "sf_redirec", "F5_S", "laravel_sessio", "MRHSessio", "TI", "ipbWWLmodpid", "PIWIK_SESSI", "_hybri", "botble_sessio", "websale_a", "xi", "MRHSequenc", "ZENDSERVERSESSI", "TWIKISI", "TNE", "_session_i", "__cfdui", "PUBLICCMS_USE", "ci_csrf_toke", "__utm", "_zendesk_cooki", "EPiServe", "nette-browse", "ushahid", "flyspray_projec", "FESESSIONI", "ahoy_trac", "phpb", "F5_HT_shrinke", "InstantCMS[logdate", "VivvoSessionI", "3dvisi", "cpsessio", "ICMSSessio", "AWSAL", "sensorsdata2015jssdkcros", "osCsi", "_solusquar", "JSESSIONI", "i_like_gite", "REVEL_SESSIO", "ImpressCM", "DokuWik", "OCSESSI", "EPiTrac", "MAKACSESSIO", "sensorsdata2015sessio", "ZM_TES", "INVENIOSESSIO", "hotaru_mobil", "wgSessio", "ASP.NET_SessionI", "PrestaSho", "F5_fullW", "com.salesforc", "JTLSHO", "cakeph", "Dynamicwe", "exp_csrf_toke", "_g", "k_visi", "CraftSessionI", "SC_ANALYTICS_GLOBAL_COOKI", "_ga", "graffitibo", "ahoy_visi", "xf_sessio", "_help_center_sessio", "SFOSWIKISI", "YII_CSRF_TOKE", "PLAY_SESSIO", "cprelogi", "fronten", "_gitlab_sessio", "_redmine_sessio", "exp_tracke", "spincms_sessio", "bblastvisi", "ci_sessio", "__derak_use", "REVEL_FLAS", "ARRAffinit", "bf_sessio", "ahoy_visito", "AWSEL", "datadom", "pinoox_sessio", "grwng_ui", "sails.si", "DotNetNukeAnonymou", "blesta_si", "Bugzilla_login_request_cooki", "exp_last_activit", "eZSESSI", "gr_user_i", "ARK_I", "CONCRETE", "_gauges", "TiPMi", "bblastactivit", "uCo", "Grand.custome", "Nop.custome", "AWSALBCOR", "MOODLEID", "VtexWorkspac", "phsi", "__derak_aut", "october_sessio", "bbsessionhas", "MOIN_SESSIO", "VtexFingerPrin", "bigWAdminI"}

// headerForApps is a HTTP header structure.
type headerForApps struct {
	HeaderName  string
	HeaderValue string
}

// cookieForApps is a HTTP cookie structure.
type cookieForApps struct {
	CookieName  string
	CookieValue string
}

// httpMetaStore is a thread safe in-memory store for interesting HTTP artifacts.
type httpMetaStore struct {

	// mapped ip address to server names
	ServerNames map[string]string

	// mapped ip address to user agents
	UserAgents map[string]string

	// mapped ip address to user agents
	Vias map[string]string

	// mapped ip address to user agents
	XPoweredBy map[string]string

	CMSHeaders map[string][]headerForApps

	CMSCookies map[string][]cookieForApps

	sync.Mutex
}

// global store for selected http meta information
// TODO: add a util to dump.
var httpStore = &httpMetaStore{
	ServerNames: make(map[string]string),
	UserAgents:  make(map[string]string),
	Vias:        make(map[string]string),
	XPoweredBy:  make(map[string]string),
	CMSHeaders:  make(map[string][]headerForApps),
	CMSCookies:  make(map[string][]cookieForApps),
}

/*
 * HTTP
 */

type httpRequest struct {
	request   *http.Request
	timestamp string
	clientIP  string
	serverIP  string
}

type httpResponse struct {
	response  *http.Response
	timestamp string
	clientIP  string
	serverIP  string
}

type httpReader struct {
	parent    *tcpConnection
	requests  []*httpRequest
	responses []*httpResponse
}

// Decode parses the stream according to the HTTP protocol.
func (h *httpReader) Decode() {
	// parse conversation
	var (
		buf         bytes.Buffer
		previousDir reassembly.TCPFlowDirection
	)

	if len(h.parent.merged) > 0 {
		previousDir = h.parent.merged[0].dir
	}

	for _, d := range h.parent.merged {
		if d.dir == previousDir {
			buf.Write(d.raw)
		} else {
			var err error

			b := bufio.NewReader(&buf)
			if previousDir == reassembly.TCPDirClientToServer {
				for !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
					err = h.readRequest(b)
				}
			} else {
				for !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
					err = h.readResponse(b)
				}
			}
			if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
				utils.DebugLog.Println("error reading HTTP", err, h.parent.ident)
			}
			buf.Reset()
			previousDir = d.dir

			buf.Write(d.raw)

			continue
		}
	}

	var err error

	b := bufio.NewReader(&buf)

	if previousDir == reassembly.TCPDirClientToServer {
		for !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
			err = h.readRequest(b)
		}
	} else {
		for !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
			err = h.readResponse(b)
		}
	}
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// iterate over responses
	for _, res := range h.responses { // populate types.HTTP with all infos from response
		ht := newHTTPFromResponse(res.response)

		_ = h.findRequest(res.response)

		atomic.AddInt64(&stats.numResponses, 1)

		// now add request information
		if res.response.Request != nil {
			if isCustomDecoderLoaded(credentialsDecoderName) {
				h.searchForLoginParams(res.response.Request)
				h.searchForBasicAuth(res.response.Request)
			}

			atomic.AddInt64(&stats.numRequests, 1)
			setRequest(ht, &httpRequest{
				request:   res.response.Request,
				timestamp: res.timestamp,
				clientIP:  res.clientIP,
				serverIP:  res.serverIP,
			})
		} else {
			// response without matching request
			// dont add to output for now
			atomic.AddInt64(&stats.numUnmatchedResp, 1)

			continue
		}

		h.parent.writeHTTP(ht)
	}

	// iterate over unanswered requests
	for _, req := range h.requests {
		if req != nil {
			ht := &types.HTTP{}
			setRequest(ht, req)

			if isCustomDecoderLoaded(credentialsDecoderName) {
				h.searchForLoginParams(req.request)
				h.searchForBasicAuth(req.request)
			}

			atomic.AddInt64(&stats.numRequests, 1)
			atomic.AddInt64(&stats.numUnansweredRequests, 1)

			h.parent.writeHTTP(ht)
		} else {
			atomic.AddInt64(&stats.numNilRequests, 1)
		}
	}
}

// search request header field for HTTP basic auth.
func (h *httpReader) searchForBasicAuth(req *http.Request) {
	if u, p, ok := req.BasicAuth(); ok {
		if u != "" || p != "" {
			writeCredentials(&types.Credentials{
				Timestamp: utils.TimeToString(h.parent.firstPacket),
				Service:   "HTTP Basic Auth",
				Flow:      h.parent.ident,
				User:      u,
				Password:  p,
			})
		}
	}
}

// search for user name and password in http url params and body params.
func (h *httpReader) searchForLoginParams(req *http.Request) {
	for name, values := range req.Form {
		if !(name == "user" || name == "username") {
			continue
		}

		var (
			pass string
			arr  []string
			ok   bool
		)

		arr, ok = req.Form["pass"]
		if !ok {
			arr = req.Form["password"]
		}

		if len(arr) > 0 {
			pass = strings.Join(arr, "; ")
		}

		writeCredentials(&types.Credentials{
			Timestamp: utils.TimeToString(h.parent.firstPacket),
			Service:   serviceHTTP,
			Flow:      h.parent.ident,
			User:      strings.Join(values, "; "),
			Password:  pass,
			Notes:     "Login Parameters",
		})
	}
}

func (t *tcpConnection) writeHTTP(h *types.HTTP) {
	httpStore.Lock()

	if h.UserAgent != "" {
		if ua, ok := httpStore.UserAgents[h.SrcIP]; ok {
			if !strings.Contains(ua, h.UserAgent) {
				httpStore.UserAgents[h.SrcIP] = ua + "| " + h.UserAgent
			}
		} else {
			httpStore.UserAgents[h.SrcIP] = h.UserAgent
		}
	}

	if h.ServerName != "" {
		if sn, ok := httpStore.ServerNames[h.DstIP]; ok {
			if !strings.Contains(sn, h.ServerName) {
				httpStore.ServerNames[h.DstIP] = sn + "| " + h.ServerName
			}
		} else {
			httpStore.ServerNames[h.DstIP] = h.ServerName
		}
	}

	if val, ok := h.ResponseHeader["Via"]; ok {
		var sn string
		if sn, ok = httpStore.Vias[h.DstIP]; ok {
			if !strings.Contains(sn, val) {
				httpStore.Vias[h.DstIP] = sn + "| " + val
			}
		} else {
			httpStore.Vias[h.DstIP] = val
		}
	}

	if val, ok := h.ResponseHeader["X-Powered-By"]; ok {
		var sn string
		if sn, ok = httpStore.XPoweredBy[h.DstIP]; ok {
			if !strings.Contains(sn, val) {
				httpStore.XPoweredBy[h.DstIP] = sn + "| " + val
			}
		} else {
			httpStore.XPoweredBy[h.DstIP] = val
		}
	}

	// Iterate over the possible CMS headers. If present, add them to the httpStore
	for _, cmsHeader := range cmsHeadersList {
		if x, ok := h.ResponseHeader[cmsHeader]; ok {
			httpStore.CMSHeaders[h.DstIP] = append(httpStore.CMSHeaders[h.DstIP], headerForApps{HeaderName: cmsHeader, HeaderValue: x})
		}
	}

	// If HTTP instructions are sent to set a cookie used by CMSs (of other apps), add the key and possible value to the httpStore
	if toSet, ok := h.ResponseHeader["Set-Cookie"]; ok {
		var (
			parsedCookie = strings.Split(toSet, "=")
			cookieKey    = parsedCookie[0]
			cookieValue  string
		)
		if len(parsedCookie) > 1 {
			cookieValue = parsedCookie[1]
		}
		for _, csmCookie := range cmsCookiesList {
			if cookieKey == csmCookie {
				httpStore.CMSCookies[h.DstIP] = append(httpStore.CMSCookies[h.DstIP], cookieForApps{CookieName: cookieKey, CookieValue: cookieValue})
			}
		}
	}

	httpStore.Unlock()

	// TODO: fixme
	// get source port and convert to integer
	// src, err := strconv.Atoi(tl.TransportFlow().Src().String())
	// if err == nil {
	// 	switch tl.LayerType() {
	// 	case layers.LayerTypeTCP:
	// 		serviceNameSrc = resolvers.LookupServiceByPort(src, typeTCP)
	// 	case layers.LayerTypeUDP:
	// 		serviceNameSrc = resolvers.LookupServiceByPort(src, typeUDP)
	// 	default:
	// 	}
	// }
	// dst, err := strconv.Atoi(tl.TransportFlow().Dst().String())
	// if err == nil {
	// 	switch tl.LayerType() {
	// 	case layers.LayerTypeTCP:
	// 		serviceNameDst = resolvers.LookupServiceByPort(dst, typeTCP)
	// 	case layers.LayerTypeUDP:
	// 		serviceNameDst = resolvers.LookupServiceByPort(dst, typeUDP)
	// 	default:
	// 	}
	// }

	if conf.IncludePayloads {
		h.RequestBody = t.client.DataSlice().bytes()
		h.ResponseBody = t.server.DataSlice().bytes()
	}

	// export metrics if configured
	if conf.ExportMetrics {
		h.Inc()
	}

	// write record to disk
	atomic.AddInt64(&httpDecoder.numRecords, 1)
	err := httpDecoder.writer.Write(h)
	if err != nil {
		errorMap.Inc(err.Error())
	}

	soft := whatSoftwareHTTP(t.ident, h)

	if len(soft) == 0 {
		return
	}

	// TODO: pass update func
	writeSoftware(soft, nil)
}

// HTTP Response

func (h *httpReader) readResponse(b *bufio.Reader) error {
	// try to read HTTP response from the buffered reader
	res, err := http.ReadResponse(b, nil)
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return err
	} else if err != nil {
		logReassemblyError("HTTP-response", "HTTP/%s Response error: %s (%v,%+v)\n", h.parent.ident, err, err, err)
		return err
	}

	body, err := ioutil.ReadAll(res.Body)
	s := len(body)
	if err != nil {
		logReassemblyError("HTTP-response-body", "HTTP/%s: failed to get body(parsed len:%d): %s\n", h.parent.ident, s, err)
	} else {
		_ = res.Body.Close()

		// Restore body so it can be read again
		res.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}
	//if h.parent.hexdump {
	//	logReassemblyInfo("Body(%d/0x%x)\n%s\n", len(body), len(body), hex.Dump(body))
	//}

	sym := ","
	if res.ContentLength > 0 && res.ContentLength != int64(s) {
		sym = "!="
	}

	// determine content type for debug log
	contentType, ok := res.Header[headerContentType]
	if !ok {
		contentType = []string{http.DetectContentType(body)}
	}

	encoding := res.Header[headerContentEncoding]
	logReassemblyInfo("HTTP/%s Response: %s (%d%s%d%s) -> %s\n", h.parent.ident, res.Status, res.ContentLength, sym, s, contentType, encoding)

	// increment counter
	stats.Lock()
	stats.responses++
	stats.Unlock()

	h.parent.Lock()
	h.responses = append(h.responses, &httpResponse{
		response:  res,
		timestamp: utils.TimeToString(h.parent.firstPacket),
		clientIP:  h.parent.net.Src().String(),
		serverIP:  h.parent.net.Dst().String(),
	})
	h.parent.Unlock()

	// write responses to disk if configured
	if (err == nil || conf.WriteIncomplete) && conf.FileStorage != "" {
		h.parent.Lock()
		var (
			name         = "unknown"
			source       = "HTTP RESPONSE"
			ctype        string
			numResponses = len(h.responses)
			numRequests  = len(h.requests)
		)
		h.parent.Unlock()

		// check if there is a matching request for the current stream
		if numRequests >= numResponses { // fetch it
			h.parent.Lock()
			req := h.requests[numResponses-1]
			h.parent.Unlock()
			if req != nil {
				name = path.Base(req.request.URL.Path)
				source += " from " + req.request.URL.Path
				ctype = strings.Join(req.request.Header[headerContentType], " ")
			}
		}

		// save file to disk
		return h.saveFile(source, name, err, body, encoding, ctype)
	}

	return nil
}

func (h *httpReader) findRequest(res *http.Response) string {
	// try to find the matching HTTP request for the response
	var (
		req    *http.Request
		reqURL string
	)

	h.parent.Lock()
	if len(h.requests) != 0 {
		// take the request from the parent stream and delete it from there
		req, h.requests = h.requests[0].request, h.requests[1:]
		reqURL = req.URL.String()
	}
	h.parent.Unlock()

	// set request instance on response
	if req != nil {
		res.Request = req
		atomic.AddInt64(&stats.numFoundRequests, 1)
	}

	return reqURL
}

func fileExtensionForContentType(typ string) string {
	parts := strings.Split(typ, ";")
	if len(typ) > 1 {
		typ = parts[0]
	}

	// types from: https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Common_types
	switch typ {
	case "application/x-gzip":
		return ".gz"
	case "image/jpg":
		return ".jpg"
	case "text/plain":
		return ".txt"
	case "text/html":
		return ".html"
	case "image/x-icon":
		return ".ico"
	case "audio/aac":
		return ".aac"
	case "application/x-abiword":
		return ".abw"
	case "application/x-freearc":
		return ".arc"
	case "video/x-msvideo":
		return ".avi"
	case "application/vnd.amazon.ebook":
		return ".azw"
	case "application/octet-stream":
		return binaryFileExtension
	case "image/bmp":
		return ".bmp"
	case "application/x-bzip":
		return ".bz"
	case "application/x-bzip2":
		return ".bz2"
	case "application/x-csh":
		return ".csh"
	case "text/css":
		return ".css"
	case "text/csv":
		return ".csv"
	case "application/msword":
		return ".doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return ".docx"
	case "application/vnd.ms-fontobject":
		return ".eot"
	case "application/epub+zip":
		return ".epub"
	case "application/gzip":
		return ".gz"
	case "image/gif":
		return ".gif"
	case "image/vnd.microsoft.icon":
		return ".ico"
	case "text/calendar":
		return ".ics"
	case "application/java-archive":
		return ".jar"
	case "image/jpeg":
		return ".jpg"
	case "text/javascript":
		return ".js"
	case "application/json":
		return ".json"
	case "application/ld+json":
		return ".jsonld"
	case "audio/midi audio/x-midi":
		return ".midi"
	case "audio/mpeg":
		return ".mp3"
	case "video/mpeg":
		return ".mpeg"
	case "text/xml":
		return ".xml"
	case "application/vnd.apple.installer+xml":
		return ".mpkg"
	case "application/vnd.oasis.opendocument.presentation":
		return ".odp"
	case "application/vnd.oasis.opendocument.spreadsheet":
		return ".ods"
	case "application/vnd.oasis.opendocument.text":
		return ".odt"
	case "audio/ogg":
		return ".oga"
	case "video/ogg":
		return ".ogv"
	case "application/ogg":
		return ".ogx"
	case "audio/opus":
		return ".opus"
	case "font/otf":
		return ".otf"
	case "image/png":
		return ".png"
	case "application/pdf":
		return ".pdf"
	case "application/php":
		return ".php"
	case "application/vnd.ms-powerpoint":
		return ".ppt"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return ".pptx"
	case "application/vnd.rar":
		return ".rar"
	case "application/rtf":
		return ".rtf"
	case "application/x-sh":
		return ".sh"
	case "image/svg+xml":
		return ".svg"
	case "application/x-shockwave-flash":
		return ".swf"
	case "application/x-tar":
		return ".tar"
	case "image/tiff":
		return ".tiff"
	case "video/mp2t":
		return ".ts"
	case "font/ttf":
		return ".ttf"
	case "application/vnd.visio":
		return ".vsd"
	case "audio/wav":
		return ".wav"
	case "audio/webm":
		return ".weba"
	case "video/webm":
		return ".webm"
	case "image/webp":
		return ".webp"
	case "font/woff":
		return ".woff"
	case "font/woff2":
		return ".woff2"
	case "application/xhtml+xml":
		return ".xhtml"
	case "application/vnd.ms-excel":
		return ".xls"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return ".xlsx"
	case "application/xml":
		return ".xml"
	case "application/vnd.mozilla.xul+xml":
		return ".xul"
	case "application/zip":
		return ".zip"
	case "video/3gpp":
		return ".3gp"
	case "video/3gpp2":
		return ".3g2"
	case "application/x-7z-compressed":
		return ".7z"
	}

	return ""
}

func trimEncoding(ctype string) string {
	parts := strings.Split(ctype, ";")
	if len(parts) > 1 {
		return parts[0]
	}
	return ctype
}

// keep track which paths for content types of extracted files have already been created.
var (
	contentTypeMap   = make(map[string]struct{})
	contentTypeMapMu sync.Mutex
)

// createContentTypePathIfRequired will create the passed in filesystem path once
// it is safe for concurrent access and will block until the path has been created on disk.
func createContentTypePathIfRequired(fsPath string) {
	contentTypeMapMu.Lock()
	if _, ok := contentTypeMap[fsPath]; !ok { // the path has not been created yet
		// add to map
		contentTypeMap[fsPath] = struct{}{}

		// create path
		err := os.MkdirAll(fsPath, defaultDirectoryPermission)
		if err != nil {
			logReassemblyError("HTTP-create-path", "Cannot create folder %s: %s\n", fsPath, err)
		}
	}
	// free lock again
	contentTypeMapMu.Unlock()
}

// TODO: write unit tests and cleanup.
func (h *httpReader) saveFile(source, name string, err error, body []byte, encoding []string, contentType string) error {
	// prevent saving zero bytes
	if len(body) == 0 {
		return nil
	}

	if name == "" || name == "/" {
		name = "unknown"
	}

	var (
		fileName string

		// detected content type
		ctype = trimEncoding(http.DetectContentType(body))

		// root path
		root = path.Join(conf.FileStorage, ctype)

		// file extension
		ext = fileExtensionForContentType(ctype)

		// file basename
		base = filepath.Clean(name+"-"+path.Base(h.parent.ident)) + ext
	)
	if err != nil {
		base = "incomplete-" + base
	}
	if filepath.Ext(name) == "" {
		fileName = name + ext
	} else {
		fileName = name
	}

	// make sure root path exists
	createContentTypePathIfRequired(root)

	// add base
	base = path.Join(root, base)
	if len(base) > 250 {
		base = base[:250] + "..."
	}
	if base == conf.FileStorage {
		base = path.Join(conf.FileStorage, "noname")
	}
	var (
		target = base
		n      = 0
	)
	for {
		_, errStat := os.Stat(target)
		if errStat != nil {
			break
		}

		if err != nil {
			target = path.Join(root, filepath.Clean("incomplete-"+name+"-"+h.parent.ident)+"-"+strconv.Itoa(n)+fileExtensionForContentType(ctype))
		} else {
			target = path.Join(root, filepath.Clean(name+"-"+h.parent.ident)+"-"+strconv.Itoa(n)+fileExtensionForContentType(ctype))
		}

		n++
	}

	utils.DebugLog.Println("saving file:", target)

	f, err := os.Create(target)
	if err != nil {
		logReassemblyError("HTTP-create", "Cannot create %s: %s\n", target, err)
		return err
	}

	var (
		// explicitly declare io.Reader interface
		r      io.Reader
		length int
		hash   string
	)

	// now assign a new buffer
	r = bytes.NewBuffer(body)
	if len(encoding) > 0 && (encoding[0] == "gzip" || encoding[0] == "deflate") {
		r, err = gzip.NewReader(r)
		if err != nil {
			logReassemblyError("HTTP-gunzip", "Failed to gzip decode: %s", err)
		}
	}

	if err == nil {
		var written int64
		written, err = io.Copy(f, r)

		if err != nil {
			logReassemblyError("HTTP-save", "%s: failed to copy %s (l:%d): %s\n", h.parent.ident, target, written, err)
		}

		if _, ok := r.(*gzip.Reader); ok {
			err = r.(*gzip.Reader).Close()
			if err != nil {
				logReassemblyError("HTTP-save", "%s: failed to close gzip reader %s (l:%d): %s\n", h.parent.ident, target, written, err)
			}
		}

		err = f.Close()
		if err != nil {
			logReassemblyError("HTTP-save", "%s: failed to close %s (l:%d): %s\n", h.parent.ident, target, written, err)
		} else {
			logReassemblyInfo("%s: Saved %s (l:%d)\n", h.parent.ident, target, written)
		}

		var data []byte

		// TODO: refactor to avoid reading the file contents into memory again
		data, err = ioutil.ReadFile(target)
		if err == nil {
			// set hash to value for decompressed content and update size
			hash = hex.EncodeToString(cryptoutils.MD5Data(data))
			length = len(data)

			// save previous content type
			ctypeOld := ctype

			// update content type
			ctype = trimEncoding(http.DetectContentType(data))

			// make sure root path exists
			createContentTypePathIfRequired(path.Join(conf.FileStorage, ctype))

			// switch the file extension and the path for the updated content type
			ext = filepath.Ext(target)

			// create new target: trim extension from old one and replace
			// and replace the old content type in the path
			newTarget := strings.Replace(strings.TrimSuffix(target, ext), ctypeOld, ctype, 1) + fileExtensionForContentType(ctype)

			err = os.Rename(target, newTarget)
			if err == nil {
				target = newTarget
			} else {
				fmt.Println("failed to rename file after decompression", err)
			}
		}
	} else {
		hash = hex.EncodeToString(cryptoutils.MD5Data(body))
		length = len(body)
	}

	// write file to disk
	writeFile(&types.File{
		Timestamp:           utils.TimeToString(h.parent.firstPacket),
		Name:                fileName,
		Length:              int64(length),
		Hash:                hash,
		Location:            target,
		Ident:               h.parent.ident,
		Source:              source,
		ContentTypeDetected: ctype,
		ContentType:         contentType,
		Context: &types.PacketContext{
			SrcIP:   h.parent.net.Src().String(),
			DstIP:   h.parent.net.Dst().String(),
			SrcPort: h.parent.transport.Src().String(),
			DstPort: h.parent.transport.Dst().String(),
		},
	})

	return nil
}

// HTTP Request

func (h *httpReader) readRequest(b *bufio.Reader) error {
	req, err := http.ReadRequest(b)
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return err
	} else if err != nil {
		logReassemblyError("HTTP-request", "HTTP/%s Request error: %s (%v,%+v)\n", h.parent.ident, err, err, err)
		return err
	}

	body, err := ioutil.ReadAll(req.Body)
	s := len(body)
	if err != nil {
		logReassemblyError("HTTP-request-body", "Got body err: %s\n", err)
		// continue execution
	} else {
		_ = req.Body.Close()

		// Restore body so it can be read again
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}
	//if h.tcpStreamReader.hexdump {
	//	logReassemblyInfo("Body(%d/0x%x)\n%s\n", len(body), len(body), hex.Dump(body))
	//}

	logReassemblyInfo("HTTP/%s Request: %s %s (body:%d)\n", h.parent.ident, req.Method, req.URL, s)

	h.parent.Lock()
	t := utils.TimeToString(h.parent.firstPacket)
	h.parent.Unlock()

	request := &httpRequest{
		request:   req,
		timestamp: t,
		clientIP:  h.parent.net.Src().String(),
		serverIP:  h.parent.net.Dst().String(),
	}

	// parse form values
	err = req.ParseForm()
	if err != nil {
		logReassemblyError("HTTP-request", "HTTP/%s failed to parse form values: %s (%v,%+v)\n", h.parent.ident, err, err, err)
	}

	// increase counter
	stats.Lock()
	stats.requests++
	stats.Unlock()

	h.parent.Lock()
	h.requests = append(h.requests, request)
	h.parent.Unlock()

	if req.Method == methodPost {
		// write request payload to disk if configured
		if (err == nil || conf.WriteIncomplete) && conf.FileStorage != "" {
			return h.saveFile("HTTP POST REQUEST to "+req.URL.Path, path.Base(req.URL.Path), err, body, req.Header[headerContentEncoding], strings.Join(req.Header[headerContentType], " "))
		}
	}

	return nil
}
