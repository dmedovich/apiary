package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>apiary — API docs</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({ url: '/openapi.yaml', dom_id: '#swagger-ui' });
  </script>
</body>
</html>
`

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", ":8080", "address to listen on")
	title := fs.String("title", "API", "API title")
	version := fs.String("version", "0.0.1", "API version")
	description := fs.String("description", "", "API description (info.description)")
	security := fs.String("security", "", "comma-separated global security schemes")
	server := fs.String("server", "", "comma-separated server URLs for info.servers")
	if err := fs.Parse(args); err != nil {
		log.Fatalf("apiary: %v", err)
	}

	set := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })
	cfg := loadConfig()

	patterns := fs.Args()
	if len(patterns) == 0 {
		patterns = cfg.Patterns
	}
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}

	opt := specOptions{
		title:       resolveStr(set["title"], *title, cfg.Title),
		version:     resolveStr(set["version"], *version, cfg.Version),
		description: resolveStr(set["description"], *description, cfg.Description),
		security:    resolveList(set["security"], splitCSV(*security), cfg.Security),
		servers:     resolveList(set["server"], splitCSV(*server), cfg.Servers),
	}

	specHandler := func(format, contentType string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			spec, n, err := buildSpec(patterns, opt)
			if err != nil {
				http.Error(w, "apiary: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if n == 0 {
				http.Error(w, "apiary: no operations found", http.StatusNotFound)
				return
			}
			data, err := encodeSpec(spec, format)
			if err != nil {
				http.Error(w, "apiary: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", contentType)
			_, _ = w.Write(data)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/openapi.yaml", specHandler("yaml", "application/yaml"))
	mux.HandleFunc("/openapi.json", specHandler("json", "application/json"))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(swaggerUIHTML))
	})

	fmt.Printf("apiary: serving Swagger UI at %s (spec at %s/openapi.yaml) — Ctrl-C to stop\n",
		displayURL(*addr), displayURL(*addr))
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("apiary: serve: %v", err)
	}
}

func displayURL(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "http://localhost" + addr
	}
	return "http://" + addr
}
