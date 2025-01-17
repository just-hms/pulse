# pulse

a blazingly fast search engine written in golang

## install

```sh
go install github.com/just-hms/pulse/pulse@latest
```

## download MSMARCO

```sh
mkdir -p data
curl -o data/dataset.tar.gz https://msmarco.blob.core.windows.net/msmarcoranking/collection.tar.gz
```

## indexing

```sh
tar xOf dataset.tar.gz | pulse spimi
```

## benchmark

```sh
pulse search "in the town where a I was born lived a man" -m TFIDF -p
pprof -http=localhost:8080 pulse.prof
```
