#!/usr/bin/env bash
#
# Exit on error. Append "|| true" if you expect an error.
set -o errexit
# Exit on error inside any functions or subshells.
set -o errtrace
# Do not allow use of undefined vars. Use ${VAR:-} to use an undefined VAR
set -o nounset
# Catch the error in case mysqldump fails (but gzip succeeds) in `mysqldump |gzip`
set -o pipefail

__nginx_dir="/etc/nginx"
__nginx_mount="/etc/nginx"
__conf_volume="deploybot_docker_conf" # named volume that contains docker configuration
__conf_mount="/root/.docker" # location at which conf volume is mounted
__compose_command="custom-docker-compose"
__push_arg="-v ${__nginx_dir}:${__nginx_mount} -v ${__conf_volume}:${__conf_mount}"

__subdomain="${1}"
__repo_url="${2}"
__machine_name="${3}"

__repo_name=$(basename ${__repo_url})
__compose_dir="${__nginx_dir}"/composes
__compose_file="docker-compose.yml"

## @brief stops the service running on given subdomain
stopService() {
  local repo_name=$1
  local machine=$2

  pushd "${__compose_dir}"/"${repo_name}"

  eval "$(docker-machine env ${machine} --shell bash)"
  VOLUMES="${__push_arg}" ${__compose_command} -f ${__compose_file} down
  eval "$(docker-machine env --shell bash -u)"

  popd
}

## @brief remove nginx entry of service
## @param $1 subdomain
removeNginxEntry() {
  local subdomain=$1

  pushd "${__nginx_dir}"

  pushd sites-available/
  rm ${subdomain}
  popd

  pushd sites-enabled/
  rm ${subdomain}
  popd

  popd
}

cleanup() {
  local repo_name=$1

  pushd ${__compose_dir}
  rm -r ${repo_name}
  popd
}

stopService "${__repo_name}" "${__machine_name}"
removeNginxEntry "${__subdomain}"
cleanup "${__repo_name}"
