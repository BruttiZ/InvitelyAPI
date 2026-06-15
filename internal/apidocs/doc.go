package apidocs

import (
	_ "embed"
	"encoding/json"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed swagger.yaml
var SwaggerYAML string

var (
	swaggerJSON     string
	swaggerJSONErr  error
	swaggerJSONOnce sync.Once
)

func SwaggerJSON() (string, error) {
	swaggerJSONOnce.Do(func() {
		var document any
		if err := yaml.Unmarshal([]byte(SwaggerYAML), &document); err != nil {
			swaggerJSONErr = err
			return
		}

		output, err := json.MarshalIndent(document, "", "  ")
		if err != nil {
			swaggerJSONErr = err
			return
		}

		swaggerJSON = string(output)
	})

	return swaggerJSON, swaggerJSONErr
}

const SwaggerHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Swagger da API Invitely</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    body { margin: 0; background: #fafafa; }
    .swagger-ui .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function () {
      SwaggerUIBundle({
        url: "/swagger/doc.json",
        dom_id: "#swagger-ui",
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>`
