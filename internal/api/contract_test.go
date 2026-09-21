package api

import (
 "context"
 "testing"
 "github.com/getkin/kin-openapi/openapi3"
)

func TestContract(t *testing.T) {
 doc, err := openapi3.NewLoader().LoadFromFile("../../api/openapi.yaml")
 if err != nil { t.Fatal(err) }
 if err := doc.Validate(context.Background()); err != nil { t.Fatal(err) }
 if doc.OpenAPI != "3.1.0" { t.Fatalf("version: %s", doc.OpenAPI) }
 for _, path := range []string{"/healthz", "/readyz", "/v1/coverage", "/v1/sources"} {
  item := doc.Paths.Find(path)
  if item == nil || item.Get == nil { t.Fatalf("missing GET %s", path) }
  for _, code := range []string{"200", "429", "503"} {
   if item.Get.Responses.Value(code) == nil { t.Fatalf("missing %s %s", path, code) }
  }
 }
 for _, name := range []string{"HealthEnvelope", "CoverageEnvelope", "SourcesEnvelope", "ErrorEnvelope"} {
  schema := doc.Components.Schemas[name].Value
  if err := schema.VisitJSON(map[string]any{"data":nil,"error":nil,"meta":nil}); err == nil {
   t.Fatalf("%s accepts missing status", name)
  }
 }
}
