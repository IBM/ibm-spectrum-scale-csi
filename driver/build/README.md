There are two Dockerfiles:

1) ```Dockerfile``` legacy for a single platform image.
2) ```multi-arch.Dockerfile``` builds multi-architecture images (for x86_64, s390x, ppc64le), utilizing features in [buildkit](https://github.com/moby/buildkit).

Each Dockerfile has a comment explaining usage.  **Most will want to use the single platform Dockerfile.**

For the multi-architecture build you'll need to use buildkit, 
or some other build engine with target ```--platform``` support.  
See [docker buildx](https://docs.docker.com/buildx/working-with-buildx/) as an easy introduction to buildkit.