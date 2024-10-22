# pulse

a blazingly fast search engine written in golang

## download MSMARCO

```shell
mkdir -p data
curl -o data/dataset.tar.gz https://msmarco.blob.core.windows.net/msmarcoranking/collection.tar.gz
```

## install TRECEVAL

```shell
git clone https://github.com/usnistgov/trec_eval.git
cd trec_eval
make
sudo mv trec_eval /usr/local/bin/
```

## todo

- [] add conjunctive & disjunctive
- [] add compression
- [] maybe use some embeddings for scoring function
- [] implement `nextGEQ`

