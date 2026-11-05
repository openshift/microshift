// Copyright 2025 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tlsutil

import "crypto/tls"

// curves maps a curve name to its ID for the elliptic curves (TLS key exchange
// groups) that Go's crypto/tls package selects from by default. Unlike cipher
// suites, crypto/tls does not expose a listing function (there is no
// tls.Curves() equivalent to tls.CipherSuites()), so the supported curves are
// enumerated here, keyed by CurveID.String().
var curves = map[string]tls.CurveID{
	tls.X25519MLKEM768.String(): tls.X25519MLKEM768,
	tls.X25519.String():         tls.X25519,
	tls.CurveP256.String():      tls.CurveP256,
	tls.CurveP384.String():      tls.CurveP384,
	tls.CurveP521.String():      tls.CurveP521,
}

// GetCurveID returns the corresponding curve ID for the given name, and a
// boolean indicating whether it is supported. Names match CurveID.String(),
// e.g. "X25519", "X25519MLKEM768", "CurveP256", "CurveP384", "CurveP521".
func GetCurveID(s string) (tls.CurveID, bool) {
	c, ok := curves[s]
	return c, ok
}
