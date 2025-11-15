package httpcontroller

import (
	"net/http"
	"os"
	"sync"
)

var (
	openAPISpec     []byte
	openAPISpecOnce sync.Once
)

// ServeOpenAPISpec возвращает OpenAPI спецификацию из корня проекта
func ServeOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	openAPISpecOnce.Do(func() {
		data, err := os.ReadFile("openapi.yml")
		if err == nil {
			openAPISpec = data
		}
	})

	if len(openAPISpec) == 0 {
		http.Error(w, "OpenAPI spec not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(openAPISpec)
}

// ServeSwaggerUI возвращает страницу Swagger UI
func ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>PR Reviewer Assignment Service - API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; padding: 0; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            window.ui = SwaggerUIBundle({
                url: "/openapi.yml",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
