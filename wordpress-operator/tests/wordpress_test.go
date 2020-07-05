package e2e

import (
	goctx "context"
	"fmt"
	"testing"
	"time"

	"github.com/renan-campos/wordpress-operator/pkg/apis"
	operator "github.com/renan-campos/wordpress-operator/pkg/apis/example/v1"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestWordpress(t *testing.T) {
	wordpressList := &operator.WordpressList{}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, wordpressList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("wordpress-group", func(t *testing.T) {
		t.Run("Cluster", WordpressCluster)
		t.Run("Cluster2", WordpressCluster)
	})
}

func WordpressCluster(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetOperatorNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "wordpress-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	if err = wordpressConfigTest(t, f, ctx); err != nil {
		t.Fatal(err)
	}
}

// Checks that the sqlRootPassword is saved.
// TODO: When this actually does something, this test should be changed
//       to ensure the sqlRootPassword is NOT saved.
func wordpressConfigTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetOperatorNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}

	// create wordpress custom resource
	exampleWordpress := &operator.Wordpress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-wordpress",
			Namespace: namespace,
		},
		Spec: operator.WordpressSpec{
			Password: "DirtyLittleSecret",
		},
	}
	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), exampleWordpress, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	createdWordpress := &operator.Wordpress{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-wordpress", Namespace: namespace}, createdWordpress)
	if err != nil {
		return err
	}

	if createdWordpress.Spec.Password != exampleWordpress.Spec.Password {
		return fmt.Errorf("Passwords do not match. %s != %s", createdWordpress.Spec.Password, exampleWordpress.Spec.Password)
	}

	return nil
}
