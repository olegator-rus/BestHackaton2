package transform

import (
	"strconv"

	"github.com/dreadl0ck/netcap/maltego"
	"github.com/dreadl0ck/netcap/types"
)

func toApplicationsForCategory() {
	maltego.IPTransform(
		nil,
		func(lt maltego.LocalTransform, trx *maltego.Transform, profile *types.DeviceProfile, min, max uint64, profilesFile string, mac string, ipaddr string) {
			if profile.MacAddr == mac {

				for _, ip := range profile.Contacts {
					if ip.Addr == ipaddr {
						category := lt.Values["description"]
						for protoName, proto := range ip.Protocols {
							if proto.Category == category {
								ent := trx.AddEntity("maltego.Service", protoName)
								ent.SetLinkLabel(strconv.FormatInt(int64(proto.Packets), 10) + " pkts")
							}
						}

						break
					}
				}
				for _, ip := range profile.DeviceIPs {
					if ip.Addr == ipaddr {
						category := lt.Values["description"]
						for protoName, proto := range ip.Protocols {
							if proto.Category == category {
								ent := trx.AddEntity("maltego.Service", protoName)
								ent.SetLinkLabel(strconv.FormatInt(int64(proto.Packets), 10) + " pkts")
							}
						}

						break
					}
				}
			}
		},
	)
}
