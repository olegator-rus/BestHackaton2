package transform

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/dsoprea/go-exif/v2"

	"github.com/dreadl0ck/netcap/maltego"
)

func toExifDataForImage() {
	var (
		lt   = maltego.ParseLocalArguments(os.Args)
		trx  = &maltego.Transform{}
		path = lt.Values["path"]
		err  error
	)

	if path == "" {
		path, err = url.QueryUnescape(lt.Values["fullImage"])
		if err != nil {
			log.Fatal(err)
		}
	}

	path = strings.TrimPrefix(path, "file://")

	log.Println("image path:", path)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	rawExif, err := exif.SearchAndExtractExif(data)
	if err != nil {
		if errors.Is(err, exif.ErrNoExif) {
			log.Println("No EXIF data found")
			trx.AddUIMessage("completed!", maltego.UIMessageInform)
			fmt.Println(trx.ReturnOutput())
			os.Exit(0)
		}
		log.Fatal(err)
	}

	entries, err := exif.GetFlatExifData(rawExif)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range entries {
		log.Printf("IFD-PATH=[%s] ID=(0x%04x) NAME=[%s] COUNT=(%d) TYPE=[%s] VALUE=[%s]\n", entry.IfdPath, entry.TagId, entry.TagName, entry.UnitCount, entry.TagTypeName, entry.Formatted)
		trx.AddEntity("netcap.ExifEntry", entry.TagName+" ("+entry.TagTypeName+") = "+entry.Formatted)
	}

	trx.AddUIMessage("completed!", maltego.UIMessageInform)
	fmt.Println(trx.ReturnOutput())
}
