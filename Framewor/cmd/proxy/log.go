/*
 * NETCAP - Traffic Analysis Framework
 * Copyright (c) 2017 Philipp Mieden <dreadl0ck [at] protonmail [dot] ch>
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package main

import (
	"log"

	"go.uber.org/zap"
)

// LogFileName holds name of the logfile
const LogFileName = "net.proxy.log"

var (
	// Log instance
	Log   *zap.Logger
	debug bool
)

// ConfigureLogger configures the logging instance
func ConfigureLogger(debug bool, outputPath string) {

	var (
		zc  zap.Config
		err error
	)

	if debug {
		// use dev config
		zc = zap.NewDevelopmentConfig()
	} else {
		// use prod config
		zc = zap.NewProductionConfig()
	}

	// append outputPath
	zc.OutputPaths = append(zc.OutputPaths, outputPath)
	Log, err = zc.Build()
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %v", err)
	}
}
