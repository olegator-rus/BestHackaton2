package transform

import (
	"github.com/dreadl0ck/netcap/maltego"
	"github.com/dreadl0ck/netcap/types"
)

func toParameterValues() {
	maltego.HTTPTransform(
		nil,
		func(lt maltego.LocalTransform, trx *maltego.Transform, http *types.HTTP, min, max uint64, profilesFile string, ipaddr string) {
			if http.SrcIP == ipaddr {
				param := lt.Values["properties.httpparameter"]
				for key, val := range http.Parameters {
					if key == param {
						trx.AddEntity("netcap.HTTPParameterValue", val)
					}
				}
			}
		},
		false,
	)
}
