#!/bin/sh
# Chasky installer: installs the latest release for the current OS/arch.
# Usage:
#   gh release download --repo jcchavezs/chasky --pattern 'install.sh' -O - | sh
# Optional env vars: VERSION (release tag, default: latest), INSTALL_DIR.

set -eu

REPO="jcchavezs/chasky"
BINARY="chasky"

die() { echo "error: $*" >&2; exit 1; }

command -v gh >/dev/null 2>&1 || die "the GitHub CLI (gh) is required: https://cli.github.com"

# Names match the goreleaser name_template (uname-compatible).
case "$(uname -s)" in
    Linux*)               OS=Linux;   EXT=tar.gz;;
    Darwin*)              OS=Darwin;  EXT=tar.gz;;
    MINGW*|MSYS*|CYGWIN*) OS=Windows; EXT=zip;;
    *) die "unsupported operating system: $(uname -s)";;
esac

case "$(uname -m)" in
    x86_64|amd64)  ARCH=x86_64;;
    i386|i686)     ARCH=i386;;
    arm64|aarch64) ARCH=arm64;;
    *) die "unsupported architecture: $(uname -m)";;
esac

ARCHIVE="${BINARY}_${OS}_${ARCH}.${EXT}"

TMP_DIR=$(mktemp -d)
trap 'rm -rf "${TMP_DIR}"' EXIT

echo "Downloading ${ARCHIVE} (${VERSION:-latest})..."
# An empty tag makes gh pick the latest release.
gh release download ${VERSION:+"${VERSION}"} --repo "${REPO}" --pattern "${ARCHIVE}" --dir "${TMP_DIR}" \
    || die "failed to download ${ARCHIVE}"

echo "Extracting..."
if [ "${EXT}" = "zip" ]; then
    unzip -q "${TMP_DIR}/${ARCHIVE}" -d "${TMP_DIR}"
    BINARY="${BINARY}.exe"
else
    tar -xzf "${TMP_DIR}/${ARCHIVE}" -C "${TMP_DIR}"
fi
[ -f "${TMP_DIR}/${BINARY}" ] || die "${BINARY} not found in ${ARCHIVE}"

if [ -z "${INSTALL_DIR:-}" ]; then
    if [ -w /usr/local/bin ]; then
        INSTALL_DIR=/usr/local/bin
    else
        INSTALL_DIR="${HOME}/.local/bin"
    fi
fi
mkdir -p "${INSTALL_DIR}" || die "cannot create ${INSTALL_DIR}"

install -m 755 "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}" \
    || die "failed to install ${BINARY} to ${INSTALL_DIR}"

echo "Installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"

case ":${PATH}:" in
    *":${INSTALL_DIR}:"*) ;;
    *) echo "Note: ${INSTALL_DIR} is not in your PATH. Add it with:"
       echo "  export PATH=\"${INSTALL_DIR}:\$PATH\"";;
esac
