all: build
.PHONY: all

# Include the library makefile
include $(addprefix ./vendor/github.com/openshift/build-machinery-go/make/, \
	golang.mk \
	targets/openshift/images.mk \
	targets/openshift/deps.mk \
)

IMAGE_REGISTRY :=registry.ci.openshift.org

# This will call a macro called "build-image" which will generate image specific targets based on the parameters:
# $0 - macro name
# $1 - target name
# $2 - image ref
# $3 - Dockerfile path
# $4 - context directory for image build# It will generate target "image-$(1)" for building the image an binding it as a prerequisite to target "images".
$(call build-image,route-controller-manager,$(IMAGE_REGISTRY)/ocp/dev:route-controller-manager, ./Dockerfile,.)

clean:
	$(RM) ./route-controller-manager
.PHONY: clean

GO_TEST_PACKAGES :=./pkg/... ./cmd/...
