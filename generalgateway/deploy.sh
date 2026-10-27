#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
EXEFILE="generalgateway"
IMAGE_REPO="${IMAGE_REPO:-gozero-demo}"

select_config() {
    case "${1:-test}" in
        online) echo "${SCRIPT_DIR}/etc/config.yaml" ;;
        test) echo "${SCRIPT_DIR}/etc/config_test.yaml" ;;
        *) echo "Unknown environment: $1" >&2; exit 1 ;;
    esac
}

build_release() {
    local environment="${1:-test}"
    local config_file
    config_file="$(select_config "${environment}")"

    rm -rf "${SCRIPT_DIR}/docker/release"
    mkdir -p "${SCRIPT_DIR}/docker/release/conf"
    (
        cd "${PROJECT_DIR}"
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \\
            go build -trimpath -ldflags="-s -w" -o generalgateway/docker/release/generalgateway ./generalgateway
    )
    cp "${SCRIPT_DIR}/start.sh" "${SCRIPT_DIR}/docker/release/start.sh"
    cp "${config_file}" "${SCRIPT_DIR}/docker/release/conf/config.yaml"
    chmod +x "${SCRIPT_DIR}/docker/release/start.sh" "${SCRIPT_DIR}/docker/release/${EXEFILE}"
    echo "${EXEFILE} build success"
}

deploy_image() {
    local version="$1"
    local environment="${2:-test}"
    local image_name="${IMAGE_REPO}/${EXEFILE}"
    if [[ "${environment}" != "online" ]]; then
        image_name="${image_name}-${environment}"
    fi

    if [[ -n "${DOCKER_USERNAME:-}" ]]; then
        : "${DOCKER_PASSWORD:?DOCKER_PASSWORD is required when DOCKER_USERNAME is set}"
        printf '%s' "${DOCKER_PASSWORD}" | docker login "${DOCKER_REGISTRY:-}" \\
            --username "${DOCKER_USERNAME}" --password-stdin
    fi
    docker build -t "${image_name}:${version}" "${SCRIPT_DIR}/docker"
    docker push "${image_name}:${version}"
    echo "Image pushed: ${image_name}:${version}"
}

case "${1:-}" in
    -b) build_release "${2:-test}" ;;
    -d)
        [[ $# -ge 2 ]] || { echo "usage: $0 -d VERSION [test|online]" >&2; exit 1; }
        build_release "${3:-test}"
        deploy_image "$2" "${3:-test}"
        ;;
    *)
        echo "usage: $0 -b [test|online] | -d VERSION [test|online]" >&2
        exit 1
        ;;
esac
