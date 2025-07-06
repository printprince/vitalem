module github.com/printprince/vitalem/gateway_service

go 1.24

require (
	github.com/labstack/echo/v4 v4.13.4
	github.com/printprince/vitalem/utils v0.0.0-00010101000000-000000000000
	github.com/swaggo/echo-swagger v1.4.1
	github.com/swaggo/swag v1.16.4
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.1 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.1 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.26.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/swaggo/files/v2 v2.0.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	golang.org/x/tools v0.34.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/printprince/vitalem/utils => ../utils

replace github.com/printprince/vitalem/logger_service => ../logger_service
