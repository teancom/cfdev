#!/bin/bash -e

set -ex

while getopts "l:" arg; do
  case $arg in
    l) linuxkit=$OPTARG
      ;;
  esac
done

if [[ -z $linuxkit ]]; then
  if ! which linuxkit; then
    echo "ERROR: either provide the linuxkit binary with the '-l' flag or add linuxkit to your path"
    exit 1
  else
    linuxkit="linuxkit"
  fi
fi


script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
linuxkit_dir=script_dir/../linuxkit

"${script_dir}/../images/kernel/build.sh"

"$linuxkit" pkg build -hash dev "$linuxkit_dir"/pkg/bosh-lite-routing
"$linuxkit" pkg build -hash dev "$linuxkit_dir"/pkg/expose-multiple-ports
"$linuxkt" pkg build -hash dev "$linuxkit_dir"/pkg/garden-runc
"$linuxkit" pkg build -hash dev "$linuxkit_dir"/pkg/openssl

"$linuxkit" build \
 -disable-content-trust \
 -name cfdev \
 -format iso-efi \
 "$linuxkit_dir"/base.yml \
 "$linuxkit_dir"/garden.yml
