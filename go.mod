module kube-globalreserve

go 1.12

replace (
	k8s.io/api => github.com/kubernetes/kubernetes/staging/src/k8s.io/api v0.0.0-20200118001809-59603c6e503c
	k8s.io/apiextensions-apiserver => github.com/kubernetes/kubernetes/staging/src/k8s.io/apiextensions-apiserver v0.0.0-20200118001809-59603c6e503c
	k8s.io/apimachinery => github.com/kubernetes/kubernetes/staging/src/k8s.io/apimachinery v0.0.0-20200118001809-59603c6e503c
	k8s.io/apiserver => github.com/kubernetes/kubernetes/staging/src/k8s.io/apiserver v0.0.0-20200118001809-59603c6e503c
	k8s.io/cli-runtime => github.com/kubernetes/kubernetes/staging/src/k8s.io/cli-runtime v0.0.0-20200118001809-59603c6e503c
	k8s.io/client-go => github.com/kubernetes/client-go v0.17.2
	k8s.io/cloud-provider => github.com/kubernetes/kubernetes/staging/src/k8s.io/cloud-provider v0.0.0-20200118001809-59603c6e503c
	k8s.io/cluster-bootstrap => github.com/kubernetes/kubernetes/staging/src/k8s.io/cluster-bootstrap v0.0.0-20200118001809-59603c6e503c
	k8s.io/code-generator => github.com/kubernetes/kubernetes/staging/src/k8s.io/code-generator v0.0.0-20200118001809-59603c6e503c
	k8s.io/component-base => github.com/kubernetes/kubernetes/staging/src/k8s.io/component-base v0.0.0-20200118001809-59603c6e503c
	k8s.io/cri-api => github.com/kubernetes/kubernetes/staging/src/k8s.io/cri-api v0.0.0-20200118001809-59603c6e503c
	k8s.io/csi-translation-lib => github.com/kubernetes/kubernetes/staging/src/k8s.io/csi-translation-lib v0.0.0-20200118001809-59603c6e503c
	k8s.io/kube-aggregator => github.com/kubernetes/kubernetes/staging/src/k8s.io/kube-aggregator v0.0.0-20200118001809-59603c6e503c
	k8s.io/kube-controller-manager => github.com/kubernetes/kubernetes/staging/src/k8s.io/kube-controller-manager v0.0.0-20200118001809-59603c6e503c
	k8s.io/kube-proxy => github.com/kubernetes/kubernetes/staging/src/k8s.io/kube-proxy v0.0.0-20200118001809-59603c6e503c
	k8s.io/kube-scheduler => github.com/kubernetes/kubernetes/staging/src/k8s.io/kube-scheduler v0.0.0-20200118001809-59603c6e503c
	k8s.io/kubectl => github.com/kubernetes/kubectl v0.17.2
	k8s.io/kubelet => github.com/kubernetes/kubernetes/staging/src/k8s.io/kubelet v0.0.0-20200118001809-59603c6e503c
	k8s.io/kubernetes => github.com/kubernetes/kubernetes v1.17.2
	k8s.io/legacy-cloud-providers => github.com/kubernetes/kubernetes/staging/src/k8s.io/legacy-cloud-providers v0.0.0-20200118001809-59603c6e503c
	k8s.io/metrics => github.com/kubernetes/kubernetes/staging/src/k8s.io/metrics v0.0.0-20200118001809-59603c6e503c
	k8s.io/sample-apiserver => github.com/kubernetes/kubernetes/staging/src/k8s.io/sample-apiserver v0.0.0-20200118001809-59603c6e503c
)

require (
	github.com/julienschmidt/httprouter v1.2.0
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/tools v0.0.0-20200423205358-59e73619c742 // indirect
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	k8s.io/component-base v0.17.2
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v1.17.2
)
