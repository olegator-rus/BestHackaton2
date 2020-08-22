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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/dreadl0ck/gopacket"
	"github.com/dreadl0ck/gopacket/layers"
	"github.com/dreadl0ck/ja3"
	"github.com/evilsocket/islazy/tui"
	"github.com/gogo/protobuf/proto"
	"github.com/ua-parser/uap-go/uaparser"

	"github.com/dreadl0ck/netcap/dpi"
	"github.com/dreadl0ck/netcap/resolvers"
	"github.com/dreadl0ck/netcap/types"
	"github.com/dreadl0ck/netcap/utils"
)

const (
	protoTCP      = "TCP"
	serviceHTTP   = "HTTP"
	serviceSSH    = "SSH"
	servicePOP3   = "POP3"
	serviceTelnet = "Telnet"
	serviceFTP    = "FTP"
)

type software struct {
	*types.Software
	sync.Mutex
}

// atomicDeviceProfileMap contains all connections and provides synchronized access.
type atomicSoftwareMap struct {
	// mapped product + version to software
	Items map[string]*software
	sync.Mutex
}

var (
	userAgentCaching = make(map[string]*userAgent)
	regExpServerName = regexp.MustCompile(`(.*?)(?:/(.*?))?(?:\s*?)(?:\((.*?)\))?$`)
	regexpXPoweredBy = regexp.MustCompile(`(.*?)(?:(?:\s|/)(.*?))?$`)
	ja3Cache         = make(map[string]string)
	jaCacheMutex     sync.Mutex
	reGenericVersion = regexp.MustCompile(`(?m)(?:^)(.*?)(\d+)\.(\d+)\.(\d+)(.*?)(?:$)`)
	hasshMap         = make(map[string][]sshSoftware)
	// Used to store CMS related information, and to do the CMS lookup.
	cmsDB                = make(map[string]interface{})
	vulnerabilitiesIndex bleve.Index
	exploitsIndex        bleve.Index
)

// Size returns the number of elements in the Items map.
func (a *atomicSoftwareMap) Size() int {
	a.Lock()
	defer a.Unlock()

	return len(a.Items)
}

var (
	// softwareStore hold all connections.
	softwareStore = &atomicSoftwareMap{
		Items: make(map[string]*software),
	}

	parser, errInitUAParser = uaparser.New("/usr/local/etc/netcap/dbs/regexes.yaml")
	pMu                     sync.Mutex

	ja3db     ja3CombinationsDB
	hasshDB   []sshHash
	hashDBMap map[string][]sshSoftware
)

type userAgent struct {
	client  *uaparser.Client
	product string
	vendor  string
	version string
	full    string
}

type process struct {
	Process string `json:"process"`
	JA3     string `json:"JA3"`
	JA3s    string `json:"JA3S"`
}

type client struct {
	Os        string    `json:"os"`
	Arch      string    `json:"arch"`
	Processes []process `json:"processes"`
}

type server struct {
	Server  string   `json:"server"`
	Clients []client `json:"clients"`
}

type ja3CombinationsDB struct {
	Servers []server `json:"servers"`
}

type sshSoftware struct {
	Version    string `json:"name"`
	Likelihood string `json:"likelyhood"` // dont remove this typo, or the hasshdb.json cannot be read!
}

type sshHash struct {
	Hash     string        `json:"hash"`
	Software []sshSoftware `json:"softwares"` // dont remove this typo, or the hasshdb.json cannot be read!
}

// process a raw user agent string and returned a structured instance.
func parseUserAgent(ua string) *userAgent {
	var (
		uaClient                       = parser.Parse(ua)
		full, product, vendor, version string
	)

	if uaClient.UserAgent != nil {
		vendor = uaClient.UserAgent.Family
		version = uaClient.UserAgent.Major
		if uaClient.UserAgent.Minor != "" {
			version += "." + uaClient.UserAgent.Minor
		}
		if uaClient.UserAgent.Patch != "" {
			version += "." + uaClient.UserAgent.Patch
		}
		full += " " + uaClient.UserAgent.Family
		full += " " + uaClient.UserAgent.Major
		full += " " + uaClient.UserAgent.Minor
		full += " " + uaClient.UserAgent.Patch

		if vendor == "Other" {
			vendor = ""
		}
	}
	if uaClient.Os != nil {
		full += " " + uaClient.Os.Family
		full += " " + uaClient.Os.Major
		full += " " + uaClient.Os.Minor
		full += " " + uaClient.Os.Patch
		full += " " + uaClient.Os.PatchMinor
	}
	if uaClient.Device != nil {
		product = uaClient.Device.Family
		full += " " + uaClient.Device.Family

		if product == "Other" {
			product = ""
		}
	}

	// if vendor could not be identified, try to determine based on product name
	// TODO: add utility for this
	if vendor == "" {
		switch product {
		case "Chrome", "Android":
			vendor = "Google"
		case "Firefox":
			vendor = "Mozilla"
		case "Internet Explorer", "IE":
			vendor = "Microsoft"
		case "Safari", "iOS", "macOS":
			vendor = "Apple"
		}
	}

	return &userAgent{
		client:  uaClient,
		product: product,
		vendor:  vendor,
		version: version,
		full:    strings.TrimSpace(full),
	}
}

// generic version harvester, scans the payload using a regular expression.
func softwareHarvester(data []byte, flowIdent string, ts time.Time, service string, dpIdent string, protos []string) (s []*software) {
	matches := reGenericVersion.FindAll(data, -1)

	//fmt.Println("got", len(matches), "matches")
	//for _, m := range matches {
	//	fmt.Println(string(m))
	//}

	if len(matches) > 0 {
		for _, v := range matches {
			s = append(s, &software{
				Software: &types.Software{
					Timestamp:      utils.TimeToString(ts),
					DeviceProfiles: []string{dpIdent},
					SourceName:     "Generic version harvester",
					SourceData:     string(data),
					Service:        service,
					DPIResults:     protos,
					Flows:          []string{flowIdent},
					Notes:          string(v),
				},
			})
		}
	}

	return s
}

// tries to determine the kind of software and version
// based on the provided input data.
func whatSoftware(dp *deviceProfile, i *packetInfo, flowIdent, serviceNameSrc, serviceNameDst, JA3, JA3s string, protos []string) (s []*software) {
	var (
		serviceIdent string
		dpIdent      = dp.MacAddr
	)

	if serviceNameSrc != "" {
		serviceIdent = serviceNameSrc
	}

	if serviceNameDst != "" {
		serviceIdent = serviceNameDst
	}

	if dp.DeviceManufacturer != "" {
		dpIdent += " <" + dp.DeviceManufacturer + ">"
	}

	// Only do JA3 fingerprinting when both ja3 and ja3s are present, aka when the server Hello is captured
	// TODO: improve this loops efficiency
	if len(JA3) > 0 && len(JA3s) > 0 {
		for _, serverIdent := range ja3db.Servers {
			serverName := serverIdent.Server

			for _, clientIdent := range serverIdent.Clients {
				clientName := clientIdent.Os + "(" + clientIdent.Arch + ")"

				for _, processInstance := range clientIdent.Processes {
					processName := processInstance.Process

					if processInstance.JA3 == JA3 && processInstance.JA3s == JA3s {
						pMu.Lock()
						values := regExpServerName.FindStringSubmatch(serverName)

						s = append(s, &software{
							Software: &types.Software{
								Timestamp:      i.timestamp,
								Product:        values[1], // Name of the server (Apache, Nginx, ...)
								Version:        values[2], // Version as found after the '/'
								Vendor:         values[3], // Often the operating system
								DeviceProfiles: []string{dpIdent},
								SourceName:     "JA3s",
								SourceData:     JA3s,
								Service:        serviceIdent,
								DPIResults:     protos,
								Flows:          []string{flowIdent},
							},
						}, &software{
							Software: &types.Software{
								Timestamp:      i.timestamp,
								Product:        processName, // Name of the browser, including version
								Vendor:         clientName,  // Name of the OS
								Version:        "",          // TODO parse client name
								DeviceProfiles: []string{dpIdent},
								SourceName:     "JA3",
								SourceData:     JA3,
								Service:        serviceIdent,
								DPIResults:     protos,
								Flows:          []string{flowIdent},
							},
						})
						pMu.Unlock()
					}
				}
			}
		}
	}

	// if nothing was found with all above attempts, try to throw the generic version number harvester at it
	// and see if this delivers anything interesting
	if len(s) == 0 {
		return softwareHarvester(i.p.Data(), flowIdent, i.p.Metadata().CaptureInfo.Timestamp, serviceIdent, dpIdent, protos)
	}

	return s
}

// TODO: pass in the device profile.
func whatSoftwareHTTP(flowIdent string, h *types.HTTP) (s []*software) {
	// HTTP User Agents
	// TODO: check for userAgents retrieved by Ja3 lookup as well
	if len(h.UserAgent) != 0 && h.UserAgent != " " {

		pMu.Lock()
		userInfo, ok := userAgentCaching[h.UserAgent]
		if !ok {
			userInfo = parseUserAgent(h.UserAgent)
			userAgentCaching[h.UserAgent] = userInfo
			utils.DebugLog.Println("UserAgent:", userInfo.full)
		}
		pMu.Unlock()

		s = append(s, &software{
			Software: &types.Software{
				Timestamp: h.Timestamp,
				Product:   userInfo.product,
				Vendor:    userInfo.vendor,
				Version:   userInfo.version,
				// DeviceProfiles: []string{dpIdent},
				SourceName: "UserAgent",
				SourceData: h.UserAgent,
				Service:    serviceHTTP,
				Flows:      []string{flowIdent},
				Notes:      userInfo.full,
			},
		})
	}

	// HTTP Server Name
	if len(h.ServerName) != 0 && h.ServerName != " " {
		values := regExpServerName.FindStringSubmatch(h.ServerName)
		s = append(s, &software{
			Software: &types.Software{
				Timestamp: h.Timestamp,
				Product:   values[1],                // Name of the server (Apache, Nginx, ...)
				Notes:     "Maybe OS: " + values[3], // potentially operating system
				Version:   values[2],                // Version as found after the '/'
				// DeviceProfiles: []string{dpIdent},
				SourceName: "ServerName",
				SourceData: h.ServerName,
				Service:    serviceHTTP,
				Flows:      []string{flowIdent},
			},
		})
	}

	// X-Powered-By HTTP Header
	if poweredBy, ok := h.ResponseHeader["X-Powered-By"]; ok {
		if len(poweredBy) != 0 && poweredBy != " " {
			values := regexpXPoweredBy.FindStringSubmatch(poweredBy)
			s = append(s, &software{
				Software: &types.Software{
					Timestamp: h.Timestamp,
					Product:   values[1], // Name of the server (Apache, Nginx, ...)
					Version:   values[2], // Version as found after the '/'
					// DeviceProfiles: []string{dpIdent},
					SourceName: "X-Powered-By",
					SourceData: poweredBy,
					Service:    serviceHTTP,
					Flows:      []string{flowIdent},
				},
			})
		}
	}

	// Try to detect apps
	// TODO: optimize this loops performance
	httpStore.Lock()
	if receivedHeaders, ok := httpStore.CMSHeaders[h.DstIP]; ok {
		httpStore.Unlock()

		var (
			headers map[string]interface{}
			hdrs    interface{}
		)

		for k, v := range cmsDB {
			if headers, ok = v.(map[string]interface{}); ok {
				if hdrs, ok = headers["headers"]; ok {
					for key, val := range hdrs.(map[string]interface{}) {
						for _, receivedHeader := range receivedHeaders {
							re, err := regexp.Compile(val.(string))
							if err != nil {
								fmt.Println("Failed to compile:    " + val.(string))
							} else if strings.EqualFold(receivedHeader.HeaderName, key) &&
								(re.MatchString(receivedHeader.HeaderValue) || val == "") {
								s = append(s, &software{
									Software: &types.Software{
										Timestamp:  h.Timestamp,
										Product:    k,
										Version:    "",
										SourceName: key,
										Service:    serviceHTTP,
										Flows:      []string{flowIdent},
									},
								})
							}
						}
					}
				}
			}
		}
	} else {
		httpStore.Unlock()
	}

	return s
}

// analyzeSoftware tries to identify software based on observations from the data
// this function first gathers as much data as possible and then calls into whatSoftware
// to determine what software the packet belongs to.
func analyzeSoftware(i *packetInfo) {
	var (
		serviceNameSrc, serviceNameDst string
		ja3Hash                        = ja3.DigestHexPacket(i.p)
		JA3s                           string
		JA3                            string
		protos                         []string
		f                              string
	)

	if ja3Hash == "" {
		ja3Hash = ja3.DigestHexPacketJa3s(i.p)
	}

	// Lookup Service For Port Numbers
	if tl := i.p.TransportLayer(); tl != nil { // set flow ident
		f = i.srcIP + ":" + tl.TransportFlow().Src().String() + "->" + i.dstIP + ":" + tl.TransportFlow().Dst().String()

		// get source port and convert to integer
		src, err := strconv.Atoi(tl.TransportFlow().Src().String())
		if err == nil {
			switch tl.LayerType() {
			case layers.LayerTypeTCP:
				serviceNameSrc = resolvers.LookupServiceByPort(src, typeTCP)
			case layers.LayerTypeUDP:
				serviceNameSrc = resolvers.LookupServiceByPort(src, typeUDP)
			default:
			}
		}
		dst, err := strconv.Atoi(tl.TransportFlow().Dst().String())
		if err == nil {
			switch tl.LayerType() {
			case layers.LayerTypeTCP:
				serviceNameDst = resolvers.LookupServiceByPort(dst, typeTCP)
			case layers.LayerTypeUDP:
				serviceNameDst = resolvers.LookupServiceByPort(dst, typeUDP)
			default:
			}
		}
	} else {
		// no transport layer
		f = i.srcIP + "->" + i.dstIP
	}

	// Deep Packet Inspection
	results := dpi.GetProtocols(i.p)
	for p := range results {
		protos = append(protos, p)
	}

	// The underlying assumption is that we will always observe a client TLS Hello before seeing a server TLS Hello
	// Assuming the packet captured corresponds to the server Hello, first try to see if a client Hello (client being the
	// destination IP) was observed. If not, this is the client. Therefore add client ja3 signature to the store.
	if len(ja3Hash) > 0 {
		var ok bool
		jaCacheMutex.Lock()
		JA3, ok = ja3Cache[i.dstIP]
		jaCacheMutex.Unlock()
		if !ok {
			jaCacheMutex.Lock()
			ja3Cache[i.srcIP] = ja3Hash
			jaCacheMutex.Unlock()
			JA3 = ""
			JA3s = ""
		} else {
			JA3s = ja3Hash
		}
	}

	// fetch the associated device profile
	dp := getDeviceProfile(i.srcMAC, i)

	// now that we have some information at hands
	// try to determine what kind of software it is
	soft := whatSoftware(dp, i, f, serviceNameSrc, serviceNameDst, JA3, JA3s, protos)
	if len(soft) == 0 {
		return
	}

	writeSoftware(soft, func(s *software) {
		updateSoftwareAuditRecord(dp, s, i)
	})
}

func writeSoftware(software []*software, update func(s *software)) {
	var newSoftwareProducts []*types.Software

	// add new audit records or update existing
	softwareStore.Lock()
	for _, s := range software {
		if s == nil {
			continue
		}
		s.Lock()
		if s.Software == nil {
			s.Unlock()
			continue
		}
		ident := s.Product + "/" + s.Version

		// trim version field if its too long
		// likely a regex matched too much text
		if len(s.Version) > 10 {
			s.Version = s.Version[:10] + "..."
		}
		s.Unlock()
		if item, ok := softwareStore.Items[ident]; ok {
			if update != nil {
				update(item)
			}
		} else {
			// fmt.Println(SoftwareStore.Items, s.Product, s.Version)
			softwareStore.Items[ident] = s

			stats.Lock()
			stats.numSoftware++
			stats.Unlock()

			newSoftwareProducts = append(newSoftwareProducts, s.Software)
		}
	}
	softwareStore.Unlock()

	if len(newSoftwareProducts) > 0 {
		// lookup known issues with identified software in the background
		go func() {
			for _, s := range newSoftwareProducts {
				vulnerabilitiesLookup(s)
				exploitsLookup(s)
			}
		}()
	}
}

// newSoftware creates a new device specific profile.
func newSoftware(i *packetInfo) *software {
	return &software{
		Software: &types.Software{
			Timestamp: i.timestamp,
		},
	}
}

func updateSoftwareAuditRecord(dp *deviceProfile, p *software, i *packetInfo) {
	dpIdent := dp.MacAddr
	if dp.DeviceManufacturer != "" {
		dpIdent += " <" + dp.DeviceManufacturer + ">"
	}

	p.Lock()
	for _, pr := range p.DeviceProfiles {
		if pr == dpIdent {
			p.Unlock()
			return
		}
	}
	p.DeviceProfiles = append(p.DeviceProfiles, dpIdent)
	tl := i.p.TransportLayer()
	if tl != nil {
		p.Flows = append(p.Flows, i.srcIP+":"+tl.TransportFlow().Src().String()+"->"+i.dstIP+":"+tl.TransportFlow().Dst().String())
	} else {
		// no transport layer
		p.Flows = append(p.Flows, i.srcIP+"->"+i.dstIP)
	}
	p.Unlock()
}

var softwareDecoder = newCustomDecoder(
	types.Type_NC_Software,
	"Software",
	"A software product that was observed on the network",
	func(d *customDecoder) error {
		if errInitUAParser != nil {
			return errInitUAParser
		}

		// Load the JSON database of JA3/JA3S combinations into memory
		data, err := ioutil.ReadFile("/usr/local/etc/netcap/dbs/ja_3_3s.json")
		if err != nil {
			return err
		}

		// unpack JSON
		err = json.Unmarshal(data, &ja3db.Servers)
		if err != nil {
			return err
		}

		// Load the JSON database of HASSH signaures
		data, err = ioutil.ReadFile("/usr/local/etc/netcap/dbs/hasshdb.json")
		if err != nil {
			return err
		}

		// unpack JSON
		err = json.Unmarshal(data, &hasshDB)
		if err != nil {
			return err
		}

		hashDBMap = make(map[string][]sshSoftware)

		for _, v := range hasshDB {
			hashDBMap[v.Hash] = v.Software
		}

		utils.DebugLog.Println("loaded", len(hashDBMap), "different HASSH digests")

		data, err = ioutil.ReadFile("/usr/local/etc/netcap/dbs/cmsdb.json")
		if err != nil {
			return err
		}

		err = json.Unmarshal(data, &cmsDB)
		if err != nil {
			return err
		}

		for _, entry := range hasshDB {
			hasshMap[entry.Hash] = entry.Software // TODO: note from Giac: Holds redundant info, but couldn't figure a more elegant way to do this
		}

		// Load vulnerabilities DB index
		indexName := filepath.Join(resolvers.DataBaseSource, vulnDBPath)
		vulnerabilitiesIndex, err = utils.OpenBleve(indexName)
		if err != nil {
			return err
		}

		utils.DebugLog.Println("loaded Ja3/ja3S database, records:", len(ja3db.Servers))

		return nil
	},
	func(p gopacket.Packet) proto.Message {
		// handle packet
		analyzeSoftware(newPacketInfo(p))

		return nil
	},
	func(e *customDecoder) error {
		httpStore.Lock()
		var rows [][]string
		for ip, ua := range httpStore.UserAgents {
			rows = append(rows, []string{ip, ua})
		}
		tui.Table(utils.DebugLogFileHandle, []string{"IP", "UserAgents"}, rows)
		rows = [][]string{}
		for ip, sn := range httpStore.ServerNames {
			rows = append(rows, []string{ip, sn})
		}
		tui.Table(utils.DebugLogFileHandle, []string{"IP", "ServerNames"}, rows)
		httpStore.Unlock()

		// teardown DPI C libs
		dpi.Destroy()

		// flush writer
		for _, item := range softwareStore.Items {
			item.Lock()
			e.write(item.Software)
			item.Unlock()
		}

		utils.CloseBleve(vulnerabilitiesIndex)

		return nil
	},
)

// TODO: move into customDecoder and use in other places to remove unnecessary package level decoders
// writeProfile writes the profile.
func (cd *customDecoder) write(r types.AuditRecord) {
	if conf.ExportMetrics {
		r.Inc()
	}

	atomic.AddInt64(&cd.numRecords, 1)
	err := cd.writer.Write(r.(proto.Message))
	if err != nil {
		log.Fatal("failed to write proto: ", err)
	}
}
