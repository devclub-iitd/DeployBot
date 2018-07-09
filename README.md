# DeployBot
Deployment Management Bot 

## Deploy

* Project Structure:
  - path => /home/devclub/projects/src/<repo-name>/{}
* Config:
    ```
    [
          - git repo -> docker-compose
          - branch
          - subdomain
          - port list
    ]
 
* Slack:
    - Add a project to config
    - Remove a project from config

## Architecture:

* Central webserver
    - listen.devclub.in
    - listen to github hooks
    - listen to slack hooks
    - contact appropriate server to deploy the respective service
* Deployment server
    - listen to deploy messages from central server
    - calls appropriate script to deploy
