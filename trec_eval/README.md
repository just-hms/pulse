# trec_eval

resorces:

- https://microsoft.github.io/msmarco/TREC-Deep-Learning-2020
- https://github.com/usnistgov/trec_eval

Used for evaluating an ad hoc retrieval run, given the results file and a standard set of judged results.

## Installing

```sh
git clone https://github.com/usnistgov/trec_eval.git
cd trec_eval
make
# add the trec_eval to a PATH folder
sudo mv trec_eval /usr/local/bin/
```

## Download standard collections

Example: the `TREC DL 2019` queries and `TREC DL 2019 qrels`, or the `TREC DL 2020` queries and `TREC DL 2020 qrels`

## Evaluation

```sh
# install pulse following the main README.md instructions
pulse search -q queries.tsv > result.tsv
trec_eval qrels.tsv results.tsv
```
