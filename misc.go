// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
package gows

import (
	"os"
)

const (
	maxBannerSize = 50 * 1024 // 50KB
)

// readBanner read contents of a banner found in the current directory.
// A banner is a .txt file which has the same name with the running application.
func readBanner() string {
	banner, err := readFileContents(os.Args[0]+".txt", maxBannerSize)
	if err != nil {
		return ""
	}
	return banner
}

// readFileContents read contents with a limit of maximum bytes
func readFileContents(file string, maxBytes int) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()
	buf := make([]byte, maxBytes)
	n, err := f.Read(buf)
	if err != nil {
		return "", err
	}
	return string(buf[0:n]), nil
}
