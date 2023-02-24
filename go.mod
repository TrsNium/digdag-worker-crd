module digdag-worker-crd

go 1.13

require (
	github.com/go-logr/logr v0.2.0
	github.com/lib/pq v1.3.0
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.7.0
	github.com/robfig/cron v1.2.0
	k8s.io/api v0.20.0-alpha.2
	k8s.io/apimachinery v0.20.0-alpha.2
	k8s.io/client-go v0.20.0-alpha.2
	sigs.k8s.io/controller-runtime v0.4.0
)
