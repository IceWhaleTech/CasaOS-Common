//go:generate bash -c "mkdir -p codegen/mod_management && go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@v2.1.0 -generate types,client -package mod_management https://raw.githubusercontent.com/IceWhaleTech/IceWhale-OpenAPI/refs/heads/main/zimaos-mod-management/mod_management/openapi.yaml > codegen/mod_management/api.go"
package interfaces
