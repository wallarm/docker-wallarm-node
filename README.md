# What is Wallarm Node

Docker container is another option to deploy Wallarm Node. Wallarm Node is Nginx-based and will operate as a reverse-proxy server which analyzes and passes through all the requests for protected application/API. 

The container contains all Wallarm Node subsystems.

# 30-second start

1. Sign up at https://my.wallarm.com. 

2. Make sure you have:

 * **/path/to/license.key** — path to the license key (check it in the registration email) on the host machine;
 * **example.com** — protected application or API;
 * **deploy@example.com** — your login for my.wallarm.com;
 * **very_secret** — password for my.wallarm.com.

3. Run a container with necessary parameters:

 ```
docker run -d -v /path/to/license.key:/etc/wallarm/license.key -e DEPLOY_USER="deploy@example.com" -e DEPLOY_PASSWORD="very_secret" -e NGINX_BACKEND=example.com -p 80:80 wallarm/node
```

As a result, the container should be running, and the protected website should be available on server port 80. New Node should be registered at Wallarm Cloud. 

For further configuration, place additional configuration files inside the container. 


# Connecting to the cloud

Every new Wallarm Node is required registering at Wallarm Cloud API. If you already tried out 30-second installation guide (abode), you're already familiar with one of the three following options: 

#### 1. Autoregistration

Set environment variables DEPLOY_USER, DEPLOY_PASSWORD with your credentials for my.wallarm.com. The container will automatically be registered in the cloud when you first start it.

By default, the container fails if one with that name exists already. To avoid it use the environment variable `DEPLOY_FORCE=true`.

```
docker run -d -v /etc/wallarm/license.key -e DEPLOY_USER="deploy@example.com" -e DEPLOY_PASSWORD="very_secret" -e NGINX_BACKEND=93.184.216.34 wallarm/node
```

#### 2. Using the known node credentials

To access the Wallarm Cloud each node uses its `uuid` and `secret` credentials. You can pass them into the environment variables `NODE_UUID` and `NODE_SECRET`.

```
docker run -d -v /etc/wallarm/license.key -e "NODE_UUID=00000000-0000-0000-0000-000000000000" -e NODE_SECRET="0000000000000000000000000000000000000000000000000000000000000000" -e NGINX_BACKEND=93.184.216.34 wallarm/node
```

#### 3. Configuration file

If you already have a `node.yaml` (configuration file), pass it into the Docker container as an external volume:

```
docker run -d -v /etc/wallarm/license.key -v /etc/wallarm/node.yaml -e NGINX_BACKEND=93.184.216.34 wallarm/node
```


# Nginx-wallarm configuration

Wallarm Node configuration is done via Nginx config file. To simplify this process in case of container, you can use environment variables  `NGINX_BACKEND` and `WALLARM_MODE`.

#### Simplified mode

* `NGINX_BACKEND` — backend address for all incoming requests.
If it doesn't have "http://" or "https://" prefix , then "http://" is used by default. Read more in proxy_pass.
* `WALLARM_MODE` — Nginx-wallarm mode. Read more in wallarm_mode.

#### Configuration files

Directories used by nginx:
* /etc/nginx-wallarm/conf.d — common settings
* /etc/nginx-wallarm/sites-enabled — virtual host settings
* /var/www/html — static files


# In-memory storage (Tarantool) setup

For behaviour-based attack detection Wallarm Node uses in-memory storage to save requests for a particular timeframe. Tarantool settings are set with the following environmental variables

* `SLAB_ALLOC_ARENA` — memory size (in gigabytes) allocated for storing serialized requests.
* `SLAB_ALLOC_MAXIMAL` — maximum size (in bytes) of the serialized request.


# Log rotation

Logs are written in the following directories:
* /var/log/nginx-wallarm/ — nginx logs
* /var/log/wallarm/ — various wallarm-specific subsystem logs 

By default, they are rotated once a day. Changing the rotation parameters by environment variables is not provided — use configuration files in /etc/logrotate.d/ instead.

# Monitoring settings

Nagios-compatible scripts for node monitoring are installed within the container. Details can be found in the documentation.

Scripts calling example:

```
docker exec -it wallarm-node /usr/lib/nagios-plugins/check_wallarm_tarantool_timeframe -w 1800 -c 900
docker exec -it wallarm-node /usr/lib/nagios-plugins/check_wallarm_export_delay -w 120 -c 300
```
