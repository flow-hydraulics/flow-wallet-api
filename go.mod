module github.com/flow-hydraulics/flow-wallet-api

go 1.17

require (
	cloud.google.com/go/kms v1.4.0
	github.com/aws/aws-sdk-go-v2 v1.13.0
	github.com/aws/aws-sdk-go-v2/config v1.13.0
	github.com/aws/aws-sdk-go-v2/service/kms v1.14.0
	github.com/caarlos0/env/v6 v6.9.1
	github.com/felixge/httpsnoop v1.0.2
	github.com/go-gormigrate/gormigrate/v2 v2.0.0
	github.com/gomodule/redigo v1.8.8
	github.com/google/go-cmp v0.5.7
	github.com/google/uuid v1.3.0
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/jpillora/backoff v1.0.0
	github.com/lib/pq v1.10.4
	github.com/onflow/cadence v0.24.0
	github.com/onflow/flow-go-sdk v0.26.0
	github.com/sirupsen/logrus v1.8.1
	go.uber.org/goleak v1.1.12
	go.uber.org/ratelimit v0.2.0
	google.golang.org/genproto v0.0.0-20220222213610-43724f9ea8cf
	google.golang.org/grpc v1.44.0
	google.golang.org/protobuf v1.27.1
	gorm.io/datatypes v1.0.5
	gorm.io/driver/mysql v1.2.3
	gorm.io/driver/postgres v1.2.3
	gorm.io/driver/sqlite v1.2.6
	gorm.io/gorm v1.22.5
)

require (
	cloud.google.com/go v0.100.2 // indirect
	cloud.google.com/go/compute v1.3.0 // indirect
	cloud.google.com/go/iam v0.1.0 // indirect
	github.com/andres-erbsen/clock v0.0.0-20160526145045-9e14626cd129 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.10.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.2.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.7.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.9.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.14.0 // indirect
	github.com/aws/smithy-go v1.10.0 // indirect
	github.com/btcsuite/btcd v0.22.0-beta // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/ethereum/go-ethereum v1.10.12 // indirect
	github.com/fxamacker/cbor/v2 v2.4.1-0.20220515183430-ad2eae63303f // indirect
	github.com/fxamacker/circlehash v0.3.0 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/go-test/deep v1.0.8 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/googleapis/gax-go/v2 v2.1.1 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.10.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.2.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.9.1 // indirect
	github.com/jackc/pgx/v4 v4.14.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.4 // indirect
	github.com/joho/godotenv v1.4.0 // indirect
	github.com/klauspost/cpuid/v2 v2.0.12 // indirect
	github.com/logrusorgru/aurora v2.0.3+incompatible // indirect
	github.com/mattn/go-sqlite3 v1.14.10 // indirect
	github.com/onflow/atree v0.3.1-0.20220531231935-525fbc26f40a // indirect
	github.com/onflow/flow-go/crypto v0.24.3 // indirect
	github.com/onflow/flow/protobuf/go/flow v0.2.5 // indirect
	github.com/onflow/sdks v0.4.4 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.2.1-0.20211004051800-57c86be7915a // indirect
	github.com/stretchr/testify v1.7.1 // indirect
	github.com/turbolent/prettier v0.0.0-20210613180524-3a3f5a5b49ba // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/zeebo/blake3 v0.2.3 // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce // indirect
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8 // indirect
	golang.org/x/sys v0.0.0-20220209214540-3681064d5158 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/api v0.70.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	gorm.io/driver/sqlserver v1.2.1 // indirect
)
