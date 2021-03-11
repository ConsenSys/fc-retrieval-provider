module github.com/ConsenSys/fc-retrieval-provider

go 1.15

require (
	github.com/ConsenSys/fc-retrieval-common v0.0.0-20210308044151-362cc2c083b8
	github.com/ConsenSys/fc-retrieval-register v0.0.0-20210305042819-da4613bcbb05
	github.com/ant0ine/go-json-rest v3.3.2+incompatible
	github.com/joho/godotenv v1.3.0
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.7.1
)

replace github.com/ConsenSys/fc-retrieval-common => ../fc-retrieval-common
