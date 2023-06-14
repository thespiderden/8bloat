module bloat

require (
	github.com/gorilla/mux v1.7.3
	github.com/mileusna/useragent v1.1.0
	github.com/tomnomnom/linkheader v0.0.0-20180905144013-02ca5825eb80
	spiderden.org/masta v0.0.0-00010101000000-000000000000
)

replace spiderden.org/masta => ../go-mastadon

go 1.13
