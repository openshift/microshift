package build

type Proxy interface {
	NewBlueprintBuild(path string, opts *PlannerOpts) (Build, error)
	NewImageFetcher(path string, opts *PlannerOpts) (Build, error)
	NewContainerfileBuild(path string, opts *PlannerOpts) (Build, error)
}

var _ Proxy = (*proxy)(nil)

func NewProxy() Proxy {
	return &proxy{}
}

type proxy struct{}

func (b *proxy) NewBlueprintBuild(path string, opts *PlannerOpts) (Build, error) {
	return NewBlueprintBuild(path, opts)
}

func (b *proxy) NewImageFetcher(path string, opts *PlannerOpts) (Build, error) {
	return NewImageFetcher(path, opts)
}

func (b *proxy) NewContainerfileBuild(path string, opts *PlannerOpts) (Build, error) {
	return NewContainerfileBuild(path, opts)
}
