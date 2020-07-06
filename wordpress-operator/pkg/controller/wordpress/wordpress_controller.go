package wordpress

import (
	"context"
	"encoding/base64"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	examplev1 "github.com/renan-campos/wordpress-operator/pkg/apis/example/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_wordpress")

// Add creates a new Wordpress Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileWordpress{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("wordpress-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Wordpress
	err = c.Watch(&source.Kind{Type: &examplev1.Wordpress{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watching for secondary resources.
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &examplev1.Wordpress{},
	})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &examplev1.Wordpress{},
	})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &examplev1.Wordpress{},
	})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &examplev1.Wordpress{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileWordpress implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileWordpress{}

// ReconcileWordpress reconciles a Wordpress object
type ReconcileWordpress struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Wordpress object and makes changes based on the state read
// and what is in the Wordpress.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileWordpress) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Wordpress")

	// Fetch the Wordpress instance
	instance := &examplev1.Wordpress{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	/***
	NOTE: Is this pattern of doing an action+requeue better than doing all the actions at once?
	      This is what the tutorial did, so I'll stick with it for now.
	  		TODO: Will ask about doing the actions all at once later.
	NOTE: All of these steps could be completed with a call to a higher-order function.
	      This will greatly reduce the amount of code and improve the clarity.
				I will do so once I study higher-order functions in Go.
	***/
	// Create Secret if it doesn't already exist
	secretFound := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, secretFound)
	if err != nil && errors.IsNotFound(err) {
		sec := r.secretForWordpress(instance)
		reqLogger.Info("Creating a new Secret", "Secret.Namespace", sec.Namespace, "Secret.Name", sec.Name)
		err = r.client.Create(context.TODO(), sec)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Secret", "Secret.Namespace", sec.Namespace, "Secret.Name", sec.Name)
			return reconcile.Result{}, err
		}
		// Secret created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Secret")
		return reconcile.Result{}, err
	}

	// Names used for other secondary resources.
	mysqlName := fmt.Sprintf("%s-mysql", instance.Name)
	wordpressName := fmt.Sprintf("%s-wordpress", instance.Name)

	/*
		// Create mysql PersistentVolumeClaim if it doesn't already exist.
		mysqlPVCFound := &corev1.PersistentVolumeClaim{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: mysqlName, Namespace: instance.Namespace}, mysqlPVCFound)
		if err != nil && errors.IsNotFound(err) {
			mysqlPVC := r.mysqlPVCForWordpress(instance)
			reqLogger.Info("Creating a new PVC", "mysqlPVC.Namespace", mysqlPVC.Namespace, "mysqlPVC.Name", mysqlPVC.Name)
			err = r.client.Create(context.TODO(), mysqlPVC)
			if err != nil {
				reqLogger.Error(err, "Failed to create new PVC", "mysqlPVC.Namespace", mysqlPVC.Namespace, "mysqlPVC.Name", mysqlPVC.Name)
				return reconcile.Result{}, err
			}
			// PVC created successfully - return and requeue
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			reqLogger.Error(err, "Failed to get mysql PVC")
			return reconcile.Result{}, err
		}
		// Create wordpress PVC if it doesn't already exist.
		wordpressPVCFound := &corev1.PersistentVolumeClaim{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: wordpressName, Namespace: instance.Namespace}, wordpressPVCFound)
		if err != nil && errors.IsNotFound(err) {
			wordpressPVC := r.wordpressPVCForWordpress(instance)
			reqLogger.Info("Creating a new PVC", "wordpressPVC.Namespace", wordpressPVC.Namespace, "wordpressPVC.Name", wordpressPVC.Name)
			err = r.client.Create(context.TODO(), wordpressPVC)
			if err != nil {
				reqLogger.Error(err, "Failed to create new PVC", "wordpressPVC.Namespace", wordpressPVC.Namespace, "wordpressPVC.Name", wordpressPVC.Name)
				return reconcile.Result{}, err
			}
			// PVC created successfully - return and requeue
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			reqLogger.Error(err, "Failed to get wordpress PVC")
			return reconcile.Result{}, err
		}
	*/

	// Create mysql deployment if it doesn't already exist.
	mysqlDepFound := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: mysqlName, Namespace: instance.Namespace}, mysqlDepFound)
	if err != nil && errors.IsNotFound(err) {
		mysqlDep := r.mysqlDeploymentForWordpress(instance)
		reqLogger.Info("Creating a new Deployment", "mysqlDep.Namespace", mysqlDep.Namespace, "mysqlDep.Name", mysqlDep.Name)
		err = r.client.Create(context.TODO(), mysqlDep)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "mysqlDep.Namespace", mysqlDep.Namespace, "mysqlDep.Name", mysqlDep.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get mysql Deployment")
		return reconcile.Result{}, err
	}
	// Create mysql service
	mysqlServiceFound := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: mysqlName, Namespace: instance.Namespace}, mysqlServiceFound)
	if err != nil && errors.IsNotFound(err) {
		mysqlService := r.mysqlServiceForWordpress(instance)
		reqLogger.Info("Creating a new Service", "mysqlService.Namespace", mysqlService.Namespace, "mysqlService.Name", mysqlService.Name)
		err = r.client.Create(context.TODO(), mysqlService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service", "mysqlService.Namespace", mysqlService.Namespace, "mysqlService.Name", mysqlService.Name)
			return reconcile.Result{}, err
		}
		// Service created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get mysql Service")
		return reconcile.Result{}, err
	}

	// Create wordpress deployment if it doesn't already exist.
	wordpressDepFound := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: wordpressName, Namespace: instance.Namespace}, wordpressDepFound)
	if err != nil && errors.IsNotFound(err) {
		wordpressDep := r.wordpressDeploymentForWordpress(instance)
		reqLogger.Info("Creating a new Deployment", "wordpressDep.Namespace", wordpressDep.Namespace, "wordpressDep.Name", wordpressDep.Name)
		err = r.client.Create(context.TODO(), wordpressDep)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "wordpressDep.Namespace", wordpressDep.Namespace, "wordpressDep.Name", wordpressDep.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get wordpress Deployment")
		return reconcile.Result{}, err
	}
	// Create wordpress service
	wordpressServiceFound := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: wordpressName, Namespace: instance.Namespace}, wordpressServiceFound)
	if err != nil && errors.IsNotFound(err) {
		wordpressService := r.wordpressServiceForWordpress(instance)
		reqLogger.Info("Creating a new Service", "wordpressService.Namespace", wordpressService.Namespace, "wordpressService.Name", wordpressService.Name)
		err = r.client.Create(context.TODO(), wordpressService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service", "wordpressService.Namespace", wordpressService.Namespace, "wordpressService.Name", wordpressService.Name)
			return reconcile.Result{}, err
		}
		// Service created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get wordpress Service")
		return reconcile.Result{}, err
	}

	// My ultra secure operator ;)
	// Probably a better way to do this, if I actually knew Go.
	s := fmt.Sprintf("Database password: %s", instance.Spec.Password)
	reqLogger.Info(s)

	return reconcile.Result{}, nil

}

func (r *ReconcileWordpress) secretForWordpress(m *examplev1.Wordpress) *corev1.Secret {
	ls := labelsForWordpress(m.Name)

	pw := []byte(m.Spec.Password)
	enc_pw := make([]byte, base64.StdEncoding.EncodedLen(len(pw)))

	base64.StdEncoding.Encode(enc_pw, pw)

	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
			Labels:    ls,
		},
		Data: map[string][]byte{"password": enc_pw},
	}

	// Set Wordpress instance as the owner and controller
	controllerutil.SetControllerReference(m, sec, r.scheme)

	return sec
}

func (r *ReconcileWordpress) mysqlPVCForWordpress(m *examplev1.Wordpress) *corev1.PersistentVolumeClaim {
	ls := labelsForWordpress(m.Name)
	ls["tier"] = "mysql"

	pvc_name := fmt.Sprintf("%s-mysql", m.Name)
	pvc_size := resource.NewQuantity(20*1024*1024*1024, resource.BinarySI)

	scn := "standard"

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvc_name,
			Namespace: m.Namespace,
			Labels:    ls,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			StorageClassName: &scn,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: *pvc_size,
				},
			},
		},
	}

	controllerutil.SetControllerReference(m, pvc, r.scheme)

	return pvc
}

func (r *ReconcileWordpress) wordpressPVCForWordpress(m *examplev1.Wordpress) *corev1.PersistentVolumeClaim {
	ls := labelsForWordpress(m.Name)
	ls["tier"] = "frontend"

	scn := "standard"

	pvc_name := fmt.Sprintf("%s-wordpress", m.Name)
	pvc_size := resource.NewQuantity(20*1024*1024*1024, resource.BinarySI)

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvc_name,
			Namespace: m.Namespace,
			Labels:    ls,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			StorageClassName: &scn,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: *pvc_size,
				},
			},
		},
	}

	controllerutil.SetControllerReference(m, pvc, r.scheme)

	return pvc
}

func (r *ReconcileWordpress) mysqlDeploymentForWordpress(m *examplev1.Wordpress) *appsv1.Deployment {
	ls := labelsForWordpress(m.Name)
	ls["tier"] = "mysql"

	//volName := fmt.Sprintf("%s-mysql", m.Name)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-mysql", m.Name),
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "mysql:5.6",
						Name:  "mysql",
						Env: []corev1.EnvVar{{
							Name: "MYSQL_ROOT_PASSWORD",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: m.Name,
									},
									Key: "password",
								},
							},
						}},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 3306,
							Name:          "mysql",
						}},
						/*
							VolumeMounts: []corev1.VolumeMount{{
								Name:      volName,
								MountPath: "/var/lib/mysql",
							}},
						*/
					}},
					/*
						Volumes: []corev1.Volume{{
							Name: volName,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: volName,
								},
							},
						}},
					*/
				},
			},
		},
	}

	controllerutil.SetControllerReference(m, dep, r.scheme)

	return dep
}

func (r *ReconcileWordpress) wordpressDeploymentForWordpress(m *examplev1.Wordpress) *appsv1.Deployment {
	ls := labelsForWordpress(m.Name)
	ls["tier"] = "frontend"

	//volName := fmt.Sprintf("%s-wordpress", m.Name)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-wordpress", m.Name),
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "wordpress:4.8-apache",
						Name:  "wordpress",
						Env: []corev1.EnvVar{{
							Name: "WORDPRESS_DB_PASSWORD",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: m.Name,
									},
									Key: "password",
								},
							},
						},
							{
								Name:  "WORDPRESS_DB_HOST",
								Value: fmt.Sprintf("%s-mysql", m.Name),
							}},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
							Name:          "wordpress",
						}},
						/*
							VolumeMounts: []corev1.VolumeMount{{
								Name:      volName,
								MountPath: "/var/www/html",
							}},
						*/
					}},
					/*
						Volumes: []corev1.Volume{{
							Name: volName,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: volName,
								},
							},
						}},
					*/
				},
			},
		},
	}

	controllerutil.SetControllerReference(m, dep, r.scheme)

	return dep
}

func (r *ReconcileWordpress) mysqlServiceForWordpress(m *examplev1.Wordpress) *corev1.Service {
	ls := labelsForWordpress(m.Name)
	ls["tier"] = "mysql"

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-mysql", m.Name),
			Namespace: m.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Port: 3306,
			}},
			Selector: ls,
			Type:     corev1.ServiceTypeClusterIP,
		},
	}

	controllerutil.SetControllerReference(m, svc, r.scheme)

	return svc
}

func (r *ReconcileWordpress) wordpressServiceForWordpress(m *examplev1.Wordpress) *corev1.Service {
	ls := labelsForWordpress(m.Name)
	ls["tier"] = "frontend"

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-wordpress", m.Name),
			Namespace: m.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Port: 80,
			}},
			Selector: ls,
			Type:     corev1.ServiceTypeLoadBalancer,
		},
	}

	controllerutil.SetControllerReference(m, svc, r.scheme)

	return svc
}

func labelsForWordpress(name string) map[string]string {
	return map[string]string{"app": "wordpress", "wordpress_cr": name}
}
