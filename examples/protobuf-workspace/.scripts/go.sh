#!/bin/bash

build() {
    local path=$1

    for d in "$path"/*; do
        if [ -d "$d" ]; then
            echo "Compiling generated module '`basename $d`'"
            (cd $d && go mod tidy && go build)
        fi
    done
}

update_deps() {
    local path=$1

    for d in "$path"/*; do
        if [ -d "$d" ]; then
            echo "Updating module dependencies '`basename $d`'"
            (cd $d && go get -u)
        fi
    done
}

while getopts btum opt; do
    case $opt in
        b)
            echo "Building generated modules"
            build "gen/go/services"
            ;;

        t)
            echo "Building generated test modules"
            build "gen/test/services"
            ;;

        u)
            update_deps "gen/go/services"
            ;;

        m)
            echo "Building generated mocks"
            build "gen/mock/services"
            ;;

        ?)
            echo $opt
            echo "Unsupported option"
            exit 1
            ;;
    esac
done

exit 0