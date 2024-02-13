# Documentation Preview Server

This directory contains the scripts and tools for running the
documentation preview server.

1. Check out `openshift/microshift` repository.
2. Set the `BRANCH` variable in
   `scripts/docs-preview/common.sh`. Typically in `main` this is set
   to the _previous_ release branch. Note that the branches in the
   docs repo use `enterprise` as a prefix instead of `release`.
3. Run `scripts/docs-preview/build.sh` to ensure the build works
   properly on your host.
4. Run `scripts/docs-preview/serve.sh` to set up the server. This
   requires port 8081 be open in the firewall, if you are serving to
   anyone else.
5. Set up a cron job to run `build.sh` regularly.
