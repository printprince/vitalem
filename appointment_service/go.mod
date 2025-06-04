module github.com/printprince/vitalem/appointment_service

go 1.23

toolchain go1.24.2

require (
	github.com/google/uuid v1.6.0
	github.com/labstack/echo/v4 v4.11.4
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/driver/postgres v1.5.2
	gorm.io/gorm v1.25.4
)

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.4.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	golang.org/x/time v0.5.0 // indirect
)

replace github.com/printprince/vitalem/utils => ../utils

replace github.com/printprince/vitalem/logger_service => ../logger_service
