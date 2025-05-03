#!/bin/bash

generate() {
    rm -rf gen
    buf generate proto
}

generate_mocks() {
    (cd gen/go/services && \
        for f in `find . -name "*_api_grpc.pb.go" -type f`; do
            generate_module_mock $f
        done
    )
}

generate_module_mock() {
    local f=$1

    echo "Generating mocks from file $f"
    PATH_USING_NEW_FOLDER=$(echo "$f" | sed 's/\(.*\).pb/\1_mock/') # replaces .pb with _mock
    PATH_USING_NEW_FILENAME=`echo "$PATH_USING_NEW_FOLDER"`

    local path
    path=$(dirname ${PATH_USING_NEW_FILENAME#*/})

    local destination
    destination="../../mock/services/$path/"`basename $PATH_USING_NEW_FILENAME`
    mockgen -source "$f" -destination $destination &
}

generate
generate_mocks

exit 0
