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

__repo_url=${1}
__repo_name=$(basename ${__repo_url} .git) #get name of git repository
__temp_dir=$(mktemp -d)
__repo_dir="${__temp_dir}/${__repo_name}"
mkdir -p "${__repo_dir}"
__access_email="devclub.iitd@gmail.com"

__hooks_repo_url="https://github.com/devclub-iitd/ServerConfig.git"
__hooks_repo_name=$(basename ${__repo_url} .git)
__hooks_temp_dir=$(mktemp -d)
__hooks_repo_dir="${__hooks_temp_dir}/${__hooks_repo_name}"
mkdir -p "${__hooks_repo_dir}"


## @brief Pulls repository in the given path
## @param $1 repo url to clone.
## @param $2 branch name of the git repository.
## @param $3 path on which to clone the repo.
pullRepository() {
  local repo_url=${1}
  local repo_path=${2}
  pushd "${repo_path}"
  if [ "$#" -gt 2 ] ; then
    local repo_branch=${3}
    git clone -b "${repo_branch}" "${repo_url}" .
  else
    git clone "${repo_url}" .
  fi
  chmod -R 755 "${repo_path}"
  echo "Repository successfully pulled - ${repo_url}"
  popd
}


initGitSecret(){
    local repo_path=${1}
    pushd "${repo_path}"
    git secret init
    git secret tell "${__access_email}"
    echo "Git Secret successfully initialized "
    popd
}

setupHooks(){
    local repo_path=${1}
    local hooks_repo_path=${2}

    pushd "${repo_path}"
    cp -r "${hooks_repo_path}""/.hooks" .
    echo "Git hooks successfully added"
    popd
}

pushRepo(){
    local repo_path=${1}
    local repo_url=${2#"git://"}
    pushd "${repo_path}"
    git add -A .
    git commit -m "[DeployBot] Initialized gitsecret and added git hooks"
    git pull
    git push https://"${GITHUB_USERNAME}":"${GITHUB_PASSWORD}"@"${repo_url}"
    echo "Git Repo successfully initialized and pushed"
    popd
}

cleanup() {
  local repo_path=$1
  rm -rf "${repo_path}"
  rm -rf "${__temp_dir}"
  rm -rf "${__hooks_temp_dir}"
  echo "Cleanup successfull"
}

pullRepository "${__repo_url}" "${__repo_dir}"
pullRepository "${__hooks_repo_url}" "${__hooks_repo_dir}" "hooks"
initGitSecret "${__repo_dir}"
setupHooks "${__repo_dir}" "${__hooks_repo_dir}"
pushRepo "${__repo_dir}" "${__repo_url}"
cleanup "${__repo_dir}"
cleanup "${__hooks_repo_dir}"
