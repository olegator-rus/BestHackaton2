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

package transform

import (
	"fmt"
	"log"
	"strings"

	"github.com/dreadl0ck/netcap/maltego"
	"github.com/dreadl0ck/netcap/resolvers"
	"github.com/dreadl0ck/netcap/types"
	"github.com/dreadl0ck/netcap/utils"
)

func lookupDHCPFingerprint() {
	// init API key
	resolvers.InitDHCPFingerprintAPIKey()
	resolvers.InitDHCPFingerprintDB()

	// read HTTP audit records and create a map of ips to useragents
	userAgentStore := make(map[string][]string)

	maltego.HTTPTransform(
		nil,
		func(lt maltego.LocalTransform, trx *maltego.Transform, http *types.HTTP, min, max uint64, profilesFile string, ipaddr string) {
			if uas, ok := userAgentStore[http.SrcIP]; ok {
				for _, u := range uas {
					if u == http.UserAgent {
						// already collected
						return
					}
				}
				// collect
				userAgentStore[http.SrcIP] = append(uas, http.UserAgent)
			} else {
				userAgentStore[http.SrcIP] = []string{http.UserAgent}
			}
		},
		true,
	)

	log.Println("userAgentStore", len(userAgentStore))
	for ip, uas := range userAgentStore {
		log.Println(ip, uas)
	}

	var (
		fp, mac string
		// mapped MAC addresses to IPs
		addrMapping          = make(map[string]string)
		mtrx                 *maltego.Transform
		messageToFingerprint *types.DHCPv4
	)

	maltego.DHCPTransform(
		nil,
		func(lt maltego.LocalTransform, trx *maltego.Transform, dhcp *types.DHCPv4, min, max uint64, profilesFile string, ipaddr string) {
			if dhcp.Operation == 2 {
				if _, ok := addrMapping[dhcp.ClientHWAddr]; !ok {
					log.Println("update addr mapping", dhcp.ClientHWAddr, dhcp.YourClientIP)
					addrMapping[dhcp.ClientHWAddr] = dhcp.YourClientIP
				}

				return
			}

			if fp == "" && mac == "" {
				mac = lt.Values["clientMac"]
				fp = lt.Values["fp"]
				log.Println("searching for mac", mac, "fp", fp)
			}
			if dhcp.ClientHWAddr == mac && dhcp.Fingerprint == fp { // deep copy
				messageToFingerprint = &types.DHCPv4{
					Timestamp:    dhcp.Timestamp,
					Operation:    dhcp.Operation,
					HardwareType: dhcp.HardwareType,
					HardwareLen:  dhcp.HardwareLen,
					HardwareOpts: dhcp.HardwareOpts,
					Xid:          dhcp.Xid,
					Secs:         dhcp.Secs,
					Flags:        dhcp.Flags,
					ClientIP:     dhcp.ClientIP,
					YourClientIP: dhcp.YourClientIP,
					NextServerIP: dhcp.NextServerIP,
					RelayAgentIP: dhcp.RelayAgentIP,
					ClientHWAddr: dhcp.ClientHWAddr,
					ServerName:   dhcp.ServerName,
					File:         dhcp.File,
					Options:      dhcp.Options,
					Context:      dhcp.Context,
					Fingerprint:  dhcp.Fingerprint,
				}
				mtrx = trx
			}
		},
		true,
	)

	if messageToFingerprint != nil { // search vendor class
		var vendor string
		for _, o := range messageToFingerprint.Options {
			if utils.IsASCII([]byte(o.Data)) && len(o.Data) > 1 {
				if o.Type == 60 {
					vendor = string(o.Data)

					break
				}
			}
		}

		ip := addrMapping[messageToFingerprint.ClientHWAddr]
		// log.Println("found ip:", ip, "for mac", messageToFingerprint.ClientHWAddr)
		userAgents := userAgentStore[ip]
		// log.Println("found user agents:", userAgents)

		res, err := resolvers.LookupDHCPFingerprint(messageToFingerprint.Fingerprint, vendor, userAgents)
		if err != nil {
			log.Fatal(err)
		}

		// log.Println("got result", res.DeviceName, "for", messageToFingerprint.ClientHWAddr)

		val := strings.ReplaceAll(res.DeviceName, "/", "\n") + "\n" + ip
		ent := mtrx.AddEntity("netcap.dhcpResult", val)

		ent.AddProperty("timestamp", "Timestamp", "strict", messageToFingerprint.Timestamp)
		ent.AddProperty("clientIP", "ClientIP", "strict", messageToFingerprint.ClientIP)
		ent.AddProperty("serverIP", "ServerIP", "strict", messageToFingerprint.NextServerIP)
	}

	mtrx.AddUIMessage("completed!", maltego.UIMessageInform)
	fmt.Println(mtrx.ReturnOutput())
}
