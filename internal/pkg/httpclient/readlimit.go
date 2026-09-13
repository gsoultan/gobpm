package httpclient

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/gsoultan/metis/internal/pkg/envvar"
)

// defaultMaxResponseBytes bounds a single outbound reply held in memory.
//
// The service-task and connector runners read a whole HTTP response into a
// []byte to map it into process variables. Without a bound, a partner (or an
// attacker who controls one, since the URL comes from a user-authored
// definition) can stream an unbounded body and exhaust the pod's memory — the
// engine's memory limit is the only thing that stops it, and it stops it by
// being OOM-killed mid-transaction. 8 MiB is generous for a JSON reply a
// process maps into variables, and small next to a 1 GiB container.
const defaultMaxResponseBytes int64 = 8 << 20

const envMaxResponseBytes = "METIS_HTTP_MAX_RESPONSE_BYTES"

// ErrResponseTooLarge is returned when an outbound reply exceeds the byte
// budget. It is a refusal, not a truncation: a process must not act on half a
// document it believes is whole.
var ErrResponseTooLarge = errors.New("httpclient: response body exceeds the configured maximum")

// MaxResponseBytes is the ceiling on a single outbound response body.
// Override with METIS_HTTP_MAX_RESPONSE_BYTES (bytes).
func MaxResponseBytes() int64 {
	if raw := envvar.Get(envMaxResponseBytes); raw != "" {
		if n, err := strconv.ParseInt(raw, 10, 64); err == nil && n > 0 {
			return n
		}
	}
	return defaultMaxResponseBytes
}

// ReadResponseBody reads up to the configured maximum and refuses anything
// larger, rather than truncating it. Reading max+1 and finding the extra byte
// is how "exactly at the limit" is told apart from "over it".
func ReadResponseBody(body io.Reader) ([]byte, error) {
	limit := MaxResponseBytes()
	raw, err := io.ReadAll(io.LimitReader(body, limit+1))
	if err != nil {
		return raw, err
	}
	if int64(len(raw)) > limit {
		return nil, fmt.Errorf("%w (%d bytes)", ErrResponseTooLarge, limit)
	}
	return raw, nil
}
