This is an example of building a custom docker image. Please remember to adjust it accordingly:
- entrypoint.sh script with your commands
- nginx.conf and other nginx configs accordingly with your settings (or at least please set tarantool domain/ IP in `upstream wallarm_tarantool`)

To build an image based on custom Dockerfile, please run
```
docker build . --build-arg WALLARM_AIO_VERSION=4.10.9-rc1 -t wallarmcustom:latest
```

Example of usage
```
docker run --rm -ti -e WALLARM_API_TOKEN=... [-e WALLARM_API_HOST=... -e WALLARM_LABELS=...] wallarmcustom:latest
```

WALLARM_API_TOKEN has to be set, usage of the rest depends on your existing configuration
