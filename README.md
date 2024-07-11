# AiO docker

Our vision of how All-In-One could be adopted as Docker by our clients.

## Build

* download desired AiO, f.e. ``curl -o aio.sh https://meganode.wallarm.com/4.10/wallarm-4.10.9.x86_64-glibc.sh``
* build ``docker build --pull --build-arg WALLARM_NODE_MODE -t aio-docker .`` or ``make build``

By default, Node combines both modes (filtering and postanalytics).  
If you need to build Node for a particular mode (say filtering), then you need to specify the WALLARM_NODE_MODE variable:  
``export WALLARM_NODE_MODE=filtering && docker build --pull --build-arg WALLARM_NODE_MODE -t aio-docker .``  
or ``export WALLARM_NODE_MODE=filtering && make build``

## Run
 Set up WALLARM_API_TOKEN, WALLARM_API_HOST and WALLARM_IGNORE_REGISTER_ERROR environment variables and run  
 ``docker run -it --rm -e WALLARM_API_TOKEN -e WALLARM_API_HOST -e WALLARM_IGNORE_REGISTER_ERROR aio-docker``

## Environment variables you MUST be aware of that are used by init scripts and services inside Wallarm

* WALLARM_API_TOKEN - your registration token for the Node.
* WALLARM_API_HOST - Wallarm cloud hostname; possible values are api.wallarm.com or us1.api.wallarm.com.
* WALLARM_IGNORE_REGISTER_ERROR - When Docker starts, the init script attempts to register the node. If it fails, Docker will continue working, but the Node will run unprotected. If this is undesirable behavior, you have to run ``export WALLARM_IGNORE_REGISTER_ERROR=false`` before starting Docker.
