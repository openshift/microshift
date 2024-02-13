# Documentation Preview Server

This directory contains the scripts and tools for running the
documentation preview server.

1. Set the `DOCS_BRANCH` environment variable. Typically in `main` this
   is set to the _previous_ release branch. Note that the branches in the
   docs repository use `enterprise` as a prefix instead of `release`.
1. Run `scripts/docs-preview/build.sh` to build the documentation.
1. Run `scripts/docs-preview/serve.sh` to set up the server. 
1. Open the http://0.0.0.0:8081/microshift_welcome/ URL.

> Notes:
> - To view up-to-date documentation, set up a cron job to run the
> `scripts/docs-preview/build.sh` script regularly.
> - Open port 8081 in the firewall if you are serving to anyone else.
> - Run `scripts/docs-preview/clean.sh` to delete all the documentation
> podman images, containers and artifacts.
