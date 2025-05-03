#!/bin/bash

install_dependencies() {
    install_plugins
    install_tools
}

install_plugins() {
    plugins=(
        github.com/mikros-dev/protoc-gen-mikros-extensions
        github.com/mikros-dev/protoc-gen-mikros-openapi
    )

    for p in "${plugins[@]}"; do
        go install $p
    done
}

install_tools() {
    go install go.uber.org/mock/mockgen@latest
    buf_install
}

buf_install() {
    if command -v buf > /dev/null 2>&1; then
        echo "buf CLI already installed"
        return
    fi

    echo "Installing buf CLI tool"

    local BIN="/usr/local/bin"
    local VERSION="1.49.0"

    curl -sSL "https://github.com/bufbuild/buf/releases/download/v${VERSION}/buf-$(uname -s)-$(uname -m)" -o "${BIN}/buf"
    chmod +x "${BIN}/buf"
}

install_dependencies

exit 0
