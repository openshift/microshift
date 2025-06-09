/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testdata

import (
	"k8s.io/klog/v2"
)

func calls() {
	klog.Infof("%s") // want `k8s.io/klog/v2.Infof format %s reads arg #1, but call has 0 args`
	klog.Infof("%s", "world")
	klog.Info("%s", "world") // want `k8s.io/klog/v2.Info call has possible formatting directive %s`
	klog.Info("world")
	klog.Infoln("%s", "world") // want `k8s.io/klog/v2.Infoln call has possible formatting directive %s`
	klog.Infoln("world")

	klog.InfofDepth(1, "%s") // want `k8s.io/klog/v2.InfofDepth format %s reads arg #1, but call has 0 args`
	klog.InfofDepth(1, "%s", "world")
	klog.InfoDepth(1, "%s", "world") // want `k8s.io/klog/v2.InfoDepth call has possible formatting directive %s`
	klog.InfoDepth(1, "world")
	klog.InfolnDepth(1, "%s", "world") // want `k8s.io/klog/v2.InfolnDepth call has possible formatting directive %s`
	klog.InfolnDepth(1, "world")

	klog.Warningf("%s") // want `k8s.io/klog/v2.Warningf format %s reads arg #1, but call has 0 args`
	klog.Warningf("%s", "world")
	klog.Warning("%s", "world") // want `k8s.io/klog/v2.Warning call has possible formatting directive %s`
	klog.Warning("world")
	klog.Warningln("%s", "world") // want `k8s.io/klog/v2.Warningln call has possible formatting directive %s`
	klog.Warningln("world")

	klog.WarningfDepth(1, "%s") // want `k8s.io/klog/v2.WarningfDepth format %s reads arg #1, but call has 0 args`
	klog.WarningfDepth(1, "%s", "world")
	klog.WarningDepth(1, "%s", "world") // want `k8s.io/klog/v2.WarningDepth call has possible formatting directive %s`
	klog.WarningDepth(1, "world")
	klog.WarninglnDepth(1, "%s", "world") // want `k8s.io/klog/v2.WarninglnDepth call has possible formatting directive %s`
	klog.WarninglnDepth(1, "world")

	klog.Errorf("%s") // want `k8s.io/klog/v2.Errorf format %s reads arg #1, but call has 0 args`
	klog.Errorf("%s", "world")
	klog.Error("%s", "world") // want `k8s.io/klog/v2.Error call has possible formatting directive %s`
	klog.Error("world")
	klog.Errorln("%s", "world") // want `k8s.io/klog/v2.Errorln call has possible formatting directive %s`
	klog.Errorln("world")

	klog.ErrorfDepth(1, "%s") // want `k8s.io/klog/v2.ErrorfDepth format %s reads arg #1, but call has 0 args`
	klog.ErrorfDepth(1, "%s", "world")
	klog.ErrorDepth(1, "%s", "world") // want `k8s.io/klog/v2.ErrorDepth call has possible formatting directive %s`
	klog.ErrorDepth(1, "world")
	klog.ErrorlnDepth(1, "%s", "world") // want `k8s.io/klog/v2.ErrorlnDepth call has possible formatting directive %s`
	klog.ErrorlnDepth(1, "world")

	klog.Fatalf("%s") // want `k8s.io/klog/v2.Fatalf format %s reads arg #1, but call has 0 args`
	klog.Fatalf("%s", "world")
	klog.Fatal("%s", "world") // want `k8s.io/klog/v2.Fatal call has possible formatting directive %s`
	klog.Fatal("world")
	klog.Fatalln("%s", "world") // want `k8s.io/klog/v2.Fatalln call has possible formatting directive %s`
	klog.Fatalln("world")

	klog.FatalfDepth(1, "%s") // want `k8s.io/klog/v2.FatalfDepth format %s reads arg #1, but call has 0 args`
	klog.FatalfDepth(1, "%s", "world")
	klog.FatalDepth(1, "%s", "world") // want `k8s.io/klog/v2.FatalDepth call has possible formatting directive %s`
	klog.FatalDepth(1, "world")
	klog.FatallnDepth(1, "%s", "world") // want `k8s.io/klog/v2.FatallnDepth call has possible formatting directive %s`
	klog.FatallnDepth(1, "world")

	klog.V(1).Infof("%s") // want `\(k8s.io/klog/v2.Verbose\).Infof format %s reads arg #1, but call has 0 args`
	klog.V(1).Infof("%s", "world")
	klog.V(1).Info("%s", "world") // want `\(k8s.io/klog/v2.Verbose\).Info call has possible formatting directive %s`
	klog.V(1).Info("world")
	klog.V(1).Infoln("%s", "world") // want `\(k8s.io/klog/v2.Verbose\).Infoln call has possible formatting directive %s`
	klog.V(1).Infoln("world")

	klog.V(1).InfofDepth(1, "%s") // want `\(k8s.io/klog/v2.Verbose\).InfofDepth format %s reads arg #1, but call has 0 args`
	klog.V(1).InfofDepth(1, "%s", "world")
	klog.V(1).InfoDepth(1, "%s", "world") // want `\(k8s.io/klog/v2.Verbose\).InfoDepth call has possible formatting directive %s`
	klog.V(1).InfoDepth(1, "world")
	klog.V(1).InfolnDepth(1, "%s", "world") // want `\(k8s.io/klog/v2.Verbose\).InfolnDepth call has possible formatting directive %s`
	klog.V(1).InfolnDepth(1, "world")

	// Detecting format specifiers for klog.InfoS and other structured logging calls would be nice,
	// but doesn't work the same way because of the extra "msg" string parameter. logcheck
	// can be used instead of "go vet".
	klog.InfoS("%s", "world")
}
