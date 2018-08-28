# DeployBot
This is a slack bot that is used for managing deployments of our docerized projects at DevClub.

# Requirements
- You need to have docker and docker-compose installed
- The following directory structure (where id_rsa is the private key of ssh-key associated with github account)
```
.
├── docker-compose.yml
├── Dockerfile
├── ignore
│   └── keys
│       └── id_rsa          (ssh private key associated with github account)
│       └── gpg_private.asc (gpg private key associated with github account)
.
.
.
```
# Initial Setup
### Copy of `docker_conf` volume is available
Nothing needs to be done in this case.
### No copy of `docker_conf` volume is present
- Comment the `read_only: true` line for `docker_conf` volume in `docker-compose.yml`
- Run: `docker-compose up -d`
- ssh into the running docker container using: `docker exec -it DeployBot_deploy /bin/bash`
- Setup docker-machine for all the child server using commands like: `docker-machine -D create -d generic --generic-ip-address <vm_dns_name> --generic-ssh-key <ssh_key_path> --generic-ssh-user <user_name> --generic-ssh-port <ssh_port> --generic-engine-port <docker_daemon_port> <server_name>`
- Uncomment the commented line in first step
- Run: `docker-compose down`

These steps will create a volume named `docker_conf` which will contain all the necessary information for our bot to deploy projects on configured servers.
NOTE: You will also have to deploy `ServerConfig` on each server for projects deployed to be accessible. 

# Environment variables
Environment variables should be present in a `.env` file. These environment variables are required:
```
GPGSECRETPASS=<GPG passphrase>
```

# Running
If initial setup has been done once, then just running `docker-compose up -d` will start the deploybot service. The bot will listen to incoming connections on port `7777` by default.

# Sample project configuration
- Each project should have a `.env` file for environment variables which should be encrypted using `git-secret` with `devclubiitd` user as collaborator. `devclubiitd`'s public gpg key can be found [here](https://api.github.com/users/devclubiitd/gpg_keys)
- Each project should have `docker-compose.yml` to build their projects.
- Sample `docker-compose.yml` is given here with name `sample_docker-compose.yml`.

Note: Make sure that you have external `reverseproxy` network in your `docker-compose.yml`. Also, never use bind mounts in your compose file.