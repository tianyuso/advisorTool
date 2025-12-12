module advisorTool

go 1.24.5

// 从主项目复制的 replace 指令
replace (
	github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos => github.com/bytebase/azure-sdk-for-go/sdk/data/azcosmos v0.0.0-20250109032656-87cf24d45689

	github.com/antlr4-go/antlr/v4 => github.com/bytebase/antlr/v4 v4.0.0-20240827034948-8c385f108920
	// Hive fix.
	github.com/beltran/gohive => github.com/bytebase/gohive v0.0.0-20240422092929-d76993a958a4
	github.com/beltran/gosasl => github.com/bytebase/gosasl v0.0.0-20240422091407-6b7481e86f08

	// copied from pingcap/tidb
	// fix potential security issue(CVE-2020-26160) introduced by indirect dependency.
	github.com/dgrijalva/jwt-go => github.com/form3tech-oss/jwt-go v3.2.6-0.20210809144907-32ab6a8243d7+incompatible
	// Other fixes.
	github.com/github/gh-ost => github.com/bytebase/gh-ost2 v1.1.7-0.20251002210738-35e5dddaad7c

	github.com/jackc/pgx/v5 => github.com/bytebase/pgx/v5 v5.0.0-20250212161523-96ff8aed8767

	github.com/mattn/go-oci8 => github.com/bytebase/go-obo v0.0.0-20231026081615-705a7fffbfd2

	github.com/microsoft/go-mssqldb => github.com/bytebase/go-mssqldb v0.0.0-20240801091126-3ff3ca07d898

	github.com/pingcap/tidb => github.com/bytebase/tidb v0.0.0-20251104040057-d29df9dd1b3b

	github.com/pingcap/tidb/pkg/parser => github.com/bytebase/tidb/pkg/parser v0.0.0-20251104040057-d29df9dd1b3b

	github.com/youmark/pkcs8 => github.com/bytebase/pkcs8 v0.0.0-20240612095628-fcd0a7484c94
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.10-20250912141014-52f32327d4b0.1
	connectrpc.com/connect v1.19.1
	github.com/antlr4-go/antlr/v4 v4.13.1
	github.com/bytebase/lsp-protocol v0.0.0-20250324071136-1586d0c10ff0
	github.com/bytebase/parser v0.0.0-20251201062756-17b16190b32d
	github.com/cenkalti/backoff/v5 v5.0.2
	github.com/cockroachdb/cockroachdb-parser v0.25.2
	github.com/go-sql-driver/mysql v1.7.1
	github.com/google/cel-go v0.26.1
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3
	github.com/hashicorp/golang-lru/v2 v2.0.7
	github.com/hjson/hjson-go/v4 v4.5.0
	github.com/jedib0t/go-pretty/v6 v6.7.7
	github.com/lib/pq v1.10.9
	github.com/microsoft/go-mssqldb v0.0.0-00010101000000-000000000000
	github.com/nyaruka/phonenumbers v1.6.6
	github.com/pingcap/tidb v0.0.0-00010101000000-000000000000
	github.com/pingcap/tidb/pkg/parser v0.0.0-20241125141335-ec8b81b98edc
	github.com/pkg/errors v0.9.1
	github.com/zeebo/xxh3 v1.0.2
	google.golang.org/genproto v0.0.0-20251103181224-f26f9409b101
	google.golang.org/genproto/googleapis/api v0.0.0-20251103181224-f26f9409b101
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251103181224-f26f9409b101
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.10
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cel.dev/expr v0.24.0 // indirect
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.2.0 // indirect
	github.com/bazelbuild/rules_go v0.49.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/biogo/store v0.0.0-20201120204734-aad293a2328f // indirect
	github.com/blevesearch/snowballstem v0.9.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cockroachdb/apd/v3 v3.2.1 // indirect
	github.com/cockroachdb/errors v1.11.3 // indirect
	github.com/cockroachdb/logtags v0.0.0-20241215232642-bb51bb14a506 // indirect
	github.com/cockroachdb/redact v1.1.5 // indirect
	github.com/cockroachdb/version v0.0.0-20250314144055-3860cd14adf2 // indirect
	github.com/dave/dst v0.27.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/elastic/gosigar v0.14.3 // indirect
	github.com/getsentry/sentry-go v0.27.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/jaegertracing/jaeger v1.18.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.11 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20230326075908-cb1d2100619a // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/opentracing/basictracer-go v1.0.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/openzipkin/zipkin-go v0.4.3 // indirect
	github.com/petermattis/goid v0.0.0-20250813065127-a731cc31b4fe // indirect
	github.com/pierrre/geohash v1.0.0 // indirect
	github.com/pingcap/errors v0.11.5-0.20250523034308-74f78ae071ee // indirect
	github.com/pingcap/failpoint v0.0.0-20240528011301-b51a646c7c86 // indirect
	github.com/pingcap/kvproto v0.0.0-20251023055424-e9d10f5dcd23 // indirect
	github.com/pingcap/log v1.1.1-0.20250917021125-19901e015dc9 // indirect
	github.com/pingcap/sysutil v1.0.1-0.20240311050922-ae81ee01f3a5 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20221212215047-62379fc7944b // indirect
	github.com/prometheus/client_golang v1.22.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.63.0 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/sasha-s/go-deadlock v0.3.5 // indirect
	github.com/shirou/gopsutil/v3 v3.24.5 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	github.com/stoewer/go-strcase v1.3.1 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	github.com/tikv/client-go/v2 v2.0.8-0.20251028065425-b7d4dfd8520e // indirect
	github.com/tikv/pd/client v0.0.0-20250703091733-dfd345b89500 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/twpayne/go-geom v1.4.1 // indirect
	github.com/twpayne/go-kml v1.5.2 // indirect
	github.com/uber/jaeger-client-go v2.22.1+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.36.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.36.0 // indirect
	go.opentelemetry.io/otel/exporters/zipkin v1.36.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/sdk v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	go.opentelemetry.io/proto/otlp v1.6.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/exp v0.0.0-20251023183803-a4bb9ffd2546 // indirect
	golang.org/x/mod v0.29.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	golang.org/x/tools v0.38.0 // indirect
	gonum.org/v1/gonum v0.16.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)
