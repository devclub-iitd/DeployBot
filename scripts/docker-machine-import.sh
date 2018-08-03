#! /bin/bash

# Exit on error. Append "|| true" if you expect an error.
set -o errexit
# Exit on error inside any functions or subshells.
set -o errtrace
# Do not allow use of undefined vars. Use ${VAR:-} to use an undefined VAR
set -o nounset
# Catch the error in case mysqldump fails (but gzip succeeds) in `mysqldump |gzip`
set -o pipefail
# Turn on traces, useful while debugging but commented out by default
# set -o xtrace


if [ -z "$1" ]; then
  echo "Usage: docker-machine-import.sh MACHINE_NAME.zip"
  echo ""
  echo "Imports an exported machine from a MACHINE_NAME.zip file"
  echo "Note: This script requires you to have the same \$MACHINE_STORAGE_PATH/certs available on all host systems"
  exit 0
fi

machine_archive="$1"
machine_name="${machine_archive/.zip/}"
MACHINE_STORAGE_PATH="${MACHINE_STORAGE_PATH:-"$HOME/.docker/machine"}"
machine_path="$MACHINE_STORAGE_PATH/machines/$machine_name"

if [ -d "$machine_path" ]; then
  echo "$machine_name already exists"
  exit 1
fi

rm -rf "$machine_name"
mkdir -p "$machine_name"
unzip "$machine_archive" -d "$machine_name" > /dev/null
perl -pi -e "s|__MACHINE__STORAGE_PATH__|$MACHINE_STORAGE_PATH|g" $machine_name/config.json
mv "$machine_name" "$MACHINE_STORAGE_PATH/machines"

echo "Imported $machine_name to docker-machine ($machine_path)"
