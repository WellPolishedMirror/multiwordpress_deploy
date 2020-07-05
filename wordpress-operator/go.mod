module github.com/renan-campos/wordpress-operator

go 1.13

require (
	github.com/operator-framework/operator-sdk v0.18.2
	github.com/operator-framework/operator-sdk-samples/go/memcached-operator v0.0.0-20200703111833-1306fcccf2a2
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)
