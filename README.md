# pulse

<!--toc:start-->
- [pulse](#pulse)
  - [download MSMARCO](#download-msmarco)
  - [profiling](#profiling)
<!--toc:end-->

a blazingly fast search engine written in golang

## download MSMARCO

```shell
mkdir -p data
curl -o data/dataset.tar.gz https://msmarco.blob.core.windows.net/msmarcoranking/collection.tar.gz
```

## profiling 

```shell
go install github.com/google/pprof@latest

go build path/to/file.go
./file --cpuprofile
pprof -http:localhost:8080 ./index ./profile.out
```

