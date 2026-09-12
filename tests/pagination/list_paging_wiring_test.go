package pagination_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	processendpoint "github.com/gsoultan/metis/server/endpoints/process"
	taskendpoint "github.com/gsoultan/metis/server/endpoints/task"
	processhttp "github.com/gsoultan/metis/server/transports/https/processes"
	taskhttp "github.com/gsoultan/metis/server/transports/https/tasks"
)

// Paging that the request struct declares, the endpoint honours, and the
// decoder never reads is paging that does not exist. ListInstancesRequest
// carried Page and PageSize from the start and ListInstancesPaged was wired to
// use them — but nothing lifted them off the query string, so every caller got
// page one at the server default and no way to ask for another.
//
// These go through the real mux and the real decoder, because the decoder is
// the part that was missing and a test against the endpoint would have passed
// throughout.

// captureInstances answers nothing and records what the decoder produced.
func captureInstances(into *processendpoint.ListInstancesRequest) processendpoint.Endpoints {
	return processendpoint.Endpoints{
		ListInstances: func(_ context.Context, request any) (any, error) {
			req, ok := request.(processendpoint.ListInstancesRequest)
			if !ok {
				return processendpoint.ListInstancesResponse{}, nil
			}
			*into = req
			return processendpoint.ListInstancesResponse{}, nil
		},
	}
}

func TestListInstancesReadsPagingFromTheQuery(t *testing.T) {
	var got processendpoint.ListInstancesRequest
	mux := http.NewServeMux()
	processhttp.RegisterHandlers(mux, captureInstances(&got), nil)

	projectID := uuid.Must(uuid.NewV7()).String()
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet,
		"/api/v1/instances?project_id="+projectID+"&page=3&page_size=25", nil)
	mux.ServeHTTP(httptest.NewRecorder(), request)

	if got.ProjectID != projectID {
		t.Errorf("project_id = %q", got.ProjectID)
	}
	if got.Page != 3 {
		t.Errorf("page = %d, want 3 — the query parameter is not reaching the endpoint", got.Page)
	}
	if got.PageSize != 25 {
		t.Errorf("page_size = %d, want 25", got.PageSize)
	}
}

// Absent paging still has to mean "no preference" rather than page zero, which
// is what a caller written before paging existed sends.
func TestListInstancesWithoutPagingAsksForNone(t *testing.T) {
	var got processendpoint.ListInstancesRequest
	mux := http.NewServeMux()
	processhttp.RegisterHandlers(mux, captureInstances(&got), nil)

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/instances?project_id=x", nil)
	mux.ServeHTTP(httptest.NewRecorder(), request)

	if got.Page != 0 || got.PageSize != 0 {
		t.Errorf("page/page_size = %d/%d, want 0/0 for a request that asked for neither", got.Page, got.PageSize)
	}
}

// A user's inbox is the listing most likely to outgrow one page, and its
// decoder had the same omission the instance listing did.
func TestListTasksByAssigneeReadsPagingFromTheQuery(t *testing.T) {
	var got taskendpoint.ListTasksByAssigneeRequest
	mux := http.NewServeMux()
	taskhttp.RegisterHandlers(mux, taskendpoint.Endpoints{
		ListTasksByAssignee: func(_ context.Context, request any) (any, error) {
			if req, ok := request.(taskendpoint.ListTasksByAssigneeRequest); ok {
				got = req
			}
			return taskendpoint.ListTasksResponse{}, nil
		},
	}, nil)

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet,
		"/api/v1/tasks/assignee/alice?page=2&page_size=20", nil)
	mux.ServeHTTP(httptest.NewRecorder(), request)

	if got.Assignee != "alice" {
		t.Errorf("assignee = %q", got.Assignee)
	}
	if got.Page != 2 || got.PageSize != 20 {
		t.Errorf("page/page_size = %d/%d, want 2/20 — the query is not reaching the endpoint", got.Page, got.PageSize)
	}
}

// The task listing gained an instance filter; the decoder has to carry it, or
// the endpoint's new branch is unreachable over HTTP.
func TestListTasksReadsInstanceIDFromTheQuery(t *testing.T) {
	var got taskendpoint.ListTasksRequest
	mux := http.NewServeMux()
	taskhttp.RegisterHandlers(mux, taskendpoint.Endpoints{
		ListTasks: func(_ context.Context, request any) (any, error) {
			if req, ok := request.(taskendpoint.ListTasksRequest); ok {
				got = req
			}
			return taskendpoint.ListTasksResponse{}, nil
		},
	}, nil)

	instanceID := uuid.Must(uuid.NewV7()).String()
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet,
		"/api/v1/tasks?instance_id="+instanceID+"&page_size=10", nil)
	mux.ServeHTTP(httptest.NewRecorder(), request)

	if got.InstanceID != instanceID {
		t.Errorf("instance_id = %q, want %q", got.InstanceID, instanceID)
	}
	if got.PageSize != 10 {
		t.Errorf("page_size = %d, want 10", got.PageSize)
	}
}
