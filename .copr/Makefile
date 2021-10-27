# This Makefile provides a wrapper script for copr [1] to consume
# [1] https://copr.fedorainfracloud.org/coprs/g/redhat-et/microshift-nightly/
#
# The process looks like:
# 1) github webhooks call copr on git pushes
# 2) copr downloads the git repository
# 3) copr executes make -f .copr/Makefile srpm outdir=.....
# 4) copr consumes the resulting source rpms to build the repositories
#
# additional details can be found here:
#   https://docs.pagure.org/copr.copr/user_documentation.html#make-srpm

outdir ?= /tmp

srpm:
	dnf -y install golang-bin git
	make srpm # from the project top level makefile
	cp packaging/rpm/_rpmbuild/SRPMS/*.src.rpm $(outdir)
