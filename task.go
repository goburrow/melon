// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package gomelon

import (
	"net/http"
)

// Task is simply a HTTP Handler.
type Task interface {
	http.Handler
}
