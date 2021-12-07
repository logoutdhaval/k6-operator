package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/grafana/k6-operator/api/v1alpha1"
	"github.com/grafana/k6-operator/pkg/cloud"
	"github.com/grafana/k6-operator/pkg/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FinishJobs waits for the pods to finish, performs finishing call for cloud output and moves state to "finished".
func FinishJobs(ctx context.Context, log logr.Logger, k6 *v1alpha1.K6, r *K6Reconciler) (ctrl.Result, error) {
	log.Info("Waiting for pods to finish")

	if cli := types.ParseCLI(&k6.Spec); !cli.HasCloudOut {
		// nothing to do
		return ctrl.Result{}, nil
	}

	selector := labels.SelectorFromSet(map[string]string{
		"app":    "k6",
		"k6_cr":  k6.Name,
		"runner": "true",
	})

	opts := &client.ListOptions{LabelSelector: selector, Namespace: k6.Namespace}
	pl := &v1.PodList{}

	if err := r.List(ctx, pl, opts); err != nil {
		log.Error(err, "Could not list pods")
		return ctrl.Result{}, err
	}

	var count int
	for _, pod := range pl.Items {
		if pod.Status.Phase == "Succeeded" ||
			pod.Status.Phase == "Failed" ||
			pod.Status.Phase == "Unknown" {
			count++
		}
	}

	log.Info(fmt.Sprintf("%d/%d pods finished", count, k6.Spec.Parallelism))

	if count >= int(k6.Spec.Parallelism) {
		if err := cloud.FinishTestRun(testRunId); err != nil {
			log.Error(err, "Could not finish test run with cloud output")
			return ctrl.Result{}, err
		}

		k6.Status.Stage = "finished"
		if err := r.Client.Status().Update(ctx, k6); err != nil {
			log.Error(err, "Could not update status of custom resource")
			return ctrl.Result{}, err
		}

		log.Info(fmt.Sprintf("Cloud test run %s was finished", testRunId))

		return ctrl.Result{}, nil
	}

	return ctrl.Result{Requeue: true}, nil
}
