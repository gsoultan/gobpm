package https

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gsoultan/gobpm/server/domains/observers/impl"
	"github.com/gsoultan/gobpm/server/domains/services"
	"github.com/gsoultan/gobpm/server/endpoints"
	"github.com/gsoultan/gobpm/server/interceptors"
	"github.com/gsoultan/gobpm/server/transports/connects"
	"github.com/gsoultan/gobpm/server/transports/https/collaboration"
	"github.com/gsoultan/gobpm/server/transports/https/common"
	"github.com/gsoultan/gobpm/server/transports/https/connectors"
	"github.com/gsoultan/gobpm/server/transports/https/decisions"
	"github.com/gsoultan/gobpm/server/transports/https/definitions"
	"github.com/gsoultan/gobpm/server/transports/https/group"
	"github.com/gsoultan/gobpm/server/transports/https/incidents"
	"github.com/gsoultan/gobpm/server/transports/https/notification"
	"github.com/gsoultan/gobpm/server/transports/https/organizations"
	"github.com/gsoultan/gobpm/server/transports/https/processes"
	"github.com/gsoultan/gobpm/server/transports/https/projects"
	"github.com/gsoultan/gobpm/server/transports/https/setup"
	"github.com/gsoultan/gobpm/server/transports/https/tasks"
	"github.com/gsoultan/gobpm/server/transports/https/users"
	"github.com/gsoultan/gobpm/ui"

	httptransport "github.com/go-kit/kit/transport/http"
)

func NewHTTPHandler(svc services.ServiceFacade, eps endpoints.Endpoints, sseObserver *impl.SSEObserver) http.Handler {
	m := http.NewServeMux()

	// Auth Middleware to extract user from token and put it in context
	f := interceptors.NewInterceptorFactory(svc)
	authMiddleware := f.NewHTTPAuth(f.NewJWTStrategy())

	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(common.EncodeError),
	}

	// SSE Endpoint
	if sseObserver != nil {
		m.HandleFunc("GET /api/v1/events", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			w.Header().Set("Access-Control-Allow-Origin", "*")

			ch := sseObserver.AddClient()
			defer sseObserver.RemoveClient(ch)

			ctx := r.Context()
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-ch:
					fmt.Fprint(w, msg)
					if flusher, ok := w.(http.Flusher); ok {
						flusher.Flush()
					}
				}
			}
		})
	}

	// Connect RPC
	path, handler := connects.NewConnectHandler(eps)
	m.Handle("/api/v1"+path, http.StripPrefix("/api/v1", handler))

	// Register Handlers
	setup.RegisterHandlers(m, eps.Setup, options)
	organizations.RegisterHandlers(m, eps.Organization, options)
	projects.RegisterHandlers(m, eps.Project, options)
	definitions.RegisterHandlers(m, eps.Definition, options)
	processes.RegisterHandlers(m, eps.Process, options)
	tasks.RegisterHandlers(m, eps.Task, options)
	users.RegisterHandlers(m, eps.User, options)
	group.RegisterHandlers(m, eps.Group, options)
	incidents.RegisterHandlers(m, eps.Incident, options)
	notification.RegisterHandlers(m, eps.Notification, options)
	decisions.RegisterHandlers(m, eps.Decision, options)
	connectors.RegisterHandlers(m, eps.Connector, options)
	collaboration.RegisterHandlers(m, eps.Collaboration, options)

	// Serve UI
	distFS := ui.Dist()
	fileServer := http.FileServerFS(distFS)

	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// If it's the root, serve index.html
		if r.URL.Path == "/" {
			fileServer.ServeHTTP(w, r)
			return
		}

		// Try to see if the file exists in the embedded FS
		_, err := fs.Stat(distFS, r.URL.Path[1:])
		if err != nil {
			// Fallback to index.html for SPA routing if it's not an asset request
			if !strings.HasPrefix(r.URL.Path, "/assets/") {
				r.URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(w, r)
	})

	// CORS middleware
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		authMiddleware.Wrap(m).ServeHTTP(w, r)
	})
}
