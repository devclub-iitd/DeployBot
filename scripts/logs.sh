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

__repo_url="$1"
__machine_name="$2"
__tail_count="$3"
__repo_name=$(basename ${__repo_url} .git)
__compose_dir="${__nginx_dir}"/composes
__compose_file="docker-compose.yml"

## @brief checks the format of tail_count
sanitize_tail_count() {
  local tail="$1"
  if ! [[ "$tail" =~ ^[0-9]+$|^all$ ]]; then
    __tail_count='100'
  fi
}

## @brief outputs the logs of the given service
get_logs() {
  local repo_name=$1
  local machine=$2
  local tail=$3

  pushd "${__compose_dir}"/"${repo_name}"

  eval "$(docker-machine env ${machine} --shell bash)"
  VOLUMES="${__push_arg}" ${__compose_command} -f ${__compose_file} logs --tail ${tail}
  eval "$(docker-machine env --shell bash -u)"

  popd
}

sanitize_tail_count "${__tail_count}"
get_logs "${__repo_name}" "${__machine_name}" "${__tail_count}"