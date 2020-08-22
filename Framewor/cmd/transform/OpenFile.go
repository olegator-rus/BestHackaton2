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
	"os"
	"os/exec"
	"runtime"

	"github.com/dreadl0ck/netcap/env"
	"github.com/dreadl0ck/netcap/maltego"
)

func openFile() {
	var (
		lt              = maltego.ParseLocalArguments(os.Args)
		trx             = &maltego.Transform{}
		openCommandName = os.Getenv(env.MaltegoOpenFileCommand)
		args            []string
	)

	// if no command has been supplied via environment variable
	// then default to:
	// - open for macOS
	// - gio open for linux
	if openCommandName == "" {
		if runtime.GOOS == platformDarwin {
			openCommandName = defaultOpenCommand
		} else { // linux
			openCommandName = defaultOpenCommandLinux
			args = append(args, "open")
		}
	}

	loc := lt.Values["location"]
	dieIfExecutable(trx, loc)

	args = append(args, loc)
	log.Println("final command for opening files:", openCommandName, args)

	out, err := exec.Command(openCommandName, args...).CombinedOutput()
	if err != nil {
		log.Println(string(out))
		log.Fatal(err)
	}

	log.Println(string(out))

	trx.AddUIMessage("completed!", maltego.UIMessageInform)
	fmt.Println(trx.ReturnOutput())
}

func isExecAny(mode os.FileMode) bool {
	return mode&0o111 != 0
}
