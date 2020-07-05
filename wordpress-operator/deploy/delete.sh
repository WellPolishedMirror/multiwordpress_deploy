kubectl delete -f crds/example.com_v1_wordpress_cr.yaml
kubectl delete -f operator.yaml
kubectl delete -f role_binding.yaml
kubectl delete -f role.yaml
kubectl delete -f service_account.yaml
kubectl delete -f crds/example.com_wordpresses_crd.yaml 
