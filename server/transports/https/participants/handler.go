package participants

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gsoultan/metis/server/endpoints/participant"
	"github.com/gsoultan/metis/server/transports/https/common"
	"github.com/rs/zerolog/log"
)

// maxUploadBytes bounds a directory upload.
//
// The body is caller-supplied and read into memory. Ten megabytes is far past
// any real directory — the import itself refuses past ten thousand rows — and
// small enough that an upload cannot exhaust the server by being large.
const maxUploadBytes = 10 << 20

func RegisterHandlers(m *http.ServeMux, eps participant.Endpoints, options []httptransport.ServerOption) {
	m.Handle("GET /api/v1/participants", httptransport.NewServer(
		eps.ListParticipants,
		decodeListParticipantsRequest,
		common.EncodeResponse,
		options...,
	))
	m.Handle("POST /api/v1/participants/import", httptransport.NewServer(
		eps.ImportParticipants,
		decodeImportParticipantsRequest,
		common.EncodeResponse,
		options...,
	))
}

func decodeListParticipantsRequest(_ context.Context, r *http.Request) (any, error) {
	return participant.ListParticipantsRequest{ProjectID: r.URL.Query().Get("project_id")}, nil
}

// decodeImportParticipantsRequest reads either an uploaded file or a described
// remote source.
//
// The content type decides, because a file has to arrive as multipart and the
// other two are naturally JSON. One route rather than two: the caller is
// answering one question, and splitting it would mean two ways to say "import"
// that return the same thing.
func decodeImportParticipantsRequest(_ context.Context, r *http.Request) (any, error) {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return decodeUpload(r)
	}

	var req participant.ImportParticipantsRequest
	if err := json.NewDecoder(http.MaxBytesReader(nil, r.Body, maxUploadBytes)).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func decodeUpload(r *http.Request) (any, error) {
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		return nil, fmt.Errorf("could not read the upload: %w", err)
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("no file was uploaded: %w", err)
	}
	defer func() {
		// Logged rather than discarded: ParseMultipartForm spills a large
		// upload to a temporary file, and one that will not close is a file
		// handle this process keeps for every import somebody runs.
		if err := file.Close(); err != nil {
			log.Warn().Err(err).Msg("Could not close an uploaded participant directory")
		}
	}()

	// Bounded again at the read: ParseMultipartForm spills past its limit to
	// disk rather than refusing, so the limit above is not by itself a bound on
	// what this reads into memory.
	body, err := common.ReadLimited(file, maxUploadBytes)
	if err != nil {
		return nil, err
	}
	return participant.ImportParticipantsRequest{
		ProjectID: r.FormValue("project_id"),
		Kind:      "csv",
		CSV:       body,
	}, nil
}
