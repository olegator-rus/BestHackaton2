package transform

import (
	"strconv"

	"github.com/dustin/go-humanize"

	"github.com/dreadl0ck/netcap/maltego"
	"github.com/dreadl0ck/netcap/types"
)

func toDevices() {
	maltego.DeviceProfileTransform(maltego.CountPacketsDevices, func(lt maltego.LocalTransform, trx *maltego.Transform, profile *types.DeviceProfile, min, max uint64, profilesFile string, mac string) {
		ident := profile.MacAddr + "\n" + profile.DeviceManufacturer
		ent := trx.AddEntity("netcap.Device", ident)

		ent.AddProperty("path", "Path", "strict", profilesFile)
		ent.AddProperty("mac", "Mac Address", "strict", profile.MacAddr)

		ent.SetLinkLabel(strconv.FormatInt(profile.NumPackets, 10) + " pkts\n" + humanize.Bytes(profile.Bytes))
		ent.SetLinkThickness(maltego.GetThickness(uint64(profile.NumPackets), min, max))
	})
}
