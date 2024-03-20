package templatingdata

// TemplatingData contains all values needed for templating Composer's Sources & Blueprints,
// and other templated artifacts within a MicroShift's test harness.
type TemplatingData struct {
	Arch string

	// Minor version of current release's RHOCP. If RHOCP is not available yet, it defaults to 0.
	RHOCPMinorY int

	// Minor version of previous release's RHOCP. If RHOCP is not available yet, it defaults to 0.
	RHOCPMinorY1 int

	// Minor version of previous previous release's RHOCP.
	RHOCPMinorY2 int

	// Current stores metadata of current release, i.e. matching currently checked out git branch.
	// If the RHOCP is not available yet, it can point to Release or Engineering candidates present on the OpenShift mirror.
	// If those are also not available, it will be empty and related composer Sources and Blueprints will not be build.
	Current Release

	// Current stores metadata of current release, i.e. matching currently checked out git branch.
	// If the RHOCP is not available yet, it can point to Release or Engineering candidates present on the OpenShift mirror.
	Previous Release

	// Current stores metadata of current release, i.e. matching currently checked out git branch.
	// Usually this should always point to RHOCP because it's two minor versions older than what we're working on currently.
	YMinus2 Release

	// Source stores metadata of RPMs built from currently checked out source code.
	Source Release

	// Source stores metadata of RPMs built from base branch of currently checked out branch.
	// Usually it can be `main` or `release-4.Y`.
	Base Release

	// Source stores metadata of RPMs built from currently checked out source code with minor version overridden
	// to be newer than what we're currently working on.
	// These are needed for various ostree upgrade tests.
	FakeNext Release

	// External stores metadata of RPMs supplied from external source, like private builds.
	External Release
}

// Release represents metadata of particular set of RPMs.
type Release struct {
	// Repository is where RPM resides. It can be local (on the disk), http (like OpenShift's mirror), or RHOCP.
	Repository string

	// Version is full version string of a RPM, e.g. 4.14.16-202403071942.p0.g4cef5f2.assembly.4.14.16.el9.
	Version string

	// Minor is minor part of the version, e.g. 14.
	Minor int

	// Images is a list of images stored in release-info RPM.
	// Currently only for local repositories.
	// TODO: Extend to remote for local image mirroring.
	Images []string
}
