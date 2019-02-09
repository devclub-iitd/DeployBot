#!/usr/bin/env bash
# This file:
#
#  - Demos BASH3 Boilerplate (change this for your script)
#
# Usage:
#
#  LOG_LEVEL=7 ./main.sh -f /tmp/x -d (change this for your script)
#
# Based on a template by BASH3 Boilerplate v2.3.0
# http://bash3boilerplate.sh/#authors
#
# The MIT License (MIT)
# Copyright (c) 2013 Kevin van Zonneveld and contributors
# You are not obligated to bundle the LICENSE file with your b3bp projects as long
# as you leave these references intact in the header comments of your source files.

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

if [[ "${BASH_SOURCE[0]}" != "${0}" ]]; then
  __i_am_main_script="0" # false

  if [[ "${__usage+x}" ]]; then
    if [[ "${BASH_SOURCE[1]}" = "${0}" ]]; then
      __i_am_main_script="1" # true
    fi

    __b3bp_external_usage="true"
    __b3bp_tmp_source_idx=1
  fi
else
  __i_am_main_script="1" # true
  [[ "${__usage+x}" ]] && unset -v __usage
  [[ "${__helptext+x}" ]] && unset -v __helptext
fi

# Set magic variables for current file, directory, os, etc.
__dir="$(cd "$(dirname "${BASH_SOURCE[${__b3bp_tmp_source_idx:-0}]}")" && pwd)"
__file="${__dir}/$(basename "${BASH_SOURCE[${__b3bp_tmp_source_idx:-0}]}")"
__base="$(basename "${__file}" .sh)"


# Define the environment variables (and their defaults) that this script depends on
LOG_LEVEL="${LOG_LEVEL:-6}" # 7 = debug -> 0 = emergency
NO_COLOR="${NO_COLOR:-}"    # true = disable color. otherwise autodetected


### Functions
##############################################################################

function __b3bp_log () {
  local log_level="${1}"
  shift

  # shellcheck disable=SC2034
  local color_debug="\x1b[35m"
  # shellcheck disable=SC2034
  local color_info="\x1b[32m"
  # shellcheck disable=SC2034
  local color_notice="\x1b[34m"
  # shellcheck disable=SC2034
  local color_warning="\x1b[33m"
  # shellcheck disable=SC2034
  local color_error="\x1b[31m"
  # shellcheck disable=SC2034
  local color_critical="\x1b[1;31m"
  # shellcheck disable=SC2034
  local color_alert="\x1b[1;33;41m"
  # shellcheck disable=SC2034
  local color_emergency="\x1b[1;4;5;33;41m"

  local colorvar="color_${log_level}"

  local color="${!colorvar:-${color_error}}"
  local color_reset="\x1b[0m"

  if [[ "${NO_COLOR:-}" = "true" ]] || ( [[ "${TERM:-}" != "xterm"* ]] && [[ "${TERM:-}" != "screen"* ]] ) || [[ ! -t 2 ]]; then
    if [[ "${NO_COLOR:-}" != "false" ]]; then
      # Don't use colors on pipes or non-recognized terminals
      color=""; color_reset=""
    fi
  fi

  # all remaining arguments are to be printed
  local log_line=""

  while IFS=$'\n' read -r log_line; do
    echo -e "$(date -u +"%Y-%m-%d %H:%M:%S UTC") ${color}$(printf "[%9s]" "${log_level}")${color_reset} ${log_line}" 1>&2
  done <<< "${@:-}"
}
function emergency () {                                __b3bp_log emergency "${@}"; exit 1; }
function alert ()     { [[ "${LOG_LEVEL:-0}" -ge 1 ]] && __b3bp_log alert "${@}"; true; }
function critical ()  { [[ "${LOG_LEVEL:-0}" -ge 2 ]] && __b3bp_log critical "${@}"; true; }
function error ()     { [[ "${LOG_LEVEL:-0}" -ge 3 ]] && __b3bp_log error "${@}"; true; }
function warning ()   { [[ "${LOG_LEVEL:-0}" -ge 4 ]] && __b3bp_log warning "${@}"; true; }
function notice ()    { [[ "${LOG_LEVEL:-0}" -ge 5 ]] && __b3bp_log notice "${@}"; true; }
function info ()      { [[ "${LOG_LEVEL:-0}" -ge 6 ]] && __b3bp_log info "${@}"; true; }
function debug ()     { [[ "${LOG_LEVEL:-0}" -ge 7 ]] && __b3bp_log debug "${@}"; true; }

function help () {
  echo "" 1>&2
  echo " ${*}" 1>&2
  echo "" 1>&2
  echo "  ${__usage:-No usage available}" 1>&2
  echo "" 1>&2

  if [[ "${__helptext:-}" ]]; then
    echo " ${__helptext}" 1>&2
    echo "" 1>&2
  fi

  exit 1
}


### Parse commandline options
##############################################################################

# Commandline options. This defines the usage page, and is used to parse cli
# opts & defaults from. The parsing is unforgiving so be precise in your syntax
# - A short option must be preset for every long option; but every short option
#   need not have a long option
# - `--` is respected as the separator between options and arguments
# - We do not bash-expand defaults, so setting '~/app' as a default will not resolve to ${HOME}.
#   you can use bash variables to work around this (so use ${HOME} instead)

# shellcheck disable=SC2015
[[ "${__usage+x}" ]] || read -r -d '' __usage <<-'EOF' || true # exits non-zero when EOF encountered
  -u --url   [arg]    Github url for the repository. Required.
  -m --machine  [arg]   Machine to deploy the project. Required.
  -b --branch  [arg] Name of the repository. Default="master"
  -s --subdomain  [arg] Subdomain of the deployed app. It will default to VIRTUAL_HOST given in docker configuration.
  -a --access  [arg] Access criteria for the service
  -v               Enable verbose mode, print script as it is executed
  -d --debug       Enables debug mode
  -h --help        This page
  -n --no-color    Disable color output
EOF

# shellcheck disable=SC2015
[[ "${__helptext+x}" ]] || read -r -d '' __helptext <<-'EOF' || true # exits non-zero when EOF encountered
 This is the deploy script by DevClub IITD. This assumes that you have docker-machine
 and docker-compose setup.
EOF

# Translate usage string -> getopts arguments, and set $arg_<flag> defaults
while read -r __b3bp_tmp_line; do
  if [[ "${__b3bp_tmp_line}" =~ ^- ]]; then
    # fetch single character version of option string
    __b3bp_tmp_opt="${__b3bp_tmp_line%% *}"
    __b3bp_tmp_opt="${__b3bp_tmp_opt:1}"

    # fetch long version if present
    __b3bp_tmp_long_opt=""

    if [[ "${__b3bp_tmp_line}" = *"--"* ]]; then
      __b3bp_tmp_long_opt="${__b3bp_tmp_line#*--}"
      __b3bp_tmp_long_opt="${__b3bp_tmp_long_opt%% *}"
    fi

    # map opt long name to+from opt short name
    printf -v "__b3bp_tmp_opt_long2short_${__b3bp_tmp_long_opt//-/_}" '%s' "${__b3bp_tmp_opt}"
    printf -v "__b3bp_tmp_opt_short2long_${__b3bp_tmp_opt}" '%s' "${__b3bp_tmp_long_opt//-/_}"

    # check if option takes an argument
    if [[ "${__b3bp_tmp_line}" =~ \[.*\] ]]; then
      __b3bp_tmp_opt="${__b3bp_tmp_opt}:" # add : if opt has arg
      __b3bp_tmp_init=""  # it has an arg. init with ""
      printf -v "__b3bp_tmp_has_arg_${__b3bp_tmp_opt:0:1}" '%s' "1"
    elif [[ "${__b3bp_tmp_line}" =~ \{.*\} ]]; then
      __b3bp_tmp_opt="${__b3bp_tmp_opt}:" # add : if opt has arg
      __b3bp_tmp_init=""  # it has an arg. init with ""
      # remember that this option requires an argument
      printf -v "__b3bp_tmp_has_arg_${__b3bp_tmp_opt:0:1}" '%s' "2"
    else
      __b3bp_tmp_init="0" # it's a flag. init with 0
      printf -v "__b3bp_tmp_has_arg_${__b3bp_tmp_opt:0:1}" '%s' "0"
    fi
    __b3bp_tmp_opts="${__b3bp_tmp_opts:-}${__b3bp_tmp_opt}"
  fi

  [[ "${__b3bp_tmp_opt:-}" ]] || continue

  if [[ "${__b3bp_tmp_line}" =~ (^|\.\ *)Default= ]]; then
    # ignore default value if option does not have an argument
    __b3bp_tmp_varname="__b3bp_tmp_has_arg_${__b3bp_tmp_opt:0:1}"

    if [[ "${!__b3bp_tmp_varname}" != "0" ]]; then
      __b3bp_tmp_init="${__b3bp_tmp_line##*Default=}"
      __b3bp_tmp_re='^"(.*)"$'
      if [[ "${__b3bp_tmp_init}" =~ ${__b3bp_tmp_re} ]]; then
        __b3bp_tmp_init="${BASH_REMATCH[1]}"
      else
        __b3bp_tmp_re="^'(.*)'$"
        if [[ "${__b3bp_tmp_init}" =~ ${__b3bp_tmp_re} ]]; then
          __b3bp_tmp_init="${BASH_REMATCH[1]}"
        fi
      fi
    fi
  fi

  if [[ "${__b3bp_tmp_line}" =~ (^|\.\ *)Required\. ]]; then
    # remember that this option requires an argument
    printf -v "__b3bp_tmp_has_arg_${__b3bp_tmp_opt:0:1}" '%s' "2"
  fi

  printf -v "arg_${__b3bp_tmp_opt:0:1}" '%s' "${__b3bp_tmp_init}"
done <<< "${__usage:-}"

# run getopts only if options were specified in __usage
if [[ "${__b3bp_tmp_opts:-}" ]]; then
  # Allow long options like --this
  __b3bp_tmp_opts="${__b3bp_tmp_opts}-:"

  # Reset in case getopts has been used previously in the shell.
  OPTIND=1

  # start parsing command line
  set +o nounset # unexpected arguments will cause unbound variables
                 # to be dereferenced
  # Overwrite $arg_<flag> defaults with the actual CLI options
  while getopts "${__b3bp_tmp_opts}" __b3bp_tmp_opt; do
    [[ "${__b3bp_tmp_opt}" = "?" ]] && help "Invalid use of script: ${*} "

    if [[ "${__b3bp_tmp_opt}" = "-" ]]; then
      # OPTARG is long-option-name or long-option=value
      if [[ "${OPTARG}" =~ .*=.* ]]; then
        # --key=value format
        __b3bp_tmp_long_opt=${OPTARG/=*/}
        # Set opt to the short option corresponding to the long option
        __b3bp_tmp_varname="__b3bp_tmp_opt_long2short_${__b3bp_tmp_long_opt//-/_}"
        printf -v "__b3bp_tmp_opt" '%s' "${!__b3bp_tmp_varname}"
        OPTARG=${OPTARG#*=}
      else
        # --key value format
        # Map long name to short version of option
        __b3bp_tmp_varname="__b3bp_tmp_opt_long2short_${OPTARG//-/_}"
        printf -v "__b3bp_tmp_opt" '%s' "${!__b3bp_tmp_varname}"
        # Only assign OPTARG if option takes an argument
        __b3bp_tmp_varname="__b3bp_tmp_has_arg_${__b3bp_tmp_opt}"
        printf -v "OPTARG" '%s' "${@:OPTIND:${!__b3bp_tmp_varname}}"
        # shift over the argument if argument is expected
        ((OPTIND+=__b3bp_tmp_has_arg_${__b3bp_tmp_opt}))
      fi
      # we have set opt/OPTARG to the short value and the argument as OPTARG if it exists
    fi
    __b3bp_tmp_varname="arg_${__b3bp_tmp_opt:0:1}"
    __b3bp_tmp_default="${!__b3bp_tmp_varname}"

    __b3bp_tmp_value="${OPTARG}"
    if [[ -z "${OPTARG}" ]] && [[ "${__b3bp_tmp_default}" = "0" ]]; then
      __b3bp_tmp_value="1"
    fi

    printf -v "${__b3bp_tmp_varname}" '%s' "${__b3bp_tmp_value}"
    debug "cli arg ${__b3bp_tmp_varname} = (${__b3bp_tmp_default}) -> ${!__b3bp_tmp_varname}"
  done
  set -o nounset # no more unbound variable references expected

  shift $((OPTIND-1))

  if [[ "${1:-}" = "--" ]] ; then
    shift
  fi
fi


### Automatic validation of required option arguments
##############################################################################

for __b3bp_tmp_varname in ${!__b3bp_tmp_has_arg_*}; do
  # validate only options which required an argument
  [[ "${!__b3bp_tmp_varname}" = "2" ]] || continue

  __b3bp_tmp_opt_short="${__b3bp_tmp_varname##*_}"
  __b3bp_tmp_varname="arg_${__b3bp_tmp_opt_short}"
  [[ "${!__b3bp_tmp_varname}" ]] && continue

  __b3bp_tmp_varname="__b3bp_tmp_opt_short2long_${__b3bp_tmp_opt_short}"
  printf -v "__b3bp_tmp_opt_long" '%s' "${!__b3bp_tmp_varname}"
  [[ "${__b3bp_tmp_opt_long:-}" ]] && __b3bp_tmp_opt_long=" (--${__b3bp_tmp_opt_long//_/-})"

  help "Option -${__b3bp_tmp_opt_short}${__b3bp_tmp_opt_long:-} requires an argument"
done


### Cleanup Environment variables
##############################################################################

for __tmp_varname in ${!__b3bp_tmp_*}; do
  unset -v "${__tmp_varname}"
done

unset -v __tmp_varname


### Externally supplied __usage. Nothing else to do here
##############################################################################

if [[ "${__b3bp_external_usage:-}" = "true" ]]; then
  unset -v __b3bp_external_usage
  return
fi


### Signal trapping and backtracing
##############################################################################

function __b3bp_cleanup_before_exit () {
  info "Cleaning up. Done"
}
trap __b3bp_cleanup_before_exit EXIT

# requires `set -o errtrace`
__b3bp_err_report() {
    local error_code
    error_code=${?}
    error "Error in ${__file} in function ${1} on line ${2}"
    exit ${error_code}
}
# Uncomment the following line for always providing an error backtrace
# trap '__b3bp_err_report "${FUNCNAME:-.}" ${LINENO}' ERR


### Command-line argument switches (like -d for debugmode, -h for showing helppage)
##############################################################################

# debug mode
if [[ "${arg_d:?}" = "1" ]]; then
  set -o xtrace
  LOG_LEVEL="7"
  # Enable error backtracing
  trap '__b3bp_err_report "${FUNCNAME:-.}" ${LINENO}' ERR
fi

# verbose mode
if [[ "${arg_v:?}" = "1" ]]; then
  set -o verbose
fi

# no color mode
if [[ "${arg_n:?}" = "1" ]]; then
  NO_COLOR="true"
fi

# help mode
if [[ "${arg_h:?}" = "1" ]]; then
  # Help exists with code 1
  help "Help using ${0}"
fi


### Validation. Error out if the things required for your script are not present
##############################################################################

[[ "${arg_m:-}" ]]     || help      "Setting a machine name with -m or --machine is required"
[[ "${arg_u:-}" ]]     || help      "Setting a repo url with -u or --url is required"
[[ "${arg_a:-}" ]]     || help      "Setting access criteria with -a or --access is required"
[[ "${LOG_LEVEL:-}" ]] || emergency "Cannot continue without LOG_LEVEL. "


### Runtime
##############################################################################

__machine_name=${arg_m}
__repo_url=${arg_u}
__repo_name=$(basename ${arg_u} .git) #get name of git repository
__repo_branch=${arg_b}
__service_access=${arg_a}
__nginx_dir="/nginx"
__build_volume="deploybot_builder" # named volume that is shared between the current docker container and the
                        # future docker-compose container
__build_mount="/scratch/" # location at which build volume is mounted
__conf_volume="deploybot_docker_conf" # named volume that contains docker configuration
__conf_mount="/root/.docker" # location at which conf volume is mounted
__temp_dir="${__build_mount}/${RANDOM}/" # scratch is important here, since this ought to be an external volume
                                 # which will be shared betweent the docker-compose container and this
                                 # container
mkdir -p ${__temp_dir}

__compose_command="custom-docker-compose"
__build_arg="-v ${__build_volume}:${__build_mount}" # VOLUME args to be given at build time to docker-compose
__push_arg="${__build_arg} -v ${__conf_volume}:${__conf_mount}" # VOLUME args to be given at build time to docker-compose

chmod -R 755 "${__temp_dir}" # mktemp gives 700 permission by default
__repo_dir="${__temp_dir}/${__repo_name}"
mkdir -p "${__repo_dir}" # Create the repo directory
__compose_file="docker-compose.yml"
__env_file=".env"
__env_file_secret=".env.secret"
__default_network="reverseproxy"

## @brief Pulls repository in the given path
## @param $1 repo url to clone.
## @param $2 branch name of the git repository.
## @param $3 path on which to clone the repo.
pullRepository() {
  local repo_url=${1}
  local repo_branch=${2}
  local repo_path=${3}
  pushd "${repo_path}"
  git clone --depth=1 -b "${repo_branch}" "${repo_url}" .
  chmod -R 755 "${repo_path}"
  info "Repository successfully pulled - ${repo_url} on branch ${repo_branch}"
  popd
}

## @brief checks if the given machine name is registered in docker-machine
## @param $1 name of machine to check the existence of
checkServerName() {
  local server_name=${1}
  info "Checking server name - ${server_name}"
  docker-machine ls --format "{{.Name}}" | grep "${server_name}"
  local result=$?
  if [ ${result} -eq 0 ] ; then
    info "server name is correct"
  else
    error "Invalid server name"
    exit 1
  fi
}

# @brief exports VIRTUAL_HOST to set custom subdomain
# @param $1 name of subdomain
# @param $2 name of machine
populateVirtualHost() {
  local subdomain=${1}
  local machine=${2}
  info "Setting subdomain: ${subdomain}"
  local ip=$(docker-machine ip ${machine})
  local domain=${subdomain}.${ip}
  info "URL: ${domain}"
  export VIRTUAL_HOST=${domain}
}

## @brief Checks if the directory structure is valid or not
## @param $1 repository path
analyzeRepository() {
  local repo_path=${1}
  pushd "${repo_path}"
  if [ ! -f "$__compose_file" ]; then
    error "Docker Compose file does not exist."
    exit 1
  else
    ## TODO: Static analyize docker file
    info "Repository structure is correct: ${repo_path}"
    return 0
  fi
  popd
}

## @brief decrypt the env file in the cloned git repository
## @param $1 path to repo
decryptEnv() {
  local repo_path=${1}
  local env_path_secret="${1}/${__env_file_secret}"
  local env_path="${1}/${__env_file}"
  if [ -f "${env_path_secret}" ]; then
    pushd ${repo_path}
    git secret reveal -f -p ${GPGSECRETPASS:-"default"}
    info "Decryption of environment variables successful"
    popd
  else
    info "No environment file given. Making an empty one"
    touch "${env_path}"
  fi
}

## @brief create the image of the repo
## @param $1 path to repo
buildImage() {
  local repo_path=${1}
  pushd "${repo_path}"
  ## This assumes that we are inside ${__volume_mount}
  VOLUMES=${__build_arg} ${__compose_command} build
  popd
  info "building image successfull"
}

## @brief pull images of other services from hub.docker.com
## @param $1 path to repo
pullImages() {
  local repo_path=${1}
  pushd "${repo_path}"
  VOLUMES=${__build_arg} ${__compose_command} pull --ignore-pull-failures
  popd
  info "Required images pulled successfully"
}

## @brief push images to hub.docker.com and/or the local registry
## @param $1 path to repo
pushImages() {
  local repo_path=${1}
  pushd "${repo_path}"
  local images=`grep '^\s*image:\s*' docker-compose.yml | sed 's/image: '\''${REGISTRY_NAME}//' | sed s'/.$//' | sort | uniq`
  
  echo "${images}" | while read line; do
    repo=$(echo "$line" | cut -d":" -f1)
    org=$(echo "$repo" | cut -d"/" -f1)
    if [[ $line == *":"* ]]; then
      tag=":"$(echo "$line" | cut -d":" -f2)
    else
      tag=""
    fi

    if [ $org = "devclubiitd" ]; then
      docker login --username ${DOCKERHUB_USERNAME} --password ${DOCKER_PASSWORD}
      VOLUMES=${__push_arg} ${__compose_command} push
      docker logout
      info "Build image pushed"
    fi

    docker tag "$repo" "$LOCAL_REGISTRY""$repo""$tag"
    docker push "$LOCAL_REGISTRY""$repo""$tag"
  done
  info "All images pushed to respective registries"
}

## @brief deploy build image to respective machine
## @param $1 path to repo
deployImage() {
  local repo_path=$1
  pushd "${repo_path}"
  access=$(echo ${__machine_name} | cut -d"-" -f2)
  if [ $access = "internal" ]; then
    export REGISTRY_NAME=$LOCAL_REGISTRY
  else
    export REGISTRY_NAME=;
  fi
  eval "$(docker-machine env ${__machine_name} --shell bash)"
  VOLUMES=${__push_arg} COMPOSE_OPTIONS="-e REGISTRY_NAME" ${__compose_command} pull
  info "Images pulled for deployment"
#  docker network create -d bridge ${__default_network} || true # create a default network if not present
  VOLUMES=${__push_arg} COMPOSE_OPTIONS="-e VIRTUAL_HOST -e REGISTRY_NAME" ${__compose_command} up -d
  eval "$(docker-machine env --shell bash -u)"
  export REGISTRY_NAME=
  info "Deployment successful"
  popd
}

nginxEntry() {
  pushd ${__nginx_dir}

  if [ -f "$1" ]; then
    rm $1
  fi

  export subdomain="$1"
  export machine_name="$2"
  if [ "$3" = "internal" ]; then
    export allowed="10.0.0.0/24"
    export denied="deny all;"
  else
    export allowed="all"
    export denied=""
  fi

  envsubst < /usr/local/bin/nginx_template > ./${subdomain}

  export subdoman=
  export access=
  export machine_name=
  export denied=

  info "nginx entry successful"
  popd
}

## @brief clean up actions after the whole build process
## @param $1 repo path
cleanup() {
  local repo_path=$1
  rm -rf "${repo_path}"
  info "Cleanup successfull"
}

info "Machine: ${__machine_name}"
info "Repo: ${__repo_url}"
info "Branch: ${__repo_branch}"
info "Name: ${__repo_name}"
checkServerName "${__machine_name}"
if [[ "${arg_s:-}" ]]; then
  ## If we are giving custom subdomain
  populateVirtualHost ${arg_s} ${__machine_name}
fi
pullRepository "${__repo_url}" "${__repo_branch}" "${__repo_dir}"
analyzeRepository "${__repo_dir}"
decryptEnv "${__repo_dir}"
buildImage "${__repo_dir}"
pullImages "${__repo_dir}"
pushImages "${__repo_dir}"
deployImage "${__repo_dir}"
nginxEntry ${arg_s} ${__machine_name} ${__service_access}
cleanup "${__temp_dir}"
