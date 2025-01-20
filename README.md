# pulse

`Multimedia Information Retrieval and Computer Vision` search engine project.

## Documentation

All information about the documentation are inside [docs.pdf](docs/docs.pdf)

## Install

```sh
go install github.com/just-hms/pulse/pulse@latest
```

## Usage

After installation write `pulse -h` to see what commands are available.

```sh
➜  pulse -h
this is the pulse search engine

Usage:
  pulse [flags]
  pulse [command]

Available Commands:
  help        Help about any command
  search      do one or more queries
  spimi       generate the indexes

Flags:
  -h, --help      help for pulse
  -p, --profile   profile the execution

Use "pulse [command] --help" for more information about a command.
```

### spimi

The spimi command index a given dataset, currently only ms-marco is supported, to use another dataset create another `spimi.ChunkReader` implementation.

```sh
➜ pulse spimi -h
generate the indexes

Usage:
  pulse spimi [flags]

Examples:
  cat dataset | pulse spimi
  pulse spimi ./path/to/dataset
  tar xOf dataset.tar.gz | pulse spimi

Flags:
  -c, --chunk uint        reader chunk size (default 50000)
  -h, --help              help for spimi
  -m, --max-memory uint   max memory used during indexing [MB] (default 3072)
      --no-compression    compress the posting list and term frequencies
      --no-stemming       remove stemming
      --no-stopwords      remove stopwords removal
  -w, --workers uint      number of workers indexing the dataset (default 16)

Global Flags:
  -p, --profile   profile the execution
```

### search

After the indexing phase use `pulse search` to query the dataset

**query**

- interactive: using `pulse search -i` the program waits for the query to be input, then it returns the top `k` results and then waits for the next query
- one-shot: using `pulse search -q` the program executes only the query passed as flag and returns the top `k` results
- file: using `pulse search -f` the program read the file passed as flag and then execute the queries present in the file
the query file must be formatted like so: `<queryID>\t<query>`

**result**

- each line the result is formatted like so: `<queryID>\tQ0<docNO><ranking>RANDOMID`
- the last line contains the query execution time in mico-seconds formatted like so `#\t<queryID>\t<elapsed_time>\t<elapsed_time_µs>`

```sh
➜  pulse search -h
do one or more queries

Usage:
  pulse search [flags]

Examples:
  pulse search -q "who is the president right now" -m TFIDF
  pulse search -i

Flags:
  -c, --conjunctive     search in conjunctive mode
  -k, --doc2ret uint    number of documents to be returned (default 10)
  -f, --file string     add a queries file
  -h, --help            help for search
  -i, --interactive     interactive search
  -m, --metric string   score metric to be used [BM25|TFIDF] (default "BM25")
  -q, --query string    single search

Global Flags:
  -p, --profile   profile the execution
```

> to visulize the benchmarking information use
> `pprof -http=localhost:8080 .pulse.prof`

## `trec_eval`

All information about `trec_eval` are inside this [README.md](trec_eval/README.md)
