package app

import (
	_ "embed"
	"net/http"

	scalargo "github.com/bdpiprava/scalar-go"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/httprule"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

//go:embed openapi.yaml
var specYAML []byte

// compilePattern parses a path template and returns a runtime.Pattern
func compilePattern(rule string) runtime.Pattern {
	tpl, err := httprule.Parse(rule)
	if err != nil {
		panic("invalid pattern: " + err.Error())
	}
	c := tpl.Compile()
	pat, err := runtime.NewPattern(1, c.OpCodes, c.Pool, c.Verb)
	if err != nil {
		panic("failed to make pattern: " + err.Error())
	}
	return pat
}

// RegisterScalarUI registers the Scalar UI <address>:<api-port>/scalar.
func (*App) RegisterScalarUI(apiSvr *api.Server) error {
	apiSvr.GRPCGatewayRouter.Handle(
		"GET",
		compilePattern("/openapi.yaml"),
		func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
			w.Header().Set("Content-Type", "plain/text; charset=utf-8")
			_, err := w.Write(specYAML)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		},
	)

	scalarHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content, err := scalargo.NewV2(
			scalargo.WithSpecURL("http://localhost:1317/openapi.yaml"),
			scalargo.WithBaseServerURL("/scalar"),
			scalargo.WithHideDownloadButton(),
			scalargo.WithTheme("kepler"),
			scalargo.WithOverrideCSS(`
				.section-flare {
					width: 100vw;
					background: radial-gradient(ellipse 80% 50% at 50% -20%, rgba(255, 123, 124, 0.4), transparent);
					height: 100vh;
				}

				.open-api-client {
					visibility: hidden;
					display: none;
				}

				span.endpoint-method {
					color: var(--scalar-color-green) !important;
				}

				.open-api-client-button {
					visibility: hidden;
					display: none;
				}

				div.flex.flex-col.gap-3.p-3.border-t.darklight-reference {
					visibility: hidden;
					display: none;
				}

				button.toggle-nested-icon {
					color: #FF7B7C !important;
					font-weight: 900 !important;
				}

				.active_page .toggle-nested-icon {
					color: #FF7B7C !important;
					font-weight: 900 !important;
				}

				p.sidebar-heading-link-title:hover {
					color: #FF7B7C !important;
				}

				.active_page.sidebar-heading {
					color: #FF7B7C !important;
					font-weight: 900 !important;
				}

				.section-header-label {
					color: #FF7B7C;
				}

				.scalar-card--contrast {
					padding: 6px !important;
				}

				span.sidebar-heading-type {
					--method-color: var(--scalar-color-green) !important;
				}

				.dark-mode {
					--scalar-color-blue: var(--scalar-color-green);
					--scalar-sidebar-search-border-color: 1px solid rgba(0, 0, 0, 0.1);
					--scalar-sidebar-search-color: #ADADAD;
					--scalar-sidebar-search-background: rgba(0, 0, 0, 0.1);
					--scalar-sidebar-font-weight-active: 900;
					--scalar-background-1: #1D1D1D;
					--scalar-background-2: #2C2C2C;
					--scalar-background-3: #202020;
					--scalar-color-1: #ADADAD;
					--scalar-color-2: #ADADAD;
					--scalar-color-3: #E8B0AE;
					--scalar-color-accent: #FF7B7C;
					--scalar-color-active: #FF7B7C;
					--scalar-sidebar-color-active: #FF7B7C;
					--scalar-sidebar-item-hover-color: #E8B0AE;
					--scalar-sidebar-item-hover-background: #3C2B29;
				}
			`),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, err = w.Write([]byte(content))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	apiSvr.GRPCGatewayRouter.Handle(
		"GET",
		compilePattern("/scalar"),
		func(
			w http.ResponseWriter,
			r *http.Request,
			pathParams map[string]string,
		) {
			http.StripPrefix("/scalar", scalarHandler).ServeHTTP(w, r)
		},
	)

	apiSvr.GRPCGatewayRouter.Handle(
		"GET",
		compilePattern("/scalar/{rest=**}"),
		func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
			http.StripPrefix("/scalar", scalarHandler).ServeHTTP(w, r)
		},
	)

	return nil
}
