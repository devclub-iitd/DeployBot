#!/usr/bin/env bash

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


## Install docker machine things
pushd /config/
for config in ls *.zip; do
    docker-machine-import.sh $PWD/${config}
done
popd

## Set git username
git config --global user.email devclub.iitd@gmail.com
git config --global user.name "DevClub IITD"

## Set git ssh key
eval "$(ssh-agent -s)"
ssh-add /root/.ssh/id_rsa
ssh-keyscan github.com >> /root/.ssh/known_hosts

## Install gnupg
gpg --pinentry-mode=loopback --passphrase $GPGSECRETPASS --import ./gpg_private.asc
