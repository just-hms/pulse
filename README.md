# pulse

a blazingly fast search engine written in golang

## install

```shell
go install github.com/just-hms/pulse/pulse@latest
```

## download MSMARCO

```shell
mkdir -p data
curl -o data/dataset.tar.gz https://msmarco.blob.core.windows.net/msmarcoranking/collection.tar.gz
```

## indexing

```shell
tar xOf dataset.tar.gz | pulse spimi
```

## benchmark

```shell
pulse search "in the town where a I was born lived a man" -m TFIDF -p
pprof -http=localhost:8080 pulse.prof
```

## todo

- [] add conjunctive & disjunctive (add tests)
- [] add compression (maybe not)
- [] implement `nextGEQ`
