#!/bin/bash

run_indexing() {
    local name=$1
    local args=$2
    local result="data/results-$name.tsv"
    
    # Measure time and run spimi
    local indexing_time=$({ time tar xOf data/dataset.tar.gz | pulse spimi $args; } 2>&1 | grep real | awk '{print $2}')

    # Measure index size
    local doc_id_size=$(du -sch data/dump/*/posting.bin | grep total | cut -f1)"B"
    local frequencies_size=$(du -sch data/dump/*/freqs.bin | grep total | cut -f1)"B"
    local vocabulary_size=$(du -sch data/dump/*/term*.bin data/dump/term*.bin | grep total | cut -f1)"B"
    local document_index_size=$(du -sch data/dump/*/doc.bin | grep total | cut -f1)"B"

    # Return metrics
    echo "indexing|$name|$indexing_time|$doc_id_size|$frequencies_size|$vocabulary_size|$document_index_size"
}

run_search(){
    local name=$1
    local args=$2
    local result="data/results-$name.tsv"

    # Run search
    pulse search -f data/msmarco-test2020-queries.tsv -k 1000 $args > $result

    # Run trec_eval
    local trec_output=$(trec_eval -m all_trec data/2020qrels-pass.tsv $result)

    local avg_query_time=$(cat $result | grep -P "#" | awk '{print $4/1000}' | awk '{s+=$1} END {print s/NR "ms"}')
    local std_dev_query_time=$(cat $result | grep -P "#" | awk '{print $4/1000}' | awk -v avg="$avg_query_time" '{sumsq+=($1-avg)^2} END {print sqrt(sumsq/NR) "ms"}')

    local p_5=$(echo "$trec_output" | grep -P "^P_5\s" | awk '{print $3}')
    local p_10=$(echo "$trec_output" | grep -P "^P_10\s" | awk '{print $3}')
    local r_1000=$(echo "$trec_output" | grep -P "^recall_1000\s" | awk '{print $3}')
    local ndcg_cut_10=$(echo "$trec_output" | grep -P "^ndcg_cut_10\s" | awk '{print $3}')
    local map_cut_10=$(echo "$trec_output" | grep -P "^map_cut_10\s" | awk '{print $3}')

    echo "search|$name|$avg_query_time|$std_dev_query_time|$p_5|$p_10|$r_1000|$ndcg_cut_10|$map_cut_10"
}

echo "indexing|name|indexing_time|doc_id_size|frequencies_size|vocabulary_size|document_index_size"
echo "search|name|avg_query_time|std_dev_query_time|p_5|p_10|r_1000|ndcg_cut_10|map_cut_10"

run_indexing "no-stemming-stopwords-removal" "--no-stemming --no-stopwords"
run_search "no-stemming-stopwords-removal-BM25" ""

run_indexing "no-compression" "--no-compression"
run_indexing "normal" ""

run_search "conjunctive-TFIDF" "--conjunctive -m TFIDF"
run_search "conjunctive-BM25" "--conjunctive"
run_search "disjunctive-TFIDF" "-m TFIDF"
run_search "disjunctive-BM25" ""
