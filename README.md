# pulse

a blazingly fast search engine written in golang

## install

```shell
go install github.com/just-hms/pulse@latest
```

## download MSMARCO

```shell
mkdir -p data
curl -o data/dataset.tar.gz https://msmarco.blob.core.windows.net/msmarcoranking/collection.tar.gz
```

## indexing

```shell
tar xOf data/dataset.tar.gz | pulse spimi
```

## benchmark

```shell
pulse search "in the town where a I was born lived a man" -m TFIDF -p
pprof -http=localhost:8080 pulse.prof
```

## install TRECEVAL

```shell
git clone https://github.com/usnistgov/trec_eval.git
cd trec_eval
make
sudo mv trec_eval /usr/local/bin/
```

## todo

- [] fix BM25
- [] launch trac eval
- [] add conjunctive & disjunctive
- [] add compression
- [] maybe use some embeddings for scoring function
- [] implement `nextGEQ`
