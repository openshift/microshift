# Cockpit Starter Kit

Scaffolding for a [Cockpit](https://cockpit-project.org/) module.

# Development dependencies

On Debian/Ubuntu:

    $ sudo apt install gettext nodejs npm make

On Fedora:

    $ sudo dnf install gettext nodejs npm make


# Getting and building the source

These commands check out the source and build it into the `dist/` directory:

```
git clone https://github.com/cockpit-project/starter-kit.git
cd starter-kit
make
```

# Installing

`make install` compiles and installs the package in `/usr/local/share/cockpit/`. The
convenience targets `srpm` and `rpm` build the source and binary rpms,
respectively. Both of these make use of the `dist` target, which is used
to generate the distribution tarball. In `production` mode, source files are
automatically minified and compressed. Set `NODE_ENV=production` if you want to
duplicate this behavior.

For development, you usually want to run your module straight out of the git
tree. To do that, run `make devel-install`, which links your checkout to the
location were cockpit-bridge looks for packages. If you prefer to do
this manually:

```
mkdir -p ~/.local/share/cockpit
ln -s `pwd`/dist ~/.local/share/cockpit/starter-kit
```

After changing the code and running `make` again, reload the Cockpit page in
your browser.

You can also use
[watch mode](https://webpack.js.org/guides/development/#using-watch-mode) to
automatically update the bundle on every code change with

    $ npm run watch

or

    $ make watch

When developing against a virtual machine, watch mode can also automatically upload
the code changes by setting the `RSYNC` environment variable to
the remote hostname.

    $ RSYNC=c make watch

When developing against a remote host as a normal user, `RSYNC_DEVEL` can be
set to upload code changes to `~/.local/share/cockpit/` instead of
`/usr/local`.

    $ RSYNC_DEVEL=example.com make watch

To "uninstall" the locally installed version, run `make devel-uninstall`, or
remove manually the symlink:

    rm ~/.local/share/cockpit/starter-kit

# Running eslint

Cockpit Starter Kit uses [ESLint](https://eslint.org/) to automatically check
JavaScript code style in `.js` and `.jsx` files.

eslint is executed within every build.

For developer convenience, the ESLint can be started explicitly by:

    $ npm run eslint

Violations of some rules can be fixed automatically by:

    $ npm run eslint:fix

Rules configuration can be found in the `.eslintrc.json` file.

## Running stylelint

Cockpit uses [Stylelint](https://stylelint.io/) to automatically check CSS code
style in `.css` and `scss` files.

styleint is executed within every build.

For developer convenience, the Stylelint can be started explicitly by:

    $ npm run stylelint

Violations of some rules can be fixed automatically by:

    $ npm run stylelint:fix

Rules configuration can be found in the `.stylelintrc.json` file.

During fast iterative development, you can also choose to not run stylelint.
This speeds up the build and avoids build failures due to e. g. ill-formatted
css or other issues:

    $ make STYLELINT=0

# Running tests locally

Run `make check` to build an RPM, install it into a standard Cockpit test VM
(centos-8-stream by default), and run the test/check-application integration test on
it. This uses Cockpit's Chrome DevTools Protocol based browser tests, through a
Python API abstraction. Note that this API is not guaranteed to be stable, so
if you run into failures and don't want to adjust tests, consider checking out
Cockpit's test/common from a tag instead of main (see the `test/common`
target in `Makefile`).

After the test VM is prepared, you can manually run the test without rebuilding
the VM, possibly with extra options for tracing and halting on test failures
(for interactive debugging):

    TEST_OS=centos-8-stream test/check-application -tvs

It is possible to setup the test environment without running the tests:

    TEST_OS=centos-8-stream make prepare-check

You can also run the test against a different Cockpit image, for example:

    TEST_OS=fedora-34 make check

# Running tests in CI

These tests can be run in [Cirrus CI](https://cirrus-ci.org/), on their free
[Linux Containers](https://cirrus-ci.org/guide/linux/) environment which
explicitly supports `/dev/kvm`. Please see [Quick
Start](https://cirrus-ci.org/guide/quick-start/) how to set up Cirrus CI for
your project after forking from starter-kit.

The included [.cirrus.yml](./.cirrus.yml) runs the integration tests for two
operating systems (Fedora and CentOS 8). Note that if/once your project grows
bigger, or gets frequent changes, you may need to move to a paid account, or
different infrastructure with more capacity.

Tests also run in [Packit](https://packit.dev/) for all currently supported
Fedora releases; see the [packit.yaml](./packit.yaml) control file. You need to
[enable Packit-as-a-service](https://packit.dev/docs/packit-service/) in your GitHub project to use this.
To run the tests in the exact same way for upstream pull requests and for
[Fedora package update gating](https://docs.fedoraproject.org/en-US/ci/), the
tests are wrapped in the [FMF metadata format](https://github.com/teemtee/fmf)
for using with the [tmt test management tool](https://docs.fedoraproject.org/en-US/ci/tmt/).
Note that Packit tests can *not* run their own virtual machine images, thus
they only run [@nondestructive tests](https://github.com/cockpit-project/cockpit/blob/main/test/common/testlib.py).

# Customizing

After cloning the Starter Kit you should rename the files, package names, and
labels to your own project's name. Use these commands to find out what to
change:

    find -iname '*starter*'
    git grep -i starter

# Automated release

Once your cloned project is ready for a release, you should consider automating
that. The intention is that the only manual step for releasing a project is to create
a signed tag for the version number, which includes a summary of the noteworthy
changes:

```
123

- this new feature
- fix bug #123
```

Pushing the release tag triggers the [release.yml](.github/workflows/release.yml.disabled)
[GitHub action](https://github.com/features/actions) workflow. This creates the
official release tarball and publishes as upstream release to GitHub. The
workflow is disabled by default -- to use it, edit the file as per the comment
at the top, and rename it to just `*.yml`.

The Fedora and COPR releases are done with [Packit](https://packit.dev/),
see the [packit.yaml](./packit.yaml) control file.

# Automated maintenance

It is important to keep your [NPM modules](./package.json) up to date, to keep
up with security updates and bug fixes. This is done with the
[npm-update bot script](https://github.com/cockpit-project/bots/blob/main/npm-update)
which is run weekly or upon [manual request](https://github.com/cockpit-project/starter-kit/actions) through the
[npm-update.yml](.github/workflows/npm-update.yml) [GitHub action](https://github.com/features/actions).

# Further reading

 * The [Starter Kit announcement](https://cockpit-project.org/blog/cockpit-starter-kit.html)
   blog post explains the rationale for this project.
 * [Cockpit Deployment and Developer documentation](https://cockpit-project.org/guide/latest/)
 * [Make your project easily discoverable](https://cockpit-project.org/blog/making-a-cockpit-application.html)
