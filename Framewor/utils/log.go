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

package utils

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	logFilePermission = 0o755
)

var (
	// ReassemblyLog is the reassembly logger.
	ReassemblyLog = log.New(ioutil.Discard, "", log.LstdFlags|log.Lmicroseconds)

	// ReassemblyLogFileHandle is the file handle for the reassembly logger.
	ReassemblyLogFileHandle *os.File

	// DebugLog is the debug logger.
	DebugLog = log.New(ioutil.Discard, "", log.LstdFlags|log.Lmicroseconds)

	// DebugLogFileHandle is the file handle for the debug logger.
	DebugLogFileHandle *os.File
)

// InitLoggers initializes the loggers for the given output path.
func InitLoggers(outpath string) {
	var err error

	DebugLogFileHandle, err = os.OpenFile(filepath.Join(outpath, "debug.log"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, logFilePermission)
	if err != nil {
		log.Fatal(err)
	}

	DebugLog.SetOutput(DebugLogFileHandle)

	ReassemblyLogFileHandle, err = os.OpenFile(filepath.Join(outpath, "reassembly.log"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, logFilePermission)
	if err != nil {
		log.Fatal(err)
	}

	ReassemblyLog.SetOutput(ReassemblyLogFileHandle)
}

// CloseLogFiles closes the logfile handles.
func CloseLogFiles() []error {
	var errs []error

	if ReassemblyLogFileHandle != nil {
		err := ReassemblyLogFileHandle.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if DebugLogFileHandle != nil {
		err := DebugLogFileHandle.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
