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


## @brief list all virtual machines running dockers
function cleanMachines(){

    IFS=', ' read -r -a machines <<< $(docker-machine ls | awk '{print $1}' | tail -n +2 | tr '\n' ', ')

    for machine_name in ${machines[@]}
    do
        cleanImages $machine_name
    done
}

## @brief docker system prune for given machine 
## @param $1 machine_name
function cleanImages(){
    local machine_name=$1;
    echo "[+] Starting cleanup for machine: $machine_name"

    eval "$(docker-machine env ${machine_name} --shell bash)"
    docker system prune -f
    
    eval "$(docker-machine env --shell bash -u)"
    echo "[+] Cleanup successful for machine: $machine_name"

}

# prune the images in the MAIN VM 
echo "[+] Cleaning up on MAIN VM"
docker system prune -f
echo "[+] Clean up on MAIN VM successful"

# prune images in other VMs
cleanMachines