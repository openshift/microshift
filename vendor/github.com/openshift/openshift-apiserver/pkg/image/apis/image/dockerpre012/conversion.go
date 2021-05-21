package dockerpre012

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/openshift/api/image/dockerpre012"
	newer "github.com/openshift/openshift-apiserver/pkg/image/apis/image"
)

// Convert docker client object to internal object, but only when this package is included
func Convert_dockerpre012_ImagePre_012_to_api_DockerImage(in *dockerpre012.ImagePre012, out *newer.DockerImage, s conversion.Scope) error {
	if err := s.Convert(in.Config, &out.Config); err != nil {
		return err
	}
	if err := s.Convert(&in.ContainerConfig, &out.ContainerConfig); err != nil {
		return err
	}
	out.ID = in.ID
	out.Parent = in.Parent
	out.Comment = in.Comment
	out.Created = metav1.NewTime(in.Created)
	out.Container = in.Container
	out.DockerVersion = in.DockerVersion
	out.Author = in.Author
	out.Architecture = in.Architecture
	out.Size = in.Size
	return nil
}
func Convert_api_DockerImage_to_dockerpre012_ImagePre_012(in *newer.DockerImage, out *dockerpre012.ImagePre012, s conversion.Scope) error {
	if err := s.Convert(&in.Config, &out.Config); err != nil {
		return err
	}
	if err := s.Convert(&in.ContainerConfig, &out.ContainerConfig); err != nil {
		return err
	}
	out.ID = in.ID
	out.Parent = in.Parent
	out.Comment = in.Comment
	out.Created = in.Created.Time
	out.Container = in.Container
	out.DockerVersion = in.DockerVersion
	out.Author = in.Author
	out.Architecture = in.Architecture
	out.Size = in.Size
	return nil
}

func addConversionFuncs(s *runtime.Scheme) error {
	if err := s.AddConversionFunc((*dockerpre012.ImagePre012)(nil), (*newer.DockerImage)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_dockerpre012_ImagePre_012_to_api_DockerImage(a.(*dockerpre012.ImagePre012), b.(*newer.DockerImage), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*newer.DockerImage)(nil), (*dockerpre012.ImagePre012)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_api_DockerImage_to_dockerpre012_ImagePre_012(a.(*newer.DockerImage), b.(*dockerpre012.ImagePre012), scope)
	}); err != nil {
		return err
	}

	return nil
}
