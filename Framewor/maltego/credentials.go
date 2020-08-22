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

package maltego

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/gogo/protobuf/proto"

	"github.com/dreadl0ck/netcap"
	"github.com/dreadl0ck/netcap/types"
)

// credentialsTransformationFunc is a transformation over Credentials profiles for a selected Credentials.
type credentialsTransformationFunc = func(lt LocalTransform, trx *Transform, profile *types.Credentials, min, max uint64, profilesFile string, mac string, ip string)

// countFunc is a function that counts something over DeviceProfiles.
type credentialsCountFunc = func(credentials *types.Credentials, mac string, min, max *uint64)

// CredentialsTransform applies a maltego transformation over Credentials profiles seen for a target Credentials.
func CredentialsTransform(count credentialsCountFunc, transform credentialsTransformationFunc) {
	lt := ParseLocalArguments(os.Args[1:])
	credentialsFile := lt.Values["path"]
	mac := lt.Values["mac"]
	ipaddr := lt.Values["ipaddr"]

	stdout := os.Stdout
	os.Stdout = os.Stderr
	netcap.PrintBuildInfo()

	f, err := os.Open(credentialsFile)
	if err != nil {
		log.Fatal(err)
	}

	// check if its an audit record file
	if !strings.HasSuffix(f.Name(), ".ncap.gz") && !strings.HasSuffix(f.Name(), ".ncap") {
		log.Fatal("input file must be an audit record file")
	}

	os.Stdout = stdout

	r, err := netcap.Open(credentialsFile, netcap.DefaultBufferSize)
	if err != nil {
		panic(err)
	}

	// read netcap header
	header, errFileHeader := r.ReadHeader()
	if errFileHeader != nil {
		log.Fatal(errFileHeader)
	}
	if header.Type != types.Type_NC_Credentials {
		panic("file does not contain Credentials records: " + header.Type.String())
	}

	var (
		credentials = new(types.Credentials)
		pm          proto.Message
		ok          bool
		trx         = Transform{}
	)
	pm = credentials

	if _, ok = pm.(types.AuditRecord); !ok {
		panic("type does not implement types.AuditRecord interface")
	}

	var (
		min uint64 = 10000000
		max uint64 = 0
	)

	if count != nil {
		for {
			err = r.Next(credentials)
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				break
			} else if err != nil {
				panic(err)
			}

			count(credentials, mac, &min, &max)
		}

		err = r.Close()
		if err != nil {
			log.Println("failed to close audit record file: ", err)
		}
	}

	r, err = netcap.Open(credentialsFile, netcap.DefaultBufferSize)
	if err != nil {
		panic(err)
	}

	// read netcap header - ignore err as it has been checked before
	_, _ = r.ReadHeader()

	for {
		err = r.Next(credentials)
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			break
		} else if err != nil {
			panic(err)
		}

		transform(lt, &trx, credentials, min, max, credentialsFile, mac, ipaddr)
	}

	err = r.Close()
	if err != nil {
		log.Println("failed to close audit record file: ", err)
	}

	trx.AddUIMessage("completed!", UIMessageInform)
	fmt.Println(trx.ReturnOutput())
}
