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

    for machineName in ${machines[@]}
    do
        cleanImages $machineName
    done
}

## @brief docker system prune for given machine 
## @param $1 machineName
function cleanImages(){
    local machineName=$1;
    echo "[+] Starting cleanup for machine: $machineName"

    eval "$(docker-machine env ${machineName} --shell bash)"
    docker system prune -f
    
    eval "$(docker-machine env --shell bash -u)"
    echo "[+] Cleanup successful for machine: $machineName"

}

# prune the images in the MAIN VM 
echo "[+] Cleaning up on MAIN VM"
docker system prune -f
echo "[+] Clean up on MAIN VM successful"

# prune images in other VMs
cleanMachines