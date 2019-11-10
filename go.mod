module github.com/metal-pod/csi-lvm

require (
	github.com/container-storage-interface/spec v1.1.0
	github.com/digitalocean/godo v1.24.1
	github.com/golang/protobuf v1.3.2
	github.com/google/lvmd v0.0.0-20190916151813-e6e28ff087f6
	github.com/kubernetes-csi/csi-test v2.0.0+incompatible
	github.com/kubernetes-csi/external-snapshotter v0.4.1
	github.com/onsi/ginkgo v1.10.3 // indirect
	github.com/onsi/gomega v1.7.1 // indirect
	github.com/sirupsen/logrus v1.0.5
	golang.org/x/crypto v0.0.0-20191108234033-bd318be0434a // indirect
	golang.org/x/net v0.0.0-20191109021931-daa7c04131f5 // indirect
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a
	golang.org/x/sys v0.0.0-20191105231009-c1f44814a5cd // indirect
	google.golang.org/genproto v0.0.0-20191108220845-16a3f7862a1a // indirect
	google.golang.org/grpc v1.23.1
	gopkg.in/yaml.v2 v2.2.5 // indirect
	k8s.io/api v0.0.0-20181117111259-46ad728b8d13
	k8s.io/apimachinery v0.0.0-20181116115711-1b0702fe2927
	k8s.io/client-go v9.0.0+incompatible
)

go 1.13
